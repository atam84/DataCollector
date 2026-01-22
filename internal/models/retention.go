package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RetentionPolicyType represents the type of retention policy
type RetentionPolicyType string

const (
	RetentionPolicyTypeGlobal     RetentionPolicyType = "global"
	RetentionPolicyTypeExchange   RetentionPolicyType = "exchange"
	RetentionPolicyTypeTimeframe  RetentionPolicyType = "timeframe"
)

// RetentionPolicy defines data retention rules
type RetentionPolicy struct {
	ID              primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	Name            string              `bson:"name" json:"name"`
	Type            RetentionPolicyType `bson:"type" json:"type"`
	Enabled         bool                `bson:"enabled" json:"enabled"`

	// Scope filters
	ExchangeID      *string             `bson:"exchange_id,omitempty" json:"exchange_id,omitempty"`
	Timeframe       *string             `bson:"timeframe,omitempty" json:"timeframe,omitempty"`

	// Retention settings
	RetentionDays   int                 `bson:"retention_days" json:"retention_days"`   // Delete data older than N days
	MaxCandles      *int64              `bson:"max_candles,omitempty" json:"max_candles,omitempty"` // Keep max N candles per symbol
	KeepLatestOnly  bool                `bson:"keep_latest_only" json:"keep_latest_only"` // Only keep most recent chunk

	// Schedule
	RunSchedule     string              `bson:"run_schedule,omitempty" json:"run_schedule,omitempty"` // cron expression
	LastRunAt       *time.Time          `bson:"last_run_at,omitempty" json:"last_run_at,omitempty"`
	NextRunAt       *time.Time          `bson:"next_run_at,omitempty" json:"next_run_at,omitempty"`

	// Metadata
	CreatedAt       time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time           `bson:"updated_at" json:"updated_at"`
}

// RetentionConfig stores global retention configuration
type RetentionConfig struct {
	ID                      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Enabled                 bool               `bson:"enabled" json:"enabled"`
	DefaultRetentionDays    int                `bson:"default_retention_days" json:"default_retention_days"`
	CleanupIntervalMinutes  int                `bson:"cleanup_interval_minutes" json:"cleanup_interval_minutes"`
	MaxChunksPerSymbol      int                `bson:"max_chunks_per_symbol" json:"max_chunks_per_symbol"`
	DeleteEmptyChunks       bool               `bson:"delete_empty_chunks" json:"delete_empty_chunks"`
	ArchiveBeforeDelete     bool               `bson:"archive_before_delete" json:"archive_before_delete"`
	UpdatedAt               time.Time          `bson:"updated_at" json:"updated_at"`
}

// DefaultRetentionConfig returns the default retention configuration
func DefaultRetentionConfig() *RetentionConfig {
	return &RetentionConfig{
		Enabled:                false, // Disabled by default for safety
		DefaultRetentionDays:   365,   // Keep data for 1 year by default
		CleanupIntervalMinutes: 1440,  // Run once a day
		MaxChunksPerSymbol:     0,     // No limit by default
		DeleteEmptyChunks:      true,
		ArchiveBeforeDelete:    false,
	}
}

// RetentionCleanupResult contains the results of a retention cleanup operation
type RetentionCleanupResult struct {
	PolicyID        string    `json:"policy_id,omitempty"`
	PolicyName      string    `json:"policy_name,omitempty"`
	ExchangeID      string    `json:"exchange_id,omitempty"`
	Timeframe       string    `json:"timeframe,omitempty"`
	ChunksDeleted   int64     `json:"chunks_deleted"`
	CandlesDeleted  int64     `json:"candles_deleted"`
	BytesFreed      int64     `json:"bytes_freed,omitempty"`
	Duration        int64     `json:"duration_ms"`
	StartedAt       time.Time `json:"started_at"`
	CompletedAt     time.Time `json:"completed_at"`
	Error           string    `json:"error,omitempty"`
}

// RetentionCleanupSummary aggregates multiple cleanup results
type RetentionCleanupSummary struct {
	TotalChunksDeleted  int64                    `json:"total_chunks_deleted"`
	TotalCandlesDeleted int64                    `json:"total_candles_deleted"`
	TotalBytesFreed     int64                    `json:"total_bytes_freed"`
	TotalDuration       int64                    `json:"total_duration_ms"`
	Results             []RetentionCleanupResult `json:"results"`
	StartedAt           time.Time                `json:"started_at"`
	CompletedAt         time.Time                `json:"completed_at"`
}

// RetentionPolicyCreateRequest for creating new retention policies
type RetentionPolicyCreateRequest struct {
	Name           string              `json:"name" validate:"required"`
	Type           RetentionPolicyType `json:"type" validate:"required"`
	Enabled        bool                `json:"enabled"`
	ExchangeID     *string             `json:"exchange_id,omitempty"`
	Timeframe      *string             `json:"timeframe,omitempty"`
	RetentionDays  int                 `json:"retention_days" validate:"required,min=1"`
	MaxCandles     *int64              `json:"max_candles,omitempty"`
	KeepLatestOnly bool                `json:"keep_latest_only"`
	RunSchedule    string              `json:"run_schedule,omitempty"`
}

// RetentionPolicyUpdateRequest for updating retention policies
type RetentionPolicyUpdateRequest struct {
	Name           *string `json:"name,omitempty"`
	Enabled        *bool   `json:"enabled,omitempty"`
	RetentionDays  *int    `json:"retention_days,omitempty"`
	MaxCandles     *int64  `json:"max_candles,omitempty"`
	KeepLatestOnly *bool   `json:"keep_latest_only,omitempty"`
	RunSchedule    *string `json:"run_schedule,omitempty"`
}

// DataUsageStats provides information about data storage usage
type DataUsageStats struct {
	ExchangeID       string    `json:"exchange_id"`
	Symbol           string    `json:"symbol,omitempty"`
	Timeframe        string    `json:"timeframe,omitempty"`
	ChunkCount       int       `json:"chunk_count"`
	TotalCandles     int64     `json:"total_candles"`
	OldestData       time.Time `json:"oldest_data,omitempty"`
	NewestData       time.Time `json:"newest_data,omitempty"`
	EstimatedSizeMB  float64   `json:"estimated_size_mb"`
}
