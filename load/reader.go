package load

import (
	"context"
	"math/rand"
	"time"

	"supafirehose/db"
	"supafirehose/metrics"
	"supafirehose/schema"

	"github.com/jackc/pgx/v5"
	"golang.org/x/time/rate"
)

// ReadWorker executes read queries against the database
type ReadWorker struct {
	connMgr   *db.ConnectionManager
	limiter   *rate.Limiter
	collector *metrics.Collector
	scenario  schema.Scenario
	churnRate float64 // Probability of churning connection per second
}

// NewReadWorker creates a new read worker
func NewReadWorker(connMgr *db.ConnectionManager, limiter *rate.Limiter, collector *metrics.Collector, scenario schema.Scenario, churnRate float64) *ReadWorker {
	return &ReadWorker{
		connMgr:   connMgr,
		limiter:   limiter,
		collector: collector,
		scenario:  scenario,
		churnRate: churnRate,
	}
}

// Run starts the read worker loop with its own connection
func (w *ReadWorker) Run(ctx context.Context) {
	// Track if scenario has been initialized (for custom scenarios)
	scenarioInitialized := false

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Create a new connection
		conn, err := w.connMgr.Connect(ctx)
		if err != nil {
			// Don't record context cancellation as error (expected during shutdown)
			if ctx.Err() != nil {
				return
			}
			// Record connection error and backoff
			w.collector.RecordRead(0, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Initialize scenario if needed (for custom scenarios that need table introspection)
		if !scenarioInitialized {
			if err := w.scenario.Initialize(ctx, conn); err != nil {
				w.collector.RecordRead(0, err)
				conn.Close(context.Background())
				w.connMgr.Release()
				time.Sleep(100 * time.Millisecond)
				continue
			}
			scenarioInitialized = true
		}

		// Run queries on this connection until churn or context done
		w.runWithConnection(ctx, conn)

		// Close connection (use background context to ensure clean close)
		conn.Close(context.Background())
		w.connMgr.Release()
	}
}

func (w *ReadWorker) runWithConnection(ctx context.Context, conn *pgx.Conn) {
	// Calculate when to churn this connection
	// If churnRate is 0.1 (10%), average connection lifetime is 10 seconds
	var churnAfter time.Time
	if w.churnRate > 0 {
		// Random lifetime based on churn rate (exponential distribution)
		avgLifetime := time.Duration(float64(time.Second) / w.churnRate)
		lifetime := time.Duration(rand.ExpFloat64() * float64(avgLifetime))
		// Cap lifetime to reasonable bounds
		if lifetime < 100*time.Millisecond {
			lifetime = 100 * time.Millisecond
		}
		if lifetime > 60*time.Second {
			lifetime = 60 * time.Second
		}
		churnAfter = time.Now().Add(lifetime)
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Check if it's time to churn
			if w.churnRate > 0 && time.Now().After(churnAfter) {
				return // Exit to churn connection
			}

			// Wait for rate limiter
			if err := w.limiter.Wait(ctx); err != nil {
				return
			}

			// Execute query
			w.executeRead(ctx, conn)
		}
	}
}

func (w *ReadWorker) executeRead(ctx context.Context, conn *pgx.Conn) {
	start := time.Now()

	err := w.scenario.ExecuteRead(ctx, conn)

	latency := time.Since(start)

	// Don't record context cancellation as an error (expected during shutdown)
	if err != nil && ctx.Err() != nil {
		return
	}
	w.collector.RecordRead(latency, err)
}
