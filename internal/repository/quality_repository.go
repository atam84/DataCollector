package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/yourusername/datacollector/internal/models"
)

// QualityRepository handles database operations for quality data
type QualityRepository struct {
	resultsCollection  *mongo.Collection
	checksCollection   *mongo.Collection
	summaryCollection  *mongo.Collection
}

// NewQualityRepository creates a new quality repository
func NewQualityRepository(db *Database) *QualityRepository {
	resultsCollection := db.GetCollection("quality_results")
	checksCollection := db.GetCollection("quality_checks")
	summaryCollection := db.GetCollection("quality_summary")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Index for quality results: unique on (exchange_id, symbol, timeframe)
	resultIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "exchange_id", Value: 1},
			{Key: "symbol", Value: 1},
			{Key: "timeframe", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	_, err := resultsCollection.Indexes().CreateOne(ctx, resultIndex)
	if err != nil {
		log.Printf("[QUALITY_REPO] Warning: Failed to create results index: %v", err)
	}

	// Index for quality checks by status
	checkStatusIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "status", Value: 1}},
	}
	_, err = checksCollection.Indexes().CreateOne(ctx, checkStatusIndex)
	if err != nil {
		log.Printf("[QUALITY_REPO] Warning: Failed to create check status index: %v", err)
	}

	// Index for summary by exchange
	summaryIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "exchange_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err = summaryCollection.Indexes().CreateOne(ctx, summaryIndex)
	if err != nil {
		log.Printf("[QUALITY_REPO] Warning: Failed to create summary index: %v", err)
	}

	return &QualityRepository{
		resultsCollection: resultsCollection,
		checksCollection:  checksCollection,
		summaryCollection: summaryCollection,
	}
}

// === Quality Results ===

// UpsertResult inserts or updates a quality result
func (r *QualityRepository) UpsertResult(ctx context.Context, result *models.DataQualityResult) error {
	filter := bson.M{
		"exchange_id": result.ExchangeID,
		"symbol":      result.Symbol,
		"timeframe":   result.Timeframe,
	}

	now := time.Now()
	result.UpdatedAt = now
	if result.CreatedAt.IsZero() {
		result.CreatedAt = now
	}

	update := bson.M{"$set": result}
	opts := options.Update().SetUpsert(true)

	_, err := r.resultsCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to upsert quality result: %w", err)
	}

	return nil
}

// FindResult finds a quality result by job identifiers
func (r *QualityRepository) FindResult(ctx context.Context, exchangeID, symbol, timeframe string) (*models.DataQualityResult, error) {
	filter := bson.M{
		"exchange_id": exchangeID,
		"symbol":      symbol,
		"timeframe":   timeframe,
	}

	var result models.DataQualityResult
	err := r.resultsCollection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find quality result: %w", err)
	}

	return &result, nil
}

// FindResultByJobID finds a quality result by job ID
func (r *QualityRepository) FindResultByJobID(ctx context.Context, jobID string) (*models.DataQualityResult, error) {
	objID, err := primitive.ObjectIDFromHex(jobID)
	if err != nil {
		return nil, fmt.Errorf("invalid job ID: %w", err)
	}

	var result models.DataQualityResult
	err = r.resultsCollection.FindOne(ctx, bson.M{"job_id": objID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find quality result: %w", err)
	}

	return &result, nil
}

// FindAllResults finds all quality results with optional filtering
func (r *QualityRepository) FindAllResults(ctx context.Context, filter bson.M) ([]*models.DataQualityResult, error) {
	opts := options.Find().SetSort(bson.D{
		{Key: "quality_status", Value: 1}, // poor first
		{Key: "completeness_score", Value: 1},
	})

	cursor, err := r.resultsCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find quality results: %w", err)
	}
	defer cursor.Close(ctx)

	var results []*models.DataQualityResult
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode quality results: %w", err)
	}

	return results, nil
}

// DeleteResult deletes a quality result
func (r *QualityRepository) DeleteResult(ctx context.Context, exchangeID, symbol, timeframe string) error {
	filter := bson.M{
		"exchange_id": exchangeID,
		"symbol":      symbol,
		"timeframe":   timeframe,
	}

	_, err := r.resultsCollection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete quality result: %w", err)
	}

	return nil
}

// === Quality Check Jobs ===

