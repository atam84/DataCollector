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
	collection      *mongo.Collection // Legacy single-document collection
	chunksCollection *mongo.Collection // New chunked storage collection
}

// NewOHLCVRepository creates a new OHLCV repository
func NewOHLCVRepository(db *Database) *OHLCVRepository {
	collection := db.GetCollection("ohlcv")
	chunksCollection := db.GetCollection("ohlcv_chunks")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Legacy index for backward compatibility
	legacyIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "exchange_id", Value: 1},
			{Key: "symbol", Value: 1},
			{Key: "timeframe", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	_, err := collection.Indexes().CreateOne(ctx, legacyIndex)
	if err != nil {
		log.Printf("[OHLCV_REPO] Warning: Failed to create legacy index: %v", err)
	}

	// Chunked storage index: unique on (exchange_id, symbol, timeframe, year_month)
	chunkIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "exchange_id", Value: 1},
			{Key: "symbol", Value: 1},
			{Key: "timeframe", Value: 1},
			{Key: "year_month", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	_, err = chunksCollection.Indexes().CreateOne(ctx, chunkIndex)
	if err != nil {
		log.Printf("[OHLCV_REPO] Warning: Failed to create chunks index: %v", err)
	}

	// Additional index for time-based queries on chunks
	timeIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "exchange_id", Value: 1},
			{Key: "symbol", Value: 1},
			{Key: "timeframe", Value: 1},
			{Key: "start_time", Value: -1},
		},
	}
	_, err = chunksCollection.Indexes().CreateOne(ctx, timeIndex)
	if err != nil {
		log.Printf("[OHLCV_REPO] Warning: Failed to create time index: %v", err)
	}

	return &OHLCVRepository{
		collection:      collection,
		chunksCollection: chunksCollection,
	}
}

// UpsertCandles inserts new candles using chunked storage (by month)
// This avoids the MongoDB 16MB document size limit
func (r *OHLCVRepository) UpsertCandles(ctx context.Context, exchangeID, symbol, timeframe string, newCandles []models.Candle) (int, error) {
	if len(newCandles) == 0 {
		return 0, nil
	}

	log.Printf("[OHLCV_REPO] Upserting %d candles for %s-%s-%s using chunked storage", len(newCandles), exchangeID, symbol, timeframe)

	// Group candles by year-month
	candlesByMonth := make(map[string][]models.Candle)
	for _, candle := range newCandles {
		yearMonth := models.GetYearMonthFromTimestamp(candle.Timestamp)
		candlesByMonth[yearMonth] = append(candlesByMonth[yearMonth], candle)
	}

	log.Printf("[OHLCV_REPO] Candles grouped into %d monthly chunks", len(candlesByMonth))

	totalUpserted := 0

	// Upsert each monthly chunk
	for yearMonth, candles := range candlesByMonth {
		count, err := r.upsertChunk(ctx, exchangeID, symbol, timeframe, yearMonth, candles)
		if err != nil {
			log.Printf("[OHLCV_REPO] Error upserting chunk %s: %v", yearMonth, err)
			// Continue with other chunks even if one fails
			continue
		}
		totalUpserted += count
	}

	log.Printf("[OHLCV_REPO] Successfully upserted %d candles across %d chunks", totalUpserted, len(candlesByMonth))
	return totalUpserted, nil
}

