package fusion

import (
	"log"
	"math"
	"sync"
	"time"

	"hybrid-metrics/pkg/metrics"
	"hybrid-metrics/pkg/state"
)

// RedisFusionAlgorithm wraps the original FusionAlgorithm with Redis state sync
type RedisFusionAlgorithm struct {
	*FusionAlgorithm  // Embed original algorithm
	redis *state.RedisStateManager
}

// NewRedisFusionAlgorithm creates a new Redis-backed fusion algorithm
func NewRedisFusionAlgorithm(config metrics.FusionConfig, redisURL string) (*RedisFusionAlgorithm, error) {
	// Create Redis state manager
	redisState, err := state.NewRedisStateManager(redisURL)
	if err != nil {
		return nil, err
	}
	
	// Create original fusion algorithm
	fusionAlg := NewFusionAlgorithm(config)
	
	return &RedisFusionAlgorithm{
		FusionAlgorithm: fusionAlg,
		redis:           redisState,
	}, nil
}

// AddComplexityData overrides original to sync with Redis
func (r *RedisFusionAlgorithm) AddComplexityData(complexity float64, timestamp time.Time) {
	// Call original method to maintain all existing functionality
	r.FusionAlgorithm.AddComplexityData(complexity, timestamp)
	
	// Sync with Redis for real-time state sharing
	if err := r.redis.SetCurrentComplexity(complexity); err != nil {
		log.Printf("⚠️ Failed to sync current complexity to Redis: %v", err)
	} else {
		log.Printf("✅ Synced current complexity %.2f to Redis", complexity)
	}
	
	if err := r.redis.SetLastComplexity(complexity); err != nil {
		log.Printf("⚠️ Failed to sync last complexity to Redis: %v", err)
	}
}

// AddSystemData overrides original to sync with Redis
func (r *RedisFusionAlgorithm) AddSystemData(systemMetrics metrics.SystemMetrics, timestamp time.Time) {
	// Call original method to maintain all existing functionality
	r.FusionAlgorithm.AddSystemData(systemMetrics, timestamp)
	
	// Sync with Redis for real-time state sharing
	if err := r.redis.SetSystemMetrics(systemMetrics); err != nil {
		log.Printf("⚠️ Failed to sync system metrics to Redis: %v", err)
	} else {
		log.Printf("✅ Synced system metrics (CPU: %.2f%%, Memory: %.2f%%) to Redis", 
			systemMetrics.CPUUsage, systemMetrics.MemoryUsage)
	}
}

// ProcessMetrics overrides original to sync decisions with Redis
func (r *RedisFusionAlgorithm) ProcessMetrics() metrics.ScalingDecision {
	// Call original method to get the intelligent scaling decision
	decision := r.FusionAlgorithm.ProcessMetrics()
	
	// Sync decision with Redis for other services to access
	if err := r.redis.SetScalingDecision(decision); err != nil {
		log.Printf("⚠️ Failed to sync scaling decision to Redis: %v", err)
	} else if decision.ShouldScale {
		log.Printf("✅ Synced scaling decision (%s, factor: %.2f) to Redis", 
			decision.Action, decision.ScalingFactor)
	}
	
	// Sync current stats with Redis for monitoring/API access
	stats := r.FusionAlgorithm.GetStats()
	if err := r.redis.SetFusionStats(stats); err != nil {
		log.Printf("⚠️ Failed to sync fusion stats to Redis: %v", err)
	}
	
	return decision
}

// GetRealTimeComplexity gets current complexity from Redis (for external access)
func (r *RedisFusionAlgorithm) GetRealTimeComplexity() (float64, error) {
	complexity, err := r.redis.GetCurrentComplexity()
	if err != nil {
		log.Printf("⚠️ Failed to get real-time complexity from Redis: %v", err)
		// Fallback to local algorithm state
		return r.FusionAlgorithm.getCurrentComplexity(), nil
	}
	return complexity, nil
}

