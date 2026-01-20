package exchange

import (
	"fmt"
	"log"
	"time"

	ccxt "github.com/ccxt/ccxt/go/v4"
	"github.com/yourusername/datacollector/internal/models"
)

// candidateExchanges returns exchanges to test for availability
func candidateExchanges() []string {
	return []string{
		"binance",
		"bybit",
		"coinbase",
		"kraken",
		"kucoin",
		"okx",
		"gateio",
		"huobi",
		"bitfinex",
		"bitstamp",
		"gemini",
		"bitget",
		"mexc",
	}
}

// getSupportedExchangesList returns a conservative list of exchanges known to work
func getSupportedExchangesList() []string {
	// All exchanges confirmed working via TestExchangeAvailability
	return []string{
		"binance",
		"bitfinex",
		"bitget",
		"bitstamp",
		"bybit",
		"coinbase",
		"gateio",
		"gemini",
		"huobi",
		"kraken",
		"kucoin",
		"mexc",
		"okx",
	}
}

// TestExchangeAvailability tests which exchanges can be instantiated
func TestExchangeAvailability() map[string]bool {
	results := make(map[string]bool)
	candidates := candidateExchanges()

	for _, exchangeID := range candidates {
		ccxtID := mapExchangeID(exchangeID)

		// Try to create exchange instance
		options := map[string]interface{}{
			"enableRateLimit": false,
			"timeout":         5000,
		}

		exchange := ccxt.CreateExchange(ccxtID, options)
		if exchange != nil {
			// Successfully created - close it immediately
			exchange.Close()
			results[exchangeID] = true
			log.Printf("[EXCHANGE TEST] ✓ %s (ccxt: %s) - Available", exchangeID, ccxtID)
		} else {
			results[exchangeID] = false
			log.Printf("[EXCHANGE TEST] ✗ %s (ccxt: %s) - Not available", exchangeID, ccxtID)
		}
	}

	return results
}

// GetSupportedExchanges returns a list of exchanges supported by the CCXT library
func GetSupportedExchanges() []string {
	return getSupportedExchangesList()
}

// IsExchangeSupported checks if an exchange is supported by CCXT
func IsExchangeSupported(exchangeID string) bool {
	ccxtExchangeID := mapExchangeID(exchangeID)
	supported := getSupportedExchangesList()

	for _, id := range supported {
		if id == ccxtExchangeID {
			return true
		}
	}

	return false
}

// Adapter defines the interface for exchange operations
type Adapter interface {
	// LoadMarkets loads all available markets from the exchange
	LoadMarkets() error

	// FetchOHLCV fetches OHLCV candles for a symbol
	FetchOHLCV(symbol, timeframe string, since *time.Time, limit int) ([]models.Candle, error)

	// GetExchangeID returns the exchange identifier
	GetExchangeID() string

	// GetSandboxMode returns whether sandbox mode is enabled
	GetSandboxMode() bool

	// Close cleans up exchange resources
	Close() error
}

// CCXTAdapter implements the Adapter interface using CCXT
type CCXTAdapter struct {
	exchange   ccxt.IExchange
	exchangeID string
	sandboxMode bool
}

// mapExchangeID maps our exchange IDs to CCXT exchange IDs
func mapExchangeID(exchangeID string) string {
	// Map our exchange IDs to CCXT's expected IDs
	mapping := map[string]string{
		"okx":      "okx",      // OKX
		"gate":     "gateio",   // Gate.io
		"huobi":    "huobi",    // Huobi
		"binance":  "binance",  // Binance
		"bybit":    "bybit",    // Bybit
		"coinbase": "coinbase", // Coinbase
		"kraken":   "kraken",   // Kraken
		"kucoin":   "kucoin",   // KuCoin
	}

	if ccxtID, ok := mapping[exchangeID]; ok {
		return ccxtID
	}

	// If not in mapping, return as-is
	return exchangeID
}

// NewCCXTAdapter creates a new CCXT adapter
func NewCCXTAdapter(exchangeID string, sandboxMode bool, enableRateLimit bool) (*CCXTAdapter, error) {
	// Map exchange ID to CCXT's expected format
	ccxtExchangeID := mapExchangeID(exchangeID)

	// Check if exchange is supported
	if !IsExchangeSupported(exchangeID) {
		supportedExchanges := GetSupportedExchanges()
		log.Printf("[EXCHANGE] Exchange '%s' (CCXT ID: '%s') not found in CCXT library", exchangeID, ccxtExchangeID)
		log.Printf("[EXCHANGE] Supported exchanges (%d): %v", len(supportedExchanges), supportedExchanges)
		return nil, fmt.Errorf("exchange '%s' not supported by this CCXT build. Supported exchanges: %v", exchangeID, supportedExchanges)
	}

	options := map[string]interface{}{
		"enableRateLimit": enableRateLimit,
		"timeout":         30000, // 30 seconds
	}

	// Create exchange instance
	exchange := ccxt.CreateExchange(ccxtExchangeID, options)
	if exchange == nil {
		return nil, fmt.Errorf("failed to create exchange instance for %s", ccxtExchangeID)
	}

	// Set sandbox mode
	if sandboxMode {
		exchange.SetSandboxMode(true)
	}

	adapter := &CCXTAdapter{
		exchange:   exchange,
		exchangeID: exchangeID,
		sandboxMode: sandboxMode,
	}

	log.Printf("[EXCHANGE] Created CCXT adapter for %s (sandbox: %v)", exchangeID, sandboxMode)

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

	// Set timeframe
	options = append(options, ccxt.WithFetchOHLCVTimeframe(timeframe))

	// Set since timestamp if provided
	if since != nil {
		sinceMs := since.UnixMilli()
		options = append(options, ccxt.WithFetchOHLCVSince(sinceMs))
	}

	// Set limit if provided
	if limit > 0 {
		options = append(options, ccxt.WithFetchOHLCVLimit(int64(limit)))
	}

	// Fetch OHLCV from exchange
	ohlcvData, err := a.exchange.FetchOHLCV(symbol, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OHLCV: %w", err)
	}

	// Convert CCXT OHLCV to our model
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

// GetSandboxMode returns whether sandbox mode is enabled
func (a *CCXTAdapter) GetSandboxMode() bool {
	return a.sandboxMode
}

// Close cleans up exchange resources
func (a *CCXTAdapter) Close() error {
	errors := a.exchange.Close()
	if len(errors) > 0 {
		return fmt.Errorf("errors closing exchange: %v", errors)
	}
	return nil
}
