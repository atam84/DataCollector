package handlers

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/yourusername/datacollector/internal/api/errors"
	"github.com/yourusername/datacollector/internal/config"
	"github.com/yourusername/datacollector/internal/models"
	"github.com/yourusername/datacollector/internal/repository"
)

// ConnectorHandler handles connector-related endpoints
type ConnectorHandler struct {
	repo     *repository.ConnectorRepository
	jobRepo  *repository.JobRepository
	ohlcvRepo *repository.OHLCVRepository
	config   *config.Config
}

// NewConnectorHandler creates a new connector handler
func NewConnectorHandler(repo *repository.ConnectorRepository, jobRepo *repository.JobRepository, cfg *config.Config) *ConnectorHandler {
	return &ConnectorHandler{
		repo:    repo,
		jobRepo: jobRepo,
		config:  cfg,
	}
}

// NewConnectorHandlerWithOHLCV creates a new connector handler with OHLCV repository
func NewConnectorHandlerWithOHLCV(repo *repository.ConnectorRepository, jobRepo *repository.JobRepository, ohlcvRepo *repository.OHLCVRepository, cfg *config.Config) *ConnectorHandler {
	return &ConnectorHandler{
		repo:      repo,
		jobRepo:   jobRepo,
		ohlcvRepo: ohlcvRepo,
		config:    cfg,
	}
}

// CreateConnector creates a new connector
// @Summary Create a new connector
// @Description Creates a new exchange connector with rate limiting configuration
// @Tags Connectors
// @Accept json
// @Produce json
// @Param request body models.ConnectorCreateRequest true "Connector configuration"
// @Success 201 {object} map[string]interface{} "Connector created"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 409 {object} map[string]interface{} "Connector already exists"
// @Router /connectors [post]
func (h *ConnectorHandler) CreateConnector(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var req models.ConnectorCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.SendError(c, errors.BadRequest("Invalid request body"))
	}

	// Validate required fields
	if req.ExchangeID == "" || req.DisplayName == "" {
		return errors.SendError(c, errors.ValidationError("Missing required fields", map[string]string{
			"exchange_id":  "required",
			"display_name": "required",
		}))
	}

	if req.RateLimit.Limit <= 0 || req.RateLimit.PeriodMs < 1000 {
		return errors.SendError(c, errors.ValidationError("Invalid rate limit configuration", map[string]interface{}{
			"rate_limit.limit":     "must be > 0",
			"rate_limit.period_ms": "must be >= 1000",
		}))
	}

	// Calculate MinDelayMs if not provided
	minDelayMs := req.RateLimit.MinDelayMs
	if minDelayMs == 0 {
		// Calculate from limit/period: e.g., 20 calls per 60000ms = 3000ms between calls
		if req.RateLimit.Limit > 0 {
			minDelayMs = req.RateLimit.PeriodMs / req.RateLimit.Limit
			// Ensure minimum of 1000ms (1 second) for safety
			if minDelayMs < 1000 {
				minDelayMs = 1000
			}
		} else {
			minDelayMs = 5000 // Default 5 seconds
		}
	}

	// Create connector model
	connector := &models.Connector{
		ExchangeID:  req.ExchangeID,
		DisplayName: req.DisplayName,
		Status:      "active",
		RateLimit: models.RateLimit{
			Limit:       req.RateLimit.Limit,
			PeriodMs:    req.RateLimit.PeriodMs,
			MinDelayMs:  minDelayMs,
			Usage:       0,
			PeriodStart: time.Now(),
		},
	}

	// Create in database
	if err := h.repo.Create(ctx, connector); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return errors.SendError(c, errors.Conflict("Connector already exists for this exchange"))
		}
		return errors.SendError(c, errors.DatabaseError("Failed to create connector"))
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    connector,
	})
}