// upsertChunk inserts or updates candles for a specific month chunk
func (r *OHLCVRepository) upsertChunk(ctx context.Context, exchangeID, symbol, timeframe, yearMonth string, candles []models.Candle) (int, error) {
	if len(candles) == 0 {
		return 0, nil
	}

	filter := bson.M{
		"exchange_id": exchangeID,
		"symbol":      symbol,
		"timeframe":   timeframe,
		"year_month":  yearMonth,
	}

	now := time.Now()

	// Find existing chunk
	var existingChunk models.OHLCVChunk
	err := r.chunksCollection.FindOne(ctx, filter).Decode(&existingChunk)

	if err == mongo.ErrNoDocuments {
		// Create new chunk
		// Sort candles by timestamp descending (newest first)
		sortCandlesDesc(candles)

		// Find start and end times
		startTime := time.UnixMilli(candles[len(candles)-1].Timestamp) // Oldest
		endTime := time.UnixMilli(candles[0].Timestamp)                 // Newest

		chunk := models.OHLCVChunk{
			ID:           primitive.NewObjectID(),
			ExchangeID:   exchangeID,
			Symbol:       symbol,
			Timeframe:    timeframe,
			YearMonth:    yearMonth,
			StartTime:    startTime,
			EndTime:      endTime,
			CreatedAt:    now,
			UpdatedAt:    now,
			CandlesCount: len(candles),
			Candles:      candles,
		}

		_, err := r.chunksCollection.InsertOne(ctx, chunk)
		if err != nil {
			return 0, fmt.Errorf("failed to insert chunk %s: %w", yearMonth, err)
		}

		log.Printf("[OHLCV_REPO] Created new chunk %s with %d candles", yearMonth, len(candles))
		return len(candles), nil
	} else if err != nil {
		return 0, fmt.Errorf("failed to find chunk: %w", err)
	}

	// Merge new candles with existing ones, avoiding duplicates
	existingTimestamps := make(map[int64]bool)
	for _, c := range existingChunk.Candles {
		existingTimestamps[c.Timestamp] = true
	}

	newUniqueCandles := make([]models.Candle, 0)
	for _, c := range candles {
		if !existingTimestamps[c.Timestamp] {
			newUniqueCandles = append(newUniqueCandles, c)
		}
	}

	if len(newUniqueCandles) == 0 {
		log.Printf("[OHLCV_REPO] No new unique candles for chunk %s", yearMonth)
		return 0, nil
	}

	// Merge and sort all candles
	allCandles := append(existingChunk.Candles, newUniqueCandles...)
	sortCandlesDesc(allCandles)

	// Find new start and end times
	startTime := time.UnixMilli(allCandles[len(allCandles)-1].Timestamp)
	endTime := time.UnixMilli(allCandles[0].Timestamp)

	// Replace the entire candles array (more efficient than $push for merging)
	update := bson.M{
		"$set": bson.M{
			"candles":       allCandles,
			"candles_count": len(allCandles),
			"start_time":    startTime,
			"end_time":      endTime,
			"updated_at":    now,
		},
	}

	_, err = r.chunksCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("failed to update chunk %s: %w", yearMonth, err)
	}

	log.Printf("[OHLCV_REPO] Updated chunk %s with %d new candles (total: %d)", yearMonth, len(newUniqueCandles), len(allCandles))
	return len(newUniqueCandles), nil
}

// sortCandlesDesc sorts candles by timestamp in descending order (newest first)
func sortCandlesDesc(candles []models.Candle) {
	for i := 0; i < len(candles)-1; i++ {
		for j := i + 1; j < len(candles); j++ {
			if candles[i].Timestamp < candles[j].Timestamp {
				candles[i], candles[j] = candles[j], candles[i]
			}
		}
	}
}

