package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Job represents an ingestion task for a symbol + timeframe
type Job struct {
	ID                  primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	ConnectorExchangeID string               `bson:"connector_exchange_id" json:"connector_exchange_id"`
	Symbol              string               `bson:"symbol" json:"symbol"`
	Timeframe           string               `bson:"timeframe" json:"timeframe"` // "1m", "5m", "1h", etc.
	Status              string               `bson:"status" json:"status"`       // "active", "paused", "error"
	CollectHistorical   bool                 `bson:"collect_historical" json:"collect_historical"`
	DependsOn           []primitive.ObjectID `bson:"depends_on,omitempty" json:"depends_on,omitempty"` // Job IDs that must complete first
	CreatedAt           time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt           time.Time            `bson:"updated_at" json:"updated_at"`

	Schedule Schedule `bson:"schedule" json:"schedule"`
	Cursor   Cursor   `bson:"cursor" json:"cursor"`
	RunState RunState `bson:"run_state" json:"run_state"`
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
	LockedUntil        *time.Time `bson:"locked_until,omitempty" json:"locked_until,omitempty"`
	LastRunTime        *time.Time `bson:"last_run_time,omitempty" json:"last_run_time,omitempty"`
	NextRunTime        *time.Time `bson:"next_run_time,omitempty" json:"next_run_time,omitempty"`
	LastError          *string    `bson:"last_error,omitempty" json:"last_error,omitempty"`
	RunsTotal          int        `bson:"runs_total" json:"runs_total"`
	ConsecutiveFailures int       `bson:"consecutive_failures" json:"consecutive_failures"`
	LastFailureTime    *time.Time `bson:"last_failure_time,omitempty" json:"last_failure_time,omitempty"`
}

// JobCreateRequest is the DTO for creating a job
type JobCreateRequest struct {
	ConnectorExchangeID string   `json:"connector_exchange_id" validate:"required"`
	Symbol              string   `json:"symbol" validate:"required"`
	Timeframe           string   `json:"timeframe" validate:"required"`
	Status              string   `json:"status" validate:"omitempty,oneof=active paused"`
	CollectHistorical   bool     `json:"collect_historical"`
	DependsOn           []string `json:"depends_on,omitempty"` // Job IDs (as strings) that must complete first
}

// JobUpdateRequest is the DTO for updating a job
type JobUpdateRequest struct {
	Status            *string   `json:"status,omitempty" validate:"omitempty,oneof=active paused error"`
	Timeframe         *string   `json:"timeframe,omitempty"`
	CollectHistorical *bool     `json:"collect_historical,omitempty"`
	DependsOn         *[]string `json:"depends_on,omitempty"` // Job IDs (as strings) - use empty array to clear dependencies
}

// JobDependency represents a dependency relationship between jobs
type JobDependency struct {
	JobID       primitive.ObjectID `json:"job_id"`
	DependsOnID primitive.ObjectID `json:"depends_on_id"`
}

// DependencyStatus represents the status of job dependencies
type DependencyStatus struct {
	JobID            string   `json:"job_id"`
	DependsOn        []string `json:"depends_on"`
	BlockedBy        []string `json:"blocked_by"`        // Dependencies that haven't completed recently
	AllDepsCompleted bool     `json:"all_deps_completed"` // Whether all dependencies completed recently
}
