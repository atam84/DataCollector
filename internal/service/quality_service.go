package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/yourusername/datacollector/internal/models"
	"github.com/yourusername/datacollector/internal/repository"
)

// QualityService handles data quality analysis
type QualityService struct {
	qualityRepo   *repository.QualityRepository
	ohlcvRepo     *repository.OHLCVRepository
	jobRepo       *repository.JobRepository
	ccxtService   *CCXTService
	connectorRepo *repository.ConnectorRepository
	rateLimiter   *RateLimiter

	// Background processing
	mu            sync.Mutex
	runningChecks map[string]bool
}

// NewQualityService creates a new quality service
func NewQualityService(
	qualityRepo *repository.QualityRepository,
	ohlcvRepo *repository.OHLCVRepository,
	jobRepo *repository.JobRepository,
	ccxtService *CCXTService,
	connectorRepo *repository.ConnectorRepository,
	rateLimiter *RateLimiter,
) *QualityService {
	return &QualityService{
		qualityRepo:   qualityRepo,
		ohlcvRepo:     ohlcvRepo,
		jobRepo:       jobRepo,
		ccxtService:   ccxtService,
		connectorRepo: connectorRepo,
		rateLimiter:   rateLimiter,
		runningChecks: make(map[string]bool),
	}
}

// AnalyzeJob analyzes data quality for a single job and stores the result
func (s *QualityService) AnalyzeJob(ctx context.Context, job *models.Job) (*models.DataQualityResult, error) {
	// Get raw quality analysis from OHLCV repository
	quality, err := s.ohlcvRepo.AnalyzeDataQuality(ctx, job.ConnectorExchangeID, job.Symbol, job.Timeframe)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze job quality: %w", err)
	}

	// Convert to result with additional fields
	result := &models.DataQualityResult{
		ExchangeID:        job.ConnectorExchangeID,
		Symbol:            job.Symbol,
		Timeframe:         job.Timeframe,
		JobID:             job.ID,
		QualityStatus:     quality.QualityStatus,
		CompletenessScore: quality.CompletenessScore,
		TotalCandles:      quality.TotalCandles,
		ExpectedCandles:   quality.ExpectedCandles,
		MissingCandles:    quality.MissingCandles,
		GapsDetected:      quality.GapsDetected,
		Gaps:              quality.Gaps,
		DataPeriodStart:   quality.OldestCandle,
		DataPeriodEnd:     quality.NewestCandle,
		DataFreshness:     quality.DataFreshness,
		FreshnessMinutes:  quality.FreshnessMinutes,
		CheckedAt:         time.Now(),
	}

	// Calculate data period days
	if !result.DataPeriodStart.IsZero() && !result.DataPeriodEnd.IsZero() {
		result.DataPeriodDays = int(result.DataPeriodEnd.Sub(result.DataPeriodStart).Hours() / 24)
		result.DataAgeDays = int(time.Since(result.DataPeriodEnd).Hours() / 24)
	}

	// Store the result
	if err := s.qualityRepo.UpsertResult(ctx, result); err != nil {
		log.Printf("[QUALITY] Warning: Failed to store quality result: %v", err)
	}

	return result, nil
}

// GetCachedResult gets the cached quality result for a job
func (s *QualityService) GetCachedResult(ctx context.Context, exchangeID, symbol, timeframe string) (*models.DataQualityResult, error) {
	return s.qualityRepo.FindResult(ctx, exchangeID, symbol, timeframe)
}

// GetCachedResultByJobID gets the cached quality result by job ID
func (s *QualityService) GetCachedResultByJobID(ctx context.Context, jobID string) (*models.DataQualityResult, error) {
	// First try to find by job ID
	result, err := s.qualityRepo.FindResultByJobID(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if result != nil {
		return result, nil
	}

	// If not found, get the job and look up by exchange/symbol/timeframe
	job, err := s.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, fmt.Errorf("job not found")
	}

	return s.qualityRepo.FindResult(ctx, job.ConnectorExchangeID, job.Symbol, job.Timeframe)
}

