package api

import (
	"encoding/json"
	"net/http"
	"time"

	"hybrid-metrics/internal/fusion"
	"hybrid-metrics/pkg/metrics"
)

// FusionAlgorithmInterface defines the interface that both FusionAlgorithm and RedisFusionAlgorithm implement
type FusionAlgorithmInterface interface {
	GetStats() metrics.FusionStats
	GetHistory() metrics.MetricHistory
	ProcessMetrics() metrics.ScalingDecision
	AddComplexityData(complexity float64, timestamp time.Time)
	AddSystemData(systemMetrics metrics.SystemMetrics, timestamp time.Time)
	HealthCheck() error
}

// Server provides HTTP API for the hybrid metrics server
type Server struct {
	fusionAlgorithm FusionAlgorithmInterface
}

// NewServer creates a new API server - now accepts interface instead of concrete type
func NewServer(fusionAlgorithm FusionAlgorithmInterface) *Server {
	return &Server{
		fusionAlgorithm: fusionAlgorithm,
	}
}

// NewServerWithFusion creates a new API server with regular FusionAlgorithm (backward compatibility)
func NewServerWithFusion(fusionAlgorithm *fusion.FusionAlgorithm) *Server {
	return &Server{
		fusionAlgorithm: fusionAlgorithm,
	}
}

// NewServerWithRedis creates a new API server with RedisFusionAlgorithm
func NewServerWithRedis(fusionAlgorithm *fusion.RedisFusionAlgorithm) *Server {
	return &Server{
		fusionAlgorithm: fusionAlgorithm,
	}
}

// Router sets up HTTP routes
func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", s.handleHealth)
	
	// Metrics endpoints
	mux.HandleFunc("/metrics", s.handleMetrics)
	mux.HandleFunc("/fusion-stats", s.handleFusionStats)
	mux.HandleFunc("/history", s.handleHistory)
	
	// Configuration endpoints
	mux.HandleFunc("/config", s.handleConfig)
	
	// Redis-specific endpoints (if using RedisFusionAlgorithm)
	mux.HandleFunc("/redis-stats", s.handleRedisStats)
	mux.HandleFunc("/real-time-complexity", s.handleRealTimeComplexity)
	
	// Add CORS middleware
	return s.corsMiddleware(mux)
}

// handleHealth provides health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if fusion algorithm supports health check
	var healthStatus string = "healthy"
	var healthDetails interface{} = nil
	
	if err := s.fusionAlgorithm.HealthCheck(); err != nil {
		healthStatus = "degraded"
		healthDetails = map[string]string{
			"redis_error": err.Error(),
		}
	}

	response := map[string]interface{}{
		"status":    healthStatus,
		"service":   "hybrid-metrics-server",
		"version":   "1.0.0",
		"timestamp": time.Now(),
		"uptime":    time.Since(time.Now().Add(-time.Duration(s.fusionAlgorithm.GetStats().UptimeSeconds) * time.Second)),
	}
	
	if healthDetails != nil {
		response["details"] = healthDetails
	}

	s.writeJSONResponse(w, response)
}

// handleMetrics provides current metrics summary
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := s.fusionAlgorithm.GetStats()
	history := s.fusionAlgorithm.GetHistory()

	// Get latest metrics
	var latestComplexity, latestCPU, latestMemory float64
	if len(history.ComplexityHistory) > 0 {
		latestComplexity = history.ComplexityHistory[len(history.ComplexityHistory)-1].Score
	}
	if len(history.SystemHistory) > 0 {
		latest := history.SystemHistory[len(history.SystemHistory)-1]
		latestCPU = latest.CPUUsage
		latestMemory = latest.MemoryUsage
	}

	response := map[string]interface{}{
		"current_metrics": map[string]interface{}{
			"complexity": latestComplexity,
			"cpu_usage":  latestCPU,
			"memory_usage": latestMemory,
		},
		"averages": map[string]interface{}{
			"complexity": stats.AverageComplexity,
			"cpu":        stats.AverageCPU,
			"memory":     stats.AverageMemory,
		},
		"correlation": stats.Correlation,
		"last_decision": stats.LastDecision,
		"timestamp": time.Now(),
	}

	s.writeJSONResponse(w, response)
}

// handleFusionStats provides detailed fusion algorithm statistics
func (s *Server) handleFusionStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := s.fusionAlgorithm.GetStats()
	s.writeJSONResponse(w, stats)
}

