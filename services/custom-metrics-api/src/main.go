package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"custom-metrics-api/pkg/state"
)

// FusionStatsResponse represents response fusion stats endpoint
type FusionStatsResponse struct {
	RealUserQueries     int     `json:"real_user_queries"`
	LastComplexity      float64 `json:"last_complexity"`
	CurrentComplexity   float64 `json:"current_complexity"`
	AverageComplexity   float64 `json:"average_complexity"`
	IsCurrentlyActive   bool    `json:"is_currently_active"`
	ActivityStatus      string  `json:"activity_status"`
	CollectorType       string  `json:"collector_type"`
}

// ComplexityStatsResponse represents response from GraphQL complexity stats
type ComplexityStatsResponse struct {
	CacheSize     int `json:"cacheSize"`
	CachedQueries []struct {
		Query      string  `json:"query"`
		Complexity float64 `json:"complexity"`
	} `json:"cachedQueries"`
	LastUpdated string `json:"lastUpdated"`
}

// ComplexityMetricsServer serves complexity metrics for Kubernetes
type ComplexityMetricsServer struct {
	hybridMetricsURL string //  HTTP API server (port 3001)
	graphqlStatsURL  string //  GraphQL REST endpoints (port 4001)
	redisURL         string //  Redis connection string
	httpClient       *http.Client
	tlsCertFile      string // Path to TLS certificate
	tlsKeyFile       string // Path to TLS private key
}

