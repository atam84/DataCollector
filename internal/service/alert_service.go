package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yourusername/datacollector/internal/models"
	"github.com/yourusername/datacollector/internal/repository"
)

// AlertService handles alert generation and management
type AlertService struct {
	alertRepo     *repository.AlertRepository
	jobRepo       *repository.JobRepository
	connectorRepo *repository.ConnectorRepository
}

// NewAlertService creates a new alert service
func NewAlertService(alertRepo *repository.AlertRepository, jobRepo *repository.JobRepository, connectorRepo *repository.ConnectorRepository) *AlertService {
	return &AlertService{
		alertRepo:     alertRepo,
		jobRepo:       jobRepo,
		connectorRepo: connectorRepo,
	}
}

// CreateAlert creates a new alert if one doesn't already exist
func (s *AlertService) CreateAlert(ctx context.Context, alertType models.AlertType, severity models.AlertSeverity, title, message string, source models.AlertSource, metadata map[string]interface{}) error {
	// Check if an active alert already exists for this source
	exists, err := s.alertRepo.ExistsActiveAlert(ctx, alertType, source.Type, source.ID)
	if err != nil {
		return fmt.Errorf("failed to check existing alert: %w", err)
	}

	if exists {
		log.Printf("[ALERT] Alert already exists for %s %s, skipping", source.Type, source.ID)
		return nil
	}

	alert := &models.Alert{
		Type:     alertType,
		Severity: severity,
		Status:   models.AlertStatusActive,
		Title:    title,
		Message:  message,
		Source:   source,
		Metadata: metadata,
	}

	if err := s.alertRepo.Create(ctx, alert); err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	log.Printf("[ALERT] Created %s alert: %s", severity, title)
	return nil
}

// AlertJobFailed creates an alert for a failed job
func (s *AlertService) AlertJobFailed(ctx context.Context, job *models.Job, errorMsg string) error {
	source := models.AlertSource{
		Type:       "job",
		ID:         job.ID.Hex(),
		ExchangeID: job.ConnectorExchangeID,
		Symbol:     job.Symbol,
		Timeframe:  job.Timeframe,
	}

	metadata := map[string]interface{}{
		"error":        errorMsg,
		"symbol":       job.Symbol,
		"timeframe":    job.Timeframe,
		"exchange_id":  job.ConnectorExchangeID,
		"last_run":     job.RunState.LastRunTime,
		"runs_total":   job.RunState.RunsTotal,
	}

	title := fmt.Sprintf("Job failed: %s/%s on %s", job.Symbol, job.Timeframe, job.ConnectorExchangeID)
	message := fmt.Sprintf("Job execution failed with error: %s", errorMsg)

	return s.CreateAlert(ctx, models.AlertTypeJobFailed, models.AlertSeverityError, title, message, source, metadata)
}

// AlertConsecutiveFailures creates an alert for consecutive job failures
func (s *AlertService) AlertConsecutiveFailures(ctx context.Context, job *models.Job, failures int) error {
	// Determine severity based on failure count
	var severity models.AlertSeverity
	if failures >= 5 {
		severity = models.AlertSeverityCritical
	} else if failures >= 3 {
		severity = models.AlertSeverityError
	} else {
		severity = models.AlertSeverityWarning
	}

	source := models.AlertSource{
		Type:       "job",
		ID:         job.ID.Hex(),
		ExchangeID: job.ConnectorExchangeID,
		Symbol:     job.Symbol,
		Timeframe:  job.Timeframe,
	}

	metadata := map[string]interface{}{
		"consecutive_failures": failures,
		"symbol":               job.Symbol,
		"timeframe":            job.Timeframe,
		"exchange_id":          job.ConnectorExchangeID,
		"last_failure_time":    job.RunState.LastFailureTime,
		"last_error":           job.RunState.LastError,
	}

	title := fmt.Sprintf("Job has %d consecutive failures: %s/%s", failures, job.Symbol, job.Timeframe)
	message := fmt.Sprintf("Job %s/%s on %s has failed %d times in a row. Last error: %s",
		job.Symbol, job.Timeframe, job.ConnectorExchangeID, failures, safeString(job.RunState.LastError))

	return s.CreateAlert(ctx, models.AlertTypeJobConsecFailures, severity, title, message, source, metadata)
}

// AlertConnectorDown creates an alert for a down connector
func (s *AlertService) AlertConnectorDown(ctx context.Context, connector *models.Connector, reason string) error {
	source := models.AlertSource{
		Type:       "connector",
		ID:         connector.ID.Hex(),
		ExchangeID: connector.ExchangeID,
	}

	metadata := map[string]interface{}{
		"exchange_id":  connector.ExchangeID,
		"display_name": connector.DisplayName,
		"status":       connector.Status,
		"reason":       reason,
	}

	title := fmt.Sprintf("Connector down: %s", connector.DisplayName)
	message := fmt.Sprintf("Connector %s (%s) is down. Reason: %s",
		connector.DisplayName, connector.ExchangeID, reason)

	return s.CreateAlert(ctx, models.AlertTypeConnectorDown, models.AlertSeverityCritical, title, message, source, metadata)
}

