package handlers

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/yourusername/datacollector/internal/api/errors"
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
// @Summary Create a new job
// @Description Creates a new data collection job for a specific symbol and timeframe
// @Tags Jobs
// @Accept json
// @Produce json
// @Param request body models.JobCreateRequest true "Job configuration"
// @Success 201 {object} map[string]interface{} "Job created"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Connector not found"
// @Failure 409 {object} map[string]interface{} "Job already exists"
// @Router /jobs [post]
func (h *JobHandler) CreateJob(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var req models.JobCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.SendError(c, errors.BadRequest("Invalid request body"))
	}

	// Validate required fields
	if req.ConnectorExchangeID == "" || req.Symbol == "" || req.Timeframe == "" {
		return errors.SendError(c, errors.ValidationError("Missing required fields", map[string]string{
			"connector_exchange_id": "required",
			"symbol":                "required",
			"timeframe":             "required",
		}))
	}

	// Verify connector exists
	_, err := h.connectorRepo.FindByExchangeID(ctx, req.ConnectorExchangeID)
	if err != nil {
		return errors.SendError(c, errors.NotFound("Connector"))
	}

	// Create job model
	status := req.Status
	if status == "" {
		status = "active"
	}

	// Calculate initial next run time
	nextRunTime := time.Now().Add(1 * time.Minute) // Start in 1 minute

	// Parse dependencies
	var dependsOn []primitive.ObjectID
	if len(req.DependsOn) > 0 {
		for _, depIDStr := range req.DependsOn {
			depID, err := primitive.ObjectIDFromHex(depIDStr)
			if err != nil {
				return errors.SendError(c, errors.ValidationError("Invalid dependency ID format", map[string]string{
					"dependency_id": depIDStr,
				}))
			}
			// Verify dependency job exists
			_, err = h.jobRepo.FindByID(ctx, depIDStr)
			if err != nil {
				return errors.SendError(c, errors.NotFound(fmt.Sprintf("Dependency job '%s'", depIDStr)))
			}
			dependsOn = append(dependsOn, depID)
		}
	}

	job := &models.Job{
		ConnectorExchangeID: req.ConnectorExchangeID,
		Symbol:              req.Symbol,
		Timeframe:           req.Timeframe,
		Status:              status,
		CollectHistorical:   req.CollectHistorical,
		DependsOn:           dependsOn,
		Schedule: models.Schedule{
			Mode: "timeframe",
		},
		RunState: models.RunState{
			NextRunTime: &nextRunTime,
		},
	}

	// Create in database
	if err := h.jobRepo.Create(ctx, job); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return errors.SendError(c, errors.Conflict("Job already exists for this exchange/symbol/timeframe combination"))
		}
		return errors.SendError(c, errors.DatabaseError("Failed to create job"))
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    job,
	})
}

// CreateJobsBatch creates multiple jobs at once
// @Summary Create multiple jobs
// @Description Creates multiple data collection jobs in a single request (max 100)
// @Tags Jobs
// @Accept json
// @Produce json
// @Param request body object{jobs=[]models.JobCreateRequest} true "List of job configurations"
// @Success 201 {object} map[string]interface{} "Jobs created"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Router /jobs/batch [post]
func (h *JobHandler) CreateJobsBatch(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var req struct {
		Jobs []models.JobCreateRequest `json:"jobs"`
	}

	if err := c.BodyParser(&req); err != nil {
		return errors.SendError(c, errors.BadRequest("Invalid request body"))
	}

	if len(req.Jobs) == 0 {
		return errors.SendError(c, errors.ValidationError("No jobs provided", nil))
	}

	if len(req.Jobs) > 100 {
		return errors.SendError(c, errors.ValidationError("Batch size exceeds limit", map[string]interface{}{
			"max_jobs": 100,
			"provided": len(req.Jobs),
		}))
	}

	// Validate all jobs first
	for i, jobReq := range req.Jobs {
		if jobReq.ConnectorExchangeID == "" || jobReq.Symbol == "" || jobReq.Timeframe == "" {
			return errors.SendError(c, errors.ValidationError(
				fmt.Sprintf("Job %d: missing required fields", i+1),
				map[string]string{
					"connector_exchange_id": "required",
					"symbol":                "required",
					"timeframe":             "required",
				},
			))
		}

		// Verify connector exists
		_, err := h.connectorRepo.FindByExchangeID(ctx, jobReq.ConnectorExchangeID)
		if err != nil {
			return errors.SendError(c, errors.NotFound(fmt.Sprintf("Job %d: Connector '%s'", i+1, jobReq.ConnectorExchangeID)))
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
// @Summary Get all jobs
// @Description Retrieves all data collection jobs with optional filtering
// @Tags Jobs
// @Accept json
// @Produce json
// @Param status query string false "Filter by status (active, paused, error)"
// @Param exchange_id query string false "Filter by exchange ID"
// @Param symbol query string false "Filter by symbol"
// @Param timeframe query string false "Filter by timeframe"
// @Success 200 {object} map[string]interface{} "List of jobs"
// @Router /jobs [get]
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
		return errors.SendError(c, errors.DatabaseError("Failed to retrieve jobs"))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    jobs,
		"count":   len(jobs),
	})
}

