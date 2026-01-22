package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/yourusername/datacollector/internal/config"
	"github.com/yourusername/datacollector/internal/models"
	"github.com/yourusername/datacollector/internal/repository"
	"github.com/yourusername/datacollector/internal/service/indicators"
)

// Retry configuration
const (
	MaxRetryAttempts     = 5
	BaseBackoffSeconds   = 30
	MaxBackoffSeconds    = 1800 // 30 minutes max backoff
	BackoffMultiplier    = 2.0
)

// JobExecutor handles job execution logic
type JobExecutor struct {
	jobRepo          *repository.JobRepository
	connectorRepo    *repository.ConnectorRepository
	ohlcvRepo        *repository.OHLCVRepository
	config           *config.Config
	ccxtService      *CCXTService
	indicatorService *indicators.Service
	rateLimiter      *RateLimiter
}

// NewJobExecutor creates a new job executor
func NewJobExecutor(jobRepo *repository.JobRepository, connectorRepo *repository.ConnectorRepository, ohlcvRepo *repository.OHLCVRepository, cfg *config.Config) *JobExecutor {
	// Create rate limiter
	rateLimiter := NewRateLimiter(connectorRepo)

	return &JobExecutor{
		jobRepo:          jobRepo,
		connectorRepo:    connectorRepo,
		ohlcvRepo:        ohlcvRepo,
		config:           cfg,
		ccxtService:      NewCCXTServiceWithRateLimiter(rateLimiter),
		indicatorService: indicators.NewService(),
		rateLimiter:      rateLimiter,
	}
}

