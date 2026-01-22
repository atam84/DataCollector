package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/yourusername/datacollector/internal/exchange"
	"github.com/yourusername/datacollector/internal/models"
)

// CCXTService handles interactions with CCXT exchange library
type CCXTService struct {
	rateLimiter *RateLimiter
}

// NewCCXTService creates a new CCXT service
func NewCCXTService() *CCXTService {
	return &CCXTService{}
}

// NewCCXTServiceWithRateLimiter creates a CCXT service with rate limiting support
func NewCCXTServiceWithRateLimiter(rateLimiter *RateLimiter) *CCXTService {
	return &CCXTService{
		rateLimiter: rateLimiter,
	}
}

// FetchOHLCVData fetches real OHLCV data from an exchange using CCXT
// First fetch (sinceMs = nil): Fetches ALL available historical data with pagination
// Subsequent fetch (sinceMs != nil): Fetches all data since timestamp
func (s *CCXTService) FetchOHLCVData(
	exchangeID string,
	symbol string,
	timeframe string,
	sinceMs *int64, // nil for first fetch (get all history), timestamp for subsequent
) ([]models.Candle, error) {
	// Use background context if no rate limiter (backward compatibility)
	return s.FetchOHLCVDataWithContext(context.Background(), exchangeID, symbol, timeframe, sinceMs)
}

// FetchOHLCVDataWithContext fetches OHLCV data with context support for rate limiting
func (s *CCXTService) FetchOHLCVDataWithContext(
	ctx context.Context,
	exchangeID string,
	symbol string,
	timeframe string,
	sinceMs *int64,
) ([]models.Candle, error) {

	// Apply rate limiting before creating adapter (LoadMarkets is an API call)
	if s.rateLimiter != nil {
		if err := s.rateLimiter.WaitForSlot(ctx, exchangeID); err != nil {
			return nil, fmt.Errorf("rate limit wait failed: %w", err)
		}
	}

	// Create adapter for the exchange dynamically
	adapter, err := exchange.NewCCXTAdapter(exchangeID, true)
	if err != nil {
		log.Printf("[CCXT] Failed to create adapter for %s: %v", exchangeID, err)
		return nil, fmt.Errorf("exchange %s not yet supported: %w", exchangeID, err)
	}
	defer adapter.Close()

	log.Printf("[CCXT] Successfully created adapter for %s", exchangeID)

	// Load markets first (this is an API call)
	if err := adapter.LoadMarkets(); err != nil {
		log.Printf("[CCXT] Failed to load markets for %s: %v", exchangeID, err)
		return nil, fmt.Errorf("failed to load markets: %w", err)
	}

	log.Printf("[CCXT] Markets loaded for %s", exchangeID)

	// Get exchange metadata for OHLCV limit
	metadata, err := exchange.GetExchangeMetadata(exchangeID)
	batchLimit := 500 // default
	if err == nil && metadata.OHLCVLimit > 0 {
		batchLimit = metadata.OHLCVLimit
	}
	log.Printf("[CCXT] Using batch limit of %d for %s", batchLimit, exchangeID)

	var allCandles []models.Candle

	if sinceMs == nil {
		// FIRST EXECUTION: Fetch ALL available historical data with pagination
		log.Printf("[CCXT] First execution - fetching ALL historical data for %s %s %s",
			exchangeID, symbol, timeframe)
		allCandles, err = s.fetchAllHistoricalDataWithContext(ctx, adapter, exchangeID, symbol, timeframe, batchLimit)
	} else {
		// SUBSEQUENT EXECUTION: Fetch data since timestamp
		log.Printf("[CCXT] Subsequent execution - fetching data since %d for %s %s %s",
			*sinceMs, exchangeID, symbol, timeframe)
		allCandles, err = s.fetchDataSinceWithContext(ctx, adapter, exchangeID, symbol, timeframe, *sinceMs, batchLimit)
	}

	if err != nil {
		return nil, err
	}

	log.Printf("[CCXT] Total candles fetched: %d", len(allCandles))

	// Reverse order so newest candles are at index 0
	reversed := make([]models.Candle, len(allCandles))
	for i, candle := range allCandles {
		reversed[len(allCandles)-1-i] = candle
	}

	log.Printf("[CCXT] Returning %d candles (newest first)", len(reversed))
	return reversed, nil
}

// fetchAllHistoricalData fetches all available historical data using pagination
// DEPRECATED: Use fetchAllHistoricalDataWithContext instead
func (s *CCXTService) fetchAllHistoricalData(
	adapter *exchange.CCXTAdapter,
	symbol string,
	timeframe string,
	batchLimit int,
) ([]models.Candle, error) {
	return s.fetchAllHistoricalDataWithContext(context.Background(), adapter, "", symbol, timeframe, batchLimit)
}