// GetJob retrieves a job by ID
// @Summary Get a job by ID
// @Description Retrieves a specific data collection job by its ID
// @Tags Jobs
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} map[string]interface{} "Job details"
// @Failure 400 {object} map[string]interface{} "Invalid job ID"
// @Failure 404 {object} map[string]interface{} "Job not found"
// @Router /jobs/{id} [get]
func (h *JobHandler) GetJob(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	job, err := h.jobRepo.FindByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid job ID format"))
		}
		return errors.SendError(c, errors.NotFound("Job"))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    job,
	})
}

// GetJobsByConnector retrieves all jobs for a connector
// GET /api/v1/connectors/:exchangeId/jobs
func (h *JobHandler) GetJobsByConnector(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exchangeID := c.Params("exchangeId")

	jobs, err := h.jobRepo.FindByConnector(ctx, exchangeID)
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to retrieve jobs for connector"))
	}

	return c.JSON(fiber.Map{
		"success":     true,
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
		return errors.SendError(c, errors.BadRequest("Invalid request body"))
	}

	// Build update map
	update := bson.M{}

	if req.Status != nil {
		validStatuses := map[string]bool{"active": true, "paused": true, "error": true}
		if !validStatuses[*req.Status] {
			return errors.SendError(c, errors.ValidationError("Invalid status value", map[string]interface{}{
				"status":         *req.Status,
				"allowed_values": []string{"active", "paused", "error"},
			}))
		}
		update["status"] = *req.Status
	}

	if req.Timeframe != nil {
		update["timeframe"] = *req.Timeframe
	}

	if req.CollectHistorical != nil {
		update["collect_historical"] = *req.CollectHistorical
	}

	// Handle dependencies update
	if req.DependsOn != nil {
		var dependsOn []primitive.ObjectID
		for _, depIDStr := range *req.DependsOn {
			depID, err := primitive.ObjectIDFromHex(depIDStr)
			if err != nil {
				return errors.SendError(c, errors.ValidationError("Invalid dependency ID format", map[string]string{
					"dependency_id": depIDStr,
				}))
			}
			// Verify dependency job exists
			_, err = h.jobRepo.FindByID(ctx, depIDStr)
			if err != nil {
				return errors.SendError(c, errors.NotFound(fmt.Sprintf("Dependency job '%s'", depIDStr)))
			}
			// Check for circular dependency
			hasCycle, err := h.jobRepo.CheckCircularDependency(ctx, id, depIDStr)
			if err != nil {
				return errors.SendError(c, errors.DatabaseError("Failed to check circular dependency"))
			}
			if hasCycle {
				return errors.SendError(c, errors.ValidationError("Circular dependency detected", map[string]string{
					"dependency_id": depIDStr,
					"error":         "Adding this dependency would create a cycle",
				}))
			}
			dependsOn = append(dependsOn, depID)
		}
		update["depends_on"] = dependsOn
	}

	if len(update) == 0 {
		return errors.SendError(c, errors.BadRequest("No fields to update"))
	}

	// Update job
	if err := h.jobRepo.Update(ctx, id, update); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return errors.SendError(c, errors.NotFound("Job"))
		}
		return errors.SendError(c, errors.DatabaseError("Failed to update job"))
	}

	// Fetch updated job
	job, err := h.jobRepo.FindByID(ctx, id)
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to retrieve updated job"))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    job,
	})
}