// ExecuteJob executes a job by fetching OHLCV data from the exchange
func (e *JobExecutor) ExecuteJob(ctx context.Context, jobID string) (*models.JobExecutionResult, error) {
	startTime := time.Now()

	// Fetch job
	job, err := e.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to find job: %w", err)
	}

	// Check job dependencies before execution
	if len(job.DependsOn) > 0 {
		depStatus, err := e.jobRepo.GetDependencyStatus(ctx, jobID, 1*time.Hour)
		if err != nil {
			log.Printf("[Job %s] Warning: failed to check dependencies: %v", jobID, err)
		} else if !depStatus.AllDepsCompleted {
			errorMsg := fmt.Sprintf("waiting for dependencies: %v", depStatus.BlockedBy)
			log.Printf("[Job %s] Blocked by dependencies: %v", jobID, depStatus.BlockedBy)
			return &models.JobExecutionResult{
				Success:         false,
				Message:         "Job blocked by dependencies",
				RecordsFetched:  0,
				ExecutionTimeMs: time.Since(startTime).Milliseconds(),
				Error:           &errorMsg,
			}, nil
		}
	}

	// Fetch connector
	connector, err := e.connectorRepo.FindByExchangeID(ctx, job.ConnectorExchangeID)
	if err != nil {
		return nil, fmt.Errorf("failed to find connector: %w", err)
	}

	// Check if connector is active
	if connector.Status != "active" {
		errorMsg := "connector is not active"
		return &models.JobExecutionResult{
			Success:         false,
			Message:         "Job execution failed",
			RecordsFetched:  0,
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
			Error:           &errorMsg,
		}, nil
	}

	// Note: Rate limiting is now handled by the RateLimiter inside CCXTService
	// Each API call will wait for a rate limit slot before executing
	// This provides proper throttling at the API call level, not just at job start

	// Acquire job lock
	lockDuration := 5 * time.Minute
	locked, err := e.jobRepo.AcquireLock(ctx, jobID, lockDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire job lock: %w", err)
	}

	if !locked {
		errorMsg := "job is currently locked by another process"
		return &models.JobExecutionResult{
			Success:         false,
			Message:         "Job is locked",
			RecordsFetched:  0,
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
			Error:           &errorMsg,
		}, nil
	}

	// Ensure lock is released at the end
	defer e.jobRepo.ReleaseLock(ctx, jobID)

	// Fetch OHLCV data from exchange
	log.Printf("[EXEC] About to call FetchOHLCVData for %s", jobID)
	candles, err := e.FetchOHLCVData(connector, job)
	log.Printf("[EXEC] FetchOHLCVData returned %d candles, err=%v", len(candles), err)
	if err != nil {
		// Handle error with retry logic
		return e.handleExecutionError(ctx, job, err, startTime)
	}

	// Success - reset consecutive failures
	if job.RunState.ConsecutiveFailures > 0 {
		if resetErr := e.jobRepo.ResetConsecutiveFailures(ctx, jobID); resetErr != nil {
			log.Printf("[EXEC] Warning: Failed to reset consecutive failures: %v", resetErr)
		}
	}

	// Calculate ALL indicators for the fetched candles
	if len(candles) > 0 {
		log.Printf("[EXEC] Calculating all indicators for %d candles", len(candles))

		// Calculate all indicators with default periods
		candles, err = e.indicatorService.CalculateAll(candles)
		if err != nil {
			log.Printf("[EXEC] Warning: Indicator calculation failed: %v", err)
			// Continue with storing candles even if indicator calculation fails
		} else {
			log.Printf("[EXEC] All indicators calculated successfully")
		}
	}

	// Store OHLCV data in database using new UpsertCandles method
	recordsStored := 0
	if len(candles) > 0 {
		recordsStored, err = e.ohlcvRepo.UpsertCandles(ctx, connector.ExchangeID, job.Symbol, job.Timeframe, candles)
		if err != nil {
			errorMsg := err.Error()
			nextRunTime := e.calculateNextRunTime(job)
			if recErr := e.jobRepo.RecordRun(ctx, jobID, false, &nextRunTime, &errorMsg); recErr != nil {
				errorMsg = fmt.Sprintf("%s (also failed to record run: %v)", errorMsg, recErr)
			}

			return &models.JobExecutionResult{
				Success:         false,
				Message:         "Failed to store OHLCV data",
				RecordsFetched:  len(candles),
				ExecutionTimeMs: time.Since(startTime).Milliseconds(),
				Error:           &errorMsg,
			}, nil
		}

		// Update cursor with the timestamp of the most recent candle
		if len(candles) > 0 {
			// Get the most recent candle (first in array after reversal - index 0)
			mostRecentCandle := candles[0]
			lastCandleTime := time.UnixMilli(mostRecentCandle.Timestamp)
			_ = e.jobRepo.UpdateCursor(ctx, jobID, lastCandleTime)
			log.Printf("[EXEC] Updated cursor to most recent candle timestamp: %s", lastCandleTime.Format("2006-01-02 15:04:05"))
		}
	}

	// Calculate next run time
	nextRunTime := e.calculateNextRunTime(job)

	// Record successful run
	fmt.Printf("[DEBUG EXECUTOR] About to call RecordRun for job %s, nextRunTime=%v\n", jobID, nextRunTime)
	if err := e.jobRepo.RecordRun(ctx, jobID, true, &nextRunTime, nil); err != nil {
		// Log error but don't fail the execution
		errorMsg := fmt.Sprintf("Failed to record run: %v", err)
		fmt.Printf("[DEBUG EXECUTOR] RecordRun failed: %s\n", errorMsg)
		return &models.JobExecutionResult{
			Success:         true,
			Message:         "Job executed but failed to record run state",
			RecordsFetched:  recordsStored,
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
			NextRunTime:     nextRunTime,
			Error:           &errorMsg,
		}, nil
	}
	fmt.Printf("[DEBUG EXECUTOR] RecordRun succeeded for job %s\n", jobID)

	return &models.JobExecutionResult{
		Success:         true,
		Message:         "Job executed successfully",
		RecordsFetched:  recordsStored,
		ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		NextRunTime:     nextRunTime,
	}, nil
}

