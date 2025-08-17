package config

import (
	"os"
	"time"
	"hybrid-metrics/pkg/metrics"
)

type Config struct {
	GraphQLURL    string
	Port          string
	FusionConfig  metrics.FusionConfig
}

func Load() *Config {
	return &Config{
		GraphQLURL: getGraphQLURL(),
		Port:       getPort(),
		FusionConfig: metrics.FusionConfig{
			// field names from metrics package
			ComplexityWeight:    0.6,               // ← Correct field name
			CPUWeight:          0.3,               // ← Correct field name  
			MemoryWeight:       0.1,               // ← Correct field name
			ScalingThreshold:   0.7,               // Threshold to trigger scaling
			CooldownPeriod:     5 * time.Minute,   // Wait time between scaling decisions
			HistoryWindowSize:  100,               // Number of metrics to keep in history
			MinSamplesRequired: 5,                 // Minimum samples needed for decisions
		},
	}
}

func getGraphQLURL() string {
	if url := os.Getenv("GRAPHQL_URL"); url != "" {
		return url
	}
	return "http://graphql-server:4000"
}

func getPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "3001"
}