// UpsertCandlesLegacy inserts candles into the legacy single-document storage
// DEPRECATED: Use UpsertCandles for chunked storage
func (r *OHLCVRepository) UpsertCandlesLegacy(ctx context.Context, exchangeID, symbol, timeframe string, newCandles []models.Candle) (int, error) {
	if len(newCandles) == 0 {
		return 0, nil
	}

	log.Printf("[OHLCV_REPO] [LEGACY] Upserting %d candles for %s-%s-%s", len(newCandles), exchangeID, symbol, timeframe)

	filter := bson.M{
		"exchange_id": exchangeID,
		"symbol":      symbol,
		"timeframe":   timeframe,
	}

	now := time.Now()

	var existingDoc models.OHLCVDocument
	err := r.collection.FindOne(ctx, filter).Decode(&existingDoc)

	if err == mongo.ErrNoDocuments {
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

	update := bson.M{
		"$push": bson.M{
			"candles": bson.M{
				"$each":     newCandles,
				"$position": 0,
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
		return 0, nil
	}

	return len(newCandles), nil
}

// FindByJob retrieves all candles for a specific job by aggregating chunks
// Returns an OHLCVDocument for backward compatibility
func (r *OHLCVRepository) FindByJob(ctx context.Context, exchangeID, symbol, timeframe string) (*models.OHLCVDocument, error) {
	filter := bson.M{
		"exchange_id": exchangeID,
		"symbol":      symbol,
		"timeframe":   timeframe,
	}

	// First try chunked storage
	chunks, err := r.findAllChunks(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(chunks) > 0 {
		// Aggregate all candles from chunks
		var allCandles []models.Candle
		for _, chunk := range chunks {
			allCandles = append(allCandles, chunk.Candles...)
		}

		// Sort by timestamp descending (newest first)
		sortCandlesDesc(allCandles)

		return &models.OHLCVDocument{
			ExchangeID:   exchangeID,
			Symbol:       symbol,
			Timeframe:    timeframe,
			CandlesCount: len(allCandles),
			Candles:      allCandles,
			UpdatedAt:    chunks[0].UpdatedAt,
		}, nil
	}

	// Fall back to legacy storage
	var doc models.OHLCVDocument
	err = r.collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No document yet
		}
		return nil, fmt.Errorf("failed to find OHLCV document: %w", err)
	}

	return &doc, nil
}

// findAllChunks retrieves all chunks for a given filter
func (r *OHLCVRepository) findAllChunks(ctx context.Context, filter bson.M) ([]models.OHLCVChunk, error) {
	// Sort by year_month descending to get newest chunks first
	opts := options.Find().SetSort(bson.D{{Key: "year_month", Value: -1}})

	cursor, err := r.chunksCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find chunks: %w", err)
	}
	defer cursor.Close(ctx)

	var chunks []models.OHLCVChunk
	if err := cursor.All(ctx, &chunks); err != nil {
		return nil, fmt.Errorf("failed to decode chunks: %w", err)
	}

	return chunks, nil
}

// GetLastCandle retrieves the most recent candle for a job
// Since newest candles are at index 0, we return candles[0]
func (r *OHLCVRepository) GetLastCandle(ctx context.Context, exchangeID, symbol, timeframe string) (*models.Candle, error) {
	// First try chunked storage - get most recent chunk
	filter := bson.M{
		"exchange_id": exchangeID,
		"symbol":      symbol,
		"timeframe":   timeframe,
	}

	opts := options.FindOne().SetSort(bson.D{{Key: "end_time", Value: -1}})

	var chunk models.OHLCVChunk
	err := r.chunksCollection.FindOne(ctx, filter, opts).Decode(&chunk)
	if err == nil && len(chunk.Candles) > 0 {
		return &chunk.Candles[0], nil // Newest candle is first
	}

	// Fall back to legacy storage
	doc, err := r.FindByJob(ctx, exchangeID, symbol, timeframe)
	if err != nil {
		return nil, err
	}

	if doc == nil || len(doc.Candles) == 0 {
		return nil, nil // No candles yet
	}

	return &doc.Candles[0], nil
}

// GetCandlesCount returns the number of candles for a specific job
func (r *OHLCVRepository) GetCandlesCount(ctx context.Context, exchangeID, symbol, timeframe string) (int, error) {
	filter := bson.M{
		"exchange_id": exchangeID,
		"symbol":      symbol,
		"timeframe":   timeframe,
	}

	// First try chunked storage - sum up all chunk counts
	chunks, err := r.findAllChunks(ctx, filter)
	if err != nil {
		return 0, err
	}

	if len(chunks) > 0 {
		totalCount := 0
		for _, chunk := range chunks {
			totalCount += chunk.CandlesCount
		}
		return totalCount, nil
	}

	// Fall back to legacy storage
	var doc models.OHLCVDocument
	err = r.collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, nil
		}
		return 0, err
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

	// First try chunked storage
	chunks, err := r.findAllChunks(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(chunks) > 0 {
		// Collect candles from chunks until we have enough
		var result []models.Candle
		for _, chunk := range chunks {
			for _, candle := range chunk.Candles {
				result = append(result, candle)
				if len(result) >= limit {
					return result[:limit], nil
				}
			}
		}
		return result, nil
	}

	// Fall back to legacy storage
	projection := bson.M{
		"candles": bson.M{
			"$slice": limit,
		},
	}

	opts := options.FindOne().SetProjection(projection)

	var doc models.OHLCVDocument
	err = r.collection.FindOne(ctx, filter, opts).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return []models.Candle{}, nil
		}
		return nil, fmt.Errorf("failed to find candles: %w", err)
	}

	return doc.Candles, nil
}

