package service

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/yourusername/datacollector/internal/repository"
	"github.com/yourusername/datacollector/internal/service/indicators"
)

// RecalculatorService handles on-demand recalculation of indicators
type RecalculatorService struct {
	jobRepo          *repository.JobRepository
	connectorRepo    *repository.ConnectorRepository
	ohlcvRepo        *repository.OHLCVRepository
	indicatorService *indicators.Service
}

// NewRecalculatorService creates a new recalculator service
func NewRecalculatorService(
	jobRepo *repository.JobRepository,
	connectorRepo *repository.ConnectorRepository,
	ohlcvRepo *repository.OHLCVRepository,
) *RecalculatorService {
	return &RecalculatorService{
		jobRepo:          jobRepo,
		connectorRepo:    connectorRepo,
		ohlcvRepo:        ohlcvRepo,
		indicatorService: indicators.NewService(),
	}
}

// RecalculateJob recalculates all indicators for a specific job
// This fetches all existing candles, recalculates indicators, and updates the database
func (r *RecalculatorService) RecalculateJob(ctx context.Context, jobID string) error {
	log.Printf("[RECALC] Starting recalculation for job %s", jobID)

	// Fetch job
	job, err := r.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to find job: %w", err)
	}

	// Fetch connector to get exchange ID
	connector, err := r.connectorRepo.FindByExchangeID(ctx, job.ConnectorExchangeID)
	if err != nil {
		return fmt.Errorf("failed to find connector: %w", err)
	}

	// Fetch all candles for this job
	ohlcvDoc, err := r.ohlcvRepo.FindByJob(ctx, connector.ExchangeID, job.Symbol, job.Timeframe)
	if err != nil {
		return fmt.Errorf("failed to fetch candles: %w", err)
	}

	if ohlcvDoc == nil || len(ohlcvDoc.Candles) == 0 {
		log.Printf("[RECALC] No candles found for job %s", jobID)
		return nil
	}

	log.Printf("[RECALC] Found %d candles for job %s", len(ohlcvDoc.Candles), jobID)

	// Recalculate all indicators for all candles
	candles, err := r.indicatorService.CalculateAll(ohlcvDoc.Candles)
	if err != nil {
		return fmt.Errorf("failed to calculate indicators: %w", err)
	}

	// Update candles in database
	recordsUpdated, err := r.ohlcvRepo.UpsertCandles(ctx, connector.ExchangeID, job.Symbol, job.Timeframe, candles)
	if err != nil {
		return fmt.Errorf("failed to update candles: %w", err)
	}

	log.Printf("[RECALC] Successfully recalculated indicators for job %s (%d candles updated)", jobID, recordsUpdated)
	return nil
}

// RecalculateConnector recalculates all indicators for all jobs using a specific connector
// This is useful when connector-level indicator configuration changes
func (r *RecalculatorService) RecalculateConnector(ctx context.Context, connectorExchangeID string) error {
	log.Printf("[RECALC] Starting recalculation for all jobs on connector %s", connectorExchangeID)

	// Find all jobs for this connector
	jobs, err := r.jobRepo.FindAll(ctx, bson.M{"connector_exchange_id": connectorExchangeID})
	if err != nil {
		return fmt.Errorf("failed to find jobs: %w", err)
	}

	connectorJobs := jobs

	if len(connectorJobs) == 0 {
		log.Printf("[RECALC] No jobs found for connector %s", connectorExchangeID)
		return nil
	}

	log.Printf("[RECALC] Found %d jobs for connector %s", len(connectorJobs), connectorExchangeID)

	// Recalculate each job
	successCount := 0
	errorCount := 0
	for _, job := range connectorJobs {
		if err := r.RecalculateJob(ctx, job.ID.Hex()); err != nil {
			log.Printf("[RECALC] Error recalculating job %s: %v", job.ID.Hex(), err)
			errorCount++
		} else {
			successCount++
		}
	}

	log.Printf("[RECALC] Connector recalculation complete: %d succeeded, %d failed", successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("recalculation completed with %d errors", errorCount)
	}

	return nil
}

// RecalculateAll recalculates indicators for all jobs in the system
// WARNING: This can be resource-intensive for large datasets
func (r *RecalculatorService) RecalculateAll(ctx context.Context) error {
	log.Printf("[RECALC] Starting recalculation for ALL jobs")

	// Find all jobs
	jobs, err := r.jobRepo.FindAll(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to find jobs: %w", err)
	}

	log.Printf("[RECALC] Found %d total jobs", len(jobs))

	// Recalculate each job
	successCount := 0
	errorCount := 0
	for _, job := range jobs {
		if err := r.RecalculateJob(ctx, job.ID.Hex()); err != nil {
			log.Printf("[RECALC] Error recalculating job %s: %v", job.ID.Hex(), err)
			errorCount++
		} else {
			successCount++
		}
	}

	log.Printf("[RECALC] Full recalculation complete: %d succeeded, %d failed", successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("recalculation completed with %d errors", errorCount)
	}

	return nil
}

// RecalculationProgress holds progress information for long-running recalculations
type RecalculationProgress struct {
	TotalJobs      int    `json:"total_jobs"`
	CompletedJobs  int    `json:"completed_jobs"`
	FailedJobs     int    `json:"failed_jobs"`
	CurrentJobID   string `json:"current_job_id,omitempty"`
	Status         string `json:"status"` // "running", "completed", "failed"
	ErrorMessage   string `json:"error_message,omitempty"`
}