// GetAllCachedResults gets all cached quality results
func (s *QualityService) GetAllCachedResults(ctx context.Context, exchangeID, qualityStatus string) ([]*models.DataQualityResult, error) {
	filter := bson.M{}
	if exchangeID != "" {
		filter["exchange_id"] = exchangeID
	}
	if qualityStatus != "" {
		filter["quality_status"] = qualityStatus
	}

	return s.qualityRepo.FindAllResults(ctx, filter)
}

// GetCachedSummary gets the cached quality summary
func (s *QualityService) GetCachedSummary(ctx context.Context, exchangeID string) (*models.QualitySummaryCache, error) {
	summary, err := s.qualityRepo.FindSummary(ctx, exchangeID)
	if err != nil {
		return nil, err
	}

	// If no cached summary, compute from results
	if summary == nil {
		return s.qualityRepo.ComputeSummaryFromResults(ctx, exchangeID)
	}

	return summary, nil
}

// StartQualityCheck starts a background quality check job
func (s *QualityService) StartQualityCheck(ctx context.Context, checkType models.QualityCheckType, exchangeID, symbol, timeframe string) (*models.QualityCheckJob, error) {
	// Create check job
	checkJob := &models.QualityCheckJob{
		Type:       checkType,
		Status:     models.QualityCheckPending,
		ExchangeID: exchangeID,
		Symbol:     symbol,
		Timeframe:  timeframe,
	}

	// Get jobs to check based on type
	filter := bson.M{}
	if exchangeID != "" {
		filter["connector_exchange_id"] = exchangeID
	}
	if symbol != "" {
		filter["symbol"] = symbol
	}
	if timeframe != "" {
		filter["timeframe"] = timeframe
	}

	jobs, err := s.jobRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find jobs: %w", err)
	}

	checkJob.TotalJobs = len(jobs)
	for _, job := range jobs {
		checkJob.JobIDs = append(checkJob.JobIDs, job.ID)
	}

	// Save the check job
	if err := s.qualityRepo.CreateCheckJob(ctx, checkJob); err != nil {
		return nil, fmt.Errorf("failed to create check job: %w", err)
	}

	// Start background processing
	go s.processQualityCheck(checkJob.ID.Hex())

	return checkJob, nil
}

// processQualityCheck processes a quality check job in the background
func (s *QualityService) processQualityCheck(checkJobID string) {
	// Prevent duplicate processing
	s.mu.Lock()
	if s.runningChecks[checkJobID] {
		s.mu.Unlock()
		return
	}
	s.runningChecks[checkJobID] = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.runningChecks, checkJobID)
		s.mu.Unlock()
	}()

	ctx := context.Background()

	// Get check job
	checkJob, err := s.qualityRepo.FindCheckJob(ctx, checkJobID)
	if err != nil || checkJob == nil {
		log.Printf("[QUALITY] Failed to find check job %s: %v", checkJobID, err)
		return
	}

	// Update status to running
	now := time.Now()
	checkJob.Status = models.QualityCheckRunning
	checkJob.StartedAt = &now
	if err := s.qualityRepo.UpdateCheckJob(ctx, checkJob); err != nil {
		log.Printf("[QUALITY] Failed to update check job status: %v", err)
	}

	log.Printf("[QUALITY] Starting quality check %s for %d jobs", checkJobID, checkJob.TotalJobs)

	// Process each job
	for i, jobID := range checkJob.JobIDs {
		job, err := s.jobRepo.FindByID(ctx, jobID.Hex())
		if err != nil || job == nil {
			checkJob.FailedJobs++
			checkJob.Errors = append(checkJob.Errors, fmt.Sprintf("Job %s not found", jobID.Hex()))
			continue
		}

		// Analyze the job
		result, err := s.AnalyzeJob(ctx, job)
		if err != nil {
			checkJob.FailedJobs++
			checkJob.LastError = err.Error()
			checkJob.Errors = append(checkJob.Errors, fmt.Sprintf("Job %s: %v", jobID.Hex(), err))
			log.Printf("[QUALITY] Failed to analyze job %s: %v", jobID.Hex(), err)
		} else {
			checkJob.CompletedJobs++

			// Update counts
			switch result.QualityStatus {
			case "excellent":
				checkJob.ExcellentCount++
			case "good":
				checkJob.GoodCount++
			case "fair":
				checkJob.FairCount++
			case "poor":
				checkJob.PoorCount++
			}
		}

		// Update progress
		checkJob.Progress = float64(i+1) / float64(checkJob.TotalJobs) * 100

		// Update check job periodically (every 10 jobs or at end)
		if (i+1)%10 == 0 || i == len(checkJob.JobIDs)-1 {
			if err := s.qualityRepo.UpdateCheckJob(ctx, checkJob); err != nil {
				log.Printf("[QUALITY] Failed to update check job progress: %v", err)
			}
		}
	}

	// Mark as completed
	completedAt := time.Now()
	checkJob.Status = models.QualityCheckCompleted
	checkJob.CompletedAt = &completedAt
	checkJob.Progress = 100

	if err := s.qualityRepo.UpdateCheckJob(ctx, checkJob); err != nil {
		log.Printf("[QUALITY] Failed to mark check job as completed: %v", err)
	}

	// Update summary cache
	s.updateSummaryCache(ctx, checkJob.ExchangeID)

	log.Printf("[QUALITY] Completed quality check %s: %d/%d jobs analyzed",
		checkJobID, checkJob.CompletedJobs, checkJob.TotalJobs)
}