// FetchOHLCVData fetches OHLCV data from the exchange using CCXT
func (e *JobExecutor) FetchOHLCVData(connector *models.Connector, job *models.Job) ([]models.Candle, error) {
	// Use background context with timeout for the fetch operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute) // Allow 30 minutes for large historical fetches
	defer cancel()

	return e.FetchOHLCVDataWithContext(ctx, connector, job)
}

// FetchOHLCVDataWithContext fetches OHLCV data with context support for cancellation and rate limiting
func (e *JobExecutor) FetchOHLCVDataWithContext(ctx context.Context, connector *models.Connector, job *models.Job) ([]models.Candle, error) {
	log.Printf("[FETCH_START] Starting FetchOHLCVData for %s/%s", job.Symbol, job.Timeframe)

	var sinceMs *int64
	var isFirstExecution bool

	// STRATEGY: First execution vs Subsequent execution
	if job.Cursor.LastCandleTime == nil {
		// FIRST EXECUTION: Fetch ALL available data (no since, no limit)
		isFirstExecution = true
		sinceMs = nil
		log.Printf("[FETCH] First execution for %s/%s - fetching ALL available data (no since, no limit)",
			job.Symbol, job.Timeframe)
	} else {
		// SUBSEQUENT EXECUTION: Fetch only NEW candles from last candle timestamp
		isFirstExecution = false
		// Add one timeframe duration to avoid fetching the last candle again
		timeframeDuration, _ := parseTimeframe(job.Timeframe)
		nextTimestamp := job.Cursor.LastCandleTime.Add(timeframeDuration)
		sinceTimestamp := nextTimestamp.UnixMilli()
		sinceMs = &sinceTimestamp
		log.Printf("[FETCH] Subsequent execution for %s/%s - fetching from timestamp %d (after last candle, no limit)",
			job.Symbol, job.Timeframe, sinceTimestamp)
	}

	// Fetch real OHLCV data from exchange using CCXT with rate limiting
	candles, err := e.ccxtService.FetchOHLCVDataWithContext(
		ctx,
		connector.ExchangeID,
		job.Symbol,
		job.Timeframe,
		sinceMs, // nil for first, timestamp for subsequent
	)

	if err != nil {
		return nil, fmt.Errorf("CCXT fetch failed: %w", err)
	}

	if isFirstExecution {
		log.Printf("[FETCH] First execution complete - fetched %d historical candles", len(candles))
	} else {
		log.Printf("[FETCH] Subsequent execution complete - fetched %d new candles", len(candles))
	}

	return candles, nil
}

// calculateNextRunTime calculates when the job should run next
func (e *JobExecutor) calculateNextRunTime(job *models.Job) time.Time {
	// Parse timeframe to duration
	duration, err := parseTimeframe(job.Timeframe)
	if err != nil {
		// Default to 5 minutes if parsing fails
		return time.Now().Add(5 * time.Minute)
	}

	return time.Now().Add(duration)
}

// parseTimeframe converts a timeframe string to a duration
func parseTimeframe(timeframe string) (time.Duration, error) {
	switch timeframe {
	case "1m":
		return 1 * time.Minute, nil
	case "5m":
		return 5 * time.Minute, nil
	case "15m":
		return 15 * time.Minute, nil
	case "30m":
		return 30 * time.Minute, nil
	case "1h":
		return 1 * time.Hour, nil
	case "4h":
		return 4 * time.Hour, nil
	case "1d":
		return 24 * time.Hour, nil
	case "1w":
		return 7 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown timeframe: %s", timeframe)
	}
}

