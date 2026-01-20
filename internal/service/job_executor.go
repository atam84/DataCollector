package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yourusername/datacollector/internal/config"
	"github.com/yourusername/datacollector/internal/models"
	"github.com/yourusername/datacollector/internal/repository"
	"github.com/yourusername/datacollector/internal/service/indicators"
)

// JobExecutor handles job execution logic
type JobExecutor struct {
	jobRepo          *repository.JobRepository
	connectorRepo    *repository.ConnectorRepository
	ohlcvRepo        *repository.OHLCVRepository
	config           *config.Config
	ccxtService      *CCXTService
	indicatorService *indicators.Service
}

// NewJobExecutor creates a new job executor
func NewJobExecutor(jobRepo *repository.JobRepository, connectorRepo *repository.ConnectorRepository, ohlcvRepo *repository.OHLCVRepository, cfg *config.Config) *JobExecutor {
	return &JobExecutor{
		jobRepo:          jobRepo,
		connectorRepo:    connectorRepo,
		ohlcvRepo:        ohlcvRepo,
		config:           cfg,
		ccxtService:      NewCCXTService(cfg.Exchange.SandboxMode),
		indicatorService: indicators.NewService(),
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

	// Acquire rate limit token
	acquired, err := e.connectorRepo.AcquireRateLimitToken(ctx, connector.ExchangeID, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire rate limit token: %w", err)
	}

	if !acquired {
		errorMsg := "rate limit exceeded"
		return &models.JobExecutionResult{
			Success:         false,
			Message:         "Rate limit exceeded",
			RecordsFetched:  0,
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
			Error:           &errorMsg,
		}, nil
	}

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
		// Record failed run
		errorMsg := err.Error()
		nextRunTime := e.calculateNextRunTime(job)
		if recErr := e.jobRepo.RecordRun(ctx, jobID, false, &nextRunTime, &errorMsg); recErr != nil {
			errorMsg = fmt.Sprintf("%s (also failed to record run: %v)", errorMsg, recErr)
		}

		return &models.JobExecutionResult{
			Success:         false,
			Message:         "Failed to fetch OHLCV data",
			RecordsFetched:  0,
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
			Error:           &errorMsg,
		}, nil
	}

	// Calculate indicators for the fetched candles
	if len(candles) > 0 {
		log.Printf("[EXEC] Calculating indicators for %d candles", len(candles))

		// Get merged indicator configuration (Job overrides Connector)
		indicatorConfig := indicators.GetEffectiveConfig(connector.IndicatorConfig, job.IndicatorConfig)

		// Validate configuration
		if err := e.indicatorService.ValidateConfig(indicatorConfig); err != nil {
			log.Printf("[EXEC] Warning: Invalid indicator configuration: %v", err)
			indicatorConfig = indicators.DefaultConfig() // Fallback to defaults
		}

		// Calculate indicators
		candles, err = e.indicatorService.CalculateAll(candles, indicatorConfig)
		if err != nil {
			log.Printf("[EXEC] Warning: Indicator calculation failed: %v", err)
			// Continue with storing candles even if indicator calculation fails
		} else {
			log.Printf("[EXEC] Indicators calculated successfully with config priority: Job > Connector > Defaults")
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

	// Fetch real OHLCV data from exchange using CCXT
	candles, err := e.ccxtService.FetchOHLCVData(
		connector.ExchangeID,
		job.Symbol,
		job.Timeframe,
		sinceMs, // nil for first, timestamp for subsequent
		connector.SandboxMode,
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