// AlertRateLimitExceeded creates an alert for rate limit issues
func (s *AlertService) AlertRateLimitExceeded(ctx context.Context, connector *models.Connector) error {
	source := models.AlertSource{
		Type:       "connector",
		ID:         connector.ID.Hex(),
		ExchangeID: connector.ExchangeID,
	}

	usagePercent := 0
	if connector.RateLimit.Limit > 0 {
		usagePercent = (connector.RateLimit.Usage * 100) / connector.RateLimit.Limit
	}

	metadata := map[string]interface{}{
		"exchange_id":    connector.ExchangeID,
		"display_name":   connector.DisplayName,
		"limit":          connector.RateLimit.Limit,
		"usage":          connector.RateLimit.Usage,
		"usage_percent":  usagePercent,
		"period_ms":      connector.RateLimit.PeriodMs,
	}

	title := fmt.Sprintf("Rate limit at %d%%: %s", usagePercent, connector.DisplayName)
	message := fmt.Sprintf("Connector %s has used %d/%d API calls (%d%%) in the current period",
		connector.DisplayName, connector.RateLimit.Usage, connector.RateLimit.Limit, usagePercent)

	return s.CreateAlert(ctx, models.AlertTypeRateLimitExceeded, models.AlertSeverityWarning, title, message, source, metadata)
}

// AlertNoDataCollected creates an alert when no data has been collected for a while
func (s *AlertService) AlertNoDataCollected(ctx context.Context, job *models.Job, sinceMinutes int) error {
	source := models.AlertSource{
		Type:       "job",
		ID:         job.ID.Hex(),
		ExchangeID: job.ConnectorExchangeID,
		Symbol:     job.Symbol,
		Timeframe:  job.Timeframe,
	}

	metadata := map[string]interface{}{
		"symbol":           job.Symbol,
		"timeframe":        job.Timeframe,
		"exchange_id":      job.ConnectorExchangeID,
		"minutes_since":    sinceMinutes,
		"last_candle_time": job.Cursor.LastCandleTime,
	}

	title := fmt.Sprintf("No data collected: %s/%s for %d minutes", job.Symbol, job.Timeframe, sinceMinutes)
	message := fmt.Sprintf("Job %s/%s on %s has not collected any new data for %d minutes",
		job.Symbol, job.Timeframe, job.ConnectorExchangeID, sinceMinutes)

	return s.CreateAlert(ctx, models.AlertTypeNoDataCollected, models.AlertSeverityWarning, title, message, source, metadata)
}

// ResolveJobAlerts resolves all active alerts for a job (called on successful execution)
func (s *AlertService) ResolveJobAlerts(ctx context.Context, jobID string) error {
	count, err := s.alertRepo.ResolveBySource(ctx, "job", jobID)
	if err != nil {
		return fmt.Errorf("failed to resolve job alerts: %w", err)
	}
	if count > 0 {
		log.Printf("[ALERT] Resolved %d alerts for job %s", count, jobID)
	}
	return nil
}

// ResolveConnectorAlerts resolves all active alerts for a connector
func (s *AlertService) ResolveConnectorAlerts(ctx context.Context, connectorID string) error {
	count, err := s.alertRepo.ResolveBySource(ctx, "connector", connectorID)
	if err != nil {
		return fmt.Errorf("failed to resolve connector alerts: %w", err)
	}
	if count > 0 {
		log.Printf("[ALERT] Resolved %d alerts for connector %s", count, connectorID)
	}
	return nil
}

// CheckJobsForAlerts checks all jobs and generates alerts as needed
func (s *AlertService) CheckJobsForAlerts(ctx context.Context) error {
	config, err := s.alertRepo.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get alert config: %w", err)
	}

	// Get jobs with consecutive failures
	jobs, err := s.jobRepo.GetJobsWithFailures(ctx)
	if err != nil {
		return fmt.Errorf("failed to get jobs with failures: %w", err)
	}

	for _, job := range jobs {
		if job.RunState.ConsecutiveFailures >= config.ConsecutiveFailureThreshold {
			if err := s.AlertConsecutiveFailures(ctx, job, job.RunState.ConsecutiveFailures); err != nil {
				log.Printf("[ALERT] Failed to create consecutive failures alert: %v", err)
			}
		}
	}

	return nil
}

// CleanupOldAlerts removes resolved alerts older than the specified duration
func (s *AlertService) CleanupOldAlerts(ctx context.Context, olderThan time.Duration) (int64, error) {
	count, err := s.alertRepo.DeleteOlderThan(ctx, olderThan)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old alerts: %w", err)
	}
	if count > 0 {
		log.Printf("[ALERT] Cleaned up %d old resolved alerts", count)
	}
	return count, nil
}

// Helper function to safely dereference string pointers
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
