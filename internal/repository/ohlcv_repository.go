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

// OHLCVRepository handles database operations for OHLCV data
type OHLCVRepository struct {
	collection *mongo.Collection
}

// NewOHLCVRepository creates a new OHLCV repository
func NewOHLCVRepository(db *Database) *OHLCVRepository {
	collection := db.GetCollection("ohlcv")

	// Create unique compound index on (exchange_id, symbol, timeframe)
	// One document per job combination
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "exchange_id", Value: 1},
			{Key: "symbol", Value: 1},
			{Key: "timeframe", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Printf("[OHLCV_REPO] Warning: Failed to create index: %v", err)
	}

	return &OHLCVRepository{
		collection: collection,
	}
}

// UpsertCandles inserts new candles into the document for a specific job
// Candles are prepended to the beginning of the array (newest first)
func (r *OHLCVRepository) UpsertCandles(ctx context.Context, exchangeID, symbol, timeframe string, newCandles []models.Candle) (int, error) {
	if len(newCandles) == 0 {
		return 0, nil
	}

	log.Printf("[OHLCV_REPO] Upserting %d candles for %s-%s-%s", len(newCandles), exchangeID, symbol, timeframe)

	// Filter to find the document
	filter := bson.M{
		"exchange_id": exchangeID,
		"symbol":      symbol,
		"timeframe":   timeframe,
	}

	now := time.Now()

	// Check if document exists
	var existingDoc models.OHLCVDocument
	err := r.collection.FindOne(ctx, filter).Decode(&existingDoc)

	if err == mongo.ErrNoDocuments {
		// Document doesn't exist - create new one
		log.Printf("[OHLCV_REPO] Creating new document for %s-%s-%s with %d candles", exchangeID, symbol, timeframe, len(newCandles))

		doc := models.OHLCVDocument{
			ID:           primitive.NewObjectID(),
			ExchangeID:   exchangeID,
			Symbol:       symbol,
			Timeframe:    timeframe,
			CreatedAt:    now,
			UpdatedAt:    now,
			CandlesCount: len(newCandles),
			Candles:      newCandles,
		}

		_, err := r.collection.InsertOne(ctx, doc)
		if err != nil {
			return 0, fmt.Errorf("failed to insert new OHLCV document: %w", err)
		}

		return len(newCandles), nil
	} else if err != nil {
		return 0, fmt.Errorf("failed to find existing document: %w", err)
	}

	// Document exists - prepend new candles to the beginning
	log.Printf("[OHLCV_REPO] Prepending %d candles to existing document (current count: %d)", len(newCandles), existingDoc.CandlesCount)

	// Update operation: prepend candles, update count and timestamp
	update := bson.M{
		"$push": bson.M{
			"candles": bson.M{
				"$each":     newCandles,
				"$position": 0, // Prepend to beginning (index 0)
			},
		},
		"$inc": bson.M{
			"candles_count": len(newCandles),
		},
		"$set": bson.M{
			"updated_at": now,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("failed to update OHLCV document: %w", err)
	}

	if result.ModifiedCount == 0 {
		log.Printf("[OHLCV_REPO] Warning: No documents modified")
		return 0, nil
	}

	log.Printf("[OHLCV_REPO] Successfully upserted %d candles (total now: %d)", len(newCandles), existingDoc.CandlesCount+len(newCandles))
	return len(newCandles), nil
}

// FindByJob retrieves the OHLCV document for a specific job
func (r *OHLCVRepository) FindByJob(ctx context.Context, exchangeID, symbol, timeframe string) (*models.OHLCVDocument, error) {
	filter := bson.M{
		"exchange_id": exchangeID,
		"symbol":      symbol,
		"timeframe":   timeframe,
	}

	var doc models.OHLCVDocument
	err := r.collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No document yet
		}
		return nil, fmt.Errorf("failed to find OHLCV document: %w", err)
	}

	return &doc, nil
}

// GetLastCandle retrieves the most recent candle for a job
// Since newest candles are at index 0, we return candles[0]
func (r *OHLCVRepository) GetLastCandle(ctx context.Context, exchangeID, symbol, timeframe string) (*models.Candle, error) {
	doc, err := r.FindByJob(ctx, exchangeID, symbol, timeframe)
	if err != nil {
		return nil, err
	}

	if doc == nil || len(doc.Candles) == 0 {
		return nil, nil // No candles yet
	}

	// Return the first candle (most recent)
	return &doc.Candles[0], nil
}

// GetCandlesCount returns the number of candles for a specific job
func (r *OHLCVRepository) GetCandlesCount(ctx context.Context, exchangeID, symbol, timeframe string) (int, error) {
	doc, err := r.FindByJob(ctx, exchangeID, symbol, timeframe)
	if err != nil {
		return 0, err
	}

	if doc == nil {
		return 0, nil
	}

	return doc.CandlesCount, nil
}

// GetRecentCandles retrieves the N most recent candles for a job
func (r *OHLCVRepository) GetRecentCandles(ctx context.Context, exchangeID, symbol, timeframe string, limit int) ([]models.Candle, error) {
	filter := bson.M{
		"exchange_id": exchangeID,
		"symbol":      symbol,
		"timeframe":   timeframe,
	}

	// Use projection to get only the first N candles
	projection := bson.M{
		"candles": bson.M{
			"$slice": limit,
		},
	}

	opts := options.FindOne().SetProjection(projection)

	var doc models.OHLCVDocument
	err := r.collection.FindOne(ctx, filter, opts).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return []models.Candle{}, nil
		}
		return nil, fmt.Errorf("failed to find candles: %w", err)
	}

	return doc.Candles, nil
}

// DeleteByJob deletes the OHLCV document for a specific job
func (r *OHLCVRepository) DeleteByJob(ctx context.Context, exchangeID, symbol, timeframe string) error {
	filter := bson.M{
		"exchange_id": exchangeID,
		"symbol":      symbol,
		"timeframe":   timeframe,
	}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete OHLCV document: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("no document found to delete")
	}

	return nil
}

// BulkInsert is deprecated - use UpsertCandles instead
// Kept for backward compatibility during migration
func (r *OHLCVRepository) BulkInsert(ctx context.Context, ohlcvList []models.Candle, exchangeID, symbol, timeframe string) (int, error) {
	log.Printf("[OHLCV_REPO] BulkInsert called (deprecated) - redirecting to UpsertCandles")
	return r.UpsertCandles(ctx, exchangeID, symbol, timeframe, ohlcvList)
}
