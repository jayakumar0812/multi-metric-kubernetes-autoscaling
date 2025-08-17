package state

import (
    "context"
    "encoding/json"
    "fmt"
    "strconv"
    "time"
    
    "github.com/redis/go-redis/v9"
)

type RedisStateManager struct {
    client *redis.Client
    ctx    context.Context
}

// Redis keys matching your hybrid-metrics service
const (
    // Core complexity data
    KeyCurrentComplexity = "hpa:complexity:current"
    KeyLastComplexity    = "hpa:complexity:last" 
    KeyComplexityHistory = "hpa:complexity:history"
    
    // Fusion algorithm results
    KeyFusionScore       = "hpa:fusion:score"
    KeyFusionStats       = "hpa:fusion:stats"
    KeyCorrelationData   = "hpa:correlation:results"
    
    // System metrics for context
    KeySystemMetrics     = "hpa:system:current"
    KeyScalingDecision   = "hpa:scaling:last"
    
    // GraphQL specific metrics
    KeyGraphQLQueries    = "hpa:graphql:queries"
    KeyQueryComplexity   = "hpa:graphql:complexity"
)

// Simple data structures for what custom-metrics-api needs
type ComplexityScore struct {
    Current   float64   `json:"current"`
    Previous  float64   `json:"previous"`
    Timestamp time.Time `json:"timestamp"`
}

type FusionResult struct {
    Score      float64   `json:"score"`
    Confidence float64   `json:"confidence"`
    Reason     string    `json:"reason"`
    Timestamp  time.Time `json:"timestamp"`
}

type CorrelationData struct {
    ComplexityCPU    float64   `json:"complexity_cpu"`
    ComplexityMemory float64   `json:"complexity_memory"`
    SampleSize       int       `json:"sample_size"`
    LastUpdated      time.Time `json:"last_updated"`
}

func NewRedisStateManager(redisURL string) (*RedisStateManager, error) {
    if redisURL == "" {
        redisURL = "redis-service:6379" // Kubernetes service name
    }
    
    rdb := redis.NewClient(&redis.Options{
        Addr:     redisURL,
        Password: "", // No password for internal K8s service
        DB:       0,  // Default DB
        PoolSize: 10,
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
    })
    
    ctx := context.Background()
    
    // Test connection
    if err := rdb.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("failed to connect to Redis: %w", err)
    }
    
    return &RedisStateManager{
        client: rdb,
        ctx:    ctx,
    }, nil
}

// Primary method for HPA - gets the current complexity score
func (r *RedisStateManager) GetCurrentComplexity() (float64, error) {
    val, err := r.client.Get(r.ctx, KeyCurrentComplexity).Float64()
    if err == redis.Nil {
        return 20.0, nil // Default baseline for GraphQL complexity
    }
    return val, err
}

// Get fusion algorithm result
func (r *RedisStateManager) GetFusionScore() (float64, error) {
    val, err := r.client.Get(r.ctx, KeyFusionScore).Float64()
    if err == redis.Nil {
        // Fallback to complexity if fusion not available
        return r.GetCurrentComplexity()
    }
    return val, err
}

// Get complete complexity data with context
func (r *RedisStateManager) GetComplexityData() (ComplexityScore, error) {
    current, err := r.GetCurrentComplexity()
    if err != nil {
        return ComplexityScore{}, err
    }
    
    previous, _ := r.client.Get(r.ctx, KeyLastComplexity).Float64()
    if previous == 0 {
        previous = current // If no previous data
    }
    
    return ComplexityScore{
        Current:   current,
        Previous:  previous,
        Timestamp: time.Now(),
    }, nil
}

