package handlers

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/yourusername/datacollector/internal/api/errors"
	"github.com/yourusername/datacollector/internal/models"
	"github.com/yourusername/datacollector/internal/repository"
	"github.com/yourusername/datacollector/internal/service"
)

// AlertHandler handles alert-related endpoints
type AlertHandler struct {
	alertRepo    *repository.AlertRepository
	alertService *service.AlertService
}

// NewAlertHandler creates a new alert handler
func NewAlertHandler(alertRepo *repository.AlertRepository, alertService *service.AlertService) *AlertHandler {
	return &AlertHandler{
		alertRepo:    alertRepo,
		alertService: alertService,
	}
}

// GetAlerts retrieves all alerts with optional filters
// GET /api/v1/alerts
func (h *AlertHandler) GetAlerts(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}

	// Filter by status
	if status := c.Query("status"); status != "" {
		filter["status"] = models.AlertStatus(status)
	}

	// Filter by severity
	if severity := c.Query("severity"); severity != "" {
		filter["severity"] = models.AlertSeverity(severity)
	}

	// Filter by type
	if alertType := c.Query("type"); alertType != "" {
		filter["type"] = models.AlertType(alertType)
	}

	// Filter by exchange
	if exchangeID := c.Query("exchange_id"); exchangeID != "" {
		filter["source.exchange_id"] = exchangeID
	}

	// Limit results
	limit := int64(c.QueryInt("limit", 100))
	if limit > 500 {
		limit = 500
	}

	alerts, err := h.alertRepo.FindAll(ctx, filter, limit)
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to retrieve alerts"))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    alerts,
		"count":   len(alerts),
	})
}

// GetActiveAlerts retrieves all active alerts
// GET /api/v1/alerts/active
func (h *AlertHandler) GetActiveAlerts(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	alerts, err := h.alertRepo.FindActive(ctx)
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to retrieve active alerts"))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    alerts,
		"count":   len(alerts),
	})
}

// GetAlertSummary retrieves alert summary statistics
// GET /api/v1/alerts/summary
func (h *AlertHandler) GetAlertSummary(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	summary, err := h.alertRepo.GetSummary(ctx)
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to retrieve alert summary"))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    summary,
	})
}

// GetAlert retrieves a single alert by ID
// GET /api/v1/alerts/:id
func (h *AlertHandler) GetAlert(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	alert, err := h.alertRepo.FindByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid alert ID format"))
		}
		return errors.SendError(c, errors.NotFound("Alert"))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    alert,
	})
}

// GetAlertsByJob retrieves all alerts for a specific job
// GET /api/v1/jobs/:id/alerts
func (h *AlertHandler) GetAlertsByJob(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	jobID := c.Params("id")

	alerts, err := h.alertRepo.FindByJobID(ctx, jobID)
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to retrieve job alerts"))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    alerts,
		"count":   len(alerts),
	})
}

// GetAlertsByConnector retrieves all alerts for a specific connector/exchange
// GET /api/v1/connectors/:exchangeId/alerts
func (h *AlertHandler) GetAlertsByConnector(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exchangeID := c.Params("exchangeId")

	alerts, err := h.alertRepo.FindByExchangeID(ctx, exchangeID)
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to retrieve connector alerts"))
	}

	return c.JSON(fiber.Map{
		"success":     true,
		"data":        alerts,
		"count":       len(alerts),
		"exchange_id": exchangeID,
	})
}

// AcknowledgeAlert acknowledges an alert
// POST /api/v1/alerts/:id/acknowledge
func (h *AlertHandler) AcknowledgeAlert(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	var req struct {
		AcknowledgedBy string `json:"acknowledged_by"`
	}

	if err := c.BodyParser(&req); err != nil {
		req.AcknowledgedBy = "system"
	}

	if req.AcknowledgedBy == "" {
		req.AcknowledgedBy = "user"
	}

	if err := h.alertRepo.Acknowledge(ctx, id, req.AcknowledgedBy); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid alert ID format"))
		}
		if strings.Contains(err.Error(), "not found") {
			return errors.SendError(c, errors.NotFound("Alert"))
		}
		return errors.SendError(c, errors.DatabaseError("Failed to acknowledge alert"))
	}

	alert, _ := h.alertRepo.FindByID(ctx, id)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Alert acknowledged successfully",
		"data":    alert,
	})
}

