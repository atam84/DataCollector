package exchange

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	ccxt "github.com/ccxt/ccxt/go/v4"
	"github.com/yourusername/datacollector/internal/models"
)

// metadataCache caches exchange metadata to avoid repeated instantiation
var (
	metadataCache     = make(map[string]*ExchangeMetadata)
	metadataCacheLock sync.RWMutex
	supportedCache    []string
	supportedCacheLock sync.RWMutex
)

// candidateExchanges returns all exchanges available in CCXT to test
func candidateExchanges() []string {
	return ccxt.Exchanges
}

// discoverSupportedExchanges dynamically discovers which exchanges work
func discoverSupportedExchanges() []string {
	supportedCacheLock.RLock()
	if len(supportedCache) > 0 {
		defer supportedCacheLock.RUnlock()
		return supportedCache
	}
	supportedCacheLock.RUnlock()

	supportedCacheLock.Lock()
	defer supportedCacheLock.Unlock()

	// Double-check after acquiring write lock
	if len(supportedCache) > 0 {
		return supportedCache
	}

	candidates := candidateExchanges()
	log.Printf("[EXCHANGE] Discovering supported exchanges from CCXT (total candidates: %d)...", len(candidates))
	log.Printf("[EXCHANGE] All CCXT exchanges: %v", candidates)

	supported := make([]string, 0)
	failedCreate := make([]string, 0)
	noOHLCV := make([]string, 0)

	for _, exchangeID := range candidates {
		options := map[string]interface{}{
			"enableRateLimit": false,
			"timeout":         5000,
		}

		exchange := ccxt.CreateExchange(exchangeID, options)
		if exchange == nil {
			failedCreate = append(failedCreate, exchangeID)
			continue
		}

		// Check if it has OHLCV support
		has := exchange.GetHas()
		hasFetchOHLCV, ok := has["fetchOHLCV"]
		if !ok || hasFetchOHLCV != true {
			noOHLCV = append(noOHLCV, exchangeID)
			exchange.Close()
			continue
		}

		supported = append(supported, exchangeID)
		exchange.Close()
	}

	sort.Strings(supported)
	supportedCache = supported

	log.Printf("[EXCHANGE] Discovery complete:")
	log.Printf("[EXCHANGE]   - Supported with OHLCV: %d exchanges", len(supported))
	log.Printf("[EXCHANGE]   - Failed to create: %d exchanges: %v", len(failedCreate), failedCreate)
	log.Printf("[EXCHANGE]   - No OHLCV support: %d exchanges: %v", len(noOHLCV), noOHLCV)
	log.Printf("[EXCHANGE]   - Supported list: %v", supported)

	return supportedCache
}

// TestExchangeAvailability tests which exchanges can be instantiated
func TestExchangeAvailability() map[string]bool {
	results := make(map[string]bool)
	candidates := candidateExchanges()

	for _, exchangeID := range candidates {
		options := map[string]interface{}{
			"enableRateLimit": false,
			"timeout":         5000,
		}

		exchange := ccxt.CreateExchange(exchangeID, options)
		if exchange != nil {
			exchange.Close()
			results[exchangeID] = true
		} else {
			results[exchangeID] = false
		}
	}

	return results
}

// GetSupportedExchanges returns a list of exchanges supported by the CCXT library
func GetSupportedExchanges() []string {
	return discoverSupportedExchanges()
}

// IsExchangeSupported checks if an exchange is supported by CCXT
func IsExchangeSupported(exchangeID string) bool {
	ccxtExchangeID := mapExchangeID(exchangeID)
	supported := GetSupportedExchanges()

	for _, id := range supported {
		if id == ccxtExchangeID {
			return true
		}
	}

	return false
}

// ClearCache clears the metadata and supported exchanges cache
func ClearCache() {
	metadataCacheLock.Lock()
	metadataCache = make(map[string]*ExchangeMetadata)
	metadataCacheLock.Unlock()

	supportedCacheLock.Lock()
	supportedCache = nil
	supportedCacheLock.Unlock()

	log.Println("[EXCHANGE] Cache cleared")
}