// handleExecutionError handles execution errors with retry logic
func (e *JobExecutor) handleExecutionError(ctx context.Context, job *models.Job, err error, startTime time.Time) (*models.JobExecutionResult, error) {
	errorMsg := err.Error()
	jobID := job.ID.Hex()

	// Determine if error is transient (can be retried)
	isTransient := isTransientError(err)

	// Increment consecutive failures
	consecutiveFailures := job.RunState.ConsecutiveFailures + 1
	if incErr := e.jobRepo.IncrementConsecutiveFailures(ctx, jobID); incErr != nil {
		log.Printf("[EXEC] Warning: Failed to increment consecutive failures: %v", incErr)
	}

	var nextRunTime time.Time
	var message string

	if isTransient && consecutiveFailures < MaxRetryAttempts {
		// Calculate backoff time
		backoffSeconds := calculateBackoff(consecutiveFailures)
		nextRunTime = time.Now().Add(time.Duration(backoffSeconds) * time.Second)
		message = fmt.Sprintf("Transient error (attempt %d/%d), will retry in %ds",
			consecutiveFailures, MaxRetryAttempts, backoffSeconds)
		log.Printf("[EXEC] %s: %s - %s", jobID, message, errorMsg)
	} else if consecutiveFailures >= MaxRetryAttempts {
		// Max retries exceeded - use normal schedule but keep error
		nextRunTime = e.calculateNextRunTime(job)
		message = fmt.Sprintf("Max retry attempts (%d) exceeded, resuming normal schedule", MaxRetryAttempts)
		log.Printf("[EXEC] %s: %s", jobID, message)

		// Reset consecutive failures to allow future retries
		if resetErr := e.jobRepo.ResetConsecutiveFailures(ctx, jobID); resetErr != nil {
			log.Printf("[EXEC] Warning: Failed to reset consecutive failures: %v", resetErr)
		}
	} else {
		// Non-transient error - use normal schedule
		nextRunTime = e.calculateNextRunTime(job)
		message = "Non-transient error, scheduled for normal retry"
		log.Printf("[EXEC] %s: %s - %s", jobID, message, errorMsg)
	}

	// Record the failed run
	fullError := fmt.Sprintf("%s: %s", message, errorMsg)
	if recErr := e.jobRepo.RecordRun(ctx, jobID, false, &nextRunTime, &fullError); recErr != nil {
		fullError = fmt.Sprintf("%s (also failed to record run: %v)", fullError, recErr)
	}

	return &models.JobExecutionResult{
		Success:         false,
		Message:         message,
		RecordsFetched:  0,
		ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		NextRunTime:     nextRunTime,
		Error:           &fullError,
	}, nil
}

// isTransientError determines if an error is transient and can be retried
func isTransientError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// Transient errors that should be retried
	transientPatterns := []string{
		"rate limit",
		"too many requests",
		"429",           // HTTP 429 Too Many Requests
		"timeout",
		"timed out",
		"connection reset",
		"connection refused",
		"temporary failure",
		"service unavailable",
		"503",           // HTTP 503 Service Unavailable
		"502",           // HTTP 502 Bad Gateway
		"504",           // HTTP 504 Gateway Timeout
		"network",
		"dns",
		"eof",
		"broken pipe",
		"context deadline exceeded",
		"context canceled",
	}

	for _, pattern := range transientPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	// Permanent errors that should NOT be retried
	permanentPatterns := []string{
		"invalid symbol",
		"symbol not found",
		"market not found",
		"authentication",
		"unauthorized",
		"forbidden",
		"400",           // HTTP 400 Bad Request
		"401",           // HTTP 401 Unauthorized
		"403",           // HTTP 403 Forbidden
		"404",           // HTTP 404 Not Found
		"invalid api key",
		"insufficient permissions",
	}

	for _, pattern := range permanentPatterns {
		if strings.Contains(errStr, pattern) {
			return false
		}
	}

	// Default to transient for unknown errors
	return true
}

// calculateBackoff calculates exponential backoff time in seconds
func calculateBackoff(attempt int) int {
	backoff := float64(BaseBackoffSeconds) * pow(BackoffMultiplier, float64(attempt-1))
	if backoff > float64(MaxBackoffSeconds) {
		return MaxBackoffSeconds
	}
	return int(backoff)
}

// pow calculates base^exp for float64
func pow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}