// updateSummaryCache updates the cached summary
func (s *QualityService) updateSummaryCache(ctx context.Context, exchangeID string) {
	// Update global summary
	globalSummary, err := s.qualityRepo.ComputeSummaryFromResults(ctx, "")
	if err != nil {
		log.Printf("[QUALITY] Failed to compute global summary: %v", err)
	} else {
		if err := s.qualityRepo.UpsertSummary(ctx, globalSummary); err != nil {
			log.Printf("[QUALITY] Failed to save global summary: %v", err)
		}
	}

	// Update exchange-specific summary if applicable
	if exchangeID != "" {
		exchSummary, err := s.qualityRepo.ComputeSummaryFromResults(ctx, exchangeID)
		if err != nil {
			log.Printf("[QUALITY] Failed to compute exchange summary: %v", err)
		} else {
			if err := s.qualityRepo.UpsertSummary(ctx, exchSummary); err != nil {
				log.Printf("[QUALITY] Failed to save exchange summary: %v", err)
			}
		}
	}
}

// GetCheckJobStatus gets the status of a quality check job
func (s *QualityService) GetCheckJobStatus(ctx context.Context, checkJobID string) (*models.QualityCheckJob, error) {
	return s.qualityRepo.FindCheckJob(ctx, checkJobID)
}

// GetActiveCheckJobs gets all active check jobs
func (s *QualityService) GetActiveCheckJobs(ctx context.Context) ([]*models.QualityCheckJob, error) {
	return s.qualityRepo.FindActiveCheckJobs(ctx)
}

// GetRecentCheckJobs gets recent check jobs
func (s *QualityService) GetRecentCheckJobs(ctx context.Context, limit int) ([]*models.QualityCheckJob, error) {
	return s.qualityRepo.FindRecentCheckJobs(ctx, limit)
}

