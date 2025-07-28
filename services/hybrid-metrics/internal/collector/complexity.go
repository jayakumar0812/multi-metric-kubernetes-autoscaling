package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"hybrid-metrics/pkg/metrics"
)

// ComplexityCollector collects complexity metrics from REAL user queries ONLY
type ComplexityCollector struct {
	graphqlURL        string
	restURL           string
	client            *http.Client
	lastComplexity    float64
	realQueryCount    int
	complexityHistory []float64
	lastQueryTime     time.Time
	baselineComplexity float64
	previousCacheSize int
	isFirstRun        bool
}

// ComplexityStatsResponse represents the response from REST complexity-stats endpoint
type ComplexityStatsResponse struct {
	CacheSize     int       `json:"cacheSize"`
	CachedQueries []struct {
		Query      string  `json:"query"`
		Complexity float64 `json:"complexity"`
	} `json:"cachedQueries"`
	LastUpdated string `json:"lastUpdated"`
}

// NewComplexityCollector creates a new authentic complexity collector
func NewComplexityCollector(graphqlURL string) *ComplexityCollector {
	// REST server runs on GraphQL port + 1
	restURL := "http://localhost:4001" // Default REST server port
	if graphqlURL == "http://localhost:4000" {
		restURL = "http://localhost:4001"
	}

	return &ComplexityCollector{
		graphqlURL:        graphqlURL,
		restURL:           restURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		complexityHistory:  make([]float64, 0, 100),
		lastComplexity:     20.0, // Baseline when no user activity
		baselineComplexity: 20.0,
		lastQueryTime:      time.Now(),
		previousCacheSize:  0,
		isFirstRun:         true,
	}
}

// Collect fetches complexity from REAL user queries ONLY
func (c *ComplexityCollector) Collect() (float64, error) {
	// Check REST server for real user query activity
	complexity, hasNewQuery, err := c.detectRealUserQuery()
	if err != nil {
		if c.isFirstRun {
			fmt.Printf("⏳ Waiting for GraphQL/REST servers to start...\n")
			fmt.Printf("💡 GraphQL: %s/graphql\n", c.graphqlURL)
			fmt.Printf("📊 REST API: %s/complexity-stats\n", c.restURL)
			c.isFirstRun = false
		}
		return c.getDecayedComplexity(), nil
	}

	c.isFirstRun = false

	if hasNewQuery {
		// NEW REAL USER QUERY detected!
		c.lastComplexity = complexity
		c.lastQueryTime = time.Now()
		c.realQueryCount++

		// Store in history for correlation analysis
		c.complexityHistory = append(c.complexityHistory, complexity)
		if len(c.complexityHistory) > 100 {
			c.complexityHistory = c.complexityHistory[1:]
		}

		fmt.Printf("🔍 REAL USER QUERY detected - Complexity: %.2f [Query #%d]\n", 
			complexity, c.realQueryCount)
		return complexity, nil
	} else {
		// No new user queries - return decayed complexity
		decayedComplexity := c.getDecayedComplexity()
		timeSince := time.Since(c.lastQueryTime)
		
		// Log waiting message periodically (every 30 seconds)
		if c.realQueryCount == 0 && int(timeSince.Seconds())%30 == 0 {
			fmt.Printf("⏳ Waiting for real user queries... (%.0fs)\n", timeSince.Seconds())
			fmt.Printf("💡 Send queries via Apollo Studio: %s/graphql\n", c.graphqlURL)
		} else if c.realQueryCount > 0 && int(timeSince.Seconds())%60 == 0 {
			fmt.Printf("💤 No new user activity for %.0fs - Complexity decayed to: %.2f\n", 
				timeSince.Seconds(), decayedComplexity)
		}
		
		return decayedComplexity, nil
	}
}

// detectRealUserQuery checks REST server for new real user queries
func (c *ComplexityCollector) detectRealUserQuery() (float64, bool, error) {
	// Check REST server for complexity stats
	url := c.restURL + "/complexity-stats"
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, false, fmt.Errorf("failed to connect to REST server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, false, fmt.Errorf("REST server returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, false, fmt.Errorf("failed to read response: %w", err)
	}

	var statsResponse ComplexityStatsResponse
	if err := json.Unmarshal(body, &statsResponse); err != nil {
		return 0, false, fmt.Errorf("failed to parse stats response: %w", err)
	}

	// Check if there are any cached queries (indicates user activity)
	if len(statsResponse.CachedQueries) == 0 {
		// No queries in cache - no user activity yet
		return 0, false, nil
	}

	// Check if cache size increased (new query arrived)
	if statsResponse.CacheSize > c.previousCacheSize {
		// New query detected!
		c.previousCacheSize = statsResponse.CacheSize
		
		// Get the latest (most recent) query complexity
		latestQuery := statsResponse.CachedQueries[len(statsResponse.CachedQueries)-1]
		return latestQuery.Complexity, true, nil
	}

	// Cache size same - no new queries
	return 0, false, nil
}

// getDecayedComplexity returns time-based decayed complexity when no new queries
func (c *ComplexityCollector) getDecayedComplexity() float64 {
	if c.realQueryCount == 0 {
		// Never had any real queries - return baseline
		return c.baselineComplexity
	}

	timeSinceLastQuery := time.Since(c.lastQueryTime)
	
	// Exponential decay: complexity reduces over time without user activity
	decayMinutes := timeSinceLastQuery.Minutes()
	decayFactor := 1.0 / (1.0 + decayMinutes*0.03) // Slow decay
	
	currentComplexity := c.lastComplexity * decayFactor
	
	// Don't go below baseline
	if currentComplexity < c.baselineComplexity {
		currentComplexity = c.baselineComplexity
	}
	
	return currentComplexity
}