// GetConnectors retrieves all connectors
// @Summary Get all connectors
// @Description Retrieves all exchange connectors with job counts
// @Tags Connectors
// @Accept json
// @Produce json
// @Param status query string false "Filter by status (active, suspended)"
// @Success 200 {object} map[string]interface{} "List of connectors"
// @Router /connectors [get]
func (h *ConnectorHandler) GetConnectors(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Optional filter by status
	filter := bson.M{}
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}

	connectors, err := h.repo.FindAll(ctx, filter)
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to retrieve connectors"))
	}

	// Enhance connectors with job counts
	responses := make([]models.ConnectorResponse, 0, len(connectors))
	for _, connector := range connectors {
		jobCount, _ := h.jobRepo.CountByConnector(ctx, connector.ExchangeID)
		activeJobCount, _ := h.jobRepo.CountActiveByConnector(ctx, connector.ExchangeID)

		response := models.ConnectorResponse{
			Connector:      *connector,
			JobCount:       jobCount,
			ActiveJobCount: activeJobCount,
		}
		responses = append(responses, response)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    responses,
		"count":   len(responses),
	})
}

// GetConnector retrieves a connector by ID
// GET /api/v1/connectors/:id
func (h *ConnectorHandler) GetConnector(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	connector, err := h.repo.FindByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid connector ID format"))
		}
		return errors.SendError(c, errors.NotFound("Connector"))
	}

	// Enhance with job counts
	jobCount, _ := h.jobRepo.CountByConnector(ctx, connector.ExchangeID)
	activeJobCount, _ := h.jobRepo.CountActiveByConnector(ctx, connector.ExchangeID)

	response := models.ConnectorResponse{
		Connector:      *connector,
		JobCount:       jobCount,
		ActiveJobCount: activeJobCount,
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// UpdateConnector updates a connector
// PUT /api/v1/connectors/:id
func (h *ConnectorHandler) UpdateConnector(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	var req models.ConnectorUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.SendError(c, errors.BadRequest("Invalid request body"))
	}

	// Build update map
	update := bson.M{}

	if req.DisplayName != nil {
		update["display_name"] = *req.DisplayName
	}

	if req.Status != nil {
		if *req.Status != "active" && *req.Status != "disabled" {
			return errors.SendError(c, errors.ValidationError("Invalid status value", map[string]interface{}{
				"status":         *req.Status,
				"allowed_values": []string{"active", "disabled"},
			}))
		}
		update["status"] = *req.Status
	}

	if req.RateLimit != nil {
		if req.RateLimit.Limit != nil {
			update["rate_limit.limit"] = *req.RateLimit.Limit
		}
		if req.RateLimit.PeriodMs != nil {
			update["rate_limit.period_ms"] = *req.RateLimit.PeriodMs
		}
		if req.RateLimit.MinDelayMs != nil {
			update["rate_limit.min_delay_ms"] = *req.RateLimit.MinDelayMs
		}
	}

	if len(update) == 0 {
		return errors.SendError(c, errors.BadRequest("No fields to update"))
	}

	// Update connector
	if err := h.repo.Update(ctx, id, update); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return errors.SendError(c, errors.NotFound("Connector"))
		}
		return errors.SendError(c, errors.DatabaseError("Failed to update connector"))
	}

	// Fetch updated connector
	connector, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to retrieve updated connector"))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    connector,
	})
}

// DeleteConnector deletes a connector
// DELETE /api/v1/connectors/:id
func (h *ConnectorHandler) DeleteConnector(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	if err := h.repo.Delete(ctx, id); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid connector ID format"))
		}
		if strings.Contains(err.Error(), "not found") {
			return errors.SendError(c, errors.NotFound("Connector"))
		}
		return errors.SendError(c, errors.DatabaseError("Failed to delete connector"))
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// SuspendConnector suspends a connector and all its jobs
// POST /api/v1/connectors/:id/suspend
func (h *ConnectorHandler) SuspendConnector(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	// Get connector first to get exchange_id
	connector, err := h.repo.FindByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid connector ID format"))
		}
		return errors.SendError(c, errors.NotFound("Connector"))
	}

	// Update connector status to disabled (suspended)
	update := bson.M{"status": "disabled"}
	if err := h.repo.Update(ctx, id, update); err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to suspend connector"))
	}

	// Suspend all jobs attached to this connector
	if err := h.jobRepo.UpdateStatusByConnector(ctx, connector.ExchangeID, "paused"); err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to suspend attached jobs"))
	}

	// Fetch updated connector with job counts
	connector, _ = h.repo.FindByID(ctx, id)
	jobCount, _ := h.jobRepo.CountByConnector(ctx, connector.ExchangeID)
	activeJobCount, _ := h.jobRepo.CountActiveByConnector(ctx, connector.ExchangeID)

	response := models.ConnectorResponse{
		Connector:      *connector,
		JobCount:       jobCount,
		ActiveJobCount: activeJobCount,
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Connector and all attached jobs suspended successfully",
		"data":    response,
	})
}

