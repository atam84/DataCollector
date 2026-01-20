package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Connector represents an exchange integration
type Connector struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ExchangeID  string             `bson:"exchange_id" json:"exchange_id"`
	DisplayName string             `bson:"display_name" json:"display_name"`
	Status      string             `bson:"status" json:"status"` // "active", "disabled"
	SandboxMode bool               `bson:"sandbox_mode" json:"sandbox_mode"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`

	RateLimit RateLimit `bson:"rate_limit" json:"rate_limit"`

	// Credentials reference (keys stored in environment variables)
	CredentialsRef CredentialsRef `bson:"credentials_ref,omitempty" json:"credentials_ref,omitempty"`

	// Indicator configuration (default for all jobs using this connector)
	IndicatorConfig *IndicatorConfig `bson:"indicator_config,omitempty" json:"indicator_config,omitempty"`
}

// RateLimit holds rate limiting configuration and state
type RateLimit struct {
	Limit         int       `bson:"limit" json:"limit"`                   // Max requests per period
	PeriodMs      int       `bson:"period_ms" json:"period_ms"`           // Period in milliseconds
	Usage         int       `bson:"usage" json:"usage"`                   // Current usage count
	PeriodStart   time.Time `bson:"period_start" json:"period_start"`     // Start of current period
	LastJobRunAt  *time.Time `bson:"last_job_run_at,omitempty" json:"last_job_run_at,omitempty"`
}

// CredentialsRef references API credentials stored in environment
type CredentialsRef struct {
	Mode string   `bson:"mode" json:"mode"` // "env", "vault"
	Keys []string `bson:"keys" json:"keys"` // e.g., ["BINANCE_API_KEY", "BINANCE_API_SECRET"]
}

// ConnectorCreateRequest is the DTO for creating a connector
type ConnectorCreateRequest struct {
	ExchangeID  string `json:"exchange_id" validate:"required"`
	DisplayName string `json:"display_name" validate:"required"`
	SandboxMode bool   `json:"sandbox_mode"`
	RateLimit   struct {
		Limit    int `json:"limit" validate:"required,min=1"`
		PeriodMs int `json:"period_ms" validate:"required,min=1000"`
	} `json:"rate_limit"`
}

// ConnectorUpdateRequest is the DTO for updating a connector
type ConnectorUpdateRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Status      *string `json:"status,omitempty" validate:"omitempty,oneof=active disabled"`
	SandboxMode *bool   `json:"sandbox_mode,omitempty"`
	RateLimit   *struct {
		Limit    *int `json:"limit,omitempty" validate:"omitempty,min=1"`
		PeriodMs *int `json:"period_ms,omitempty" validate:"omitempty,min=1000"`
	} `json:"rate_limit,omitempty"`
}

// ConnectorResponse is the enhanced DTO with additional computed fields
type ConnectorResponse struct {
	Connector
	JobCount       int64  `json:"job_count"`
	ActiveJobCount int64  `json:"active_job_count"`
	LastExecution  *int64 `json:"last_execution,omitempty"`
}