// GetRealTimeStats gets current stats from Redis (for external access)
func (r *RedisFusionAlgorithm) GetRealTimeStats() (metrics.FusionStats, error) {
	stats, err := r.redis.GetFusionStats()
	if err != nil {
		log.Printf("⚠️ Failed to get real-time stats from Redis: %v", err)
		// Fallback to local algorithm state
		return r.FusionAlgorithm.GetStats(), nil
	}
	return stats, nil
}

// GetRealTimeSystemMetrics gets current system metrics from Redis
func (r *RedisFusionAlgorithm) GetRealTimeSystemMetrics() (metrics.SystemMetrics, error) {
	return r.redis.GetSystemMetrics()
}

// GetRealTimeScalingDecision gets latest scaling decision from Redis
func (r *RedisFusionAlgorithm) GetRealTimeScalingDecision() (metrics.ScalingDecision, error) {
	return r.redis.GetScalingDecision()
}

// HealthCheck checks both algorithm and Redis health
func (r *RedisFusionAlgorithm) HealthCheck() error {
	return r.redis.HealthCheck()
}

// Close closes Redis connection
func (r *RedisFusionAlgorithm) Close() error {
	if r.redis != nil {
		return r.redis.Close()
	}
	return nil
}

// SyncFromRedis loads state from Redis (useful for restarts/recovery)
func (r *RedisFusionAlgorithm) SyncFromRedis() error {
	// Get current complexity from Redis
	complexity, err := r.redis.GetCurrentComplexity()
	if err != nil {
		log.Printf("⚠️ Could not sync complexity from Redis: %v", err)
	} else {
		log.Printf("📊 Synced complexity %.2f from Redis", complexity)
	}
	
	// Get system metrics from Redis
	systemMetrics, err := r.redis.GetSystemMetrics()
	if err != nil {
		log.Printf("⚠️ Could not sync system metrics from Redis: %v", err)
	} else {
		log.Printf("📊 Synced system metrics from Redis (CPU: %.2f%%, Memory: %.2f%%)", 
			systemMetrics.CPUUsage, systemMetrics.MemoryUsage)
	}
	
	return nil
}

// GetRedisKeys returns current Redis keys (for debugging)
func (r *RedisFusionAlgorithm) GetRedisKeys() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	if complexity, err := r.redis.GetCurrentComplexity(); err == nil {
		result["current_complexity"] = complexity
	}
	
	if lastComplexity, err := r.redis.GetLastComplexity(); err == nil {
		result["last_complexity"] = lastComplexity
	}
	
	if stats, err := r.redis.GetFusionStats(); err == nil {
		result["fusion_stats"] = stats
	}
	
	if sysMetrics, err := r.redis.GetSystemMetrics(); err == nil {
		result["system_metrics"] = sysMetrics
	}
	
	if decision, err := r.redis.GetScalingDecision(); err == nil {
		result["scaling_decision"] = decision
	}
	
	return result, nil
}

// ========================================
// Original FusionAlgorithm (unchanged)
// ========================================

// FusionAlgorithm implements the hybrid auto-scaling algorithm
type FusionAlgorithm struct {
	config              metrics.FusionConfig
	history             metrics.MetricHistory
	correlationResults  metrics.CorrelationResults
	lastScalingDecision time.Time
	startTime           time.Time
	stats               metrics.FusionStats
	mutex               sync.RWMutex
}

// NewFusionAlgorithm creates a new fusion algorithm instance
func NewFusionAlgorithm(config metrics.FusionConfig) *FusionAlgorithm {
	return &FusionAlgorithm{
		config:    config,
		startTime: time.Now(),
		history: metrics.MetricHistory{
			ComplexityHistory: make([]metrics.ComplexityMetrics, 0, config.HistoryWindowSize),
			SystemHistory:     make([]metrics.SystemMetrics, 0, config.HistoryWindowSize),
			DecisionHistory:   make([]metrics.ScalingDecision, 0, config.HistoryWindowSize),
			MaxSize:          config.HistoryWindowSize,
		},
		stats: metrics.FusionStats{
			TotalDecisions:     0,
			ScaleUpDecisions:   0,
			ScaleDownDecisions: 0,
			NoActionDecisions:  0,
		},
	}
}

