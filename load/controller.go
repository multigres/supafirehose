package load

import (
	"context"
	"sync"

	"supafirehose/db"
	"supafirehose/metrics"
	"supafirehose/schema"

	"golang.org/x/time/rate"
)

// Config holds the load generator configuration
type Config struct {
	Connections int    `json:"connections"`
	ReadQPS     int    `json:"read_qps"`
	WriteQPS    int    `json:"write_qps"`
	ChurnRate   int    `json:"churn_rate"`   // Connections churned per second
	Scenario    string `json:"scenario"`     // Scenario name (simple, jsonb, wide, fk, custom)
	CustomTable string `json:"custom_table"` // Table name for custom scenario
}

// Controller manages the load generation workers
type Controller struct {
	mu sync.RWMutex

	running bool
	config  Config

	// Rate limiters (shared across workers)
	readLimiter  *rate.Limiter
	writeLimiter *rate.Limiter

	// Dependencies
	connMgr   *db.ConnectionManager
	collector *metrics.Collector
	registry  *schema.Registry

	// Current scenario
	scenario schema.Scenario

	// Worker management
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewController creates a new load controller
func NewController(connMgr *db.ConnectionManager, collector *metrics.Collector) *Controller {
	registry := schema.NewRegistry()
	defaultScenario, _ := registry.Get("simple")

	return &Controller{
		connMgr:      connMgr,
		collector:    collector,
		registry:     registry,
		scenario:     defaultScenario,
		readLimiter:  rate.NewLimiter(rate.Limit(100), 100),
		writeLimiter: rate.NewLimiter(rate.Limit(10), 10),
	}
}

// GetRegistry returns the scenario registry
func (c *Controller) GetRegistry() *schema.Registry {
	return c.registry
}

// GetScenario returns the current scenario
func (c *Controller) GetScenario() schema.Scenario {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.scenario
}

// SetScenario sets the current scenario by name
func (c *Controller) SetScenario(name string, customTable string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var newScenario schema.Scenario

	if name == "custom" {
		// customTable can be empty for auto-discovery
		newScenario = c.registry.CreateCustomScenario(customTable)
	} else {
		s, ok := c.registry.Get(name)
		if !ok {
			// Default to simple if not found
			s, _ = c.registry.Get("simple")
		}
		newScenario = s
	}

	c.scenario = newScenario
	c.config.Scenario = name
	c.config.CustomTable = customTable

	return nil
}

// Start begins load generation with the current configuration
func (c *Controller) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return
	}

	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.running = true

	// Calculate churn rate per connection
	// If we have 1000 connections and want 100 churns/sec,
	// each connection has a 0.1 probability of churning per second
	var churnRate float64
	if c.config.Connections > 0 && c.config.ChurnRate > 0 {
		churnRate = float64(c.config.ChurnRate) / float64(c.config.Connections)
	}

	// Split connections between readers and writers (80/20)
	numReaders := (c.config.Connections * 80) / 100
	if numReaders < 1 && c.config.Connections > 0 {
		numReaders = 1
	}
	numWriters := c.config.Connections - numReaders
	if numWriters < 0 {
		numWriters = 0
	}

	// Get current scenario
	scenario := c.scenario

	// Start read workers
	for i := 0; i < numReaders; i++ {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			worker := NewReadWorker(c.connMgr, c.readLimiter, c.collector, scenario, churnRate)
			worker.Run(c.ctx)
		}()
	}

	// Start write workers
	for i := 0; i < numWriters; i++ {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			worker := NewWriteWorker(c.connMgr, c.writeLimiter, c.collector, scenario, churnRate)
			worker.Run(c.ctx)
		}()
	}
}

// Stop gracefully stops all workers
func (c *Controller) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return
	}

	c.cancel()
	c.wg.Wait()
	c.running = false
}

// UpdateConfig updates the load configuration
func (c *Controller) UpdateConfig(cfg Config) {
	c.mu.Lock()
	oldConfig := c.config
	c.config = cfg

	// Update rate limiters immediately (burst = QPS for smooth rate)
	c.readLimiter.SetLimit(rate.Limit(cfg.ReadQPS))
	c.readLimiter.SetBurst(max(cfg.ReadQPS, 1))
	c.writeLimiter.SetLimit(rate.Limit(cfg.WriteQPS))
	c.writeLimiter.SetBurst(max(cfg.WriteQPS, 1))

	// Update scenario if changed
	scenarioChanged := oldConfig.Scenario != cfg.Scenario || oldConfig.CustomTable != cfg.CustomTable
	if scenarioChanged {
		if cfg.Scenario == "custom" {
			c.scenario = c.registry.CreateCustomScenario(cfg.CustomTable)
		} else if cfg.Scenario != "" {
			if s, ok := c.registry.Get(cfg.Scenario); ok {
				c.scenario = s
			}
		}
	}

	// If running and connection count, churn, or scenario changed, restart workers
	needsRestart := c.running && (oldConfig.Connections != cfg.Connections ||
		oldConfig.ChurnRate != cfg.ChurnRate ||
		scenarioChanged)
	c.mu.Unlock()

	if needsRestart {
		c.Stop()
		c.Start()
	}
}

// GetConfig returns the current configuration
func (c *Controller) GetConfig() Config {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
}

// IsRunning returns whether the load generator is running
func (c *Controller) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

// SetConfig sets the initial configuration without restarting
func (c *Controller) SetConfig(cfg Config) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.config = cfg
	c.readLimiter.SetLimit(rate.Limit(cfg.ReadQPS))
	c.readLimiter.SetBurst(max(cfg.ReadQPS, 1))
	c.writeLimiter.SetLimit(rate.Limit(cfg.WriteQPS))
	c.writeLimiter.SetBurst(max(cfg.WriteQPS, 1))
}
