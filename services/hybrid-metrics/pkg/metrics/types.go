package metrics

import "time"

// SystemMetrics represents system resource usage
type SystemMetrics struct {
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	Timestamp   time.Time `json:"timestamp"`
}

// ComplexityMetrics represents GraphQL query complexity data
type ComplexityMetrics struct {
	Score     float64   `json:"score"`
	QueryType string    `json:"query_type,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// HybridMetrics combines all metric types for fusion algorithm
type HybridMetrics struct {
	Complexity ComplexityMetrics `json:"complexity"`
	System     SystemMetrics     `json:"system"`
	Timestamp  time.Time         `json:"timestamp"`
}

// ScalingDecision represents the output of the fusion algorithm
type ScalingDecision struct {
	ShouldScale   bool    `json:"should_scale"`
	Action        string  `json:"action"` // "scale_up", "scale_down", "no_action"
	ScalingFactor float64 `json:"scaling_factor"`
	Reason        string  `json:"reason"`
	Confidence    float64 `json:"confidence"`
	Timestamp     time.Time `json:"timestamp"`
}

// CorrelationResults represents correlation analysis results
type CorrelationResults struct {
	ComplexityCPU    float64 `json:"complexity_cpu_correlation"`
	ComplexityMemory float64 `json:"complexity_memory_correlation"`
	SampleSize       int     `json:"sample_size"`
	LastUpdated      time.Time `json:"last_updated"`
}

// FusionConfig holds configuration for the fusion algorithm
type FusionConfig struct {
	ComplexityWeight    float64       `json:"complexity_weight"`
	CPUWeight          float64       `json:"cpu_weight"`
	MemoryWeight       float64       `json:"memory_weight"`
	ScalingThreshold   float64       `json:"scaling_threshold"`
	CooldownPeriod     time.Duration `json:"cooldown_period"`
	HistoryWindowSize  int           `json:"history_window_size"`
	MinSamplesRequired int           `json:"min_samples_required"`
}

// MetricHistory stores historical metric data
type MetricHistory struct {
	ComplexityHistory []ComplexityMetrics `json:"complexity_history"`
	SystemHistory     []SystemMetrics     `json:"system_history"`
	DecisionHistory   []ScalingDecision   `json:"decision_history"`
	MaxSize           int                 `json:"max_size"`
}

// FusionStats provides insights into the fusion algorithm performance
type FusionStats struct {
	TotalDecisions      int               `json:"total_decisions"`
	ScaleUpDecisions    int               `json:"scale_up_decisions"`
	ScaleDownDecisions  int               `json:"scale_down_decisions"`
	NoActionDecisions   int               `json:"no_action_decisions"`
	AverageComplexity   float64           `json:"average_complexity"`
	AverageCPU          float64           `json:"average_cpu"`
	AverageMemory       float64           `json:"average_memory"`
	Correlation         CorrelationResults `json:"correlation"`
	LastDecision        *ScalingDecision  `json:"last_decision,omitempty"`
	UptimeSeconds       int64             `json:"uptime_seconds"`
}

// GraphQLComplexityResponse represents the response from GraphQL complexity endpoint
type GraphQLComplexityResponse struct {
	CacheSize     int `json:"cacheSize"`
	CachedQueries []struct {
		Query      string  `json:"query"`
		Complexity float64 `json:"complexity"`
	} `json:"cachedQueries"`
}