// Get fusion algorithm results with confidence scoring
func (r *RedisStateManager) GetFusionResult() (FusionResult, error) {
    // Try to get structured fusion result
    data, err := r.client.Get(r.ctx, KeyScalingDecision).Result()
    if err == redis.Nil {
        // Fallback to simple fusion score
        score, _ := r.GetFusionScore()
        return FusionResult{
            Score:      score,
            Confidence: 0.5,
            Reason:     "fallback_mode",
            Timestamp:  time.Now(),
        }, nil
    }
    if err != nil {
        return FusionResult{}, err
    }
    
    var result map[string]interface{}
    if err := json.Unmarshal([]byte(data), &result); err != nil {
        return FusionResult{}, err
    }
    
    // Extract values safely
    score := 20.0
    confidence := 0.5
    reason := "default"
    
    if s, ok := result["scaling_factor"].(float64); ok {
        score = s * 50 // Convert scaling factor to complexity-like score
    }
    if c, ok := result["confidence"].(float64); ok {
        confidence = c
    }
    if r, ok := result["reason"].(string); ok {
        reason = r
    }
    
    return FusionResult{
        Score:      score,
        Confidence: confidence,
        Reason:     reason,
        Timestamp:  time.Now(),
    }, nil
}

// Get correlation data
func (r *RedisStateManager) GetCorrelationData() (CorrelationData, error) {
    data, err := r.client.Get(r.ctx, KeyCorrelationData).Result()
    if err == redis.Nil {
        return CorrelationData{
            ComplexityCPU:    0.0,
            ComplexityMemory: 0.0,
            SampleSize:       0,
            LastUpdated:      time.Now(),
        }, nil
    }
    if err != nil {
        return CorrelationData{}, err
    }
    
    var correlation CorrelationData
    err = json.Unmarshal([]byte(data), &correlation)
    return correlation, err
}

// Get recent complexity history for trend analysis
func (r *RedisStateManager) GetComplexityHistory(limit int) ([]float64, error) {
    if limit <= 0 {
        limit = 10 // Default to last 10 values
    }
    
    vals, err := r.client.LRange(r.ctx, KeyComplexityHistory, 0, int64(limit-1)).Result()
    if err != nil {
        return nil, err
    }
    
    complexities := make([]float64, 0, len(vals))
    for _, val := range vals {
        if complexity, parseErr := strconv.ParseFloat(val, 64); parseErr == nil {
            complexities = append(complexities, complexity)
        }
    }
    
    return complexities, nil
}

// Get the best metric for HPA scaling decisions
func (r *RedisStateManager) GetOptimalMetricForHPA() (float64, string, error) {
    fusionResult, err := r.GetFusionResult()
    if err == nil && fusionResult.Confidence > 0.7 {
        return fusionResult.Score, fmt.Sprintf("fusion_algorithm (confidence: %.2f)", fusionResult.Confidence), nil
    }
    
    // Try fusion score (simpler)
    fusionScore, err := r.GetFusionScore()
    if err == nil && fusionScore > 0 {
        return fusionScore, "fusion_score", nil
    }
    
    // Fallback to raw complexity
    complexity, err := r.GetCurrentComplexity()
    if err == nil {
        return complexity, "raw_complexity", nil
    }
    
    // Ultimate fallback
    return 25.0, "default_fallback", fmt.Errorf("all metrics unavailable")
}

// Health check for the GraphQL HPA system
func (r *RedisStateManager) HealthCheck() error {
    return r.client.Ping(r.ctx).Err()
}

// System status for debugging
func (r *RedisStateManager) GetSystemStatus() map[string]interface{} {
    status := make(map[string]interface{})
    
    // Check connectivity
    status["redis_connected"] = r.client.Ping(r.ctx).Err() == nil
    
    // Check data availability
    complexity, err := r.GetCurrentComplexity()
    status["complexity_available"] = err == nil
    status["current_complexity"] = complexity
    
    fusionScore, err := r.GetFusionScore()
    status["fusion_available"] = err == nil
    status["fusion_score"] = fusionScore
    
    // Check data freshness
    correlationData, _ := r.GetCorrelationData()
    status["correlation_samples"] = correlationData.SampleSize
    status["last_correlation_update"] = correlationData.LastUpdated
    
    return status
}

// Close connection
func (r *RedisStateManager) Close() error {
    return r.client.Close()
}