// NewComplexityMetricsServer creates a new server
func NewComplexityMetricsServer() *ComplexityMetricsServer {
	// Get URLs from environment variables with fallbacks
	hybridMetricsURL := os.Getenv("HYBRID_METRICS_URL")
	if hybridMetricsURL == "" {
		hybridMetricsURL = "http://hybrid-metrics-server:3001"
	}
	
	graphqlStatsURL := os.Getenv("GRAPHQL_STATS_URL")
	if graphqlStatsURL == "" {
		graphqlStatsURL = "http://graphql-server:4001"
	}

	// Get Redis URL
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis-service:6379"
	}

	// Get TLS certificate paths
	tlsCertFile := os.Getenv("TLS_CERT_FILE")
	if tlsCertFile == "" {
		tlsCertFile = "/etc/certs/tls.crt"
	}
	
	tlsKeyFile := os.Getenv("TLS_KEY_FILE")
	if tlsKeyFile == "" {
		tlsKeyFile = "/etc/certs/tls.key"
	}

	log.Printf("📊 Hybrid Metrics URL: %s", hybridMetricsURL)
	log.Printf("📊 GraphQL Stats URL: %s", graphqlStatsURL)
	log.Printf("🔗 Redis URL: %s", redisURL)
	log.Printf("🔒 TLS Certificate: %s", tlsCertFile)
	log.Printf("🔒 TLS Private Key: %s", tlsKeyFile)

	return &ComplexityMetricsServer{
		hybridMetricsURL: hybridMetricsURL,
		graphqlStatsURL:  graphqlStatsURL,
		redisURL:         redisURL,
		tlsCertFile:      tlsCertFile,
		tlsKeyFile:       tlsKeyFile,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// checkTLSFiles verifies that TLS certificate files exist and are readable
func (s *ComplexityMetricsServer) checkTLSFiles() error {
	// Check certificate file
	if _, err := os.Stat(s.tlsCertFile); os.IsNotExist(err) {
		return fmt.Errorf("TLS certificate file not found: %s", s.tlsCertFile)
	}
	
	// Check private key file
	if _, err := os.Stat(s.tlsKeyFile); os.IsNotExist(err) {
		return fmt.Errorf("TLS private key file not found: %s", s.tlsKeyFile)
	}
	
	// Try to load the certificate pair to validate
	_, err := tls.LoadX509KeyPair(s.tlsCertFile, s.tlsKeyFile)
	if err != nil {
		return fmt.Errorf("failed to load TLS certificate pair: %w", err)
	}
	
	log.Printf("✅ TLS certificate files validated successfully")
	return nil
}

// fetchComplexityFromRedis gets complexity from Redis
func (s *ComplexityMetricsServer) fetchComplexityFromRedis() (float64, error) {
	redisState, err := state.NewRedisStateManager(s.redisURL)
	if err != nil {
		return 0, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	defer redisState.Close()
	
	// Get real-time complexity from Redis
	complexity, err := redisState.GetCurrentComplexity()
	if err != nil {
		return 0, fmt.Errorf("failed to get complexity from Redis: %w", err)
	}
	
	log.Printf("📊 Redis - Current complexity: %.2f", complexity)
	return complexity, nil
}

// fetchComplexityFromServers gets complexity from servers with Redis priority
func (s *ComplexityMetricsServer) fetchComplexityFromServers() (float64, error) {
	// PRIORITY 1: Try Redis first (real-time state)
	complexity, err := s.fetchComplexityFromRedis()
	if err == nil && complexity > 0 {
		log.Printf("✅ Using Redis complexity: %.2f", complexity)
		return complexity, nil
	}
	log.Printf("⚠️ Redis failed: %v", err)

	// PRIORITY 2: Try HTTP API fusion stats (fallback)
	complexity, err = s.fetchFromFusionStats()
	if err == nil && complexity > 0 {
		log.Printf("⚠️ Using HTTP fusion stats complexity: %.2f", complexity)
		return complexity, nil
	}
	log.Printf("⚠️ Fusion stats failed: %v", err)

	// PRIORITY 3: Try GraphQL complexity stats (second fallback)
	complexity, err = s.fetchFromGraphQLStats()
	if err == nil && complexity > 0 {
		log.Printf("⚠️ Using GraphQL stats complexity: %.2f", complexity)
		return complexity, nil
	}
	log.Printf("⚠️ GraphQL stats failed: %v", err)

	// PRIORITY 4: Final fallback - check if servers are healthy
	complexity, err = s.tryHealthEndpoints()
	if err == nil {
		log.Printf("⚠️ Using baseline complexity: %.2f", complexity)
		return complexity, nil
	}

	return 0, fmt.Errorf("all data sources failed - Redis: %v", err)
}

// fetchFromFusionStats gets data from HTTP API server fusion stats
func (s *ComplexityMetricsServer) fetchFromFusionStats() (float64, error) {
	url := fmt.Sprintf("%s/fusion-stats", s.hybridMetricsURL)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create fusion stats request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to call fusion stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("fusion stats returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read fusion stats response: %w", err)
	}

	var fusionData FusionStatsResponse
	if err := json.Unmarshal(body, &fusionData); err != nil {
		return 0, fmt.Errorf("failed to parse fusion stats response: %w", err)
	}

	complexity := fusionData.CurrentComplexity
	if complexity <= 0 {
		complexity = fusionData.LastComplexity
	}
	if complexity <= 0 {
		complexity = fusionData.AverageComplexity
	}

	log.Printf("📊 Fusion Stats - Complexity: %.2f (Active: %v, Queries: %d)", 
		complexity, fusionData.IsCurrentlyActive, fusionData.RealUserQueries)

	return complexity, nil
}

// fetchFromGraphQLStats gets data from GraphQL complexity stats
func (s *ComplexityMetricsServer) fetchFromGraphQLStats() (float64, error) {
	url := fmt.Sprintf("%s/complexity-stats", s.graphqlStatsURL)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create complexity stats request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to call complexity stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("complexity stats returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read complexity stats response: %w", err)
	}

	var complexityData ComplexityStatsResponse
	if err := json.Unmarshal(body, &complexityData); err != nil {
		return 0, fmt.Errorf("failed to parse complexity stats response: %w", err)
	}

	// Get the latest query complexity
	if len(complexityData.CachedQueries) > 0 {
		latestQuery := complexityData.CachedQueries[len(complexityData.CachedQueries)-1]
		complexity := latestQuery.Complexity
		
		log.Printf("📈 GraphQL Stats - Latest complexity: %.2f (Cache size: %d)", 
			complexity, complexityData.CacheSize)
		
		return complexity, nil
	}

	return 0, fmt.Errorf("no cached queries found")
}

// tryHealthEndpoints checks if servers are healthy and returns baseline
func (s *ComplexityMetricsServer) tryHealthEndpoints() (float64, error) {
	// Check HTTP API server health
	resp1, err1 := s.httpClient.Get(fmt.Sprintf("%s/health", s.hybridMetricsURL))
	if err1 == nil && resp1.StatusCode == http.StatusOK {
		resp1.Body.Close()
		log.Printf("⚠️ HTTP API server healthy but no stats available. Using baseline.")
		return 50.0, nil // Baseline complexity
	}

	// Check GraphQL server health
	resp2, err2 := s.httpClient.Get(fmt.Sprintf("%s/health", s.graphqlStatsURL))
	if err2 == nil && resp2.StatusCode == http.StatusOK {
		resp2.Body.Close()
		log.Printf("⚠️ GraphQL server healthy but no complexity stats. Using baseline.")
		return 50.0, nil // Baseline complexity
	}

	return 0, fmt.Errorf("both servers unreachable - HTTP API: %v, GraphQL: %v", err1, err2)
}