// StartGapFill starts a background gap fill job
func (s *QualityService) StartGapFill(ctx context.Context, jobID string, fillAll bool) (*models.GapFillJob, error) {
	job, err := s.jobRepo.FindByID(ctx, jobID)
	if err != nil || job == nil {
		return nil, fmt.Errorf("job not found")
	}

	// Check if there's already an active gap fill job for this job
	existingJob, err := s.qualityRepo.FindActiveGapFillJobForJob(ctx, job.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing gap fill job: %w", err)
	}
	if existingJob != nil {
		return existingJob, nil // Return existing job
	}

	// Get current quality to find gaps
	result, err := s.GetCachedResult(ctx, job.ConnectorExchangeID, job.Symbol, job.Timeframe)
	if err != nil || result == nil {
		// Analyze first if no cached result
		result, err = s.AnalyzeJob(ctx, job)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze job: %w", err)
		}
	}

	// Determine number of gaps to fill
	totalGaps := len(result.Gaps)
	if !fillAll && totalGaps > 5 {
		totalGaps = 5
	}

	// Create gap fill job
	gapFillJob := &models.GapFillJob{
		JobID:      job.ID,
		ExchangeID: job.ConnectorExchangeID,
		Symbol:     job.Symbol,
		Timeframe:  job.Timeframe,
		Status:     models.GapFillPending,
		FillAll:    fillAll,
		TotalGaps:  totalGaps,
	}

	// Save the job
	if err := s.qualityRepo.CreateGapFillJob(ctx, gapFillJob); err != nil {
		return nil, fmt.Errorf("failed to create gap fill job: %w", err)
	}

	// Start background processing
	go s.processGapFill(gapFillJob.ID.Hex())

	return gapFillJob, nil
}

