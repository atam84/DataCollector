package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/yourusername/datacollector/internal/api/errors"
	"github.com/yourusername/datacollector/internal/models"
	"github.com/yourusername/datacollector/internal/repository"
	"github.com/yourusername/datacollector/internal/service"
)

// QualityHandler handles quality-related endpoints
type QualityHandler struct {
	qualityService *service.QualityService
	jobRepo        *repository.JobRepository
}

// NewQualityHandler creates a new quality handler
func NewQualityHandler(qualityService *service.QualityService, jobRepo *repository.JobRepository) *QualityHandler {
	return &QualityHandler{
		qualityService: qualityService,
		jobRepo:        jobRepo,
	}
}

// GetCachedSummary returns the cached quality summary
// GET /api/v1/quality/summary
func (h *QualityHandler) GetCachedSummary(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exchangeID := c.Query("exchange_id")

	summary, err := h.qualityService.GetCachedSummary(ctx, exchangeID)
	if err != nil {
		return errors.SendError(c, errors.InternalError("Failed to get quality summary: "+err.Error()))
	}

	if summary == nil {
		// Return empty summary if none exists
		summary = &models.QualitySummaryCache{
			ExchangeID: exchangeID,
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    summary,
	})
}

// GetCachedResults returns all cached quality results
// GET /api/v1/quality
func (h *QualityHandler) GetCachedResults(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	exchangeID := c.Query("exchange_id")
	qualityStatus := c.Query("quality_status")

	results, err := h.qualityService.GetAllCachedResults(ctx, exchangeID, qualityStatus)
	if err != nil {
		return errors.SendError(c, errors.InternalError("Failed to get quality results: "+err.Error()))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    results,
		"count":   len(results),
	})
}

// GetJobQuality returns the cached quality for a specific job
// GET /api/v1/jobs/:id/quality
func (h *QualityHandler) GetJobQuality(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	jobID := c.Params("id")

	result, err := h.qualityService.GetCachedResultByJobID(ctx, jobID)
	if err != nil {
		return errors.SendError(c, errors.InternalError("Failed to get job quality: "+err.Error()))
	}

	// If no cached result, analyze on-demand (for single job it's fast)
	if result == nil {
		job, err := h.jobRepo.FindByID(ctx, jobID)
		if err != nil {
			return errors.SendError(c, errors.NotFound("Job"))
		}

		result, err = h.qualityService.AnalyzeJob(ctx, job)
		if err != nil {
			return errors.SendError(c, errors.InternalError("Failed to analyze job quality: "+err.Error()))
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// StartQualityCheck starts a background quality check
// POST /api/v1/quality/check
func (h *QualityHandler) StartQualityCheck(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var req struct {
		Type       string `json:"type"`        // single, all, exchange
		ExchangeID string `json:"exchange_id"` // For exchange type
		Symbol     string `json:"symbol"`      // For single type
		Timeframe  string `json:"timeframe"`   // For single type
	}

	if err := c.BodyParser(&req); err != nil {
		// Default to checking all if no body
		req.Type = "all"
	}

	var checkType models.QualityCheckType
	switch req.Type {
	case "single":
		checkType = models.QualityCheckTypeSingle
	case "exchange":
		checkType = models.QualityCheckTypeExchange
	default:
		checkType = models.QualityCheckTypeAll
	}

	checkJob, err := h.qualityService.StartQualityCheck(ctx, checkType, req.ExchangeID, req.Symbol, req.Timeframe)
	if err != nil {
		return errors.SendError(c, errors.InternalError("Failed to start quality check: "+err.Error()))
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"success": true,
		"message": "Quality check started",
		"data":    checkJob,
	})
}

// GetCheckJobStatus returns the status of a quality check job
// GET /api/v1/quality/checks/:id
func (h *QualityHandler) GetCheckJobStatus(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	checkID := c.Params("id")

	checkJob, err := h.qualityService.GetCheckJobStatus(ctx, checkID)
	if err != nil {
		return errors.SendError(c, errors.InternalError("Failed to get check status: "+err.Error()))
	}

	if checkJob == nil {
		return errors.SendError(c, errors.NotFound("Quality check job"))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    checkJob,
	})
}

// GetActiveCheckJobs returns all active check jobs
// GET /api/v1/quality/checks/active
func (h *QualityHandler) GetActiveCheckJobs(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	jobs, err := h.qualityService.GetActiveCheckJobs(ctx)
	if err != nil {
		return errors.SendError(c, errors.InternalError("Failed to get active checks: "+err.Error()))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    jobs,
		"count":   len(jobs),
	})
}

// GetRecentCheckJobs returns recent check jobs
// GET /api/v1/quality/checks
func (h *QualityHandler) GetRecentCheckJobs(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	limit := c.QueryInt("limit", 20)
	if limit > 100 {
		limit = 100
	}

	jobs, err := h.qualityService.GetRecentCheckJobs(ctx, limit)
	if err != nil {
		return errors.SendError(c, errors.InternalError("Failed to get recent checks: "+err.Error()))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    jobs,
		"count":   len(jobs),
	})
}

