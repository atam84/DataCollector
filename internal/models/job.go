package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Job represents an ingestion task for a symbol + timeframe
type Job struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ConnectorExchangeID string             `bson:"connector_exchange_id" json:"connector_exchange_id"`
	Symbol             string             `bson:"symbol" json:"symbol"`
	Timeframe          string             `bson:"timeframe" json:"timeframe"` // "1m", "5m", "1h", etc.
	Status             string             `bson:"status" json:"status"`       // "active", "paused", "error"
	CreatedAt          time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"updated_at"`

	Schedule Schedule `bson:"schedule" json:"schedule"`
	Cursor   Cursor   `bson:"cursor" json:"cursor"`
	RunState RunState `bson:"run_state" json:"run_state"`

	// Indicator configuration (overrides connector defaults if specified)
	IndicatorConfig *IndicatorConfig `bson:"indicator_config,omitempty" json:"indicator_config,omitempty"`
}

// Schedule defines when the job should run
type Schedule struct {
	Mode string  `bson:"mode" json:"mode"` // "timeframe", "cron"
	Cron *string `bson:"cron,omitempty" json:"cron,omitempty"`
}

// Cursor tracks the job's progress
type Cursor struct {
	LastCandleTime *time.Time `bson:"last_candle_time,omitempty" json:"last_candle_time,omitempty"`
}

// RunState tracks execution state
type RunState struct {
	LockedUntil  *time.Time `bson:"locked_until,omitempty" json:"locked_until,omitempty"`
	LastRunTime  *time.Time `bson:"last_run_time,omitempty" json:"last_run_time,omitempty"`
	NextRunTime  *time.Time `bson:"next_run_time,omitempty" json:"next_run_time,omitempty"`
	LastError    *string    `bson:"last_error,omitempty" json:"last_error,omitempty"`
	RunsTotal    int        `bson:"runs_total" json:"runs_total"`
}

// JobCreateRequest is the DTO for creating a job
type JobCreateRequest struct {
	ConnectorExchangeID string `json:"connector_exchange_id" validate:"required"`
	Symbol              string `json:"symbol" validate:"required"`
	Timeframe           string `json:"timeframe" validate:"required,oneof=1m 5m 15m 30m 1h 4h 1d 1w"`
	Status              string `json:"status" validate:"omitempty,oneof=active paused"`
}

// JobUpdateRequest is the DTO for updating a job
type JobUpdateRequest struct {
	Status    *string `json:"status,omitempty" validate:"omitempty,oneof=active paused error"`
	Timeframe *string `json:"timeframe,omitempty" validate:"omitempty,oneof=1m 5m 15m 30m 1h 4h 1d 1w"`
}
