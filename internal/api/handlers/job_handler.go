package handlers

import (
	"context"
	"fmt"
	"math"
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
	ohlcvRepo     *repository.OHLCVRepository
	jobExecutor   *service.JobExecutor
}

// NewJobHandler creates a new job handler
func NewJobHandler(jobRepo *repository.JobRepository, connectorRepo *repository.ConnectorRepository, ohlcvRepo *repository.OHLCVRepository, jobExecutor *service.JobExecutor) *JobHandler {
	return &JobHandler{
		jobRepo:       jobRepo,
		connectorRepo: connectorRepo,
		ohlcvRepo:     ohlcvRepo,
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
		CollectHistorical:   req.CollectHistorical,
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

// CreateJobsBatch creates multiple jobs at once
// POST /api/v1/jobs/batch
func (h *JobHandler) CreateJobsBatch(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var req struct {
		Jobs []models.JobCreateRequest `json:"jobs"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(req.Jobs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No jobs provided",
		})
	}

	if len(req.Jobs) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot create more than 100 jobs at once",
		})
	}

	// Validate all jobs first
	for i, jobReq := range req.Jobs {
		if jobReq.ConnectorExchangeID == "" || jobReq.Symbol == "" || jobReq.Timeframe == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Job %d: connector_exchange_id, symbol, and timeframe are required", i+1),
			})
		}

		// Verify connector exists
		_, err := h.connectorRepo.FindByExchangeID(ctx, jobReq.ConnectorExchangeID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Job %d: Connector not found for exchange: %s", i+1, jobReq.ConnectorExchangeID),
			})
		}
	}

	// Create all jobs
	createdJobs := make([]*models.Job, 0, len(req.Jobs))
	failed := make([]string, 0)

	for i, jobReq := range req.Jobs {
		status := jobReq.Status
		if status == "" {
			status = "active"
		}

		// Calculate initial next run time with slight offset to avoid simultaneous starts
		nextRunTime := time.Now().Add(time.Duration(1+i) * time.Minute)

		job := &models.Job{
			ConnectorExchangeID: jobReq.ConnectorExchangeID,
			Symbol:              jobReq.Symbol,
			Timeframe:           jobReq.Timeframe,
			Status:              status,
			CollectHistorical:   jobReq.CollectHistorical,
			Schedule: models.Schedule{
				Mode: "timeframe",
			},
			RunState: models.RunState{
				NextRunTime: &nextRunTime,
			},
		}

		// Create in database
		if err := h.jobRepo.Create(ctx, job); err != nil {
			failed = append(failed, fmt.Sprintf("%s/%s: %v", job.Symbol, job.Timeframe, err))
			continue
		}

		createdJobs = append(createdJobs, job)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success":      true,
		"created":      len(createdJobs),
		"failed":       len(failed),
		"failed_jobs":  failed,
		"jobs":         createdJobs,
	})
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

	if req.CollectHistorical != nil {
		update["collect_historical"] = *req.CollectHistorical
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

// GetJobOHLCVData retrieves paginated OHLCV data for a job
// GET /api/v1/jobs/:id/ohlcv?page=1&limit=50
func (h *JobHandler) GetJobOHLCVData(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	id := c.Params("id")

	// Get job to verify it exists and get details
	job, err := h.jobRepo.FindByID(ctx, id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Job not found",
		})
	}

	// Get pagination parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 1000 {
		limit = 50
	}

	skip := (page - 1) * limit

	// Query OHLCV data
	filter := bson.M{
		"exchange_id": job.ConnectorExchangeID,
		"symbol":      job.Symbol,
		"timeframe":   job.Timeframe,
	}

	// Count total records
	total, err := h.ohlcvRepo.Count(ctx, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to count records",
		})
	}

	// Fetch data with pagination
	data, err := h.ohlcvRepo.FindWithPagination(ctx, filter, int64(skip), int64(limit))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch OHLCV data",
		})
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	return c.JSON(fiber.Map{
		"data": data,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// ExportJobData exports job data in CSV or JSON format
// GET /api/v1/jobs/:id/export?format=csv
func (h *JobHandler) ExportJobData(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	id := c.Params("id")
	format := c.Query("format", "csv")

	// Get job details
	job, err := h.jobRepo.FindByID(ctx, id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Job not found",
		})
	}

	// Fetch all OHLCV data for this job
	filter := bson.M{
		"exchange_id": job.ConnectorExchangeID,
		"symbol":      job.Symbol,
		"timeframe":   job.Timeframe,
	}

	data, err := h.ohlcvRepo.FindWithPagination(ctx, filter, 0, 10000) // Limit to 10000 records
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch data for export",
		})
	}

	if format == "json" {
		c.Set("Content-Type", "application/json")
		c.Set("Content-Disposition", "attachment; filename="+job.Symbol+"_"+job.Timeframe+"_export.json")
		return c.JSON(data)
	}

	// CSV format
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename="+job.Symbol+"_"+job.Timeframe+"_export.csv")

	// Build CSV
	csv := "timestamp,open,high,low,close,volume\n"
	for _, record := range data {
		timestamp := time.Unix(record.Timestamp/1000, (record.Timestamp%1000)*1000000)
		csv += timestamp.Format(time.RFC3339) + ","
		csv += formatFloat(record.Open) + ","
		csv += formatFloat(record.High) + ","
		csv += formatFloat(record.Low) + ","
		csv += formatFloat(record.Close) + ","
		csv += formatFloat(record.Volume) + "\n"
	}

	return c.SendString(csv)
}

// ExportJobDataForML exports job data optimized for machine learning
// GET /api/v1/jobs/:id/export/ml
func (h *JobHandler) ExportJobDataForML(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	id := c.Params("id")

	// Get job details
	job, err := h.jobRepo.FindByID(ctx, id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Job not found",
		})
	}

	// Fetch all OHLCV data
	filter := bson.M{
		"exchange_id": job.ConnectorExchangeID,
		"symbol":      job.Symbol,
		"timeframe":   job.Timeframe,
	}

	data, err := h.ohlcvRepo.FindWithPagination(ctx, filter, 0, 10000)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch data for ML export",
		})
	}

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename="+job.Symbol+"_"+job.Timeframe+"_ml.csv")

	// Build ML-optimized CSV with additional features
	csv := "timestamp,open,high,low,close,volume,returns,log_returns,volatility,price_change\n"

	for i, record := range data {
		timestamp := time.Unix(record.Timestamp/1000, (record.Timestamp%1000)*1000000)
		csv += timestamp.Format(time.RFC3339) + ","
		csv += formatFloat(record.Open) + ","
		csv += formatFloat(record.High) + ","
		csv += formatFloat(record.Low) + ","
		csv += formatFloat(record.Close) + ","
		csv += formatFloat(record.Volume) + ","

		// Calculate features
		if i > 0 {
			prevClose := data[i-1].Close
			returns := (record.Close - prevClose) / prevClose
			csv += formatFloat(returns) + ","
			if prevClose > 0 && record.Close > 0 {
				csv += formatFloat(math.Log(record.Close / prevClose)) + ","
			} else {
				csv += "0,"
			}
			csv += formatFloat(record.High - record.Low) + "," // volatility proxy
			csv += formatFloat(record.Close - prevClose) + "\n" // price change
		} else {
			csv += "0,0,0,0\n"
		}
	}

	return c.SendString(csv)
}

// Helper function
func formatFloat(f float64) string {
	return fmt.Sprintf("%.8f", f)
}