// RefreshJobQuality refreshes the quality for a specific job
// POST /api/v1/jobs/:id/quality/refresh
func (h *QualityHandler) RefreshJobQuality(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	jobID := c.Params("id")

	job, err := h.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		return errors.SendError(c, errors.NotFound("Job"))
	}

	result, err := h.qualityService.AnalyzeJob(ctx, job)
	if err != nil {
		return errors.SendError(c, errors.InternalError("Failed to analyze job quality: "+err.Error()))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Quality refreshed",
		"data":    result,
	})
}

// FillJobGaps starts a background gap fill job
// POST /api/v1/jobs/:id/quality/fill-gaps
func (h *QualityHandler) FillJobGaps(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	jobID := c.Params("id")

	var req struct {
		FillAll bool `json:"fill_all"`
	}

	if err := c.BodyParser(&req); err != nil {
		req.FillAll = true // Default to fill all
	}

	gapFillJob, err := h.qualityService.StartGapFill(ctx, jobID, req.FillAll)
	if err != nil {
		return errors.SendError(c, errors.InternalError("Failed to start gap fill: "+err.Error()))
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"success": true,
		"message": "Gap fill started",
		"data":    gapFillJob,
	})
}

// GetGapFillStatus returns the status of a gap fill job
// GET /api/v1/jobs/:id/quality/fill-gaps/status
func (h *QualityHandler) GetGapFillStatus(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	jobID := c.Params("id")

	// First check if there's an active gap fill job
	activeJob, err := h.qualityService.GetActiveGapFillJobForJob(ctx, jobID)
	if err != nil {
		return errors.SendError(c, errors.InternalError("Failed to get gap fill status: "+err.Error()))
	}

	if activeJob != nil {
		return c.JSON(fiber.Map{
			"success": true,
			"data":    activeJob,
		})
	}

	// If no active job, get the most recent completed one
	recentJobs, err := h.qualityService.GetRecentGapFillJobsForJob(ctx, jobID, 1)
	if err != nil {
		return errors.SendError(c, errors.InternalError("Failed to get gap fill status: "+err.Error()))
	}

	if len(recentJobs) > 0 {
		return c.JSON(fiber.Map{
			"success": true,
			"data":    recentJobs[0],
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    nil,
	})
}

// GetGapFillHistory returns the gap fill history for a job
// GET /api/v1/jobs/:id/quality/fill-gaps/history
func (h *QualityHandler) GetGapFillHistory(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	jobID := c.Params("id")
	limit := c.QueryInt("limit", 10)
	if limit > 50 {
		limit = 50
	}

	jobs, err := h.qualityService.GetRecentGapFillJobsForJob(ctx, jobID, limit)
	if err != nil {
		return errors.SendError(c, errors.InternalError("Failed to get gap fill history: "+err.Error()))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    jobs,
		"count":   len(jobs),
	})
}

// StartBackfill starts a background backfill job to fetch historical data
// POST /api/v1/jobs/:id/quality/backfill
func (h *QualityHandler) StartBackfill(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	jobID := c.Params("id")

	var req struct {
		MonthsBack int    `json:"months_back"`  // How many months back to fetch
		TargetDate string `json:"target_date"`  // Specific target date (YYYY-MM-DD)
	}

	if err := c.BodyParser(&req); err != nil {
		// Default values will be used
	}

	backfillJob, err := h.qualityService.StartBackfill(ctx, jobID, req.MonthsBack, req.TargetDate)
	if err != nil {
		return errors.SendError(c, errors.InternalError("Failed to start backfill: "+err.Error()))
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"success": true,
		"message": "Backfill started",
		"data":    backfillJob,
	})
}

// GetBackfillStatus returns the status of a backfill job
// GET /api/v1/jobs/:id/quality/backfill/status
func (h *QualityHandler) GetBackfillStatus(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	jobID := c.Params("id")

	// First check if there's an active backfill job
	activeJob, err := h.qualityService.GetActiveBackfillJobForJob(ctx, jobID)
	if err != nil {
		return errors.SendError(c, errors.InternalError("Failed to get backfill status: "+err.Error()))
	}

	if activeJob != nil {
		return c.JSON(fiber.Map{
			"success": true,
			"data":    activeJob,
		})
	}

	// If no active job, get the most recent completed one
	recentJobs, err := h.qualityService.GetRecentBackfillJobsForJob(ctx, jobID, 1)
	if err != nil {
		return errors.SendError(c, errors.InternalError("Failed to get backfill status: "+err.Error()))
	}

	if len(recentJobs) > 0 {
		return c.JSON(fiber.Map{
			"success": true,
			"data":    recentJobs[0],
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    nil,
	})
}

// GetBackfillHistory returns the backfill history for a job
// GET /api/v1/jobs/:id/quality/backfill/history
func (h *QualityHandler) GetBackfillHistory(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	jobID := c.Params("id")
	limit := c.QueryInt("limit", 10)
	if limit > 50 {
		limit = 50
	}

	jobs, err := h.qualityService.GetRecentBackfillJobsForJob(ctx, jobID, limit)
	if err != nil {
		return errors.SendError(c, errors.InternalError("Failed to get backfill history: "+err.Error()))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    jobs,
		"count":   len(jobs),
	})
}