// ResumeConnector resumes a suspended connector and all its jobs
// POST /api/v1/connectors/:id/resume
func (h *ConnectorHandler) ResumeConnector(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	// Get connector first to get exchange_id
	connector, err := h.repo.FindByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid connector ID format"))
		}
		return errors.SendError(c, errors.NotFound("Connector"))
	}

	// Update connector status to active
	update := bson.M{"status": "active"}
	if err := h.repo.Update(ctx, id, update); err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to resume connector"))
	}

	// Resume all jobs attached to this connector
	if err := h.jobRepo.UpdateStatusByConnector(ctx, connector.ExchangeID, "active"); err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to resume attached jobs"))
	}

	// Fetch updated connector with job counts
	connector, _ = h.repo.FindByID(ctx, id)
	jobCount, _ := h.jobRepo.CountByConnector(ctx, connector.ExchangeID)
	activeJobCount, _ := h.jobRepo.CountActiveByConnector(ctx, connector.ExchangeID)

	response := models.ConnectorResponse{
		Connector:      *connector,
		JobCount:       jobCount,
		ActiveJobCount: activeJobCount,
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Connector and all attached jobs resumed successfully",
		"data":    response,
	})
}

// GetRateLimitStatus returns the current rate limit status for a connector
// GET /api/v1/connectors/:id/rate-limit
func (h *ConnectorHandler) GetRateLimitStatus(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	connector, err := h.repo.FindByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid connector ID format"))
		}
		return errors.SendError(c, errors.NotFound("Connector"))
	}

	// Calculate current status
	now := time.Now()
	periodElapsed := now.Sub(connector.RateLimit.PeriodStart).Milliseconds()
	periodRemaining := int64(connector.RateLimit.PeriodMs) - periodElapsed
	if periodRemaining < 0 {
		periodRemaining = 0
	}

	// Calculate effective min delay
	minDelayMs := connector.RateLimit.MinDelayMs
	if minDelayMs == 0 && connector.RateLimit.Limit > 0 && connector.RateLimit.PeriodMs > 0 {
		minDelayMs = connector.RateLimit.PeriodMs / connector.RateLimit.Limit
		if minDelayMs < 1000 {
			minDelayMs = 1000
		}
	}
	if minDelayMs == 0 {
		minDelayMs = 5000 // Default 5 seconds
	}

	// Calculate time since last API call
	var lastCallMs int64 = -1
	var canCallNow bool = true
	if connector.RateLimit.LastAPICallAt != nil {
		lastCallMs = now.Sub(*connector.RateLimit.LastAPICallAt).Milliseconds()
		canCallNow = lastCallMs >= int64(minDelayMs) && connector.RateLimit.Usage < connector.RateLimit.Limit
	}

	return c.JSON(fiber.Map{
		"exchange_id":         connector.ExchangeID,
		"limit":               connector.RateLimit.Limit,
		"usage":               connector.RateLimit.Usage,
		"period_ms":           connector.RateLimit.PeriodMs,
		"period_remaining_ms": periodRemaining,
		"min_delay_ms":        minDelayMs,
		"last_call_ms":        lastCallMs,
		"can_call_now":        canCallNow,
		"last_api_call_at":    connector.RateLimit.LastAPICallAt,
		"period_start":        connector.RateLimit.PeriodStart,
	})
}

// ResetRateLimitUsage resets the rate limit usage counter for a connector
// POST /api/v1/connectors/:id/rate-limit/reset
func (h *ConnectorHandler) ResetRateLimitUsage(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	connector, err := h.repo.FindByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid connector ID format"))
		}
		return errors.SendError(c, errors.NotFound("Connector"))
	}

	// Reset the rate limit period
	if err := h.repo.ResetRateLimitPeriod(ctx, connector.ExchangeID); err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to reset rate limit"))
	}

	return c.JSON(fiber.Map{
		"success":     true,
		"message":     "Rate limit usage reset successfully",
		"exchange_id": connector.ExchangeID,
	})
}