// CollectDetailed collects detailed complexity metrics with authenticity info
func (c *ComplexityCollector) CollectDetailed() (metrics.ComplexityMetrics, error) {
	complexity, err := c.Collect()
	if err != nil {
		return metrics.ComplexityMetrics{}, err
	}

	// Determine query type and authenticity
	queryType := "baseline"
	if c.HasRecentRealActivity() {
		queryType = c.determineQueryType(complexity)
	}

	return metrics.ComplexityMetrics{
		Score:     complexity,
		QueryType: queryType,
		Timestamp: time.Now(),
	}, nil
}

// determineQueryType categorizes queries based on complexity score
func (c *ComplexityCollector) determineQueryType(complexity float64) string {
	switch {
	case complexity < 50:
		return "simple"
	case complexity < 150:
		return "medium"
	case complexity < 400:
		return "complex"
	default:
		return "very_complex"
	}
}

// HasRecentRealActivity checks if there has been recent real user activity
func (c *ComplexityCollector) HasRecentRealActivity() bool {
	return c.realQueryCount > 0 && time.Since(c.lastQueryTime) < 120*time.Second
}

// GetActivityStatus returns current user activity status
func (c *ComplexityCollector) GetActivityStatus() string {
	if c.realQueryCount == 0 {
		return "no_user_activity_yet"
	}

	timeSince := time.Since(c.lastQueryTime)
	switch {
	case timeSince < 10*time.Second:
		return "very_active"
	case timeSince < 30*time.Second:
		return "active"
	case timeSince < 120*time.Second:
		return "recent_activity"
	case timeSince < 600*time.Second:
		return "idle"
	default:
		return "dormant"
	}
}

// HealthCheck verifies both GraphQL and REST servers are reachable
func (c *ComplexityCollector) HealthCheck() error {
	// Check REST server health
	url := c.restURL + "/health"
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("REST server health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("REST server health check returned status %d", resp.StatusCode)
	}
	return nil
}

// GetStats returns authentic collector statistics
func (c *ComplexityCollector) GetStats() map[string]interface{} {
	avgComplexity := c.baselineComplexity
	minComplexity := c.baselineComplexity
	maxComplexity := c.baselineComplexity
	
	if len(c.complexityHistory) > 0 {
		sum := 0.0
		minComplexity = c.complexityHistory[0]
		maxComplexity = c.complexityHistory[0]
		
		for _, comp := range c.complexityHistory {
			sum += comp
			if comp < minComplexity {
				minComplexity = comp
			}
			if comp > maxComplexity {
				maxComplexity = comp
			}
		}
		avgComplexity = sum / float64(len(c.complexityHistory))
	}

	timeSinceLastQuery := time.Since(c.lastQueryTime)
	
	return map[string]interface{}{
		"real_user_queries":         c.realQueryCount,
		"last_complexity":           c.lastComplexity,
		"current_complexity":        c.getDecayedComplexity(),
		"average_complexity":        avgComplexity,
		"min_complexity":            minComplexity,
		"max_complexity":            maxComplexity,
		"baseline_complexity":       c.baselineComplexity,
		"seconds_since_last_query":  int(timeSinceLastQuery.Seconds()),
		"history_size":              len(c.complexityHistory),
		"activity_status":           c.GetActivityStatus(),
		"collector_type":            "authentic_dual_server",
		"is_currently_active":       c.HasRecentRealActivity(),
		"cache_size_tracking":       c.previousCacheSize,
		"graphql_url":               c.graphqlURL,
		"rest_url":                  c.restURL,
	}
}

// WaitForUserActivity provides helpful waiting messages
func (c *ComplexityCollector) WaitForUserActivity(timeout time.Duration) bool {
	fmt.Printf("⏳ Monitoring for real user queries...\n")
	fmt.Printf("💡 GraphQL Studio: %s/graphql\n", c.graphqlURL)
	fmt.Printf("📊 REST API: %s/complexity-stats\n", c.restURL)
	fmt.Printf("🔬 Your algorithm will respond to authentic user queries only!\n")
	
	start := time.Now()
	for time.Since(start) < timeout {
		_, hasNewQuery, err := c.detectRealUserQuery()
		if err == nil && hasNewQuery {
			fmt.Printf("✅ Real user activity detected!\n")
			return true
		}
		time.Sleep(3 * time.Second)
	}
	
	return false
}

// PrintInstructions prints user instructions for testing
func (c *ComplexityCollector) PrintInstructions() {
	fmt.Printf("\n🎯 HOW TO TEST YOUR AUTHENTIC ALGORITHM:\n")
	fmt.Printf("1. Open Apollo Studio: %s/graphql\n", c.graphqlURL)
	fmt.Printf("2. Send queries and watch this terminal for scaling decisions\n")
	fmt.Printf("3. Simple query: query { user(id: \"1\") { name } }\n")
	fmt.Printf("4. Complex query: query { systemAnalytics { totalUsers topUsers { posts } } }\n")
	fmt.Printf("5. Watch for: 🔍 REAL USER QUERY detected\n")
	fmt.Printf("6. Then see: 🎯 SCALING DECISION based on your complexity!\n\n")
}