// DeleteByJob deletes all OHLCV data (chunks and legacy) for a specific job
func (r *OHLCVRepository) DeleteByJob(ctx context.Context, exchangeID, symbol, timeframe string) error {
	filter := bson.M{
		"exchange_id": exchangeID,
		"symbol":      symbol,
		"timeframe":   timeframe,
	}

	// Delete from chunked storage
	chunksResult, err := r.chunksCollection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete OHLCV chunks: %w", err)
	}
	log.Printf("[OHLCV_REPO] Deleted %d chunks for %s-%s-%s", chunksResult.DeletedCount, exchangeID, symbol, timeframe)

	// Also delete from legacy storage
	legacyResult, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete legacy OHLCV document: %w", err)
	}
	log.Printf("[OHLCV_REPO] Deleted %d legacy documents for %s-%s-%s", legacyResult.DeletedCount, exchangeID, symbol, timeframe)

	if chunksResult.DeletedCount == 0 && legacyResult.DeletedCount == 0 {
		return fmt.Errorf("no data found to delete")
	}

	return nil
}

// BulkInsert is deprecated - use UpsertCandles instead
// Kept for backward compatibility during migration
func (r *OHLCVRepository) BulkInsert(ctx context.Context, ohlcvList []models.Candle, exchangeID, symbol, timeframe string) (int, error) {
	log.Printf("[OHLCV_REPO] BulkInsert called (deprecated) - redirecting to UpsertCandles")
	return r.UpsertCandles(ctx, exchangeID, symbol, timeframe, ohlcvList)
}

// Count returns the total number of candles for the given filter
func (r *OHLCVRepository) Count(ctx context.Context, filter bson.M) (int64, error) {
	// First try chunked storage
	chunks, err := r.findAllChunks(ctx, filter)
	if err != nil {
		return 0, err
	}

	if len(chunks) > 0 {
		var total int64
		for _, chunk := range chunks {
			total += int64(chunk.CandlesCount)
		}
		return total, nil
	}

	// Fall back to legacy storage
	var doc models.OHLCVDocument
	err = r.collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to count candles: %w", err)
	}

	return int64(len(doc.Candles)), nil
}

