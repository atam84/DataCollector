package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AlertSeverity represents the severity level of an alert
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityError    AlertSeverity = "error"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeJobFailed         AlertType = "job_failed"
	AlertTypeJobConsecFailures AlertType = "job_consecutive_failures"
	AlertTypeConnectorDown     AlertType = "connector_down"
	AlertTypeRateLimitExceeded AlertType = "rate_limit_exceeded"
	AlertTypeNoDataCollected   AlertType = "no_data_collected"
	AlertTypeSystemError       AlertType = "system_error"
)

// AlertStatus represents the status of an alert
type AlertStatus string

const (
	AlertStatusActive       AlertStatus = "active"
	AlertStatusAcknowledged AlertStatus = "acknowledged"
	AlertStatusResolved     AlertStatus = "resolved"
)

// Alert represents a system alert
type Alert struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type         AlertType          `bson:"type" json:"type"`
	Severity     AlertSeverity      `bson:"severity" json:"severity"`
	Status       AlertStatus        `bson:"status" json:"status"`
	Title        string             `bson:"title" json:"title"`
	Message      string             `bson:"message" json:"message"`
	Source       AlertSource        `bson:"source" json:"source"`
	Metadata     map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
	AcknowledgedAt *time.Time       `bson:"acknowledged_at,omitempty" json:"acknowledged_at,omitempty"`
	AcknowledgedBy *string          `bson:"acknowledged_by,omitempty" json:"acknowledged_by,omitempty"`
	ResolvedAt   *time.Time         `bson:"resolved_at,omitempty" json:"resolved_at,omitempty"`
}

// AlertSource identifies the source of the alert
type AlertSource struct {
	Type       string `bson:"type" json:"type"` // "job", "connector", "system"
	ID         string `bson:"id,omitempty" json:"id,omitempty"`
	ExchangeID string `bson:"exchange_id,omitempty" json:"exchange_id,omitempty"`
	Symbol     string `bson:"symbol,omitempty" json:"symbol,omitempty"`
	Timeframe  string `bson:"timeframe,omitempty" json:"timeframe,omitempty"`
}

// AlertConfig stores alert configuration settings
type AlertConfig struct {
	ID                      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ConsecutiveFailureThreshold int            `bson:"consecutive_failure_threshold" json:"consecutive_failure_threshold"`
	NoDataAlertAfterMinutes     int            `bson:"no_data_alert_after_minutes" json:"no_data_alert_after_minutes"`
	RateLimitAlertThreshold     int            `bson:"rate_limit_alert_threshold" json:"rate_limit_alert_threshold"` // percentage
	EnabledAlertTypes       []AlertType        `bson:"enabled_alert_types" json:"enabled_alert_types"`
	UpdatedAt               time.Time          `bson:"updated_at" json:"updated_at"`
}

// DefaultAlertConfig returns the default alert configuration
func DefaultAlertConfig() *AlertConfig {
	return &AlertConfig{
		ConsecutiveFailureThreshold: 3,
		NoDataAlertAfterMinutes:     60,
		RateLimitAlertThreshold:     90,
		EnabledAlertTypes: []AlertType{
			AlertTypeJobFailed,
			AlertTypeJobConsecFailures,
			AlertTypeConnectorDown,
			AlertTypeRateLimitExceeded,
		},
	}
}

// AlertSummary provides a summary of alerts by status
type AlertSummary struct {
	Active       int64 `json:"active"`
	Acknowledged int64 `json:"acknowledged"`
	Total        int64 `json:"total"`
	BySeverity   map[AlertSeverity]int64 `json:"by_severity"`
	ByType       map[AlertType]int64     `json:"by_type"`
}

// AlertCreateRequest for creating new alerts programmatically
type AlertCreateRequest struct {
	Type     AlertType     `json:"type" validate:"required"`
	Severity AlertSeverity `json:"severity" validate:"required"`
	Title    string        `json:"title" validate:"required"`
	Message  string        `json:"message" validate:"required"`
	Source   AlertSource   `json:"source"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AlertUpdateRequest for updating alert status
type AlertUpdateRequest struct {
	Status         *AlertStatus `json:"status,omitempty"`
	AcknowledgedBy *string      `json:"acknowledged_by,omitempty"`
}