// GetConnectorStats returns statistics for a connector including data volume
// GET /api/v1/connectors/:id/stats
func (h *ConnectorHandler) GetConnectorStats(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	id := c.Params("id")

	connector, err := h.repo.FindByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid connector ID format"))
		}
		return errors.SendError(c, errors.NotFound("Connector"))
	}

	// Get job counts
	jobCount, _ := h.jobRepo.CountByConnector(ctx, connector.ExchangeID)
	activeJobCount, _ := h.jobRepo.CountActiveByConnector(ctx, connector.ExchangeID)

	// Get OHLCV stats if repository is available
	var ohlcvStats *models.OHLCVStats
	if h.ohlcvRepo != nil {
		ohlcvStats, _ = h.ohlcvRepo.GetStatsByExchange(ctx, connector.ExchangeID)
	}

	// Get last run time from jobs
	jobs, _ := h.jobRepo.FindByConnector(ctx, connector.ExchangeID)
	var lastRunTime *time.Time
	var failedJobs int
	for _, job := range jobs {
		if job.RunState.LastRunTime != nil {
			if lastRunTime == nil || job.RunState.LastRunTime.After(*lastRunTime) {
				lastRunTime = job.RunState.LastRunTime
			}
		}
		if job.RunState.LastError != nil && *job.RunState.LastError != "" {
			failedJobs++
		}
	}

	response := fiber.Map{
		"connector_id":     connector.ID.Hex(),
		"exchange_id":      connector.ExchangeID,
		"display_name":     connector.DisplayName,
		"status":           connector.Status,
		"job_count":        jobCount,
		"active_job_count": activeJobCount,
		"failed_jobs":      failedJobs,
		"last_run_time":    lastRunTime,
		"rate_limit": fiber.Map{
			"limit":        connector.RateLimit.Limit,
			"usage":        connector.RateLimit.Usage,
			"period_ms":    connector.RateLimit.PeriodMs,
			"min_delay_ms": connector.RateLimit.MinDelayMs,
		},
	}

	if ohlcvStats != nil {
		response["data_stats"] = fiber.Map{
			"total_candles":     ohlcvStats.TotalCandles,
			"total_chunks":      ohlcvStats.TotalChunks,
			"unique_symbols":    ohlcvStats.UniqueSymbols,
			"unique_timeframes": ohlcvStats.UniqueTimeframes,
			"oldest_data":       ohlcvStats.OldestData,
			"newest_data":       ohlcvStats.NewestData,
		}
	}

	return c.JSON(response)
}

// GetAllStats returns aggregate statistics across all connectors
// GET /api/v1/stats
func (h *ConnectorHandler) GetAllStats(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get connector stats
	connectors, err := h.repo.FindAll(ctx, bson.M{})
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to retrieve connector statistics"))
	}

	activeConnectors := 0
	for _, conn := range connectors {
		if conn.Status == "active" {
			activeConnectors++
		}
	}

	// Get job stats by counting all jobs
	allJobs, _ := h.jobRepo.FindAll(ctx, bson.M{})
	totalJobs := int64(len(allJobs))
	activeJobs := int64(0)
	for _, job := range allJobs {
		if job.Status == "active" {
			activeJobs++
		}
	}

	// Get OHLCV stats if repository is available
	var ohlcvStats *models.OHLCVStats
	if h.ohlcvRepo != nil {
		ohlcvStats, _ = h.ohlcvRepo.GetAllStats(ctx)
	}

	response := fiber.Map{
		"connectors": fiber.Map{
			"total":  len(connectors),
			"active": activeConnectors,
		},
		"jobs": fiber.Map{
			"total":  totalJobs,
			"active": activeJobs,
		},
	}

	if ohlcvStats != nil {
		response["data"] = fiber.Map{
			"total_candles":     ohlcvStats.TotalCandles,
			"total_chunks":      ohlcvStats.TotalChunks,
			"unique_exchanges":  ohlcvStats.UniqueExchanges,
			"unique_symbols":    ohlcvStats.UniqueSymbols,
			"unique_timeframes": ohlcvStats.UniqueTimeframes,
			"oldest_data":       ohlcvStats.OldestData,
			"newest_data":       ohlcvStats.NewestData,
		}
	}

	return c.JSON(response)
}