// Adapter defines the interface for exchange operations
type Adapter interface {
	LoadMarkets() error
	FetchOHLCV(symbol, timeframe string, since *time.Time, limit int) ([]models.Candle, error)
	GetExchangeID() string
	Close() error
}

// CCXTAdapter implements the Adapter interface using CCXT
type CCXTAdapter struct {
	exchange   ccxt.IExchange
	exchangeID string
}

// mapExchangeID maps our exchange IDs to CCXT exchange IDs
func mapExchangeID(exchangeID string) string {
	mapping := map[string]string{
		"gate": "gateio",
	}

	if ccxtID, ok := mapping[exchangeID]; ok {
		return ccxtID
	}

	return exchangeID
}

// DebugExchange provides detailed debugging info for a specific exchange
func DebugExchange(exchangeID string) map[string]interface{} {
	result := map[string]interface{}{
		"input_id":    exchangeID,
		"mapped_id":   mapExchangeID(exchangeID),
		"in_ccxt_list": false,
		"can_create":  false,
		"has_ohlcv":   false,
		"error":       nil,
	}

	ccxtID := mapExchangeID(exchangeID)

	// Check if in CCXT exchanges list
	for _, id := range ccxt.Exchanges {
		if id == ccxtID {
			result["in_ccxt_list"] = true
			break
		}
	}

	// Try to create instance
	options := map[string]interface{}{
		"enableRateLimit": false,
		"timeout":         10000,
	}

	exchange := ccxt.CreateExchange(ccxtID, options)
	if exchange == nil {
		result["error"] = "CreateExchange returned nil"
		log.Printf("[EXCHANGE DEBUG] %s: CreateExchange returned nil", exchangeID)
		return result
	}
	defer exchange.Close()

	result["can_create"] = true

	// Check OHLCV support
	has := exchange.GetHas()
	result["has_map"] = has

	if hasFetchOHLCV, ok := has["fetchOHLCV"]; ok {
		result["fetchOHLCV_value"] = hasFetchOHLCV
		result["fetchOHLCV_type"] = fmt.Sprintf("%T", hasFetchOHLCV)
		if hasFetchOHLCV == true {
			result["has_ohlcv"] = true
		}
	} else {
		result["fetchOHLCV_value"] = "key not found"
	}

	// Get timeframes
	timeframes := exchange.GetTimeframes()
	result["timeframes"] = timeframes

	log.Printf("[EXCHANGE DEBUG] %s: %+v", exchangeID, result)
	return result
}

// NewCCXTAdapter creates a new CCXT adapter
func NewCCXTAdapter(exchangeID string, enableRateLimit bool) (*CCXTAdapter, error) {
	ccxtExchangeID := mapExchangeID(exchangeID)
	log.Printf("[EXCHANGE] NewCCXTAdapter called for '%s' (mapped to '%s')", exchangeID, ccxtExchangeID)

	// Debug: Check if exchange is in CCXT list
	inList := false
	for _, id := range ccxt.Exchanges {
		if id == ccxtExchangeID {
			inList = true
			break
		}
	}
	log.Printf("[EXCHANGE] Exchange '%s' in ccxt.Exchanges list: %v", ccxtExchangeID, inList)

	if !IsExchangeSupported(exchangeID) {
		supportedExchanges := GetSupportedExchanges()
		log.Printf("[EXCHANGE] Exchange '%s' (mapped: '%s') NOT in supported list", exchangeID, ccxtExchangeID)
		log.Printf("[EXCHANGE] Supported exchanges (%d): %v", len(supportedExchanges), supportedExchanges)

		// Debug why it's not supported
		debugInfo := DebugExchange(exchangeID)
		log.Printf("[EXCHANGE] Debug info for '%s': %+v", exchangeID, debugInfo)

		return nil, fmt.Errorf("exchange '%s' not yet supported", exchangeID)
	}

	options := map[string]interface{}{
		"enableRateLimit": enableRateLimit,
		"timeout":         30000,
	}

	exchange := ccxt.CreateExchange(ccxtExchangeID, options)
	if exchange == nil {
		log.Printf("[EXCHANGE] CreateExchange returned nil for '%s'", ccxtExchangeID)
		return nil, fmt.Errorf("failed to create exchange instance for %s", ccxtExchangeID)
	}

	adapter := &CCXTAdapter{
		exchange:   exchange,
		exchangeID: exchangeID,
	}

	log.Printf("[EXCHANGE] Successfully created CCXT adapter for %s", exchangeID)

	return adapter, nil
}