// CreateCheckJob creates a new quality check job
func (r *QualityRepository) CreateCheckJob(ctx context.Context, checkJob *models.QualityCheckJob) error {
	now := time.Now()
	checkJob.CreatedAt = now
	checkJob.UpdatedAt = now
	checkJob.Status = models.QualityCheckPending

	result, err := r.checksCollection.InsertOne(ctx, checkJob)
	if err != nil {
		return fmt.Errorf("failed to create quality check job: %w", err)
	}

	checkJob.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// UpdateCheckJob updates a quality check job
func (r *QualityRepository) UpdateCheckJob(ctx context.Context, checkJob *models.QualityCheckJob) error {
	checkJob.UpdatedAt = time.Now()

	_, err := r.checksCollection.UpdateOne(
		ctx,
		bson.M{"_id": checkJob.ID},
		bson.M{"$set": checkJob},
	)
	if err != nil {
		return fmt.Errorf("failed to update quality check job: %w", err)
	}

	return nil
}

// FindCheckJob finds a quality check job by ID
func (r *QualityRepository) FindCheckJob(ctx context.Context, id string) (*models.QualityCheckJob, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid check job ID: %w", err)
	}

	var checkJob models.QualityCheckJob
	err = r.checksCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&checkJob)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find quality check job: %w", err)
	}

	return &checkJob, nil
}

// FindActiveCheckJobs finds check jobs that are pending or running
func (r *QualityRepository) FindActiveCheckJobs(ctx context.Context) ([]*models.QualityCheckJob, error) {
	filter := bson.M{
		"status": bson.M{
			"$in": []models.QualityCheckStatus{
				models.QualityCheckPending,
				models.QualityCheckRunning,
			},
		},
	}

	cursor, err := r.checksCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find active check jobs: %w", err)
	}
	defer cursor.Close(ctx)

	var jobs []*models.QualityCheckJob
	if err := cursor.All(ctx, &jobs); err != nil {
		return nil, fmt.Errorf("failed to decode check jobs: %w", err)
	}

	return jobs, nil
}

// FindRecentCheckJobs finds recent check jobs
func (r *QualityRepository) FindRecentCheckJobs(ctx context.Context, limit int) ([]*models.QualityCheckJob, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.checksCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find recent check jobs: %w", err)
	}
	defer cursor.Close(ctx)

	var jobs []*models.QualityCheckJob
	if err := cursor.All(ctx, &jobs); err != nil {
		return nil, fmt.Errorf("failed to decode check jobs: %w", err)
	}

	return jobs, nil
}

// === Quality Summary ===

// UpsertSummary inserts or updates the quality summary
func (r *QualityRepository) UpsertSummary(ctx context.Context, summary *models.QualitySummaryCache) error {
	filter := bson.M{"exchange_id": summary.ExchangeID}

	now := time.Now()
	summary.UpdatedAt = now
	if summary.CreatedAt.IsZero() {
		summary.CreatedAt = now
	}

	update := bson.M{"$set": summary}
	opts := options.Update().SetUpsert(true)

	_, err := r.summaryCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to upsert quality summary: %w", err)
	}

	return nil
}

// FindSummary finds the quality summary (global or by exchange)
func (r *QualityRepository) FindSummary(ctx context.Context, exchangeID string) (*models.QualitySummaryCache, error) {
	filter := bson.M{"exchange_id": exchangeID}

	var summary models.QualitySummaryCache
	err := r.summaryCollection.FindOne(ctx, filter).Decode(&summary)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find quality summary: %w", err)
	}

	return &summary, nil
}

// FindAllSummaries finds all quality summaries
func (r *QualityRepository) FindAllSummaries(ctx context.Context) ([]*models.QualitySummaryCache, error) {
	cursor, err := r.summaryCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to find quality summaries: %w", err)
	}
	defer cursor.Close(ctx)

	var summaries []*models.QualitySummaryCache
	if err := cursor.All(ctx, &summaries); err != nil {
		return nil, fmt.Errorf("failed to decode quality summaries: %w", err)
	}

	return summaries, nil
}

// ComputeSummaryFromResults computes summary from stored results
func (r *QualityRepository) ComputeSummaryFromResults(ctx context.Context, exchangeID string) (*models.QualitySummaryCache, error) {
	filter := bson.M{}
	if exchangeID != "" {
		filter["exchange_id"] = exchangeID
	}

	results, err := r.FindAllResults(ctx, filter)
	if err != nil {
		return nil, err
	}

	summary := &models.QualitySummaryCache{
		ExchangeID:  exchangeID,
		TotalJobs:   len(results),
		LastCheckAt: time.Now(),
	}

	var totalCompleteness float64

	for _, result := range results {
		totalCompleteness += result.CompletenessScore
		summary.TotalCandles += result.TotalCandles
		summary.TotalMissingCandles += result.MissingCandles
		summary.TotalGaps += result.GapsDetected

		switch result.QualityStatus {
		case "excellent":
			summary.ExcellentQuality++
		case "good":
			summary.GoodQuality++
		case "fair":
			summary.FairQuality++
		case "poor":
			summary.PoorQuality++
		}

		switch result.DataFreshness {
		case "fresh":
			summary.FreshDataJobs++
		case "stale":
			summary.StaleDataJobs++
		case "very_stale":
			summary.VeryStaleDataJobs++
		}
	}

	if summary.TotalJobs > 0 {
		summary.AverageCompleteness = totalCompleteness / float64(summary.TotalJobs)
	}

	return summary, nil
}