// AddComplexityData adds new complexity data to the algorithm
func (f *FusionAlgorithm) AddComplexityData(complexity float64, timestamp time.Time) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	complexityMetric := metrics.ComplexityMetrics{
		Score:     complexity,
		Timestamp: timestamp,
	}

	f.history.ComplexityHistory = append(f.history.ComplexityHistory, complexityMetric)
	
	// Keep history within bounds
	if len(f.history.ComplexityHistory) > f.history.MaxSize {
		f.history.ComplexityHistory = f.history.ComplexityHistory[1:]
	}
}

// AddSystemData adds new system metrics to the algorithm
func (f *FusionAlgorithm) AddSystemData(systemMetrics metrics.SystemMetrics, timestamp time.Time) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.history.SystemHistory = append(f.history.SystemHistory, systemMetrics)
	
	// Keep history within bounds
	if len(f.history.SystemHistory) > f.history.MaxSize {
		f.history.SystemHistory = f.history.SystemHistory[1:]
	}
}

// ProcessMetrics runs the fusion algorithm and makes scaling decisions
func (f *FusionAlgorithm) ProcessMetrics() metrics.ScalingDecision {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Check if we have enough data
	if len(f.history.ComplexityHistory) < f.config.MinSamplesRequired ||
		len(f.history.SystemHistory) < f.config.MinSamplesRequired {
		return f.createDecision("no_action", 0, "insufficient_data", 0)
	}

	// Check cooldown period
	if time.Since(f.lastScalingDecision) < f.config.CooldownPeriod {
		return f.createDecision("no_action", 0, "cooldown_period", 0)
	}

	// Step 1: Update correlation analysis
	f.updateCorrelationAnalysis()

	// Step 2: Get current metrics
	currentComplexity := f.getCurrentComplexity()
	currentCPU := f.getCurrentCPU()
	currentMemory := f.getCurrentMemory()

	// Step 3: Apply fusion algorithm
	fusionScore := f.calculateFusionScore(currentComplexity, currentCPU, currentMemory)

	// Step 4: Predict future resource usage
	predictedCPU := f.predictCPUUsage(currentComplexity)
	predictedMemory := f.predictMemoryUsage(currentComplexity)

	// Step 5: Make scaling decision
	decision := f.makeScalingDecision(fusionScore, predictedCPU, predictedMemory, currentComplexity)

	// Step 6: Update statistics
	f.updateStats(decision, currentComplexity, currentCPU, currentMemory)

	return decision
}

// calculateFusionScore - FUSION ALGORITHM!
func (f *FusionAlgorithm) calculateFusionScore(complexity, cpu, memory float64) float64 {
	// Normalize metrics to 0-1 scale
	normalizedComplexity := f.normalizeComplexity(complexity)
	normalizedCPU := cpu / 100.0
	normalizedMemory := memory / 100.0

	// Apply weights based on correlation strength
	complexityWeight := f.config.ComplexityWeight
	cpuWeight := f.config.CPUWeight
	memoryWeight := f.config.MemoryWeight

	// Adjust weights based on correlation analysis
	if f.correlationResults.ComplexityCPU > 0.7 {
		// Strong correlation - increase complexity weight
		complexityWeight *= 1.2
		cpuWeight *= 0.8
	}

	// Calculate weighted fusion score
	fusionScore := (normalizedComplexity * complexityWeight) +
		(normalizedCPU * cpuWeight) +
		(normalizedMemory * memoryWeight)

	// Apply non-linear scaling for extreme values
	if normalizedComplexity > 0.8 {
		// Very high complexity gets exponential boost
		fusionScore *= (1 + math.Pow(normalizedComplexity, 2))
	}

	return fusionScore
}