// ResolveAlert resolves an alert
// POST /api/v1/alerts/:id/resolve
func (h *AlertHandler) ResolveAlert(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	if err := h.alertRepo.Resolve(ctx, id); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid alert ID format"))
		}
		if strings.Contains(err.Error(), "not found") {
			return errors.SendError(c, errors.NotFound("Alert"))
		}
		return errors.SendError(c, errors.DatabaseError("Failed to resolve alert"))
	}

	alert, _ := h.alertRepo.FindByID(ctx, id)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Alert resolved successfully",
		"data":    alert,
	})
}

// DeleteAlert deletes an alert
// DELETE /api/v1/alerts/:id
func (h *AlertHandler) DeleteAlert(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	if err := h.alertRepo.Delete(ctx, id); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid alert ID format"))
		}
		if strings.Contains(err.Error(), "not found") {
			return errors.SendError(c, errors.NotFound("Alert"))
		}
		return errors.SendError(c, errors.DatabaseError("Failed to delete alert"))
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// AcknowledgeAllAlerts acknowledges all active alerts
// POST /api/v1/alerts/acknowledge-all
func (h *AlertHandler) AcknowledgeAllAlerts(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var req struct {
		AcknowledgedBy string `json:"acknowledged_by"`
	}

	if err := c.BodyParser(&req); err != nil {
		req.AcknowledgedBy = "system"
	}

	if req.AcknowledgedBy == "" {
		req.AcknowledgedBy = "user"
	}

	// Get all active alerts
	activeAlerts, err := h.alertRepo.FindActive(ctx)
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to retrieve active alerts"))
	}

	acknowledged := 0
	for _, alert := range activeAlerts {
		if err := h.alertRepo.Acknowledge(ctx, alert.ID.Hex(), req.AcknowledgedBy); err == nil {
			acknowledged++
		}
	}

	return c.JSON(fiber.Map{
		"success":      true,
		"message":      "Alerts acknowledged",
		"acknowledged": acknowledged,
		"total":        len(activeAlerts),
	})
}

// GetAlertConfig retrieves the alert configuration
// GET /api/v1/alerts/config
func (h *AlertHandler) GetAlertConfig(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := h.alertRepo.GetConfig(ctx)
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to retrieve alert configuration"))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    config,
	})
}

// UpdateAlertConfig updates the alert configuration
// PUT /api/v1/alerts/config
func (h *AlertHandler) UpdateAlertConfig(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var config models.AlertConfig
	if err := c.BodyParser(&config); err != nil {
		return errors.SendError(c, errors.BadRequest("Invalid request body"))
	}

	// Validate
	if config.ConsecutiveFailureThreshold < 1 {
		return errors.SendError(c, errors.ValidationError("Invalid configuration", map[string]string{
			"consecutive_failure_threshold": "must be at least 1",
		}))
	}

	if err := h.alertRepo.SaveConfig(ctx, &config); err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to save alert configuration"))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Alert configuration updated",
		"data":    config,
	})
}

// TriggerAlertCheck manually triggers an alert check
// POST /api/v1/alerts/check
func (h *AlertHandler) TriggerAlertCheck(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := h.alertService.CheckJobsForAlerts(ctx); err != nil {
		return errors.SendError(c, errors.InternalError("Failed to check for alerts: "+err.Error()))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Alert check completed",
	})
}

// CleanupAlerts removes old resolved alerts
// POST /api/v1/alerts/cleanup
func (h *AlertHandler) CleanupAlerts(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Default to 7 days
	days := c.QueryInt("days", 7)
	if days < 1 {
		days = 1
	}

	olderThan := time.Duration(days) * 24 * time.Hour

	count, err := h.alertService.CleanupOldAlerts(ctx, olderThan)
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to cleanup alerts"))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Alert cleanup completed",
		"deleted": count,
	})
}