// Date range fallback durations (in months) for handling "date too wide" errors
var dateRangeFallbacks = []int{60, 12, 6, 3, 1} // 5 years, 1 year, 6 months, 3 months, 1 month

// isDateRangeError checks if the error is a "date too wide" or similar range error
func isDateRangeError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// Check for common date range error patterns across exchanges
	return strings.Contains(errStr, "date of query is too wide") ||
		strings.Contains(errStr, "date range") ||
		strings.Contains(errStr, "time range") ||
		strings.Contains(errStr, "range too large") ||
		strings.Contains(errStr, "max limit") ||
		strings.Contains(errStr, "exceeds maximum")
}

// fetchAllHistoricalDataWithContext fetches all available historical data with rate limiting
// Strategy: Start from oldest available and fetch forward to present
// Implements date range fallback for exchanges with range limits
func (s *CCXTService) fetchAllHistoricalDataWithContext(
	ctx context.Context,
	adapter *exchange.CCXTAdapter,
	exchangeID string,
	symbol string,
	timeframe string,
	batchLimit int,
) ([]models.Candle, error) {
	// Calculate timeframe duration in milliseconds for pagination
	tfDuration := getTimeframeDurationMs(timeframe)

	// Try with progressively shorter date ranges if we hit range limits
	for fallbackIdx, months := range dateRangeFallbacks {
		startTime := time.Now().AddDate(0, -months, 0)
		log.Printf("[CCXT] Attempting historical fetch with %d month range (fallback %d/%d)", months, fallbackIdx+1, len(dateRangeFallbacks))

		candles, err := s.fetchHistoricalFromDate(ctx, adapter, exchangeID, symbol, timeframe, startTime, batchLimit, tfDuration)

		if err != nil {
			if isDateRangeError(err) {
				log.Printf("[CCXT] Date range error with %d months, trying shorter range: %v", months, err)
				continue // Try next shorter range
			}
			// Non-range error, return what we have or the error
			if len(candles) > 0 {
				return candles, nil
			}
			return nil, err
		}

		// Success!
		return candles, nil
	}

	return nil, fmt.Errorf("failed to fetch historical data: all date range fallbacks exhausted")
}

// fetchHistoricalFromDate fetches historical data starting from a specific date
func (s *CCXTService) fetchHistoricalFromDate(
	ctx context.Context,
	adapter *exchange.CCXTAdapter,
	exchangeID string,
	symbol string,
	timeframe string,
	startTime time.Time,
	batchLimit int,
	tfDuration int64,
) ([]models.Candle, error) {
	var allCandles []models.Candle
	currentSince := startTime

	log.Printf("[CCXT] Starting historical fetch from %s", startTime.Format("2006-01-02"))

	maxIterations := 1000 // Safety limit to prevent infinite loops
	iteration := 0

	for iteration < maxIterations {
		iteration++

		// Check context cancellation
		select {
		case <-ctx.Done():
			log.Printf("[CCXT] Context cancelled, returning %d candles collected so far", len(allCandles))
			if len(allCandles) > 0 {
				return allCandles, nil
			}
			return nil, ctx.Err()
		default:
		}

		// Apply rate limiting before each API call
		if s.rateLimiter != nil && exchangeID != "" {
			if err := s.rateLimiter.WaitForSlot(ctx, exchangeID); err != nil {
				log.Printf("[CCXT] Rate limit wait failed: %v", err)
				if len(allCandles) > 0 {
					return allCandles, nil
				}
				return nil, fmt.Errorf("rate limit wait failed: %w", err)
			}
		}

		log.Printf("[CCXT] Fetching batch %d from %s", iteration, currentSince.Format("2006-01-02 15:04:05"))

		candles, err := adapter.FetchOHLCV(symbol, timeframe, &currentSince, batchLimit)
		if err != nil {
			log.Printf("[CCXT] Error fetching batch %d: %v", iteration, err)
			// If we already have some data, return it despite the error
			if len(allCandles) > 0 {
				log.Printf("[CCXT] Returning %d candles collected before error", len(allCandles))
				return allCandles, nil
			}
			return nil, fmt.Errorf("failed to fetch OHLCV: %w", err)
		}

		if len(candles) == 0 {
			log.Printf("[CCXT] No more candles returned, stopping pagination")
			break
		}

		log.Printf("[CCXT] Batch %d: fetched %d candles", iteration, len(candles))

		// Append candles (they come in chronological order - oldest first)
		allCandles = append(allCandles, candles...)

		// Get the timestamp of the last (newest) candle in this batch
		lastCandle := candles[len(candles)-1]
		lastTimestamp := time.UnixMilli(lastCandle.Timestamp)

		// Check if we've reached the present (last candle is recent)
		if time.Since(lastTimestamp) < time.Duration(tfDuration)*time.Millisecond*2 {
			log.Printf("[CCXT] Reached present time, stopping pagination")
			break
		}

		// If we got fewer candles than the limit, we've likely reached the end
		if len(candles) < batchLimit {
			log.Printf("[CCXT] Received fewer candles than limit (%d < %d), likely at end", len(candles), batchLimit)
			break
		}

		// Move forward: set next fetch to start after the last candle
		currentSince = lastTimestamp.Add(time.Duration(tfDuration) * time.Millisecond)

		// Note: Rate limiting is now handled by the RateLimiter, no fixed sleep needed
		// If no rate limiter is configured, use a minimal safety delay
		if s.rateLimiter == nil {
			time.Sleep(1 * time.Second) // Safety delay when no rate limiter
		}
	}

	log.Printf("[CCXT] Historical fetch complete: %d total candles in %d batches", len(allCandles), iteration)
	return allCandles, nil
}