// LoadMarkets loads all available markets from the exchange
func (a *CCXTAdapter) LoadMarkets() error {
	_, err := a.exchange.LoadMarkets()
	return err
}

// FetchOHLCV fetches OHLCV candles for a symbol
func (a *CCXTAdapter) FetchOHLCV(symbol, timeframe string, since *time.Time, limit int) ([]models.Candle, error) {
	var options []ccxt.FetchOHLCVOptions

	options = append(options, ccxt.WithFetchOHLCVTimeframe(timeframe))

	if since != nil {
		sinceMs := since.UnixMilli()
		options = append(options, ccxt.WithFetchOHLCVSince(sinceMs))
	}

	if limit > 0 {
		options = append(options, ccxt.WithFetchOHLCVLimit(int64(limit)))
	}

	ohlcvData, err := a.exchange.FetchOHLCV(symbol, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OHLCV: %w", err)
	}

	candles := make([]models.Candle, 0, len(ohlcvData))
	for _, bar := range ohlcvData {
		candle := models.Candle{
			Timestamp: bar.Timestamp,
			Open:      bar.Open,
			High:      bar.High,
			Low:       bar.Low,
			Close:     bar.Close,
			Volume:    bar.Volume,
		}
		candles = append(candles, candle)
	}

	return candles, nil
}

// GetExchangeID returns the exchange identifier
func (a *CCXTAdapter) GetExchangeID() string {
	return a.exchangeID
}

// Close cleans up exchange resources
func (a *CCXTAdapter) Close() error {
	errors := a.exchange.Close()
	if len(errors) > 0 {
		return fmt.Errorf("errors closing exchange: %v", errors)
	}
	return nil
}

// ExchangeMetadata holds exchange metadata information
type ExchangeMetadata struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Countries   []string          `json:"countries,omitempty"`
	RateLimit   int               `json:"rate_limit"`
	Timeout     int               `json:"timeout"`
	Timeframes  map[string]string `json:"timeframes"`
	HasOHLCV    bool              `json:"has_ohlcv"`
	OHLCVLimit  int               `json:"ohlcv_limit"`
	Symbols     []string          `json:"symbols,omitempty"`
	SymbolCount int               `json:"symbol_count"`
}