// processGapFill processes a gap fill job in the background
func (s *QualityService) processGapFill(gapFillJobID string) {
	// Prevent duplicate processing
	s.mu.Lock()
	if s.runningChecks[gapFillJobID] {
		s.mu.Unlock()
		return
	}
	s.runningChecks[gapFillJobID] = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.runningChecks, gapFillJobID)
		s.mu.Unlock()
	}()

	ctx := context.Background()

	// Get gap fill job
	gapFillJob, err := s.qualityRepo.FindGapFillJob(ctx, gapFillJobID)
	if err != nil || gapFillJob == nil {
		log.Printf("[GAP_FILL] Failed to find gap fill job %s: %v", gapFillJobID, err)
		return
	}

	// Update status to running
	now := time.Now()
	gapFillJob.Status = models.GapFillRunning
	gapFillJob.StartedAt = &now
	if err := s.qualityRepo.UpdateGapFillJob(ctx, gapFillJob); err != nil {
		log.Printf("[GAP_FILL] Failed to update gap fill job status: %v", err)
	}

	log.Printf("[GAP_FILL] Starting gap fill %s for %s %s %s", gapFillJobID, gapFillJob.ExchangeID, gapFillJob.Symbol, gapFillJob.Timeframe)

	// Get job and connector
	job, err := s.jobRepo.FindByID(ctx, gapFillJob.JobID.Hex())
	if err != nil || job == nil {
		gapFillJob.Status = models.GapFillFailed
		gapFillJob.LastError = "Job not found"
		completedAt := time.Now()
		gapFillJob.CompletedAt = &completedAt
		s.qualityRepo.UpdateGapFillJob(ctx, gapFillJob)
		return
	}

	connector, err := s.connectorRepo.FindByExchangeID(ctx, job.ConnectorExchangeID)
	if err != nil || connector == nil {
		gapFillJob.Status = models.GapFillFailed
		gapFillJob.LastError = "Connector not found"
		completedAt := time.Now()
		gapFillJob.CompletedAt = &completedAt
		s.qualityRepo.UpdateGapFillJob(ctx, gapFillJob)
		return
	}

	// Get current quality to find gaps
	result, err := s.GetCachedResult(ctx, job.ConnectorExchangeID, job.Symbol, job.Timeframe)
	if err != nil || result == nil {
		result, err = s.AnalyzeJob(ctx, job)
		if err != nil {
			gapFillJob.Status = models.GapFillFailed
			gapFillJob.LastError = fmt.Sprintf("Failed to analyze job: %v", err)
			completedAt := time.Now()
			gapFillJob.CompletedAt = &completedAt
			s.qualityRepo.UpdateGapFillJob(ctx, gapFillJob)
			return
		}
	}

	if len(result.Gaps) == 0 {
		gapFillJob.Status = models.GapFillCompleted
		gapFillJob.Progress = 100
		completedAt := time.Now()
		gapFillJob.CompletedAt = &completedAt
		s.qualityRepo.UpdateGapFillJob(ctx, gapFillJob)
		log.Printf("[GAP_FILL] No gaps to fill for %s", gapFillJobID)
		return
	}

	// Determine which gaps to fill
	var gapsToFill []models.DataGap
	if gapFillJob.FillAll {
		gapsToFill = result.Gaps
	} else {
		limit := 5
		if len(result.Gaps) < limit {
			limit = len(result.Gaps)
		}
		gapsToFill = result.Gaps[:limit]
	}

	gapFillJob.TotalGaps = len(gapsToFill)

	// Fill each gap
	for i, gap := range gapsToFill {
		gapFillJob.GapsAttempted++

		// Fetch data for the gap period
		candles, err := s.ccxtService.FetchOHLCVRange(
			ctx,
			connector,
			job.Symbol,
			job.Timeframe,
			gap.StartTime.UnixMilli(),
			gap.EndTime.UnixMilli(),
		)

		if err != nil {
			gapFillJob.LastError = fmt.Sprintf("Gap %s-%s: %v",
				gap.StartTime.Format(time.RFC3339), gap.EndTime.Format(time.RFC3339), err)
			gapFillJob.Errors = append(gapFillJob.Errors, gapFillJob.LastError)
		} else if len(candles) > 0 {
			// Store the fetched candles
			_, err = s.ohlcvRepo.UpsertCandles(ctx, job.ConnectorExchangeID, job.Symbol, job.Timeframe, candles)
			if err != nil {
				gapFillJob.LastError = fmt.Sprintf("Failed to store candles: %v", err)
				gapFillJob.Errors = append(gapFillJob.Errors, gapFillJob.LastError)
			} else {
				gapFillJob.CandlesFetched += len(candles)
				gapFillJob.GapsFilled++
			}
		}

		// Update progress
		gapFillJob.Progress = float64(i+1) / float64(len(gapsToFill)) * 100

		// Update job periodically (every gap or at end)
		if err := s.qualityRepo.UpdateGapFillJob(ctx, gapFillJob); err != nil {
			log.Printf("[GAP_FILL] Failed to update gap fill job progress: %v", err)
		}

		// Small delay between gaps to avoid rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	// Mark as completed
	completedAt := time.Now()
	gapFillJob.Status = models.GapFillCompleted
	gapFillJob.CompletedAt = &completedAt
	gapFillJob.Progress = 100

	if err := s.qualityRepo.UpdateGapFillJob(ctx, gapFillJob); err != nil {
		log.Printf("[GAP_FILL] Failed to mark gap fill job as completed: %v", err)
	}

	// Re-analyze quality after filling
	_, _ = s.AnalyzeJob(ctx, job)

	log.Printf("[GAP_FILL] Completed gap fill %s: %d/%d gaps filled, %d candles fetched",
		gapFillJobID, gapFillJob.GapsFilled, gapFillJob.GapsAttempted, gapFillJob.CandlesFetched)
}

// GetGapFillJobStatus gets the status of a gap fill job
func (s *QualityService) GetGapFillJobStatus(ctx context.Context, gapFillJobID string) (*models.GapFillJob, error) {
	return s.qualityRepo.FindGapFillJob(ctx, gapFillJobID)
}

// GetActiveGapFillJobs gets all active gap fill jobs
func (s *QualityService) GetActiveGapFillJobs(ctx context.Context) ([]*models.GapFillJob, error) {
	return s.qualityRepo.FindActiveGapFillJobs(ctx)
}

// GetActiveGapFillJobForJob gets the active gap fill job for a specific data job
func (s *QualityService) GetActiveGapFillJobForJob(ctx context.Context, jobID string) (*models.GapFillJob, error) {
	job, err := s.jobRepo.FindByID(ctx, jobID)
	if err != nil || job == nil {
		return nil, fmt.Errorf("job not found")
	}
	return s.qualityRepo.FindActiveGapFillJobForJob(ctx, job.ID)
}

