package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/yourusername/datacollector/internal/models"
)

// RetentionRepository handles database operations for retention policies
type RetentionRepository struct {
	policyCollection *mongo.Collection
	configCollection *mongo.Collection
	ohlcvCollection  *mongo.Collection
}

// NewRetentionRepository creates a new retention repository
func NewRetentionRepository(db *Database) *RetentionRepository {
	policyCollection := db.GetCollection("retention_policies")
	configCollection := db.GetCollection("retention_config")
	ohlcvCollection := db.GetCollection("ohlcv_chunks")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create indexes for policies
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "type", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "enabled", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "exchange_id", Value: 1}},
		},
	}
	_, _ = policyCollection.Indexes().CreateMany(ctx, indexes)

	return &RetentionRepository{
		policyCollection: policyCollection,
		configCollection: configCollection,
		ohlcvCollection:  ohlcvCollection,
	}
}

// CreatePolicy creates a new retention policy
func (r *RetentionRepository) CreatePolicy(ctx context.Context, policy *models.RetentionPolicy) error {
	policy.ID = primitive.NewObjectID()
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	_, err := r.policyCollection.InsertOne(ctx, policy)
	if err != nil {
		return fmt.Errorf("failed to create retention policy: %w", err)
	}

	return nil
}

// FindPolicyByID retrieves a policy by ID
func (r *RetentionRepository) FindPolicyByID(ctx context.Context, id string) (*models.RetentionPolicy, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid policy ID: %w", err)
	}

	var policy models.RetentionPolicy
	err = r.policyCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&policy)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("policy not found")
		}
		return nil, fmt.Errorf("failed to find policy: %w", err)
	}

	return &policy, nil
}

// FindAllPolicies retrieves all retention policies
func (r *RetentionRepository) FindAllPolicies(ctx context.Context) ([]*models.RetentionPolicy, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.policyCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find policies: %w", err)
	}
	defer cursor.Close(ctx)

	var policies []*models.RetentionPolicy
	if err := cursor.All(ctx, &policies); err != nil {
		return nil, fmt.Errorf("failed to decode policies: %w", err)
	}

	return policies, nil
}

// FindEnabledPolicies retrieves all enabled retention policies
func (r *RetentionRepository) FindEnabledPolicies(ctx context.Context) ([]*models.RetentionPolicy, error) {
	filter := bson.M{"enabled": true}
	opts := options.Find().SetSort(bson.D{{Key: "type", Value: 1}})

	cursor, err := r.policyCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find enabled policies: %w", err)
	}
	defer cursor.Close(ctx)

	var policies []*models.RetentionPolicy
	if err := cursor.All(ctx, &policies); err != nil {
		return nil, fmt.Errorf("failed to decode policies: %w", err)
	}

	return policies, nil
}

// UpdatePolicy updates a retention policy
func (r *RetentionRepository) UpdatePolicy(ctx context.Context, id string, update bson.M) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid policy ID: %w", err)
	}

	update["updated_at"] = time.Now()

	result, err := r.policyCollection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": update},
	)

	if err != nil {
		return fmt.Errorf("failed to update policy: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("policy not found")
	}

	return nil
}

// DeletePolicy removes a retention policy
func (r *RetentionRepository) DeletePolicy(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid policy ID: %w", err)
	}

	result, err := r.policyCollection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("policy not found")
	}

	return nil
}

// RecordPolicyRun updates the last run timestamp for a policy
func (r *RetentionRepository) RecordPolicyRun(ctx context.Context, id string) error {
	now := time.Now()
	update := bson.M{
		"last_run_at": now,
		"updated_at":  now,
	}
	return r.UpdatePolicy(ctx, id, update)
}

// GetConfig retrieves the retention configuration
func (r *RetentionRepository) GetConfig(ctx context.Context) (*models.RetentionConfig, error) {
	var config models.RetentionConfig
	err := r.configCollection.FindOne(ctx, bson.M{}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return models.DefaultRetentionConfig(), nil
		}
		return nil, fmt.Errorf("failed to get retention config: %w", err)
	}
	return &config, nil
}

// SaveConfig saves the retention configuration
func (r *RetentionRepository) SaveConfig(ctx context.Context, config *models.RetentionConfig) error {
	config.UpdatedAt = time.Now()

	opts := options.Update().SetUpsert(true)
	filter := bson.M{}
	if config.ID != primitive.NilObjectID {
		filter["_id"] = config.ID
	}

	update := bson.M{"$set": config}

	_, err := r.configCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to save retention config: %w", err)
	}

	return nil
}