// GetConnectorHealth returns health status for a specific connector
// @Summary Get connector health status
// @Description Returns health metrics and status for a specific connector
// @Tags Connectors
// @Accept json
// @Produce json
// @Param id path string true "Connector ID"
// @Success 200 {object} map[string]interface{} "Health status"
// @Failure 404 {object} map[string]interface{} "Connector not found"
// @Router /connectors/{id}/health [get]
func (h *ConnectorHandler) GetConnectorHealth(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	connector, err := h.repo.FindByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid connector ID format"))
		}
		return errors.SendError(c, errors.NotFound("Connector"))
	}

	// Get job counts for context
	jobCount, _ := h.jobRepo.CountByConnector(ctx, connector.ExchangeID)
	activeJobCount, _ := h.jobRepo.CountActiveByConnector(ctx, connector.ExchangeID)

	// Determine health status description
	healthDescription := "All systems operational"
	if connector.Health.Status == "degraded" {
		healthDescription = "Some API calls failing, monitoring"
	} else if connector.Health.Status == "unhealthy" {
		healthDescription = "Multiple consecutive failures detected"
	}

	// Calculate error rate
	errorRate := float64(0)
	if connector.Health.TotalCalls > 0 {
		errorRate = (float64(connector.Health.TotalFailures) / float64(connector.Health.TotalCalls)) * 100
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"exchange_id":           connector.ExchangeID,
			"display_name":          connector.DisplayName,
			"connector_status":      connector.Status,
			"health":                connector.Health,
			"health_description":    healthDescription,
			"error_rate_percentage": errorRate,
			"job_count":             jobCount,
			"active_job_count":      activeJobCount,
		},
	})
}

// GetAllConnectorsHealth returns health status for all connectors
// @Summary Get all connectors health status
// @Description Returns health metrics and status for all connectors
// @Tags Connectors
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Health status for all connectors"
// @Router /connectors/health [get]
func (h *ConnectorHandler) GetAllConnectorsHealth(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	connectors, err := h.repo.FindAll(ctx, bson.M{})
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to retrieve connectors"))
	}

	// Build health report for each connector
	healthReports := make([]fiber.Map, 0, len(connectors))
	healthySummary := 0
	degradedSummary := 0
	unhealthySummary := 0

	for _, conn := range connectors {
		// Calculate error rate
		errorRate := float64(0)
		if conn.Health.TotalCalls > 0 {
			errorRate = (float64(conn.Health.TotalFailures) / float64(conn.Health.TotalCalls)) * 100
		}

		// Determine health status (default to healthy if no status set yet)
		healthStatus := conn.Health.Status
		if healthStatus == "" {
			healthStatus = "healthy"
		}

		switch healthStatus {
		case "healthy":
			healthySummary++
		case "degraded":
			degradedSummary++
		case "unhealthy":
			unhealthySummary++
		}

		// Get job counts
		jobCount, _ := h.jobRepo.CountByConnector(ctx, conn.ExchangeID)
		activeJobCount, _ := h.jobRepo.CountActiveByConnector(ctx, conn.ExchangeID)

		healthReports = append(healthReports, fiber.Map{
			"exchange_id":           conn.ExchangeID,
			"display_name":          conn.DisplayName,
			"connector_status":      conn.Status,
			"health_status":         healthStatus,
			"total_calls":           conn.Health.TotalCalls,
			"total_failures":        conn.Health.TotalFailures,
			"consecutive_failures":  conn.Health.ConsecutiveFailures,
			"error_rate_percentage": errorRate,
			"uptime_percentage":     conn.Health.UptimePercentage,
			"average_response_ms":   conn.Health.AverageResponseMs,
			"last_successful_call":  conn.Health.LastSuccessfulCall,
			"last_failed_call":      conn.Health.LastFailedCall,
			"last_error":            conn.Health.LastError,
			"job_count":             jobCount,
			"active_job_count":      activeJobCount,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"summary": fiber.Map{
			"total":     len(connectors),
			"healthy":   healthySummary,
			"degraded":  degradedSummary,
			"unhealthy": unhealthySummary,
		},
		"connectors": healthReports,
	})
}

