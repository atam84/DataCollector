package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/yourusername/datacollector/internal/models"
	"github.com/yourusername/datacollector/internal/repository"
	"github.com/yourusername/datacollector/internal/service"
)

// JobHandler handles job-related endpoints
type JobHandler struct {
	jobRepo       *repository.JobRepository
	connectorRepo *repository.ConnectorRepository
	jobExecutor   *service.JobExecutor
}

// NewJobHandler creates a new job handler
func NewJobHandler(jobRepo *repository.JobRepository, connectorRepo *repository.ConnectorRepository, jobExecutor *service.JobExecutor) *JobHandler {
	return &JobHandler{
		jobRepo:       jobRepo,
		connectorRepo: connectorRepo,
		jobExecutor:   jobExecutor,
	}
}

// CreateJob creates a new job
// POST /api/v1/jobs
func (h *JobHandler) CreateJob(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var req models.JobCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.ConnectorExchangeID == "" || req.Symbol == "" || req.Timeframe == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "connector_exchange_id, symbol, and timeframe are required",
		})
	}

	// Verify connector exists
	_, err := h.connectorRepo.FindByExchangeID(ctx, req.ConnectorExchangeID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Connector not found for exchange: " + req.ConnectorExchangeID,
		})
	}

	// Create job model
	status := req.Status
	if status == "" {
		status = "active"
	}

	// Calculate initial next run time
	nextRunTime := time.Now().Add(1 * time.Minute) // Start in 1 minute

	job := &models.Job{
		ConnectorExchangeID: req.ConnectorExchangeID,
		Symbol:              req.Symbol,
		Timeframe:           req.Timeframe,
		Status:              status,
		Schedule: models.Schedule{
			Mode: "timeframe",
		},
		RunState: models.RunState{
			NextRunTime: &nextRunTime,
		},
	}

	// Create in database
	if err := h.jobRepo.Create(ctx, job); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(job)
}

// GetJobs retrieves all jobs
// GET /api/v1/jobs
func (h *JobHandler) GetJobs(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Optional filters
	filter := bson.M{}

	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}

	if exchangeID := c.Query("exchange_id"); exchangeID != "" {
		filter["connector_exchange_id"] = exchangeID
	}

	if symbol := c.Query("symbol"); symbol != "" {
		filter["symbol"] = symbol
	}

	if timeframe := c.Query("timeframe"); timeframe != "" {
		filter["timeframe"] = timeframe
	}

	jobs, err := h.jobRepo.FindAll(ctx, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":  jobs,
		"count": len(jobs),
	})
}

// GetJob retrieves a job by ID
// GET /api/v1/jobs/:id
func (h *JobHandler) GetJob(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	job, err := h.jobRepo.FindByID(ctx, id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(job)
}

// GetJobsByConnector retrieves all jobs for a connector
// GET /api/v1/connectors/:exchangeId/jobs
func (h *JobHandler) GetJobsByConnector(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exchangeID := c.Params("exchangeId")

	jobs, err := h.jobRepo.FindByConnector(ctx, exchangeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":        jobs,
		"count":       len(jobs),
		"exchange_id": exchangeID,
	})
}

// UpdateJob updates a job
// PUT /api/v1/jobs/:id
func (h *JobHandler) UpdateJob(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	var req models.JobUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Build update map
	update := bson.M{}

	if req.Status != nil {
		validStatuses := map[string]bool{"active": true, "paused": true, "error": true}
		if !validStatuses[*req.Status] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "status must be 'active', 'paused', or 'error'",
			})
		}
		update["status"] = *req.Status
	}

	if req.Timeframe != nil {
		update["timeframe"] = *req.Timeframe
	}

	if len(update) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No fields to update",
		})
	}

	// Update job
	if err := h.jobRepo.Update(ctx, id, update); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Fetch updated job
	job, err := h.jobRepo.FindByID(ctx, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(job)
}

// PauseJob pauses a job
// POST /api/v1/jobs/:id/pause
func (h *JobHandler) PauseJob(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	if err := h.jobRepo.UpdateStatus(ctx, id, "paused"); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	job, err := h.jobRepo.FindByID(ctx, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Job paused successfully",
		"job":     job,
	})
}

