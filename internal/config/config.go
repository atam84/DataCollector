package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Exchange ExchangeConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port string
	Host string
}

// DatabaseConfig holds MongoDB configuration
type DatabaseConfig struct {
	URI      string
	Database string
}

// ExchangeConfig holds exchange-related configuration
type ExchangeConfig struct {
	SandboxMode      bool
	EnableRateLimit  bool
	RequestTimeout   int // in milliseconds
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			URI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGODB_DATABASE", "datacollector"),
		},
		Exchange: ExchangeConfig{
			SandboxMode:     getEnvBool("EXCHANGE_SANDBOX_MODE", true), // Default to sandbox for safety
			EnableRateLimit: getEnvBool("EXCHANGE_ENABLE_RATE_LIMIT", true),
			RequestTimeout:  getEnvInt("EXCHANGE_REQUEST_TIMEOUT", 30000),
		},
	}

	// Validate required fields
	if cfg.Database.URI == "" {
		return nil, fmt.Errorf("MONGODB_URI is required")
	}

	return cfg, nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool retrieves a boolean environment variable or returns a default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return boolVal
	}
	return defaultValue
}

// getEnvInt retrieves an integer environment variable or returns a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}
		return intVal
	}
	return defaultValue
}
