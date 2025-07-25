package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jayakumar0812/multi-metric-kubernetes-autoscaling/pkg/metrics"
)

// ComplexityCollector collects complexity metrics from GraphQL server
type ComplexityCollector struct {
	graphqlURL string
	client     *http.Client
}

// NewComplexityCollector creates a new complexity collector
func NewComplexityCollector(graphqlURL string) *ComplexityCollector {
	return &ComplexityCollector{
		graphqlURL: graphqlURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Collect fetches the latest complexity metrics from GraphQL server
func (c *ComplexityCollector) Collect() (float64, error) {
	// Get complexity stats from GraphQL server
	url := fmt.Sprintf("%s/complexity-stats", c.graphqlURL)
	
	resp, err := c.client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch complexity stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("GraphQL server returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	var complexityResponse metrics.GraphQLComplexityResponse
	if err := json.Unmarshal(body, &complexityResponse); err != nil {
		return 0, fmt.Errorf("failed to parse complexity response: %w", err)
	}

	// Calculate average complexity from cached queries
	if len(complexityResponse.CachedQueries) == 0 {
		// No queries in cache, return default low complexity
		return 10.0, nil
	}

	// Calculate weighted average (recent queries have more weight)
	totalComplexity := 0.0
	totalWeight := 0.0
	
	for i, query := range complexityResponse.CachedQueries {
		// Give more weight to recent queries (later in the slice)
		weight := float64(i+1) / float64(len(complexityResponse.CachedQueries))
		totalComplexity += query.Complexity * weight
		totalWeight += weight
	}

	averageComplexity := totalComplexity / totalWeight
	
	// Apply some smoothing to avoid wild fluctuations
	smoothedComplexity := c.smoothComplexity(averageComplexity)
	
	return smoothedComplexity, nil
}

// smoothComplexity applies smoothing to complexity values
func (c *ComplexityCollector) smoothComplexity(complexity float64) float64 {
	// Simple exponential smoothing
	// In a real implementation, you'd maintain state for this
	
	// For now, just ensure complexity is within reasonable bounds
	if complexity < 1 {
		return 1
	}
	if complexity > 1000 {
		return 1000
	}
	
	return complexity
}

// CollectDetailed collects detailed complexity metrics (for future use)
func (c *ComplexityCollector) CollectDetailed() (metrics.ComplexityMetrics, error) {
	complexity, err := c.Collect()
	if err != nil {
		return metrics.ComplexityMetrics{}, err
	}

	return metrics.ComplexityMetrics{
		Score:     complexity,
		QueryType: c.determineQueryType(complexity),
		Timestamp: time.Now(),
	}, nil
}

// determineQueryType categorizes queries based on complexity
func (c *ComplexityCollector) determineQueryType(complexity float64) string {
	switch {
	case complexity < 50:
		return "simple"
	case complexity < 200:
		return "medium"
	case complexity < 500:
		return "complex"
	default:
		return "very_complex"
	}
}

// HealthCheck verifies that the GraphQL server is reachable
func (c *ComplexityCollector) HealthCheck() error {
	url := fmt.Sprintf("%s/health", c.graphqlURL)
	
	resp, err := c.client.Get(url)
	if err != nil {
		return fmt.Errorf("GraphQL server health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GraphQL server health check returned status %d", resp.StatusCode)
	}

	return nil
}