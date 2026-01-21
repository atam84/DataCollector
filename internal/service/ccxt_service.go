package service

import (
	"fmt"
	"log"
	"time"

	"github.com/yourusername/datacollector/internal/exchange"
	"github.com/yourusername/datacollector/internal/models"
)

// CCXTService handles interactions with CCXT exchange library
type CCXTService struct{}

// NewCCXTService creates a new CCXT service
func NewCCXTService() *CCXTService {
	return &CCXTService{}
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

	// Create adapter for the exchange dynamically
	adapter, err := exchange.NewCCXTAdapter(exchangeID, true)
	if err != nil {
		log.Printf("[CCXT] Failed to create adapter for %s: %v", exchangeID, err)
		return nil, fmt.Errorf("exchange %s not yet supported: %w", exchangeID, err)
	}
	defer adapter.Close()

	log.Printf("[CCXT] Successfully created adapter for %s", exchangeID)

	// Load markets first
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
		allCandles, err = s.fetchAllHistoricalData(adapter, symbol, timeframe, batchLimit)
	} else {
		// SUBSEQUENT EXECUTION: Fetch data since timestamp
		log.Printf("[CCXT] Subsequent execution - fetching data since %d for %s %s %s",
			*sinceMs, exchangeID, symbol, timeframe)
		allCandles, err = s.fetchDataSince(adapter, symbol, timeframe, *sinceMs, batchLimit)
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
// Strategy: Start from oldest available and fetch forward to present
func (s *CCXTService) fetchAllHistoricalData(
	adapter *exchange.CCXTAdapter,
	symbol string,
	timeframe string,
	batchLimit int,
) ([]models.Candle, error) {
	var allCandles []models.Candle

	// Calculate timeframe duration in milliseconds for pagination
	tfDuration := getTimeframeDurationMs(timeframe)

	// Start from a long time ago (5 years) to get all available data
	// Most exchanges don't have data that old, so they'll return from their earliest available
	startTime := time.Now().AddDate(-5, 0, 0) // 5 years ago
	currentSince := startTime

	log.Printf("[CCXT] Starting historical fetch from %s", startTime.Format("2006-01-02"))

	maxIterations := 1000 // Safety limit to prevent infinite loops
	iteration := 0

	for iteration < maxIterations {
		iteration++

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

		// Small delay to respect rate limits
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("[CCXT] Historical fetch complete: %d total candles in %d batches", len(allCandles), iteration)
	return allCandles, nil
}

// fetchDataSince fetches data from a specific timestamp to present
func (s *CCXTService) fetchDataSince(
	adapter *exchange.CCXTAdapter,
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
		time.Sleep(100 * time.Millisecond)
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