// predictCPUUsage predicts future CPU usage based on complexity
func (f *FusionAlgorithm) predictCPUUsage(complexity float64) float64 {
	if f.correlationResults.ComplexityCPU == 0 {
		return 0 // No prediction without correlation data
	}

	// Simple linear prediction based on correlation
	normalizedComplexity := f.normalizeComplexity(complexity)
	predictedCPU := normalizedComplexity * f.correlationResults.ComplexityCPU * 100

	// Add trend analysis
	if len(f.history.ComplexityHistory) >= 3 {
		trend := f.calculateComplexityTrend()
		predictedCPU += trend * 20 // Trend influence
	}

	return math.Max(0, math.Min(100, predictedCPU))
}

// predictMemoryUsage predicts future memory usage based on complexity
func (f *FusionAlgorithm) predictMemoryUsage(complexity float64) float64 {
	if f.correlationResults.ComplexityMemory == 0 {
		return 0
	}

	normalizedComplexity := f.normalizeComplexity(complexity)
	predictedMemory := normalizedComplexity * f.correlationResults.ComplexityMemory * 100

	return math.Max(0, math.Min(100, predictedMemory))
}

// makeScalingDecision decides whether to scale based on fusion score
func (f *FusionAlgorithm) makeScalingDecision(fusionScore, predictedCPU, predictedMemory, complexity float64) metrics.ScalingDecision {
	threshold := f.config.ScalingThreshold

	// Scenario-based decision making
	if fusionScore > threshold && complexity > 200 {
		// High complexity scenario - proactive scaling
		scalingFactor := math.Min(2.0, fusionScore/threshold)
		return f.createDecision("scale_up", scalingFactor, "high_complexity_predicted", 0.9)
	}

	if predictedCPU > 85 || predictedMemory > 90 {
		// Resource pressure predicted
		scalingFactor := 1.5
		return f.createDecision("scale_up", scalingFactor, "resource_pressure_predicted", 0.8)
	}

	if fusionScore < threshold*0.3 && f.getCurrentCPU() < 30 {
		// Scale down opportunity
		scalingFactor := 0.8
		return f.createDecision("scale_down", scalingFactor, "low_utilization", 0.6)
	}

	return f.createDecision("no_action", 1.0, "within_thresholds", 0.7)
}

// Helper methods for calculations
func (f *FusionAlgorithm) normalizeComplexity(complexity float64) float64 {
	// Normalize complexity to 0-1 scale (assuming max complexity of 1000)
	return math.Min(1.0, complexity/1000.0)
}

func (f *FusionAlgorithm) getCurrentComplexity() float64 {
	if len(f.history.ComplexityHistory) == 0 {
		return 0
	}
	return f.history.ComplexityHistory[len(f.history.ComplexityHistory)-1].Score
}

func (f *FusionAlgorithm) getCurrentCPU() float64 {
	if len(f.history.SystemHistory) == 0 {
		return 0
	}
	return f.history.SystemHistory[len(f.history.SystemHistory)-1].CPUUsage
}

func (f *FusionAlgorithm) getCurrentMemory() float64 {
	if len(f.history.SystemHistory) == 0 {
		return 0
	}
	return f.history.SystemHistory[len(f.history.SystemHistory)-1].MemoryUsage
}

func (f *FusionAlgorithm) calculateComplexityTrend() float64 {
	if len(f.history.ComplexityHistory) < 3 {
		return 0
	}

	recent := len(f.history.ComplexityHistory)
	latest := f.history.ComplexityHistory[recent-1].Score
	previous := f.history.ComplexityHistory[recent-2].Score
	older := f.history.ComplexityHistory[recent-3].Score

	trend := (latest - previous) + (previous - older)
	return trend / 2.0 // Average trend
}