// fetchDataSince fetches data from a specific timestamp to present
// DEPRECATED: Use fetchDataSinceWithContext instead
func (s *CCXTService) fetchDataSince(
	adapter *exchange.CCXTAdapter,
	symbol string,
	timeframe string,
	sinceMs int64,
	batchLimit int,
) ([]models.Candle, error) {
	return s.fetchDataSinceWithContext(context.Background(), adapter, "", symbol, timeframe, sinceMs, batchLimit)
}

// fetchDataSinceWithContext fetches data from a specific timestamp with rate limiting
func (s *CCXTService) fetchDataSinceWithContext(
	ctx context.Context,
	adapter *exchange.CCXTAdapter,
	exchangeID string,
	symbol string,
	timeframe string,
	sinceMs int64,
	batchLimit int,
) ([]models.Candle, error) {
	var allCandles []models.Candle

	tfDuration := getTimeframeDurationMs(timeframe)
	currentSince := time.UnixMilli(sinceMs)

	log.Printf("[CCXT] Fetching data since %s", currentSince.Format("2006-01-02 15:04:05"))

	maxIterations := 100 // Fewer iterations needed for incremental updates
	iteration := 0

	for iteration < maxIterations {
		iteration++

		// Check context cancellation
		select {
		case <-ctx.Done():
			log.Printf("[CCXT] Context cancelled, returning %d candles collected so far", len(allCandles))
			if len(allCandles) > 0 {
				return allCandles, nil
			}
			return nil, ctx.Err()
		default:
		}

		// Apply rate limiting before each API call
		if s.rateLimiter != nil && exchangeID != "" {
			if err := s.rateLimiter.WaitForSlot(ctx, exchangeID); err != nil {
				log.Printf("[CCXT] Rate limit wait failed: %v", err)
				if len(allCandles) > 0 {
					return allCandles, nil
				}
				return nil, fmt.Errorf("rate limit wait failed: %w", err)
			}
		}

		candles, err := adapter.FetchOHLCV(symbol, timeframe, &currentSince, batchLimit)
		if err != nil {
			if len(allCandles) > 0 {
				return allCandles, nil
			}
			return nil, fmt.Errorf("failed to fetch OHLCV: %w", err)
		}

		if len(candles) == 0 {
			break
		}

		log.Printf("[CCXT] Batch %d: fetched %d candles", iteration, len(candles))
		allCandles = append(allCandles, candles...)

		// Check if we've reached present
		lastCandle := candles[len(candles)-1]
		lastTimestamp := time.UnixMilli(lastCandle.Timestamp)

		if time.Since(lastTimestamp) < time.Duration(tfDuration)*time.Millisecond*2 {
			break
		}

		if len(candles) < batchLimit {
			break
		}

		currentSince = lastTimestamp.Add(time.Duration(tfDuration) * time.Millisecond)

		// Note: Rate limiting is now handled by the RateLimiter, no fixed sleep needed
		// If no rate limiter is configured, use a minimal safety delay
		if s.rateLimiter == nil {
			time.Sleep(1 * time.Second) // Safety delay when no rate limiter
		}
	}

	log.Printf("[CCXT] Incremental fetch complete: %d candles", len(allCandles))
	return allCandles, nil
}