// KUBERNETES API DISCOVERY ENDPOINT
func (s *ComplexityMetricsServer) apiDiscoveryHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔍 Kubernetes API discovery request from: %s (HTTPS)", r.RemoteAddr)
	
	response := map[string]interface{}{
		"kind":         "APIResourceList",
		"apiVersion":   "v1",
		"groupVersion": "external.metrics.k8s.io/v1beta1",
		"resources": []map[string]interface{}{
			{
				"name":         "graphql_complexity_score",
				"singularName": "",
				"namespaced":   true,
				"kind":         "ExternalMetricValueList",
				"verbs":        []string{"get"},
			},
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("❌ Failed to encode API discovery response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// KUBERNETES EXTERNAL METRICS ENDPOINT
func (s *ComplexityMetricsServer) kubernetesExternalMetricsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("📊 Kubernetes external metrics request from: %s (HTTPS)", r.RemoteAddr)
	
	complexity, err := s.fetchComplexityFromServers()
	if err != nil {
		complexity = 50.0 // Use baseline if fetch fails
		log.Printf("⚠️ Using baseline complexity for Kubernetes: %.2f (error: %v)", complexity, err)
	} else {
		log.Printf("✅ Returning complexity to Kubernetes HPA: %.2f", complexity)
	}
	
	response := map[string]interface{}{
		"kind":       "ExternalMetricValueList",
		"apiVersion": "external.metrics.k8s.io/v1beta1",
		"metadata":   map[string]interface{}{},
		"items": []map[string]interface{}{
			{
				"metricName": "graphql_complexity_score",
				"metricLabels": map[string]string{
					"service": "graphql-server",
				},
				"timestamp": time.Now().Format(time.RFC3339),
				"value": fmt.Sprintf("%d", int(complexity*1000)), // K8s expects string, multiply by 1000 for precision
			},
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("❌ Failed to encode external metrics response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// Health check endpoint (supports both /health and /healthz)
func (s *ComplexityMetricsServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	complexity, err := s.fetchComplexityFromServers()
	
	w.Header().Set("Content-Type", "application/json")
	
	status := map[string]interface{}{
		"service":            "custom-metrics-api-server",
		"protocol":           "HTTPS",
		"hybrid_metrics_url": s.hybridMetricsURL,
		"graphql_stats_url":  s.graphqlStatsURL,
		"redis_url":          s.redisURL,
		"timestamp":          time.Now().Format(time.RFC3339),
	}

	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		status["status"] = "unhealthy"
		status["error"] = err.Error()
		status["current_complexity"] = 50.0
		status["complexity_source"] = "baseline_fallback"
	} else {
		status["status"] = "healthy"
		status["current_complexity"] = complexity
		status["complexity_source"] = "live_servers"
	}

	json.NewEncoder(w).Encode(status)
}

// Redis health check endpoint
func (s *ComplexityMetricsServer) redisHealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Test Redis connection
	complexity, err := s.fetchComplexityFromRedis()
	
	status := map[string]interface{}{
		"service":   "custom-metrics-api-server",
		"redis_url": s.redisURL,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		status["status"] = "redis_unavailable"
		status["error"] = err.Error()
	} else {
		status["status"] = "redis_connected"
		status["current_complexity"] = complexity
	}

	json.NewEncoder(w).Encode(status)
}

// Simple readiness check
func (s *ComplexityMetricsServer) readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "ready",
		"service":  "custom-metrics-api-server",
		"protocol": "HTTPS",
	})
}

// Status endpoint showing all current info
func (s *ComplexityMetricsServer) statusHandler(w http.ResponseWriter, r *http.Request) {
	complexity, err := s.fetchComplexityFromServers()
	
	status := map[string]interface{}{
		"service":            "custom-metrics-api-server",
		"version":            "1.0.0-redis",
		"protocol":           "HTTPS",
		"hybrid_metrics_url": s.hybridMetricsURL,
		"graphql_stats_url":  s.graphqlStatsURL,
		"redis_url":          s.redisURL,
		"timestamp":          time.Now().Format(time.RFC3339),
		"tls_cert_file":      s.tlsCertFile,
		"endpoints": map[string]string{
			"fusion_stats":     s.hybridMetricsURL + "/fusion-stats",
			"complexity_stats": s.graphqlStatsURL + "/complexity-stats",
			"http_health":      s.hybridMetricsURL + "/health",
			"graphql_health":   s.graphqlStatsURL + "/health",
			"redis_health":     "/redis-health",
		},
	}

	if err != nil {
		status["status"] = "partial"
		status["error"] = err.Error()
		status["current_complexity"] = 50.0
		status["complexity_source"] = "baseline_fallback"
	} else {
		status["status"] = "healthy"
		status["current_complexity"] = complexity
		status["complexity_source"] = "live_servers"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func main() {
	log.Printf("🚀 Starting Custom Metrics API Server with Redis for GraphQL Complexity (HTTPS)...")
	
	// Create server
	server := NewComplexityMetricsServer()
	
	log.Printf("📊 HTTP API Server (Fusion): %s", server.hybridMetricsURL)
	log.Printf("📈 GraphQL Stats Server: %s", server.graphqlStatsURL)
	log.Printf("🔗 Redis Server: %s", server.redisURL)
	
	// Check TLS certificate files
	if err := server.checkTLSFiles(); err != nil {
		log.Fatalf("❌ TLS setup failed: %v", err)
	}
	
	// Test connections to other services
	complexity, err := server.fetchComplexityFromServers()
	if err != nil {
		log.Printf("⚠️  Warning: Cannot connect to servers: %v", err)
		log.Printf("💡 Server will start anyway and use baseline complexity (50.0)")
		log.Printf("🔍 Check these endpoints:")
		log.Printf("   - Redis: %s", server.redisURL)
		log.Printf("   - Fusion stats: %s/fusion-stats", server.hybridMetricsURL)
		log.Printf("   - Complexity stats: %s/complexity-stats", server.graphqlStatsURL)
	} else {
		log.Printf("✅ Successfully connected to servers. Current complexity: %.2f", complexity)
	}

	// KUBERNETES API ENDPOINTS (Required for HPA)
	http.HandleFunc("/apis/external.metrics.k8s.io/v1beta1", server.apiDiscoveryHandler)
	http.HandleFunc("/apis/external.metrics.k8s.io/v1beta1/namespaces/default/graphql_complexity_score", server.kubernetesExternalMetricsHandler)
	
	// HEALTH CHECK ENDPOINTS (Support both formats)
	http.HandleFunc("/health", server.healthHandler)
	http.HandleFunc("/healthz", server.healthHandler)  // K8s probes
	http.HandleFunc("/ready", server.readinessHandler)
	http.HandleFunc("/readyz", server.readinessHandler) // K8s probes
	http.HandleFunc("/redis-health", server.redisHealthHandler) // Redis-specific health
	
	// OTHER ENDPOINTS
	http.HandleFunc("/status", server.statusHandler)
	
	// Root handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			server.statusHandler(w, r)
			return
		}
		http.NotFound(w, r)
	})

	// Get port from environment or use default HTTPS port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8443" // Changed to HTTPS port
	}
	
	log.Printf("🔒 Custom Metrics API Server starting on HTTPS port %s", port)
	log.Printf("🔍 Health check: https://localhost:%s/health", port)
	log.Printf("🔍 Health check (K8s): https://localhost:%s/healthz", port)
	log.Printf("🔗 Redis health: https://localhost:%s/redis-health", port)
	log.Printf("📋 Status: https://localhost:%s/status", port)
	log.Printf("🏷️  Kubernetes API discovery: https://localhost:%s/apis/external.metrics.k8s.io/v1beta1", port)
	log.Printf("📊 Kubernetes metrics: https://localhost:%s/apis/external.metrics.k8s.io/v1beta1/namespaces/default/graphql_complexity_score", port)
	
	log.Printf("🎯 Custom Metrics API Server ready! (HTTPS + Redis)")
	log.Printf("📈 Bridging complexity from Redis → HTTP APIs → Kubernetes format")
	
	// Start the HTTPS server
	if err := http.ListenAndServeTLS(":"+port, server.tlsCertFile, server.tlsKeyFile, nil); err != nil {
		log.Fatalf("❌ HTTPS Server failed to start: %v", err)
	}
}