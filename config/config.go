package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Database
	DatabaseURL string

	// Server
	HTTPPort int

	// Load defaults
	DefaultConnections int
	DefaultReadQPS     int
	DefaultWriteQPS    int

	// Limits
	MaxConnections int
	MaxReadQPS     int
	MaxWriteQPS    int

	// Metrics
	MetricsInterval time.Duration

	// Schema scenario
	DefaultScenario string
	CustomTable     string
}

// Load reads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://localhost:5432/pooler_demo"),
		HTTPPort:           getEnvInt("HTTP_PORT", 8080),
		DefaultConnections: getEnvInt("DEFAULT_CONNECTIONS", 10),
		DefaultReadQPS:     getEnvInt("DEFAULT_READ_QPS", 100),
		DefaultWriteQPS:    getEnvInt("DEFAULT_WRITE_QPS", 10),
		MaxConnections:     getEnvInt("MAX_CONNECTIONS", 20000),
		MaxReadQPS:         getEnvInt("MAX_READ_QPS", 500000),
		MaxWriteQPS:        getEnvInt("MAX_WRITE_QPS", 500000),
		MetricsInterval:    getEnvDuration("METRICS_INTERVAL", 100*time.Millisecond),
		DefaultScenario:    getEnv("DEFAULT_SCENARIO", "simple"),
		CustomTable:        getEnv("CUSTOM_TABLE", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