// ResumeJob resumes a paused job
// POST /api/v1/jobs/:id/resume
func (h *JobHandler) ResumeJob(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	// Get job to check connector
	job, err := h.jobRepo.FindByID(ctx, id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Check if connector is active
	connector, err := h.connectorRepo.FindByExchangeID(ctx, job.ConnectorExchangeID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Connector not found",
		})
	}

	if connector.Status != "active" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot resume job: connector is suspended. Please resume the connector first.",
		})
	}

	// Resume the job
	if err := h.jobRepo.UpdateStatus(ctx, id, "active"); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	job, err = h.jobRepo.FindByID(ctx, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Job resumed successfully",
		"job":     job,
	})
}

// DeleteJob deletes a job
// DELETE /api/v1/jobs/:id
func (h *JobHandler) DeleteJob(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	if err := h.jobRepo.Delete(ctx, id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ExecuteJob executes a job manually
// POST /api/v1/jobs/:id/execute
func (h *JobHandler) ExecuteJob(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	id := c.Params("id")

	// Execute the job
	result, err := h.jobExecutor.ExecuteJob(ctx, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Return result
	if !result.Success {
		return c.Status(fiber.StatusOK).JSON(result)
	}

	return c.JSON(result)
}

// GetQueue retrieves upcoming job executions
// GET /api/v1/jobs/queue
func (h *JobHandler) GetQueue(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get all active jobs sorted by next run time
	filter := bson.M{"status": "active"}
	jobs, err := h.jobRepo.FindAll(ctx, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Sort jobs by next_run_time (ascending)
	// Filter out jobs without next_run_time
	queuedJobs := make([]*models.Job, 0)
	for _, job := range jobs {
		if job.RunState.NextRunTime != nil {
			queuedJobs = append(queuedJobs, job)
		}
	}

	// Sort by next run time
	for i := 0; i < len(queuedJobs); i++ {
		for j := i + 1; j < len(queuedJobs); j++ {
			if queuedJobs[i].RunState.NextRunTime.After(*queuedJobs[j].RunState.NextRunTime) {
				queuedJobs[i], queuedJobs[j] = queuedJobs[j], queuedJobs[i]
			}
		}
	}

	return c.JSON(fiber.Map{
		"data":  queuedJobs,
		"count": len(queuedJobs),
	})
}

// GetIndicatorConfig retrieves the indicator configuration for a job
// GET /api/v1/jobs/:id/indicators/config
func (h *JobHandler) GetIndicatorConfig(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	job, err := h.jobRepo.FindByID(ctx, id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Job not found",
		})
	}

	return c.JSON(fiber.Map{
		"job_id": id,
		"config": job.IndicatorConfig,
	})
}

// UpdateIndicatorConfig updates the indicator configuration for a job
// PUT /api/v1/jobs/:id/indicators/config
func (h *JobHandler) UpdateIndicatorConfig(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	var config models.IndicatorConfig
	if err := c.BodyParser(&config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update the job's indicator config
	update := bson.M{
		"indicator_config": config,
	}

	if err := h.jobRepo.Update(ctx, id, update); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update indicator configuration",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Indicator configuration updated successfully",
		"job_id":  id,
		"config":  config,
	})
}

// PatchIndicatorConfig partially updates the indicator configuration for a job
// PATCH /api/v1/jobs/:id/indicators/config
func (h *JobHandler) PatchIndicatorConfig(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	// Get current job
	job, err := h.jobRepo.FindByID(ctx, id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Job not found",
		})
	}

	// Parse partial update
	var partialConfig map[string]interface{}
	if err := c.BodyParser(&partialConfig); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Build update with dot notation for nested fields
	update := bson.M{}
	for key, value := range partialConfig {
		update["indicator_config."+key] = value
	}

	if err := h.jobRepo.Update(ctx, id, update); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update indicator configuration",
		})
	}

	// Fetch updated job
	job, _ = h.jobRepo.FindByID(ctx, id)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Indicator configuration updated successfully",
		"job_id":  id,
		"config":  job.IndicatorConfig,
	})
}