// GetRecentGapFillJobsForJob gets recent gap fill jobs for a specific data job
func (s *QualityService) GetRecentGapFillJobsForJob(ctx context.Context, jobID string, limit int) ([]*models.GapFillJob, error) {
	job, err := s.jobRepo.FindByID(ctx, jobID)
	if err != nil || job == nil {
		return nil, fmt.Errorf("job not found")
	}
	return s.qualityRepo.FindRecentGapFillJobsForJob(ctx, job.ID, limit)
}

// FillGaps attempts to fill gaps in a job's data (DEPRECATED - use StartGapFill for background processing)
func (s *QualityService) FillGaps(ctx context.Context, jobID string, fillAll bool, startTime, endTime time.Time) (*models.GapFillResult, error) {
	job, err := s.jobRepo.FindByID(ctx, jobID)
	if err != nil || job == nil {
		return nil, fmt.Errorf("job not found")
	}

	connector, err := s.connectorRepo.FindByExchangeID(ctx, job.ConnectorExchangeID)
	if err != nil || connector == nil {
		return nil, fmt.Errorf("connector not found")
	}

	// Get current quality to find gaps
	result, err := s.GetCachedResult(ctx, job.ConnectorExchangeID, job.Symbol, job.Timeframe)
	if err != nil || result == nil {
		// Analyze first if no cached result
		result, err = s.AnalyzeJob(ctx, job)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze job: %w", err)
		}
	}

	fillResult := &models.GapFillResult{
		JobID:     jobID,
		StartedAt: time.Now(),
	}

	if len(result.Gaps) == 0 {
		fillResult.CompletedAt = time.Now()
		return fillResult, nil
	}

	// Determine which gaps to fill
	var gapsToFill []models.DataGap
	if fillAll {
		gapsToFill = result.Gaps
	} else if !startTime.IsZero() && !endTime.IsZero() {
		// Find gaps within the specified time range
		for _, gap := range result.Gaps {
			if gap.StartTime.After(startTime) && gap.EndTime.Before(endTime) {
				gapsToFill = append(gapsToFill, gap)
			}
		}
	} else {
		// Default: fill first 5 gaps
		limit := 5
		if len(result.Gaps) < limit {
			limit = len(result.Gaps)
		}
		gapsToFill = result.Gaps[:limit]
	}

	fillResult.GapsAttempted = len(gapsToFill)

	// Fill each gap
	for _, gap := range gapsToFill {
		// Fetch data for the gap period
		candles, err := s.ccxtService.FetchOHLCVRange(
			ctx,
			connector,
			job.Symbol,
			job.Timeframe,
			gap.StartTime.UnixMilli(),
			gap.EndTime.UnixMilli(),
		)

		if err != nil {
			fillResult.Errors = append(fillResult.Errors, fmt.Sprintf("Gap %s-%s: %v",
				gap.StartTime.Format(time.RFC3339), gap.EndTime.Format(time.RFC3339), err))
			continue
		}

		if len(candles) > 0 {
			// Store the fetched candles
			_, err = s.ohlcvRepo.UpsertCandles(ctx, job.ConnectorExchangeID, job.Symbol, job.Timeframe, candles)
			if err != nil {
				fillResult.Errors = append(fillResult.Errors, fmt.Sprintf("Failed to store candles: %v", err))
				continue
			}

			fillResult.CandlesFetched += len(candles)
			fillResult.GapsFilled++
		}
	}

	fillResult.CompletedAt = time.Now()

	// Re-analyze quality after filling
	_, _ = s.AnalyzeJob(ctx, job)

	return fillResult, nil
}

// RunScheduledCheck runs a scheduled quality check for all jobs
func (s *QualityService) RunScheduledCheck(ctx context.Context) error {
	log.Println("[QUALITY] Starting scheduled quality check")

	_, err := s.StartQualityCheck(ctx, models.QualityCheckTypeScheduled, "", "", "")
	if err != nil {
		return fmt.Errorf("failed to start scheduled check: %w", err)
	}

	return nil
}