// PauseJob pauses a job
// POST /api/v1/jobs/:id/pause
func (h *JobHandler) PauseJob(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	if err := h.jobRepo.UpdateStatus(ctx, id, "paused"); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return errors.SendError(c, errors.NotFound("Job"))
		}
		return errors.SendError(c, errors.DatabaseError("Failed to pause job"))
	}

	job, err := h.jobRepo.FindByID(ctx, id)
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to retrieve job"))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Job paused successfully",
		"data":    job,
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
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid job ID format"))
		}
		return errors.SendError(c, errors.NotFound("Job"))
	}

	// Check if connector is active
	connector, err := h.connectorRepo.FindByExchangeID(ctx, job.ConnectorExchangeID)
	if err != nil {
		return errors.SendError(c, errors.NotFound("Connector"))
	}

	if connector.Status != "active" {
		return errors.SendError(c, errors.ConnectorInactive(connector.ExchangeID).WithDetails(map[string]string{
			"action": "Please resume the connector first before resuming this job",
		}))
	}

	// Resume the job
	if err := h.jobRepo.UpdateStatus(ctx, id, "active"); err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to resume job"))
	}

	job, err = h.jobRepo.FindByID(ctx, id)
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to retrieve job"))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Job resumed successfully",
		"data":    job,
	})
}

// DeleteJob deletes a job
// DELETE /api/v1/jobs/:id
func (h *JobHandler) DeleteJob(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	if err := h.jobRepo.Delete(ctx, id); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid job ID format"))
		}
		if strings.Contains(err.Error(), "not found") {
			return errors.SendError(c, errors.NotFound("Job"))
		}
		return errors.SendError(c, errors.DatabaseError("Failed to delete job"))
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ExecuteJob executes a job manually
// @Summary Execute a job manually
// @Description Triggers immediate execution of a data collection job
// @Tags Jobs
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} map[string]interface{} "Execution result"
// @Failure 404 {object} map[string]interface{} "Job not found"
// @Failure 409 {object} map[string]interface{} "Job is locked"
// @Failure 500 {object} map[string]interface{} "Execution failed"
// @Router /jobs/{id}/execute [post]
func (h *JobHandler) ExecuteJob(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	id := c.Params("id")

	// Execute the job
	result, err := h.jobExecutor.ExecuteJob(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return errors.SendError(c, errors.NotFound("Job"))
		}
		if strings.Contains(err.Error(), "locked") {
			return errors.SendError(c, errors.JobLocked(id))
		}
		return errors.SendError(c, errors.ExchangeError("Job execution failed: "+err.Error()))
	}

	// Return result with success flag
	return c.JSON(fiber.Map{
		"success": result.Success,
		"data":    result,
	})
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
		return errors.SendError(c, errors.DatabaseError("Failed to retrieve job queue"))
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
		"success": true,
		"data":    queuedJobs,
		"count":   len(queuedJobs),
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
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid job ID format"))
		}
		return errors.SendError(c, errors.NotFound("Job"))
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
		return errors.SendError(c, errors.DatabaseError("Failed to count OHLCV records"))
	}

	// Fetch data with pagination
	data, err := h.ohlcvRepo.FindWithPagination(ctx, filter, int64(skip), int64(limit))
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to fetch OHLCV data"))
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	return c.JSON(fiber.Map{
		"success": true,
		"data":    data,
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

	// Validate format
	if format != "csv" && format != "json" {
		return errors.SendError(c, errors.ValidationError("Invalid export format", map[string]interface{}{
			"format":         format,
			"allowed_values": []string{"csv", "json"},
		}))
	}

	// Get job details
	job, err := h.jobRepo.FindByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid job ID format"))
		}
		return errors.SendError(c, errors.NotFound("Job"))
	}

	// Fetch all OHLCV data for this job
	filter := bson.M{
		"exchange_id": job.ConnectorExchangeID,
		"symbol":      job.Symbol,
		"timeframe":   job.Timeframe,
	}

	data, err := h.ohlcvRepo.FindWithPagination(ctx, filter, 0, 10000) // Limit to 10000 records
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to fetch data for export"))
	}

	if len(data) == 0 {
		return errors.SendError(c, errors.NoData(fmt.Sprintf("%s/%s", job.Symbol, job.Timeframe)))
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
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid job ID format"))
		}
		return errors.SendError(c, errors.NotFound("Job"))
	}

	// Fetch all OHLCV data
	filter := bson.M{
		"exchange_id": job.ConnectorExchangeID,
		"symbol":      job.Symbol,
		"timeframe":   job.Timeframe,
	}

	data, err := h.ohlcvRepo.FindWithPagination(ctx, filter, 0, 10000)
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to fetch data for ML export"))
	}

	if len(data) == 0 {
		return errors.SendError(c, errors.NoData(fmt.Sprintf("%s/%s", job.Symbol, job.Timeframe)))
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

// GetJobDependencies retrieves the dependencies for a job
// GET /api/v1/jobs/:id/dependencies
func (h *JobHandler) GetJobDependencies(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	// Get dependency status
	status, err := h.jobRepo.GetDependencyStatus(ctx, id, 1*time.Hour)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid job ID format"))
		}
		if strings.Contains(err.Error(), "not found") {
			return errors.SendError(c, errors.NotFound("Job"))
		}
		return errors.SendError(c, errors.DatabaseError("Failed to get dependency status"))
	}

	// Get detailed info about each dependency
	job, _ := h.jobRepo.FindByID(ctx, id)
	var dependencyJobs []*models.Job
	if len(job.DependsOn) > 0 {
		dependencyJobs, _ = h.jobRepo.FindByIDs(ctx, job.DependsOn)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"status":           status,
			"dependency_jobs":  dependencyJobs,
			"dependency_count": len(job.DependsOn),
		},
	})
}

