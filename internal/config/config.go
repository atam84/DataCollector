package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server         ServerConfig
	Database       DatabaseConfig
	Exchange       ExchangeConfig
	HistoricalData HistoricalDataConfig
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

// HistoricalDataConfig holds configuration for historical data fetching
type HistoricalDataConfig struct {
	// How many days back to fetch on first execution (by timeframe)
	Start1m  int // 1-minute candles
	Start5m  int // 5-minute candles
	Start15m int // 15-minute candles
	Start1h  int // 1-hour candles
	Start4h  int // 4-hour candles
	Start1d  int // 1-day candles
	Start1w  int // 1-week candles

	// Maximum candles to fetch per execution (rate limit protection)
	MaxCandlesPerFetch int

	// Backfill batch size for large gaps
	BackfillBatchSize int
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
		HistoricalData: HistoricalDataConfig{
			Start1m:            getEnvInt("HISTORICAL_START_1m", 7),     // 7 days for 1m candles
			Start5m:            getEnvInt("HISTORICAL_START_5m", 30),    // 30 days for 5m candles
			Start15m:           getEnvInt("HISTORICAL_START_15m", 90),   // 90 days for 15m candles
			Start1h:            getEnvInt("HISTORICAL_START_1h", 180),   // 180 days for 1h candles
			Start4h:            getEnvInt("HISTORICAL_START_4h", 365),   // 1 year for 4h candles
			Start1d:            getEnvInt("HISTORICAL_START_1d", 1095),  // 3 years for 1d candles
			Start1w:            getEnvInt("HISTORICAL_START_1w", 1825),  // 5 years for 1w candles
			MaxCandlesPerFetch: getEnvInt("MAX_CANDLES_PER_FETCH", 1000),
			BackfillBatchSize:  getEnvInt("BACKFILL_BATCH_SIZE", 500),
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

// GetHistoricalStartDate returns the start date for historical data fetching based on timeframe
func (h *HistoricalDataConfig) GetHistoricalStartDate(timeframe string) time.Time {
	var daysBack int

	switch timeframe {
	case "1m":
		daysBack = h.Start1m
	case "5m":
		daysBack = h.Start5m
	case "15m":
		daysBack = h.Start15m
	case "1h":
		daysBack = h.Start1h
	case "4h":
		daysBack = h.Start4h
	case "1d":
		daysBack = h.Start1d
	case "1w":
		daysBack = h.Start1w
	default:
		// Default to 30 days for unknown timeframes
		daysBack = 30
	}

	return time.Now().Add(-time.Duration(daysBack) * 24 * time.Hour)
}
