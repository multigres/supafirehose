package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// Collector aggregates metrics from workers and provides snapshots
type Collector struct {
	mu sync.RWMutex

	// Current window histograms
	readLatencies  *Histogram
	writeLatencies *Histogram

	// Window counters (reset each interval)
	readCount   int64
	writeCount  int64
	readErrors  int64
	writeErrors int64

	// Total counters (never reset except via Reset())
	totalQueries atomic.Int64
	totalErrors  atomic.Int64

	// Recent errors list (for UI visibility)
	recentErrors    []ErrorEntry
	lastErrorTime   time.Time
	maxRecentErrors int

	// Pool stats function
	poolStatsFunc func() PoolStats

	// Start time for uptime calculation
	startTime time.Time
}

// NewCollector creates a new metrics collector
func NewCollector(poolStatsFunc func() PoolStats) *Collector {
	return &Collector{
		readLatencies:   NewHistogram(),
		writeLatencies:  NewHistogram(),
		poolStatsFunc:   poolStatsFunc,
		startTime:       time.Now(),
		recentErrors:    make([]ErrorEntry, 0),
		maxRecentErrors: 10, // Keep last 10 errors
	}
}

// RecordRead records a read operation
func (c *Collector) RecordRead(latency time.Duration, err error) {
	c.readLatencies.Record(latency)
	atomic.AddInt64(&c.readCount, 1)
	c.totalQueries.Add(1)

	if err != nil {
		atomic.AddInt64(&c.readErrors, 1)
		c.totalErrors.Add(1)
		c.addError("read: " + err.Error())
	}
}

// RecordWrite records a write operation
func (c *Collector) RecordWrite(latency time.Duration, err error) {
	c.writeLatencies.Record(latency)
	atomic.AddInt64(&c.writeCount, 1)
	c.totalQueries.Add(1)

	if err != nil {
		atomic.AddInt64(&c.writeErrors, 1)
		c.totalErrors.Add(1)
		c.addError("write: " + err.Error())
	}
}

// addError adds an error to the recent errors list (rate limited to 1 per 10 seconds)
func (c *Collector) addError(errMsg string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Rate limit: only add 1 error per 10 seconds
	if time.Since(c.lastErrorTime) < 10*time.Second {
		return
	}
	c.lastErrorTime = time.Now()

	// Add new error
	entry := ErrorEntry{
		Timestamp: time.Now().UnixMilli(),
		Message:   errMsg,
	}
	c.recentErrors = append(c.recentErrors, entry)

	// Trim to max size
	if len(c.recentErrors) > c.maxRecentErrors {
		c.recentErrors = c.recentErrors[len(c.recentErrors)-c.maxRecentErrors:]
	}
}

// Snapshot returns current metrics and resets window counters
func (c *Collector) Snapshot(interval time.Duration) MetricsSnapshot {
	// Get histogram snapshots (this also resets them)
	readHist := c.readLatencies.SnapshotAndReset()
	writeHist := c.writeLatencies.SnapshotAndReset()

	// Get and reset window counters
	readCount := atomic.SwapInt64(&c.readCount, 0)
	writeCount := atomic.SwapInt64(&c.writeCount, 0)
	readErrors := atomic.SwapInt64(&c.readErrors, 0)
	writeErrors := atomic.SwapInt64(&c.writeErrors, 0)

	// Calculate QPS based on actual interval
	intervalSec := interval.Seconds()
	readQPS := float64(readCount) / intervalSec
	writeQPS := float64(writeCount) / intervalSec

	// Get totals
	totalQueries := c.totalQueries.Load()
	totalErrors := c.totalErrors.Load()

	// Calculate error rate
	var errorRate float64
	if totalQueries > 0 {
		errorRate = float64(totalErrors) / float64(totalQueries)
	}

	// Get pool stats
	var poolStats PoolStats
	if c.poolStatsFunc != nil {
		poolStats = c.poolStatsFunc()
	}

	// Get recent errors
	c.mu.RLock()
	recentErrors := make([]ErrorEntry, len(c.recentErrors))
	copy(recentErrors, c.recentErrors)
	c.mu.RUnlock()

	return MetricsSnapshot{
		Timestamp: time.Now().UnixMilli(),
		Reads: OperationStats{
			QPS:        readQPS,
			LatencyP50: readHist.P50,
			LatencyP99: readHist.P99,
			LatencyAvg: readHist.Avg,
			Errors:     readErrors,
		},
		Writes: OperationStats{
			QPS:        writeQPS,
			LatencyP50: writeHist.P50,
			LatencyP99: writeHist.P99,
			LatencyAvg: writeHist.Avg,
			Errors:     writeErrors,
		},
		Totals: TotalStats{
			Queries:   totalQueries,
			Errors:    totalErrors,
			ErrorRate: errorRate,
		},
		Pool:         poolStats,
		RecentErrors: recentErrors,
	}
}

// Reset clears all metrics
func (c *Collector) Reset() {
	c.readLatencies.SnapshotAndReset()
	c.writeLatencies.SnapshotAndReset()
	atomic.StoreInt64(&c.readCount, 0)
	atomic.StoreInt64(&c.writeCount, 0)
	atomic.StoreInt64(&c.readErrors, 0)
	atomic.StoreInt64(&c.writeErrors, 0)
	c.totalQueries.Store(0)
	c.totalErrors.Store(0)
	c.startTime = time.Now()

	// Clear recent errors
	c.mu.Lock()
	c.recentErrors = make([]ErrorEntry, 0)
	c.lastErrorTime = time.Time{}
	c.mu.Unlock()
}

// Uptime returns the duration since the collector was created or reset
func (c *Collector) Uptime() time.Duration {
	return time.Since(c.startTime)
}
