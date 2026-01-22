package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yourusername/datacollector/internal/models"
	"github.com/yourusername/datacollector/internal/repository"
)

// RetentionService handles data retention and cleanup operations
type RetentionService struct {
	retentionRepo *repository.RetentionRepository
}

// NewRetentionService creates a new retention service
func NewRetentionService(retentionRepo *repository.RetentionRepository) *RetentionService {
	return &RetentionService{
		retentionRepo: retentionRepo,
	}
}

// RunCleanup executes cleanup based on enabled policies
func (s *RetentionService) RunCleanup(ctx context.Context) (*models.RetentionCleanupSummary, error) {
	startTime := time.Now()

	config, err := s.retentionRepo.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get retention config: %w", err)
	}

	if !config.Enabled {
		log.Println("[RETENTION] Cleanup disabled, skipping")
		return &models.RetentionCleanupSummary{
			StartedAt:   startTime,
			CompletedAt: time.Now(),
			Results:     []models.RetentionCleanupResult{},
		}, nil
	}

	// Get enabled policies
	policies, err := s.retentionRepo.FindEnabledPolicies(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled policies: %w", err)
	}

	summary := &models.RetentionCleanupSummary{
		StartedAt: startTime,
		Results:   make([]models.RetentionCleanupResult, 0, len(policies)),
	}

	// Execute each policy
	for _, policy := range policies {
		result := s.executePolicy(ctx, policy)
		summary.Results = append(summary.Results, result)
		summary.TotalChunksDeleted += result.ChunksDeleted
		summary.TotalCandlesDeleted += result.CandlesDeleted
		summary.TotalBytesFreed += result.BytesFreed
		summary.TotalDuration += result.Duration

		// Record policy run
		if err := s.retentionRepo.RecordPolicyRun(ctx, policy.ID.Hex()); err != nil {
			log.Printf("[RETENTION] Warning: Failed to record policy run: %v", err)
		}
	}

	// Also delete empty chunks if configured
	if config.DeleteEmptyChunks {
		emptyDeleted, err := s.retentionRepo.DeleteEmptyChunks(ctx)
		if err != nil {
			log.Printf("[RETENTION] Warning: Failed to delete empty chunks: %v", err)
		} else if emptyDeleted > 0 {
			log.Printf("[RETENTION] Deleted %d empty chunks", emptyDeleted)
			summary.TotalChunksDeleted += emptyDeleted
		}
	}

	summary.CompletedAt = time.Now()
	log.Printf("[RETENTION] Cleanup completed: %d chunks, %d candles deleted in %dms",
		summary.TotalChunksDeleted, summary.TotalCandlesDeleted, summary.TotalDuration)

	return summary, nil
}

// executePolicy executes a single retention policy
func (s *RetentionService) executePolicy(ctx context.Context, policy *models.RetentionPolicy) models.RetentionCleanupResult {
	startTime := time.Now()

	result := models.RetentionCleanupResult{
		PolicyID:   policy.ID.Hex(),
		PolicyName: policy.Name,
		StartedAt:  startTime,
	}

	// Calculate cutoff time
	cutoffTime := time.Now().AddDate(0, 0, -policy.RetentionDays)

	// Determine scope
	exchangeID := ""
	timeframe := ""

	if policy.ExchangeID != nil {
		exchangeID = *policy.ExchangeID
		result.ExchangeID = exchangeID
	}
	if policy.Timeframe != nil {
		timeframe = *policy.Timeframe
		result.Timeframe = timeframe
	}

	log.Printf("[RETENTION] Executing policy '%s': delete data older than %s (exchange=%s, timeframe=%s)",
		policy.Name, cutoffTime.Format("2006-01-02"), exchangeID, timeframe)

	// Delete old chunks
	chunksDeleted, err := s.retentionRepo.DeleteChunksOlderThan(ctx, cutoffTime, exchangeID, timeframe)
	if err != nil {
		result.Error = err.Error()
		log.Printf("[RETENTION] Policy '%s' error: %v", policy.Name, err)
	} else {
		result.ChunksDeleted = chunksDeleted
		// Estimate candles (average ~1000 per chunk)
		result.CandlesDeleted = chunksDeleted * 1000
		// Estimate bytes (average ~100KB per chunk)
		result.BytesFreed = chunksDeleted * 100 * 1024
		log.Printf("[RETENTION] Policy '%s' completed: %d chunks deleted", policy.Name, chunksDeleted)
	}

	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(startTime).Milliseconds()

	return result
}

// RunDefaultCleanup executes cleanup based on default config only
func (s *RetentionService) RunDefaultCleanup(ctx context.Context, retentionDays int) (*models.RetentionCleanupResult, error) {
	startTime := time.Now()

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	log.Printf("[RETENTION] Running default cleanup: delete data older than %s", cutoffTime.Format("2006-01-02"))

	chunksDeleted, err := s.retentionRepo.DeleteChunksOlderThan(ctx, cutoffTime, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to delete old chunks: %w", err)
	}

	result := &models.RetentionCleanupResult{
		ChunksDeleted:  chunksDeleted,
		CandlesDeleted: chunksDeleted * 1000,
		BytesFreed:     chunksDeleted * 100 * 1024,
		StartedAt:      startTime,
		CompletedAt:    time.Now(),
		Duration:       time.Since(startTime).Milliseconds(),
	}

	log.Printf("[RETENTION] Default cleanup completed: %d chunks deleted", chunksDeleted)

	return result, nil
}

// CleanupByExchange deletes old data for a specific exchange
func (s *RetentionService) CleanupByExchange(ctx context.Context, exchangeID string, retentionDays int) (*models.RetentionCleanupResult, error) {
	startTime := time.Now()

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	log.Printf("[RETENTION] Cleaning up exchange %s: delete data older than %s", exchangeID, cutoffTime.Format("2006-01-02"))

	chunksDeleted, err := s.retentionRepo.DeleteChunksOlderThan(ctx, cutoffTime, exchangeID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to delete old chunks: %w", err)
	}

	result := &models.RetentionCleanupResult{
		ExchangeID:     exchangeID,
		ChunksDeleted:  chunksDeleted,
		CandlesDeleted: chunksDeleted * 1000,
		BytesFreed:     chunksDeleted * 100 * 1024,
		StartedAt:      startTime,
		CompletedAt:    time.Now(),
		Duration:       time.Since(startTime).Milliseconds(),
	}

	log.Printf("[RETENTION] Exchange cleanup completed: %d chunks deleted", chunksDeleted)

	return result, nil
}

// GetDataUsage returns data usage statistics
func (s *RetentionService) GetDataUsage(ctx context.Context, exchangeID string) ([]*models.DataUsageStats, error) {
	return s.retentionRepo.GetDataUsageStats(ctx, exchangeID)
}

// GetTotalUsage returns total data usage
func (s *RetentionService) GetTotalUsage(ctx context.Context) (chunks int64, candles int64, estimatedSizeMB float64, err error) {
	chunks, candles, err = s.retentionRepo.GetTotalDataUsage(ctx)
	if err != nil {
		return 0, 0, 0, err
	}
	estimatedSizeMB = float64(candles*100) / (1024 * 1024)
	return chunks, candles, estimatedSizeMB, nil
}

// DeleteEmptyChunks removes chunks with no candles
func (s *RetentionService) DeleteEmptyChunks(ctx context.Context) (int64, error) {
	return s.retentionRepo.DeleteEmptyChunks(ctx)
}
