package service

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ccxt/ccxt/go/v4"
	"github.com/yourusername/datacollector/internal/models"
)

// CCXTService handles interactions with CCXT exchange library
type CCXTService struct{}

// NewCCXTService creates a new CCXT service
func NewCCXTService() *CCXTService {
	return &CCXTService{}
}

// FetchOHLCVData fetches real OHLCV data from an exchange using CCXT
// First fetch (sinceMs = nil): Fetches ALL available data without limit
// Subsequent fetch (sinceMs != nil): Fetches all data since timestamp without limit
func (s *CCXTService) FetchOHLCVData(
	exchangeID string,
	symbol string,
	timeframe string,
	sinceMs *int64, // nil for first fetch, timestamp for subsequent
) ([]models.Candle, error) {

	if sinceMs == nil {
		log.Printf("[CCXT] Fetching %s %s %s - FIRST EXECUTION (ALL available data, no limit)",
			exchangeID, symbol, timeframe)
	} else {
		log.Printf("[CCXT] Fetching %s %s %s since timestamp %d (no limit)",
			exchangeID, symbol, timeframe, *sinceMs)
	}

	// Fetch OHLCV based on exchange
	var ohlcvData interface{}
	var err error

	switch strings.ToLower(exchangeID) {
	case "bybit":
		ohlcvData, err = s.fetchFromBybit(symbol, timeframe, sinceMs)
	case "binance":
		ohlcvData, err = s.fetchFromBinance(symbol, timeframe, sinceMs)
	default:
		return nil, fmt.Errorf("exchange %s not yet supported", exchangeID)
	}

	if err != nil {
		return nil, err
	}

	// Convert CCXT OHLCV to our model
	result, err := s.convertOHLCVData(ohlcvData)
	if err != nil {
		return nil, err
	}

	log.Printf("[CCXT] Successfully converted %d candles", len(result))
	return result, nil
}

// fetchFromBybit fetches OHLCV data from Bybit
func (s *CCXTService) fetchFromBybit(symbol string, timeframe string, sinceMs *int64) (interface{}, error) {
	exchange := ccxt.NewBybit(nil)

	// Load API credentials from environment if available
	apiKey := os.Getenv("BYBIT_API_KEY")
	apiSecret := os.Getenv("BYBIT_API_SECRET")

	if apiKey != "" && apiSecret != "" {
		log.Printf("[CCXT] Using API credentials for Bybit")
		exchange.SetApiKey(apiKey)
		exchange.SetSecret(apiSecret)
	}

	// Build options - only include since if provided, NEVER include limit
	var data interface{}
	var err error

	if sinceMs == nil {
		// First fetch - no parameters
		log.Printf("[CCXT] Calling Bybit.FetchOHLCV(%s, timeframe=%s) - NO since, NO limit", symbol, timeframe)
		data, err = exchange.FetchOHLCV(symbol,
			ccxt.WithFetchOHLCVTimeframe(timeframe),
		)
	} else {
		// Subsequent fetch - with since, no limit
		log.Printf("[CCXT] Calling Bybit.FetchOHLCV(%s, timeframe=%s, since=%d) - NO limit", symbol, timeframe, *sinceMs)
		data, err = exchange.FetchOHLCV(symbol,
			ccxt.WithFetchOHLCVTimeframe(timeframe),
			ccxt.WithFetchOHLCVSince(*sinceMs),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("bybit FetchOHLCV failed: %w", err)
	}
	return data, nil
}

// fetchFromBinance fetches OHLCV data from Binance
func (s *CCXTService) fetchFromBinance(symbol string, timeframe string, sinceMs *int64) (interface{}, error) {
	exchange := ccxt.NewBinance(nil)

	// Load API credentials from environment if available
	apiKey := os.Getenv("BINANCE_API_KEY")
	apiSecret := os.Getenv("BINANCE_API_SECRET")

	if apiKey != "" && apiSecret != "" {
		log.Printf("[CCXT] Using API credentials for Binance")
		exchange.SetApiKey(apiKey)
		exchange.SetSecret(apiSecret)
	}

	// Build options - only include since if provided, NEVER include limit
	var data interface{}
	var err error

	if sinceMs == nil {
		// First fetch - no parameters
		log.Printf("[CCXT] Calling Binance.FetchOHLCV(%s, timeframe=%s) - NO since, NO limit", symbol, timeframe)
		data, err = exchange.FetchOHLCV(symbol,
			ccxt.WithFetchOHLCVTimeframe(timeframe),
		)
	} else {
		// Subsequent fetch - with since, no limit
		log.Printf("[CCXT] Calling Binance.FetchOHLCV(%s, timeframe=%s, since=%d) - NO limit", symbol, timeframe, *sinceMs)
		data, err = exchange.FetchOHLCV(symbol,
			ccxt.WithFetchOHLCVTimeframe(timeframe),
			ccxt.WithFetchOHLCVSince(*sinceMs),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("binance FetchOHLCV failed: %w", err)
	}
	return data, nil
}

// convertOHLCVData converts CCXT OHLCV response to our Candle model
// CCXT returns candles in chronological order (oldest first)
// We reverse the order so newest candles are at index 0
func (s *CCXTService) convertOHLCVData(ohlcvData interface{}) ([]models.Candle, error) {
	// CCXT returns []ccxt.OHLCV
	dataSlice, ok := ohlcvData.([]ccxt.OHLCV)
	if !ok {
		return nil, fmt.Errorf("unexpected OHLCV data format: %T", ohlcvData)
	}

	log.Printf("[CCXT] Received %d candles from exchange", len(dataSlice))

	result := make([]models.Candle, len(dataSlice))

	// Convert in reverse order - newest candles at beginning
	for i, ccxtCandle := range dataSlice {
		// Convert CCXT OHLCV to our Candle model
		candle := models.Candle{
			Timestamp:  ccxtCandle.Timestamp,
			Open:       ccxtCandle.Open,
			High:       ccxtCandle.High,
			Low:        ccxtCandle.Low,
			Close:      ccxtCandle.Close,
			Volume:     ccxtCandle.Volume,
			Indicators: models.Indicators{}, // Empty for now
		}

		// Place at reversed position (newest first)
		result[len(dataSlice)-1-i] = candle
	}

	log.Printf("[CCXT] Converted and reversed %d candles (newest first)", len(result))
	return result, nil
}