// getTimeframeDurationMs returns the duration of a timeframe in milliseconds
func getTimeframeDurationMs(timeframe string) int64 {
	switch timeframe {
	case "1s":
		return 1000
	case "1m":
		return 60 * 1000
	case "3m":
		return 3 * 60 * 1000
	case "5m":
		return 5 * 60 * 1000
	case "15m":
		return 15 * 60 * 1000
	case "30m":
		return 30 * 60 * 1000
	case "1h":
		return 60 * 60 * 1000
	case "2h":
		return 2 * 60 * 60 * 1000
	case "4h":
		return 4 * 60 * 60 * 1000
	case "6h":
		return 6 * 60 * 60 * 1000
	case "8h":
		return 8 * 60 * 60 * 1000
	case "12h":
		return 12 * 60 * 60 * 1000
	case "1d":
		return 24 * 60 * 60 * 1000
	case "3d":
		return 3 * 24 * 60 * 60 * 1000
	case "1w":
		return 7 * 24 * 60 * 60 * 1000
	case "1M":
		return 30 * 24 * 60 * 60 * 1000
	default:
		// Default to 1 hour
		return 60 * 60 * 1000
	}
}

// FetchOHLCVRange fetches OHLCV data for a specific time range (used for gap filling)
func (s *CCXTService) FetchOHLCVRange(
	ctx context.Context,
	connector *models.Connector,
	symbol string,
	timeframe string,
	startMs int64,
	endMs int64,
) ([]models.Candle, error) {
	exchangeID := connector.ExchangeID

	// Apply rate limiting before creating adapter
	if s.rateLimiter != nil {
		if err := s.rateLimiter.WaitForSlot(ctx, exchangeID); err != nil {
			return nil, fmt.Errorf("rate limit wait failed: %w", err)
		}
	}

	// Create adapter for the exchange
	adapter, err := exchange.NewCCXTAdapter(exchangeID, true)
	if err != nil {
		return nil, fmt.Errorf("exchange %s not yet supported: %w", exchangeID, err)
	}
	defer adapter.Close()

	// Load markets
	if err := adapter.LoadMarkets(); err != nil {
		return nil, fmt.Errorf("failed to load markets: %w", err)
	}

	// Get exchange metadata for OHLCV limit
	metadata, err := exchange.GetExchangeMetadata(exchangeID)
	batchLimit := 500
	if err == nil && metadata.OHLCVLimit > 0 {
		batchLimit = metadata.OHLCVLimit
	}

	var allCandles []models.Candle
	tfDuration := getTimeframeDurationMs(timeframe)
	currentSince := time.UnixMilli(startMs)
	endTime := time.UnixMilli(endMs)

	log.Printf("[CCXT] Fetching range %s to %s for %s %s %s",
		currentSince.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"),
		exchangeID, symbol, timeframe)

	maxIterations := 100
	for i := 0; i < maxIterations; i++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			if len(allCandles) > 0 {
				return allCandles, nil
			}
			return nil, ctx.Err()
		default:
		}

		// Apply rate limiting
		if s.rateLimiter != nil {
			if err := s.rateLimiter.WaitForSlot(ctx, exchangeID); err != nil {
				if len(allCandles) > 0 {
					return allCandles, nil
				}
				return nil, fmt.Errorf("rate limit wait failed: %w", err)
			}
		}

		candles, err := adapter.FetchOHLCV(symbol, timeframe, &currentSince, batchLimit)
		if err != nil {
			if len(allCandles) > 0 {
				return allCandles, nil
			}
			return nil, fmt.Errorf("failed to fetch OHLCV: %w", err)
		}

		if len(candles) == 0 {
			break
		}

		// Filter candles within our range and append
		for _, candle := range candles {
			candleTime := time.UnixMilli(candle.Timestamp)
			if candleTime.After(endTime) {
				// We've passed the end time, stop
				goto done
			}
			allCandles = append(allCandles, candle)
		}

		// Move to next batch
		lastCandle := candles[len(candles)-1]
		lastTimestamp := time.UnixMilli(lastCandle.Timestamp)

		if lastTimestamp.After(endTime) || lastTimestamp.Equal(endTime) {
			break
		}

		if len(candles) < batchLimit {
			break
		}

		currentSince = lastTimestamp.Add(time.Duration(tfDuration) * time.Millisecond)

		if s.rateLimiter == nil {
			time.Sleep(1 * time.Second)
		}
	}

done:
	log.Printf("[CCXT] Range fetch complete: %d candles", len(allCandles))

	// Reverse order so newest candles are at index 0
	reversed := make([]models.Candle, len(allCandles))
	for i, candle := range allCandles {
		reversed[len(allCandles)-1-i] = candle
	}

	return reversed, nil
}
