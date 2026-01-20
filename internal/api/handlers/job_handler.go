package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/yourusername/datacollector/internal/models"
	"github.com/yourusername/datacollector/internal/repository"
)

// JobHandler handles job-related endpoints
type JobHandler struct {
	jobRepo       *repository.JobRepository
	connectorRepo *repository.ConnectorRepository
}

// NewJobHandler creates a new job handler
func NewJobHandler(jobRepo *repository.JobRepository, connectorRepo *repository.ConnectorRepository) *JobHandler {
	return &JobHandler{
		jobRepo:       jobRepo,
		connectorRepo: connectorRepo,
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

	job := &models.Job{
		ConnectorExchangeID: req.ConnectorExchangeID,
		Symbol:              req.Symbol,
		Timeframe:           req.Timeframe,
		Status:              status,
		Schedule: models.Schedule{
			Mode: "timeframe",
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

	if err := h.jobRepo.UpdateStatus(ctx, id, "active"); err != nil {
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