// DeleteChunksOlderThan deletes OHLCV chunks older than the specified time
func (r *RetentionRepository) DeleteChunksOlderThan(ctx context.Context, cutoffTime time.Time, exchangeID, timeframe string) (int64, error) {
	filter := bson.M{
		"year_month": bson.M{"$lt": cutoffTime.Format("2006-01")},
	}

	if exchangeID != "" {
		filter["exchange_id"] = exchangeID
	}
	if timeframe != "" {
		filter["timeframe"] = timeframe
	}

	result, err := r.ohlcvCollection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old chunks: %w", err)
	}

	return result.DeletedCount, nil
}

// GetDataUsageStats returns storage usage statistics
func (r *RetentionRepository) GetDataUsageStats(ctx context.Context, exchangeID string) ([]*models.DataUsageStats, error) {
	matchStage := bson.M{}
	if exchangeID != "" {
		matchStage["exchange_id"] = exchangeID
	}

	pipeline := []bson.M{
		{"$match": matchStage},
		{"$group": bson.M{
			"_id": bson.M{
				"exchange_id": "$exchange_id",
				"symbol":      "$symbol",
				"timeframe":   "$timeframe",
			},
			"chunk_count":   bson.M{"$sum": 1},
			"total_candles": bson.M{"$sum": bson.M{"$size": bson.M{"$ifNull": bson.A{"$candles", bson.A{}}}}},
			"oldest_data":   bson.M{"$min": "$year_month"},
			"newest_data":   bson.M{"$max": "$year_month"},
		}},
		{"$sort": bson.D{
			{Key: "_id.exchange_id", Value: 1},
			{Key: "_id.symbol", Value: 1},
			{Key: "_id.timeframe", Value: 1},
		}},
	}

	cursor, err := r.ohlcvCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate usage stats: %w", err)
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID struct {
			ExchangeID string `bson:"exchange_id"`
			Symbol     string `bson:"symbol"`
			Timeframe  string `bson:"timeframe"`
		} `bson:"_id"`
		ChunkCount   int    `bson:"chunk_count"`
		TotalCandles int64  `bson:"total_candles"`
		OldestData   string `bson:"oldest_data"`
		NewestData   string `bson:"newest_data"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode usage stats: %w", err)
	}

	stats := make([]*models.DataUsageStats, 0, len(results))
	for _, r := range results {
		stat := &models.DataUsageStats{
			ExchangeID:   r.ID.ExchangeID,
			Symbol:       r.ID.Symbol,
			Timeframe:    r.ID.Timeframe,
			ChunkCount:   r.ChunkCount,
			TotalCandles: r.TotalCandles,
			// Estimate ~100 bytes per candle
			EstimatedSizeMB: float64(r.TotalCandles*100) / (1024 * 1024),
		}

		if r.OldestData != "" {
			if t, err := time.Parse("2006-01", r.OldestData); err == nil {
				stat.OldestData = t
			}
		}
		if r.NewestData != "" {
			if t, err := time.Parse("2006-01", r.NewestData); err == nil {
				stat.NewestData = t
			}
		}

		stats = append(stats, stat)
	}

	return stats, nil
}

// GetTotalDataUsage returns aggregate data usage
func (r *RetentionRepository) GetTotalDataUsage(ctx context.Context) (chunks int64, candles int64, err error) {
	// Count chunks
	chunks, err = r.ohlcvCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, 0, fmt.Errorf("failed to count chunks: %w", err)
	}

	// Sum candles
	pipeline := []bson.M{
		{"$group": bson.M{
			"_id":           nil,
			"total_candles": bson.M{"$sum": bson.M{"$size": bson.M{"$ifNull": bson.A{"$candles", bson.A{}}}}},
		}},
	}

	cursor, err := r.ohlcvCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return chunks, 0, fmt.Errorf("failed to sum candles: %w", err)
	}
	defer cursor.Close(ctx)

	var result []struct {
		TotalCandles int64 `bson:"total_candles"`
	}
	if err := cursor.All(ctx, &result); err != nil {
		return chunks, 0, fmt.Errorf("failed to decode candle count: %w", err)
	}

	if len(result) > 0 {
		candles = result[0].TotalCandles
	}

	return chunks, candles, nil
}

// DeleteEmptyChunks removes chunks that have no candles
func (r *RetentionRepository) DeleteEmptyChunks(ctx context.Context) (int64, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"candles": bson.M{"$size": 0}},
			{"candles": bson.M{"$exists": false}},
			{"candles": nil},
		},
	}

	result, err := r.ohlcvCollection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to delete empty chunks: %w", err)
	}

	return result.DeletedCount, nil
}