// GetStatsByExchange returns aggregate statistics for an exchange
func (r *OHLCVRepository) GetStatsByExchange(ctx context.Context, exchangeID string) (*models.OHLCVStats, error) {
	filter := bson.M{"exchange_id": exchangeID}

	// Get all chunks for this exchange
	chunks, err := r.findAllChunks(ctx, filter)
	if err != nil {
		return nil, err
	}

	stats := &models.OHLCVStats{
		ExchangeID: exchangeID,
	}

	// Count unique symbols and timeframes
	symbols := make(map[string]bool)
	timeframes := make(map[string]bool)

	for _, chunk := range chunks {
		stats.TotalCandles += int64(chunk.CandlesCount)
		stats.TotalChunks++
		symbols[chunk.Symbol] = true
		timeframes[chunk.Timeframe] = true

		if stats.OldestData.IsZero() || chunk.StartTime.Before(stats.OldestData) {
			stats.OldestData = chunk.StartTime
		}
		if chunk.EndTime.After(stats.NewestData) {
			stats.NewestData = chunk.EndTime
		}
	}

	stats.UniqueSymbols = len(symbols)
	stats.UniqueTimeframes = len(timeframes)

	// Also check legacy storage
	legacyFilter := bson.M{"exchange_id": exchangeID}
	cursor, err := r.collection.Find(ctx, legacyFilter)
	if err == nil {
		defer cursor.Close(ctx)
		var legacyDocs []models.OHLCVDocument
		if cursor.All(ctx, &legacyDocs) == nil {
			for _, doc := range legacyDocs {
				stats.TotalCandles += int64(doc.CandlesCount)
				stats.LegacyDocuments++
				symbols[doc.Symbol] = true
				timeframes[doc.Timeframe] = true
			}
			stats.UniqueSymbols = len(symbols)
			stats.UniqueTimeframes = len(timeframes)
		}
	}

	return stats, nil
}

// GetAllStats returns aggregate statistics across all exchanges
func (r *OHLCVRepository) GetAllStats(ctx context.Context) (*models.OHLCVStats, error) {
	stats := &models.OHLCVStats{}

	// Get stats from chunked storage using aggregation
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":           nil,
				"total_candles": bson.M{"$sum": "$candles_count"},
				"total_chunks":  bson.M{"$sum": 1},
				"oldest_data":   bson.M{"$min": "$start_time"},
				"newest_data":   bson.M{"$max": "$end_time"},
				"exchanges":     bson.M{"$addToSet": "$exchange_id"},
				"symbols":       bson.M{"$addToSet": "$symbol"},
				"timeframes":    bson.M{"$addToSet": "$timeframe"},
			},
		},
	}

	cursor, err := r.chunksCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate stats: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}

	if len(results) > 0 {
		result := results[0]
		if v, ok := result["total_candles"].(int64); ok {
			stats.TotalCandles = v
		} else if v, ok := result["total_candles"].(int32); ok {
			stats.TotalCandles = int64(v)
		}
		if v, ok := result["total_chunks"].(int64); ok {
			stats.TotalChunks = int(v)
		} else if v, ok := result["total_chunks"].(int32); ok {
			stats.TotalChunks = int(v)
		}
		if v, ok := result["oldest_data"].(primitive.DateTime); ok {
			stats.OldestData = v.Time()
		}
		if v, ok := result["newest_data"].(primitive.DateTime); ok {
			stats.NewestData = v.Time()
		}
		if v, ok := result["exchanges"].(primitive.A); ok {
			stats.UniqueExchanges = len(v)
		}
		if v, ok := result["symbols"].(primitive.A); ok {
			stats.UniqueSymbols = len(v)
		}
		if v, ok := result["timeframes"].(primitive.A); ok {
			stats.UniqueTimeframes = len(v)
		}
	}

	return stats, nil
}

// FindWithPagination returns paginated candles for the given filter
func (r *OHLCVRepository) FindWithPagination(ctx context.Context, filter bson.M, skip, limit int64) ([]models.Candle, error) {
	// First try chunked storage
	chunks, err := r.findAllChunks(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(chunks) > 0 {
		// Aggregate all candles from chunks
		var allCandles []models.Candle
		for _, chunk := range chunks {
			allCandles = append(allCandles, chunk.Candles...)
		}

		// Sort candles by timestamp descending (newest first)
		sortCandlesDesc(allCandles)

		total := int64(len(allCandles))
		if skip >= total {
			return []models.Candle{}, nil
		}

		end := skip + limit
		if end > total {
			end = total
		}

		return allCandles[skip:end], nil
	}

	// Fall back to legacy storage
	var doc models.OHLCVDocument
	err = r.collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return []models.Candle{}, nil
		}
		return nil, fmt.Errorf("failed to find candles: %w", err)
	}

	total := int64(len(doc.Candles))
	if skip >= total {
		return []models.Candle{}, nil
	}

	end := skip + limit
	if end > total {
		end = total
	}

	// Sort candles by timestamp descending (newest first)
	candles := doc.Candles
	sortCandlesDesc(candles)

	return candles[skip:end], nil
}

