// CREATE: services/custom-metrics-api/src/main.go
// Updated to use your actual endpoints

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// FusionStatsResponse represents response from your fusion stats endpoint
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
	hybridMetricsURL string // Your HTTP API server (port 3001)
	graphqlStatsURL  string // Your GraphQL REST endpoints (port 4001)
	httpClient       *http.Client
}

// NewComplexityMetricsServer creates a new server
func NewComplexityMetricsServer() *ComplexityMetricsServer {
	// Your HTTP API server for fusion stats
	hybridMetricsURL := os.Getenv("HYBRID_METRICS_URL")
	if hybridMetricsURL == "" {
		hybridMetricsURL = "http://localhost:3001" // Your HTTP API server
	}

	// Your GraphQL server REST endpoints for complexity stats
	graphqlStatsURL := os.Getenv("GRAPHQL_STATS_URL")
	if graphqlStatsURL == "" {
		graphqlStatsURL = "http://localhost:4001" // Your GraphQL REST endpoints
	}

	return &ComplexityMetricsServer{
		hybridMetricsURL: hybridMetricsURL,
		graphqlStatsURL:  graphqlStatsURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// fetchComplexityFromServers gets complexity from your servers
func (s *ComplexityMetricsServer) fetchComplexityFromServers() (float64, error) {
	// Try to get fusion stats first (most comprehensive)
	complexity, err := s.fetchFromFusionStats()
	if err == nil && complexity > 0 {
		return complexity, nil
	}

	// Fallback to GraphQL complexity stats
	complexity, err = s.fetchFromGraphQLStats()
	if err == nil && complexity > 0 {
		return complexity, nil
	}

	// Final fallback - check if servers are healthy
	return s.tryHealthEndpoints()
}

// fetchFromFusionStats gets data from your HTTP API server fusion stats
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
		return 20.0, nil
	}

	// Check GraphQL server health
	resp2, err2 := s.httpClient.Get(fmt.Sprintf("%s/health", s.graphqlStatsURL))
	if err2 == nil && resp2.StatusCode == http.StatusOK {
		resp2.Body.Close()
		log.Printf("⚠️ GraphQL server healthy but no complexity stats. Using baseline.")
		return 20.0, nil
	}

	return 0, fmt.Errorf("both servers unreachable - HTTP API: %v, GraphQL: %v", err1, err2)
}

// Health check endpoint
func (s *ComplexityMetricsServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	complexity, err := s.fetchComplexityFromServers()
	
	w.Header().Set("Content-Type", "application/json")
	
	status := map[string]interface{}{
		"service":            "custom-metrics-api-server",
		"hybrid_metrics_url": s.hybridMetricsURL,
		"graphql_stats_url":  s.graphqlStatsURL,
		"timestamp":          time.Now().Format(time.RFC3339),
	}

	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		status["status"] = "unhealthy"
		status["error"] = err.Error()
		status["current_complexity"] = 20.0
		status["complexity_source"] = "baseline_fallback"
	} else {
		status["status"] = "healthy"
		status["current_complexity"] = complexity
		status["complexity_source"] = "live_servers"
	}

	json.NewEncoder(w).Encode(status)
}

// Prometheus-style metrics endpoint
func (s *ComplexityMetricsServer) metricsHandler(w http.ResponseWriter, r *http.Request) {
	complexity, err := s.fetchComplexityFromServers()
	if err != nil {
		complexity = 20.0 // Use baseline if fetch fails
		log.Printf("⚠️ Using baseline complexity due to error: %v", err)
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "# HELP graphql_complexity_current Current GraphQL query complexity\n")
	fmt.Fprintf(w, "# TYPE graphql_complexity_current gauge\n")
	fmt.Fprintf(w, "graphql_complexity_current %.2f\n", complexity)
	
	fmt.Fprintf(w, "# HELP graphql_complexity_for_hpa Complexity metric formatted for HPA\n")
	fmt.Fprintf(w, "# TYPE graphql_complexity_for_hpa gauge\n")
	fmt.Fprintf(w, "graphql_complexity_for_hpa %.0f\n", complexity*100) // Multiply by 100 for HPA precision
}

// External metrics endpoint (simplified K8s external metrics API format)
func (s *ComplexityMetricsServer) externalMetricsHandler(w http.ResponseWriter, r *http.Request) {
	complexity, err := s.fetchComplexityFromServers()
	if err != nil {
		complexity = 20.0 // Use baseline if fetch fails
	}

	// Simplified external metrics response for HPA
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
				"value":     fmt.Sprintf("%.0f", complexity*100), // K8s expects string for quantity
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Status endpoint showing all current info
func (s *ComplexityMetricsServer) statusHandler(w http.ResponseWriter, r *http.Request) {
	complexity, err := s.fetchComplexityFromServers()
	
	status := map[string]interface{}{
		"service":            "custom-metrics-api-server",
		"version":            "1.0.0",
		"hybrid_metrics_url": s.hybridMetricsURL,
		"graphql_stats_url":  s.graphqlStatsURL,
		"timestamp":          time.Now().Format(time.RFC3339),
		"endpoints": map[string]string{
			"fusion_stats":     s.hybridMetricsURL + "/fusion-stats",
			"complexity_stats": s.graphqlStatsURL + "/complexity-stats",
			"http_health":      s.hybridMetricsURL + "/health",
			"graphql_health":   s.graphqlStatsURL + "/health",
		},
	}

	if err != nil {
		status["status"] = "partial"
		status["error"] = err.Error()
		status["current_complexity"] = 20.0
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
	log.Printf("🚀 Starting Custom Metrics API Server for GraphQL Complexity...")
	
	// Create server
	server := NewComplexityMetricsServer()
	
	log.Printf("📊 HTTP API Server (Fusion): %s", server.hybridMetricsURL)
	log.Printf("📈 GraphQL Stats Server: %s", server.graphqlStatsURL)
	
	// Test connections
	complexity, err := server.fetchComplexityFromServers()
	if err != nil {
		log.Printf("⚠️  Warning: Cannot connect to servers: %v", err)
		log.Printf("💡 Server will start anyway and use baseline complexity")
		log.Printf("🔍 Check these endpoints:")
		log.Printf("   - Fusion stats: %s/fusion-stats", server.hybridMetricsURL)
		log.Printf("   - Complexity stats: %s/complexity-stats", server.graphqlStatsURL)
	} else {
		log.Printf("✅ Successfully connected to servers. Current complexity: %.2f", complexity)
	}

	// Set up HTTP handlers
	http.HandleFunc("/health", server.healthHandler)
	http.HandleFunc("/ready", server.healthHandler)
	http.HandleFunc("/metrics", server.metricsHandler)
	http.HandleFunc("/external-metrics", server.externalMetricsHandler)
	http.HandleFunc("/status", server.statusHandler)
	
	// Root handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			server.statusHandler(w, r)
			return
		}
		http.NotFound(w, r)
	})

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	
	log.Printf("📊 Custom Metrics API Server starting on port %s", port)
	log.Printf("🔍 Health check: http://localhost:%s/health", port)
	log.Printf("📈 Metrics: http://localhost:%s/metrics", port)
	log.Printf("🎯 External metrics: http://localhost:%s/external-metrics", port)
	log.Printf("📋 Status: http://localhost:%s/status", port)
	
	log.Printf("🎯 Custom Metrics API Server ready!")
	log.Printf("📈 Bridging complexity from ports 3001 & 4001 to Kubernetes format")
	
	// Start the server
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}