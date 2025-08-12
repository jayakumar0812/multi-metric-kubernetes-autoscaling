package state

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/redis/go-redis/v9"
    "hybrid-metrics/pkg/metrics"
)

type RedisStateManager struct {
    client *redis.Client
    ctx    context.Context
}

// Redis keys
const (
    KeyCurrentComplexity = "hpa:complexity:current"
    KeyLastComplexity    = "hpa:complexity:last"
    KeySystemMetrics     = "hpa:system:current"
    KeyScalingDecision   = "hpa:scaling:last"
    KeyFusionStats       = "hpa:fusion:stats"
)

func NewRedisStateManager(redisURL string) (*RedisStateManager, error) {
    if redisURL == "" {
        redisURL = "redis-service:6379" // Kubernetes service name
    }
    
    rdb := redis.NewClient(&redis.Options{
        Addr:         redisURL,
        Password:     "", // No password
        DB:           0,  // Default DB
        PoolSize:     10,
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

// Complexity Operations
func (r *RedisStateManager) SetCurrentComplexity(complexity float64) error {
    return r.client.Set(r.ctx, KeyCurrentComplexity, complexity, 5*time.Minute).Err()
}

func (r *RedisStateManager) GetCurrentComplexity() (float64, error) {
    val, err := r.client.Get(r.ctx, KeyCurrentComplexity).Float64()
    if err == redis.Nil {
        return 20.0, nil // Default baseline
    }
    return val, err
}

func (r *RedisStateManager) SetLastComplexity(complexity float64) error {
    return r.client.Set(r.ctx, KeyLastComplexity, complexity, 10*time.Minute).Err()
}

func (r *RedisStateManager) GetLastComplexity() (float64, error) {
    val, err := r.client.Get(r.ctx, KeyLastComplexity).Float64()
    if err == redis.Nil {
        return 20.0, nil
    }
    return val, err
}

// System Metrics Operations
func (r *RedisStateManager) SetSystemMetrics(metrics metrics.SystemMetrics) error {
    data, err := json.Marshal(metrics)
    if err != nil {
        return err
    }
    return r.client.Set(r.ctx, KeySystemMetrics, data, 2*time.Minute).Err()
}

func (r *RedisStateManager) GetSystemMetrics() (metrics.SystemMetrics, error) {
    var sysMetrics metrics.SystemMetrics
    data, err := r.client.Get(r.ctx, KeySystemMetrics).Result()
    if err == redis.Nil {
        return metrics.SystemMetrics{
            CPUUsage:    20.0,
            MemoryUsage: 30.0,
            Timestamp:   time.Now(),
        }, nil
    }
    if err != nil {
        return sysMetrics, err
    }
    
    err = json.Unmarshal([]byte(data), &sysMetrics)
    return sysMetrics, err
}

// Scaling Decision Operations
func (r *RedisStateManager) SetScalingDecision(decision metrics.ScalingDecision) error {
    data, err := json.Marshal(decision)
    if err != nil {
        return err
    }
    return r.client.Set(r.ctx, KeyScalingDecision, data, 15*time.Minute).Err()
}

func (r *RedisStateManager) GetScalingDecision() (metrics.ScalingDecision, error) {
    var decision metrics.ScalingDecision
    data, err := r.client.Get(r.ctx, KeyScalingDecision).Result()
    if err == redis.Nil {
        return metrics.ScalingDecision{
            ShouldScale:   false,
            Action:        "no_action",
            ScalingFactor: 1.0,
            Reason:        "no_data",
            Confidence:    0.0,
            Timestamp:     time.Now(),
        }, nil
    }
    if err != nil {
        return decision, err
    }
    
    err = json.Unmarshal([]byte(data), &decision)
    return decision, err
}

// Fusion Stats Operations
func (r *RedisStateManager) SetFusionStats(stats metrics.FusionStats) error {
    data, err := json.Marshal(stats)
    if err != nil {
        return err
    }
    return r.client.Set(r.ctx, KeyFusionStats, data, 5*time.Minute).Err()
}

func (r *RedisStateManager) GetFusionStats() (metrics.FusionStats, error) {
    var stats metrics.FusionStats
    data, err := r.client.Get(r.ctx, KeyFusionStats).Result()
    if err == redis.Nil {
        return metrics.FusionStats{
            TotalDecisions:     0,
            ScaleUpDecisions:   0,
            ScaleDownDecisions: 0,
            NoActionDecisions:  0,
            AverageComplexity:  20.0,
            AverageCPU:         20.0,
            AverageMemory:      30.0,
        }, nil
    }
    if err != nil {
        return stats, err
    }
    
    err = json.Unmarshal([]byte(data), &stats)
    return stats, err
}

// Health Check
func (r *RedisStateManager) HealthCheck() error {
    return r.client.Ping(r.ctx).Err()
}

// Close connection
func (r *RedisStateManager) Close() error {
    return r.client.Close()
}