// GetExchangeMetadata fetches metadata dynamically from CCXT
func GetExchangeMetadata(exchangeID string) (*ExchangeMetadata, error) {
	ccxtExchangeID := mapExchangeID(exchangeID)

	// Check cache first
	metadataCacheLock.RLock()
	if cached, ok := metadataCache[ccxtExchangeID]; ok {
		metadataCacheLock.RUnlock()
		return cached, nil
	}
	metadataCacheLock.RUnlock()

	// Create exchange instance to fetch metadata
	options := map[string]interface{}{
		"enableRateLimit": false,
		"timeout":         10000,
	}

	exchange := ccxt.CreateExchange(ccxtExchangeID, options)
	if exchange == nil {
		return nil, fmt.Errorf("failed to create exchange instance for %s", ccxtExchangeID)
	}
	defer exchange.Close()

	metadata := &ExchangeMetadata{
		ID:        ccxtExchangeID,
		Name:      capitalizeFirst(ccxtExchangeID),
		RateLimit: 1000,
		Timeout:   30000,
		HasOHLCV:  false,
		OHLCVLimit: 500,
	}

	// Get timeframes dynamically
	timeframes := exchange.GetTimeframes()
	metadata.Timeframes = make(map[string]string)
	for k, v := range timeframes {
		if vs, ok := v.(string); ok {
			metadata.Timeframes[k] = vs
		} else {
			metadata.Timeframes[k] = k
		}
	}

	// Get capabilities from GetHas()
	has := exchange.GetHas()
	if hasFetchOHLCV, ok := has["fetchOHLCV"]; ok {
		metadata.HasOHLCV = hasFetchOHLCV == true
	}

	// Get features for OHLCV limit
	features := exchange.GetFeatures()
	if spotFeatures, ok := features["spot"].(map[string]interface{}); ok {
		if fetchOHLCVFeatures, ok := spotFeatures["fetchOHLCV"].(map[string]interface{}); ok {
			if limit, ok := fetchOHLCVFeatures["limit"].(float64); ok {
				metadata.OHLCVLimit = int(limit)
			}
		}
	}

	// Fallback OHLCV limits for known exchanges
	if metadata.OHLCVLimit == 0 || metadata.OHLCVLimit == 500 {
		ohlcvLimits := map[string]int{
			"binance": 1000, "bybit": 1000, "coinbase": 300,
			"kraken": 720, "kucoin": 1500, "okx": 300,
			"gateio": 1000, "huobi": 2000, "bitfinex": 10000,
			"bitstamp": 1000, "gemini": 500, "bitget": 1000, "mexc": 1000,
		}
		if limit, ok := ohlcvLimits[ccxtExchangeID]; ok {
			metadata.OHLCVLimit = limit
		}
	}

	// Cache the result
	metadataCacheLock.Lock()
	metadataCache[ccxtExchangeID] = metadata
	metadataCacheLock.Unlock()

	log.Printf("[EXCHANGE] Fetched metadata for %s: %d timeframes, OHLCV limit: %d",
		ccxtExchangeID, len(metadata.Timeframes), metadata.OHLCVLimit)

	return metadata, nil
}

// GetExchangeSymbols fetches all available symbols for an exchange
func GetExchangeSymbols(exchangeID string) ([]string, error) {
	ccxtExchangeID := mapExchangeID(exchangeID)

	options := map[string]interface{}{
		"enableRateLimit": true,
		"timeout":         30000,
	}

	exchange := ccxt.CreateExchange(ccxtExchangeID, options)
	if exchange == nil {
		return nil, fmt.Errorf("failed to create exchange instance for %s", ccxtExchangeID)
	}
	defer exchange.Close()

	// Load markets to get symbols
	_, err := exchange.LoadMarkets()
	if err != nil {
		return nil, fmt.Errorf("failed to load markets: %w", err)
	}

	symbols := exchange.GetSymbols()
	sort.Strings(symbols)

	return symbols, nil
}

// GetAllExchangesMetadata fetches metadata for all supported exchanges
func GetAllExchangesMetadata() []ExchangeMetadata {
	exchanges := GetSupportedExchanges()
	result := make([]ExchangeMetadata, 0, len(exchanges))

	for _, exchangeID := range exchanges {
		metadata, err := GetExchangeMetadata(exchangeID)
		if err != nil {
			log.Printf("[EXCHANGE] Warning: Failed to get metadata for %s: %v", exchangeID, err)
			result = append(result, ExchangeMetadata{
				ID:        exchangeID,
				Name:      capitalizeFirst(exchangeID),
				RateLimit: 600,
				HasOHLCV:  true,
				OHLCVLimit: 500,
				Timeframes: map[string]string{
					"1m": "1m", "5m": "5m", "15m": "15m", "30m": "30m",
					"1h": "1h", "4h": "4h", "1d": "1d", "1w": "1w",
				},
			})
			continue
		}
		result = append(result, *metadata)
	}

	return result
}

// capitalizeFirst capitalizes the first letter of a string
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	if s[0] >= 'a' && s[0] <= 'z' {
		return string(s[0]-32) + s[1:]
	}
	return s
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