// AnalyzeDataQuality analyzes the data quality for a specific job's data
func (r *OHLCVRepository) AnalyzeDataQuality(ctx context.Context, exchangeID, symbol, timeframe string) (*models.DataQuality, error) {
	// Get all candles
	doc, err := r.FindByJob(ctx, exchangeID, symbol, timeframe)
	if err != nil {
		return nil, err
	}

	quality := &models.DataQuality{
		ExchangeID: exchangeID,
		Symbol:     symbol,
		Timeframe:  timeframe,
	}

	if doc == nil || len(doc.Candles) == 0 {
		quality.QualityStatus = "poor"
		quality.DataFreshness = "no_data"
		return quality, nil
	}

	quality.TotalCandles = int64(len(doc.Candles))

	// Get candles sorted by timestamp ascending for gap analysis
	candles := make([]models.Candle, len(doc.Candles))
	copy(candles, doc.Candles)
	sortCandlesAsc(candles)

	// Find oldest and newest candles
	quality.OldestCandle = time.UnixMilli(candles[0].Timestamp)
	quality.NewestCandle = time.UnixMilli(candles[len(candles)-1].Timestamp)

	// Calculate expected candles based on timeframe and date range
	timeframeDuration := models.GetTimeframeDurationMinutes(timeframe)
	totalDuration := quality.NewestCandle.Sub(quality.OldestCandle).Minutes()
	quality.ExpectedCandles = int64(totalDuration/float64(timeframeDuration)) + 1

	// Calculate missing candles
	quality.MissingCandles = quality.ExpectedCandles - quality.TotalCandles
	if quality.MissingCandles < 0 {
		quality.MissingCandles = 0 // More candles than expected (possible duplicates or overlap)
	}

	// Calculate completeness score
	if quality.ExpectedCandles > 0 {
		quality.CompletenessScore = (float64(quality.TotalCandles) / float64(quality.ExpectedCandles)) * 100
		if quality.CompletenessScore > 100 {
			quality.CompletenessScore = 100
		}
	}

	// Calculate data freshness
	freshnessMinutes := time.Since(quality.NewestCandle).Minutes()
	quality.FreshnessMinutes = int64(freshnessMinutes)

	// Freshness based on timeframe
	expectedFreshness := float64(timeframeDuration) * 2 // Allow 2x timeframe for "fresh"
	if freshnessMinutes <= expectedFreshness {
		quality.DataFreshness = "fresh"
	} else if freshnessMinutes <= expectedFreshness*5 {
		quality.DataFreshness = "stale"
	} else {
		quality.DataFreshness = "very_stale"
	}

	// Detect gaps in the data
	quality.Gaps = detectGaps(candles, timeframeDuration)
	quality.GapsDetected = len(quality.Gaps)

	// Determine overall quality status
	quality.QualityStatus = calculateQualityStatus(quality.CompletenessScore, quality.GapsDetected, quality.DataFreshness, quality.TotalCandles)

	return quality, nil
}

// sortCandlesAsc sorts candles by timestamp in ascending order (oldest first)
func sortCandlesAsc(candles []models.Candle) {
	for i := 0; i < len(candles)-1; i++ {
		for j := i + 1; j < len(candles); j++ {
			if candles[i].Timestamp > candles[j].Timestamp {
				candles[i], candles[j] = candles[j], candles[i]
			}
		}
	}
}

