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
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`

	RateLimit RateLimit `bson:"rate_limit" json:"rate_limit"`

	// Health monitoring
	Health ConnectorHealth `bson:"health" json:"health"`

	// Credentials reference (keys stored in environment variables)
	CredentialsRef CredentialsRef `bson:"credentials_ref,omitempty" json:"credentials_ref,omitempty"`
}

// ConnectorHealth tracks the health status and metrics of a connector
type ConnectorHealth struct {
	Status               string     `bson:"status" json:"status"`                                                   // "healthy", "degraded", "unhealthy"
	LastSuccessfulCall   *time.Time `bson:"last_successful_call,omitempty" json:"last_successful_call,omitempty"`   // Last successful API call
	LastFailedCall       *time.Time `bson:"last_failed_call,omitempty" json:"last_failed_call,omitempty"`           // Last failed API call
	LastError            string     `bson:"last_error,omitempty" json:"last_error,omitempty"`                       // Last error message
	ConsecutiveFailures  int        `bson:"consecutive_failures" json:"consecutive_failures"`                       // Consecutive failure count
	TotalCalls           int64      `bson:"total_calls" json:"total_calls"`                                         // Total API calls made
	TotalFailures        int64      `bson:"total_failures" json:"total_failures"`                                   // Total failed calls
	AverageResponseMs    float64    `bson:"average_response_ms" json:"average_response_ms"`                         // Average response time in ms
	LastResponseMs       int64      `bson:"last_response_ms" json:"last_response_ms"`                               // Last response time in ms
	LastHealthCheck      *time.Time `bson:"last_health_check,omitempty" json:"last_health_check,omitempty"`         // Last health check timestamp
	UptimePercentage     float64    `bson:"uptime_percentage" json:"uptime_percentage"`                             // Uptime percentage (0-100)
}

// RateLimit holds rate limiting configuration and state
type RateLimit struct {
	// Configuration
	Limit        int `bson:"limit" json:"limit"`                 // Max requests per period
	PeriodMs     int `bson:"period_ms" json:"period_ms"`         // Period in milliseconds (e.g., 60000 for 1 minute)
	MinDelayMs   int `bson:"min_delay_ms" json:"min_delay_ms"`   // Minimum delay between API calls in ms (e.g., 5000 for 5 seconds)

	// State tracking
	Usage         int        `bson:"usage" json:"usage"`                                           // Current usage count in period
	PeriodStart   time.Time  `bson:"period_start" json:"period_start"`                             // Start of current period
	LastAPICallAt *time.Time `bson:"last_api_call_at,omitempty" json:"last_api_call_at,omitempty"` // Timestamp of last API call
	LastJobRunAt  *time.Time `bson:"last_job_run_at,omitempty" json:"last_job_run_at,omitempty"`   // Timestamp of last job execution
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
	RateLimit   struct {
		Limit      int `json:"limit" validate:"required,min=1"`           // Max requests per period
		PeriodMs   int `json:"period_ms" validate:"required,min=1000"`    // Period in milliseconds
		MinDelayMs int `json:"min_delay_ms" validate:"omitempty,min=100"` // Min delay between calls (default: calculated from limit/period)
	} `json:"rate_limit"`
}

// ConnectorUpdateRequest is the DTO for updating a connector
type ConnectorUpdateRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Status      *string `json:"status,omitempty" validate:"omitempty,oneof=active disabled"`
	RateLimit   *struct {
		Limit      *int `json:"limit,omitempty" validate:"omitempty,min=1"`
		PeriodMs   *int `json:"period_ms,omitempty" validate:"omitempty,min=1000"`
		MinDelayMs *int `json:"min_delay_ms,omitempty" validate:"omitempty,min=100"`
	} `json:"rate_limit,omitempty"`
}

// ConnectorResponse is the enhanced DTO with additional computed fields
type ConnectorResponse struct {
	Connector
	JobCount       int64  `json:"job_count"`
	ActiveJobCount int64  `json:"active_job_count"`
	LastExecution  *int64 `json:"last_execution,omitempty"`
}
