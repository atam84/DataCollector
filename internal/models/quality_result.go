package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// QualityCheckStatus represents the status of a quality check
type QualityCheckStatus string

const (
	QualityCheckPending    QualityCheckStatus = "pending"
	QualityCheckRunning    QualityCheckStatus = "running"
	QualityCheckCompleted  QualityCheckStatus = "completed"
	QualityCheckFailed     QualityCheckStatus = "failed"
)

// QualityCheckType represents the type of quality check
type QualityCheckType string

const (
	QualityCheckTypeSingle   QualityCheckType = "single"    // Single job check
	QualityCheckTypeAll      QualityCheckType = "all"       // All jobs check
	QualityCheckTypeExchange QualityCheckType = "exchange"  // All jobs for an exchange
	QualityCheckTypeScheduled QualityCheckType = "scheduled" // Scheduled periodic check
)

// DataQualityResult stores the cached result of a quality analysis for a job
// Unique identifier: (exchange_id, symbol, timeframe)
type DataQualityResult struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ExchangeID        string             `bson:"exchange_id" json:"exchange_id"`
	Symbol            string             `bson:"symbol" json:"symbol"`
	Timeframe         string             `bson:"timeframe" json:"timeframe"`
	JobID             primitive.ObjectID `bson:"job_id,omitempty" json:"job_id,omitempty"`

	// Quality metrics
	QualityStatus     string    `bson:"quality_status" json:"quality_status"`       // excellent, good, fair, poor
	CompletenessScore float64   `bson:"completeness_score" json:"completeness_score"`
	TotalCandles      int64     `bson:"total_candles" json:"total_candles"`
	ExpectedCandles   int64     `bson:"expected_candles" json:"expected_candles"`
	MissingCandles    int64     `bson:"missing_candles" json:"missing_candles"`
	GapsDetected      int       `bson:"gaps_detected" json:"gaps_detected"`
	Gaps              []DataGap `bson:"gaps,omitempty" json:"gaps,omitempty"`

	// Data period
	DataPeriodStart   time.Time `bson:"data_period_start" json:"data_period_start"`
	DataPeriodEnd     time.Time `bson:"data_period_end" json:"data_period_end"`
	DataPeriodDays    int       `bson:"data_period_days" json:"data_period_days"`
	DataAgeDays       int       `bson:"data_age_days" json:"data_age_days"` // How old is the newest data

	// Freshness
	DataFreshness     string `bson:"data_freshness" json:"data_freshness"`       // fresh, stale, very_stale
	FreshnessMinutes  int64  `bson:"freshness_minutes" json:"freshness_minutes"`

	// Timestamps
	CheckedAt         time.Time `bson:"checked_at" json:"checked_at"`
	CreatedAt         time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time `bson:"updated_at" json:"updated_at"`
}

// QualityCheckJob represents a background quality check job
type QualityCheckJob struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Type          QualityCheckType     `bson:"type" json:"type"`
	Status        QualityCheckStatus   `bson:"status" json:"status"`

	// Filter for what to check
	ExchangeID    string               `bson:"exchange_id,omitempty" json:"exchange_id,omitempty"`
	Symbol        string               `bson:"symbol,omitempty" json:"symbol,omitempty"`
	Timeframe     string               `bson:"timeframe,omitempty" json:"timeframe,omitempty"`
	JobIDs        []primitive.ObjectID `bson:"job_ids,omitempty" json:"job_ids,omitempty"`

	// Progress
	TotalJobs     int       `bson:"total_jobs" json:"total_jobs"`
	CompletedJobs int       `bson:"completed_jobs" json:"completed_jobs"`
	FailedJobs    int       `bson:"failed_jobs" json:"failed_jobs"`
	Progress      float64   `bson:"progress" json:"progress"` // 0-100

	// Results summary
	ExcellentCount int `bson:"excellent_count" json:"excellent_count"`
	GoodCount      int `bson:"good_count" json:"good_count"`
	FairCount      int `bson:"fair_count" json:"fair_count"`
	PoorCount      int `bson:"poor_count" json:"poor_count"`

	// Error tracking
	LastError     string    `bson:"last_error,omitempty" json:"last_error,omitempty"`
	Errors        []string  `bson:"errors,omitempty" json:"errors,omitempty"`

	// Timestamps
	StartedAt     *time.Time `bson:"started_at,omitempty" json:"started_at,omitempty"`
	CompletedAt   *time.Time `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
	CreatedAt     time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `bson:"updated_at" json:"updated_at"`
}

// QualitySummaryCache stores the aggregated quality summary
type QualitySummaryCache struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ExchangeID          string             `bson:"exchange_id,omitempty" json:"exchange_id,omitempty"` // Empty = global

	// Counts
	TotalJobs           int     `bson:"total_jobs" json:"total_jobs"`
	ExcellentQuality    int     `bson:"excellent_quality" json:"excellent_quality"`
	GoodQuality         int     `bson:"good_quality" json:"good_quality"`
	FairQuality         int     `bson:"fair_quality" json:"fair_quality"`
	PoorQuality         int     `bson:"poor_quality" json:"poor_quality"`

	// Aggregates
	AverageCompleteness float64 `bson:"average_completeness" json:"average_completeness"`
	TotalCandles        int64   `bson:"total_candles" json:"total_candles"`
	TotalMissingCandles int64   `bson:"total_missing_candles" json:"total_missing_candles"`
	TotalGaps           int     `bson:"total_gaps" json:"total_gaps"`

	// Freshness
	FreshDataJobs       int `bson:"fresh_data_jobs" json:"fresh_data_jobs"`
	StaleDataJobs       int `bson:"stale_data_jobs" json:"stale_data_jobs"`
	VeryStaleDataJobs   int `bson:"very_stale_data_jobs" json:"very_stale_data_jobs"`

	// Timestamps
	LastCheckAt         time.Time `bson:"last_check_at" json:"last_check_at"`
	CreatedAt           time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt           time.Time `bson:"updated_at" json:"updated_at"`
}

// GapFillRequest represents a request to fill gaps in data
type GapFillRequest struct {
	JobID      string    `json:"job_id"`
	StartTime  time.Time `json:"start_time,omitempty"`
	EndTime    time.Time `json:"end_time,omitempty"`
	FillAll    bool      `json:"fill_all"` // If true, attempt to fill all detected gaps
}

// GapFillResult represents the result of a gap fill operation
type GapFillResult struct {
	JobID          string    `json:"job_id"`
	GapsAttempted  int       `json:"gaps_attempted"`
	GapsFilled     int       `json:"gaps_filled"`
	CandlesFetched int       `json:"candles_fetched"`
	Errors         []string  `json:"errors,omitempty"`
	StartedAt      time.Time `json:"started_at"`
	CompletedAt    time.Time `json:"completed_at"`
}