// updateCorrelationAnalysis calculates correlation between complexity and resources
func (f *FusionAlgorithm) updateCorrelationAnalysis() {
	if len(f.history.ComplexityHistory) < f.config.MinSamplesRequired ||
		len(f.history.SystemHistory) < f.config.MinSamplesRequired {
		return
	}

	// Align data by timestamp (simple approach - match by position)
	minLen := min(len(f.history.ComplexityHistory), len(f.history.SystemHistory))
	
	complexityValues := make([]float64, minLen)
	cpuValues := make([]float64, minLen)
	memoryValues := make([]float64, minLen)

	for i := 0; i < minLen; i++ {
		complexityValues[i] = f.history.ComplexityHistory[i].Score
		cpuValues[i] = f.history.SystemHistory[i].CPUUsage
		memoryValues[i] = f.history.SystemHistory[i].MemoryUsage
	}

	f.correlationResults = metrics.CorrelationResults{
		ComplexityCPU:    f.calculateCorrelation(complexityValues, cpuValues),
		ComplexityMemory: f.calculateCorrelation(complexityValues, memoryValues),
		SampleSize:       minLen,
		LastUpdated:      time.Now(),
	}
}

// calculateCorrelation calculates Pearson correlation coefficient
func (f *FusionAlgorithm) calculateCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0
	}

	n := float64(len(x))
	sumX, sumY, sumXY, sumX2, sumY2 := 0.0, 0.0, 0.0, 0.0, 0.0

	for i := 0; i < len(x); i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
		sumY2 += y[i] * y[i]
	}

	numerator := n*sumXY - sumX*sumY
	denominator := math.Sqrt((n*sumX2 - sumX*sumX) * (n*sumY2 - sumY*sumY))

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}

// createDecision creates a scaling decision
func (f *FusionAlgorithm) createDecision(action string, factor float64, reason string, confidence float64) metrics.ScalingDecision {
	decision := metrics.ScalingDecision{
		ShouldScale:   action != "no_action",
		Action:        action,
		ScalingFactor: factor,
		Reason:        reason,
		Confidence:    confidence,
		Timestamp:     time.Now(),
	}

	f.history.DecisionHistory = append(f.history.DecisionHistory, decision)
	if len(f.history.DecisionHistory) > f.history.MaxSize {
		f.history.DecisionHistory = f.history.DecisionHistory[1:]
	}

	if decision.ShouldScale {
		f.lastScalingDecision = time.Now()
	}

	return decision
}

// updateStats updates algorithm statistics
func (f *FusionAlgorithm) updateStats(decision metrics.ScalingDecision, complexity, cpu, memory float64) {
	f.stats.TotalDecisions++
	
	switch decision.Action {
	case "scale_up":
		f.stats.ScaleUpDecisions++
	case "scale_down":
		f.stats.ScaleDownDecisions++
	default:
		f.stats.NoActionDecisions++
	}

	// Update running averages
	total := float64(f.stats.TotalDecisions)
	f.stats.AverageComplexity = ((f.stats.AverageComplexity * (total - 1)) + complexity) / total
	f.stats.AverageCPU = ((f.stats.AverageCPU * (total - 1)) + cpu) / total
	f.stats.AverageMemory = ((f.stats.AverageMemory * (total - 1)) + memory) / total

	f.stats.Correlation = f.correlationResults
	f.stats.LastDecision = &decision
	f.stats.UptimeSeconds = int64(time.Since(f.startTime).Seconds())
}

// GetStats returns current algorithm statistics
func (f *FusionAlgorithm) GetStats() metrics.FusionStats {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.stats
}

// GetHistory returns metric history
func (f *FusionAlgorithm) GetHistory() metrics.MetricHistory {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.history
}

// HealthCheck returns nil since FusionAlgorithm has no external dependencies
func (f *FusionAlgorithm) HealthCheck() error {
	return nil
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}