// handleHistory provides metric history
func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters for filtering
	query := r.URL.Query()
	limit := 50 // default limit
	
	if limitStr := query.Get("limit"); limitStr != "" {
		if parsedLimit := parseInt(limitStr, 50); parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	history := s.fusionAlgorithm.GetHistory()

	// Limit the response size
	complexityHistory := history.ComplexityHistory
	if len(complexityHistory) > limit {
		complexityHistory = complexityHistory[len(complexityHistory)-limit:]
	}

	systemHistory := history.SystemHistory
	if len(systemHistory) > limit {
		systemHistory = systemHistory[len(systemHistory)-limit:]
	}

	decisionHistory := history.DecisionHistory
	if len(decisionHistory) > limit {
		decisionHistory = decisionHistory[len(decisionHistory)-limit:]
	}

	response := map[string]interface{}{
		"complexity_history": complexityHistory,
		"system_history":     systemHistory,
		"decision_history":   decisionHistory,
		"total_samples": map[string]interface{}{
			"complexity": len(history.ComplexityHistory),
			"system":     len(history.SystemHistory),
			"decisions":  len(history.DecisionHistory),
		},
		"timestamp": time.Now(),
	}

	s.writeJSONResponse(w, response)
}

// handleRedisStats provides Redis-specific statistics (only works with RedisFusionAlgorithm)
func (s *Server) handleRedisStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if this is a RedisFusionAlgorithm
	if redisFusion, ok := s.fusionAlgorithm.(*fusion.RedisFusionAlgorithm); ok {
		// Get real-time stats from Redis
		realTimeStats, err := redisFusion.GetRealTimeStats()
		if err != nil {
			response := map[string]interface{}{
				"error": "Failed to get Redis stats",
				"details": err.Error(),
				"fallback_stats": s.fusionAlgorithm.GetStats(),
			}
			s.writeJSONResponse(w, response)
			return
		}

		// Get Redis keys for debugging
		redisKeys, _ := redisFusion.GetRedisKeys()

		response := map[string]interface{}{
			"real_time_stats": realTimeStats,
			"redis_keys": redisKeys,
			"redis_health": "connected",
			"timestamp": time.Now(),
		}
		s.writeJSONResponse(w, response)
	} else {
		response := map[string]interface{}{
			"error": "Redis stats not available",
			"reason": "Not using RedisFusionAlgorithm",
			"algorithm_type": "FusionAlgorithm",
		}
		s.writeJSONResponse(w, response)
	}
}

// handleRealTimeComplexity provides real-time complexity from Redis
func (s *Server) handleRealTimeComplexity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if this is a RedisFusionAlgorithm
	if redisFusion, ok := s.fusionAlgorithm.(*fusion.RedisFusionAlgorithm); ok {
		complexity, err := redisFusion.GetRealTimeComplexity()
		if err != nil {
			response := map[string]interface{}{
				"error": "Failed to get real-time complexity",
				"details": err.Error(),
				"fallback_complexity": 0,
			}
			s.writeJSONResponse(w, response)
			return
		}

		response := map[string]interface{}{
			"current_complexity": complexity,
			"source": "redis",
			"timestamp": time.Now(),
		}
		s.writeJSONResponse(w, response)
	} else {
		// Fallback to local complexity
		history := s.fusionAlgorithm.GetHistory()
		var complexity float64
		if len(history.ComplexityHistory) > 0 {
			complexity = history.ComplexityHistory[len(history.ComplexityHistory)-1].Score
		}

		response := map[string]interface{}{
			"current_complexity": complexity,
			"source": "local",
			"timestamp": time.Now(),
		}
		s.writeJSONResponse(w, response)
	}
}

// handleConfig provides current configuration
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Determine algorithm type
	algorithmType := "FusionAlgorithm"
	redisEnabled := false
	if _, ok := s.fusionAlgorithm.(*fusion.RedisFusionAlgorithm); ok {
		algorithmType = "RedisFusionAlgorithm"
		redisEnabled = true
	}

	response := map[string]interface{}{
		"algorithm_type": algorithmType,
		"redis_enabled": redisEnabled,
		"fusion_algorithm": map[string]interface{}{
			"complexity_weight":     0.6,
			"cpu_weight":           0.3,
			"memory_weight":        0.1,
			"scaling_threshold":    0.75,
			"history_window_size":  100,
			"min_samples_required": 10,
		},
		"collection_intervals": map[string]interface{}{
			"complexity_seconds": 5,
			"system_seconds":     3,
			"fusion_seconds":     10,
		},
		"timestamp": time.Now(),
	}

	s.writeJSONResponse(w, response)
}

// writeJSONResponse writes a JSON response
func (s *Server) writeJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
		return
	}
}

// corsMiddleware adds CORS headers
func (s *Server) corsMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

// parseInt safely parses an integer with a default value
func parseInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	
	// Simple integer parsing (you might want to use strconv.Atoi for production)
	result := defaultValue
	if len(s) > 0 && s[0] >= '0' && s[0] <= '9' {
		result = int(s[0] - '0')
		for i := 1; i < len(s) && s[i] >= '0' && s[i] <= '9'; i++ {
			result = result*10 + int(s[i]-'0')
		}
	}
	
	return result
}