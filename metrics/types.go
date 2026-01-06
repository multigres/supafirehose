package metrics

// MetricsSnapshot represents a point-in-time snapshot of all metrics
type MetricsSnapshot struct {
	Timestamp    int64          `json:"timestamp"`
	Reads        OperationStats `json:"reads"`
	Writes       OperationStats `json:"writes"`
	Totals       TotalStats     `json:"totals"`
	Pool         PoolStats      `json:"pool"`
	RecentErrors []ErrorEntry   `json:"recent_errors,omitempty"`
}

// ErrorEntry represents a single error with timestamp
type ErrorEntry struct {
	Timestamp int64  `json:"timestamp"`
	Message   string `json:"message"`
}

// OperationStats holds metrics for a specific operation type (read/write)
type OperationStats struct {
	QPS        float64 `json:"qps"`
	LatencyP50 float64 `json:"latency_p50_ms"`
	LatencyP99 float64 `json:"latency_p99_ms"`
	LatencyAvg float64 `json:"latency_avg_ms"`
	Errors     int64   `json:"errors"`
}

// TotalStats holds aggregate metrics
type TotalStats struct {
	Queries   int64   `json:"queries"`
	Errors    int64   `json:"errors"`
	ErrorRate float64 `json:"error_rate"`
}

// PoolStats holds connection pool metrics
type PoolStats struct {
	ActiveConnections int32 `json:"active_connections"`
	IdleConnections   int32 `json:"idle_connections"`
	WaitingRequests   int32 `json:"waiting_requests"`
}