// SetJobDependencies sets the dependencies for a job
// PUT /api/v1/jobs/:id/dependencies
func (h *JobHandler) SetJobDependencies(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	var req struct {
		DependsOn []string `json:"depends_on"`
	}

	if err := c.BodyParser(&req); err != nil {
		return errors.SendError(c, errors.BadRequest("Invalid request body"))
	}

	// Parse and validate dependencies
	var dependsOn []primitive.ObjectID
	for _, depIDStr := range req.DependsOn {
		depID, err := primitive.ObjectIDFromHex(depIDStr)
		if err != nil {
			return errors.SendError(c, errors.ValidationError("Invalid dependency ID format", map[string]string{
				"dependency_id": depIDStr,
			}))
		}

		// Verify dependency job exists
		_, err = h.jobRepo.FindByID(ctx, depIDStr)
		if err != nil {
			return errors.SendError(c, errors.NotFound(fmt.Sprintf("Dependency job '%s'", depIDStr)))
		}

		// Check for circular dependency
		hasCycle, err := h.jobRepo.CheckCircularDependency(ctx, id, depIDStr)
		if err != nil {
			return errors.SendError(c, errors.DatabaseError("Failed to check circular dependency"))
		}
		if hasCycle {
			return errors.SendError(c, errors.ValidationError("Circular dependency detected", map[string]string{
				"dependency_id": depIDStr,
				"error":         "Adding this dependency would create a cycle",
			}))
		}

		dependsOn = append(dependsOn, depID)
	}

	// Set dependencies
	if err := h.jobRepo.SetDependencies(ctx, id, dependsOn); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid job ID format"))
		}
		if strings.Contains(err.Error(), "not found") {
			return errors.SendError(c, errors.NotFound("Job"))
		}
		return errors.SendError(c, errors.DatabaseError("Failed to set dependencies"))
	}

	// Get updated job
	job, err := h.jobRepo.FindByID(ctx, id)
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to retrieve job"))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("Set %d dependencies", len(dependsOn)),
		"data":    job,
	})
}

// GetJobDependents retrieves jobs that depend on a given job
// GET /api/v1/jobs/:id/dependents
func (h *JobHandler) GetJobDependents(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")

	// Verify job exists
	_, err := h.jobRepo.FindByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return errors.SendError(c, errors.BadRequest("Invalid job ID format"))
		}
		return errors.SendError(c, errors.NotFound("Job"))
	}

	// Find jobs that depend on this job
	dependents, err := h.jobRepo.FindJobsDependingOn(ctx, id)
	if err != nil {
		return errors.SendError(c, errors.DatabaseError("Failed to find dependent jobs"))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    dependents,
		"count":   len(dependents),
	})
}

// Helper function
func formatFloat(f float64) string {
	return fmt.Sprintf("%.8f", f)
}