// detectGaps finds gaps in the candle data
func detectGaps(candles []models.Candle, timeframeDurationMinutes int64) []models.DataGap {
	if len(candles) < 2 {
		return nil
	}

	var gaps []models.DataGap
	expectedGapMs := timeframeDurationMinutes * 60 * 1000 // Convert minutes to milliseconds
	tolerance := expectedGapMs + (expectedGapMs / 10)     // Allow 10% tolerance

	for i := 1; i < len(candles); i++ {
		actualGap := candles[i].Timestamp - candles[i-1].Timestamp

		// If gap is more than expected (with tolerance), we have missing candles
		if actualGap > tolerance {
			missingCount := int((actualGap / expectedGapMs) - 1)
			if missingCount > 0 {
				gap := models.DataGap{
					StartTime:       time.UnixMilli(candles[i-1].Timestamp),
					EndTime:         time.UnixMilli(candles[i].Timestamp),
					MissingCandles:  missingCount,
					DurationMinutes: actualGap / (60 * 1000),
				}
				gaps = append(gaps, gap)
			}
		}
	}

	return gaps
}

// calculateQualityStatus determines the overall quality status
// Minimum data requirements:
// - "excellent" requires at least 1000 candles
// - "good" requires at least 500 candles
// - "fair" requires at least 100 candles
// Jobs with fewer than 100 candles are always "poor" (insufficient data)
func calculateQualityStatus(completeness float64, gapsCount int, freshness string, totalCandles int64) string {
	// Minimum data requirements - insufficient data is always "poor"
	if totalCandles < 100 {
		return "poor"
	}

	if completeness >= 99 && gapsCount == 0 && freshness == "fresh" && totalCandles >= 1000 {
		return "excellent"
	}
	if completeness >= 95 && gapsCount <= 2 && freshness != "very_stale" && totalCandles >= 500 {
		return "good"
	}
	if completeness >= 80 && gapsCount <= 5 && totalCandles >= 100 {
		return "fair"
	}
	return "poor"
}

// GetDataQualitySummary returns aggregated data quality metrics for an exchange
func (r *OHLCVRepository) GetDataQualitySummary(ctx context.Context, exchangeID string) (*models.DataQualitySummary, error) {
	filter := bson.M{}
	if exchangeID != "" {
		filter["exchange_id"] = exchangeID
	}

	// Get all unique combinations of exchange/symbol/timeframe
	pipeline := []bson.M{
		{"$match": filter},
		{
			"$group": bson.M{
				"_id": bson.M{
					"exchange_id": "$exchange_id",
					"symbol":      "$symbol",
					"timeframe":   "$timeframe",
				},
			},
		},
	}

	cursor, err := r.chunksCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate jobs: %w", err)
	}
	defer cursor.Close(ctx)

	var jobs []bson.M
	if err := cursor.All(ctx, &jobs); err != nil {
		return nil, fmt.Errorf("failed to decode jobs: %w", err)
	}

	summary := &models.DataQualitySummary{
		TotalJobs: len(jobs),
	}

	var totalCompleteness float64

	for _, job := range jobs {
		id := job["_id"].(bson.M)
		exchangeID := id["exchange_id"].(string)
		symbol := id["symbol"].(string)
		timeframe := id["timeframe"].(string)

		quality, err := r.AnalyzeDataQuality(ctx, exchangeID, symbol, timeframe)
		if err != nil {
			continue
		}

		totalCompleteness += quality.CompletenessScore
		summary.TotalMissingCandles += quality.MissingCandles
		summary.TotalGaps += quality.GapsDetected

		switch quality.QualityStatus {
		case "excellent":
			summary.ExcellentQuality++
		case "good":
			summary.GoodQuality++
		case "fair":
			summary.FairQuality++
		case "poor":
			summary.PoorQuality++
		}

		if quality.DataFreshness == "fresh" {
			summary.FreshDataJobs++
		} else {
			summary.StaleDataJobs++
		}
	}

	if summary.TotalJobs > 0 {
		summary.AverageCompleteness = totalCompleteness / float64(summary.TotalJobs)
	}

	return summary, nil
}
