package config

import (
	"os"
	"strconv"
	"time"

	"github.com/jayakumar0812/multi-metric-kubernetes-autoscaling/pkg/metrics"
)

// Config holds all application configuration
type Config struct {
	Port         string                `json:"port"`
	GraphQLURL   string                `json:"graphql_url"`
	FusionConfig metrics.FusionConfig  `json:"fusion_config"`
	LogLevel     string                `json:"log_level"`
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		Port:       getEnvOrDefault("METRICS_SERVER_PORT", "3001"),
		GraphQLURL: getEnvOrDefault("GRAPHQL_URL", "http://localhost:4000"),
		LogLevel:   getEnvOrDefault("LOG_LEVEL", "info"),
		FusionConfig: metrics.FusionConfig{
			ComplexityWeight:    getEnvFloatOrDefault("COMPLEXITY_WEIGHT", 0.6),
			CPUWeight:          getEnvFloatOrDefault("CPU_WEIGHT", 0.3),
			MemoryWeight:       getEnvFloatOrDefault("MEMORY_WEIGHT", 0.1),
			ScalingThreshold:   getEnvFloatOrDefault("SCALING_THRESHOLD", 0.75),
			CooldownPeriod:     getEnvDurationOrDefault("COOLDOWN_PERIOD", 30*time.Second),
			HistoryWindowSize:  getEnvIntOrDefault("HISTORY_WINDOW_SIZE", 100),
			MinSamplesRequired: getEnvIntOrDefault("MIN_SAMPLES_REQUIRED", 10),
		},
	}
}

// Helper functions for environment variable parsing
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvFloatOrDefault(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}