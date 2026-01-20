package exchange

import (
	"fmt"
	"time"

	ccxt "github.com/ccxt/ccxt/go/v4"
	"github.com/yourusername/datacollector/internal/models"
)

// Adapter defines the interface for exchange operations
type Adapter interface {
	// LoadMarkets loads all available markets from the exchange
	LoadMarkets() error

	// FetchOHLCV fetches OHLCV candles for a symbol
	FetchOHLCV(symbol, timeframe string, since *time.Time, limit int) ([]models.OHLCV, error)

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

// NewCCXTAdapter creates a new CCXT adapter
func NewCCXTAdapter(exchangeID string, sandboxMode bool, enableRateLimit bool) (*CCXTAdapter, error) {
	options := map[string]interface{}{
		"enableRateLimit": enableRateLimit,
		"timeout":         30000, // 30 seconds
	}

	// Create exchange instance
	exchange := ccxt.CreateExchange(exchangeID, options)
	if exchange == nil {
		return nil, fmt.Errorf("unsupported exchange: %s", exchangeID)
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

	return adapter, nil
}

// LoadMarkets loads all available markets from the exchange
func (a *CCXTAdapter) LoadMarkets() error {
	_, err := a.exchange.LoadMarkets()
	return err
}

// FetchOHLCV fetches OHLCV candles for a symbol
func (a *CCXTAdapter) FetchOHLCV(symbol, timeframe string, since *time.Time, limit int) ([]models.OHLCV, error) {
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
		options = append(options, ccxt.WithFetchOHLCVLimit(limit))
	}

	// Fetch OHLCV from exchange
	ohlcvData, err := a.exchange.FetchOHLCV(symbol, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OHLCV: %w", err)
	}

	// Convert CCXT OHLCV to our model
	candles := make([]models.OHLCV, 0, len(ohlcvData))
	for _, bar := range ohlcvData {
		candle := models.OHLCV{
			ExchangeID: a.exchangeID,
			Symbol:     symbol,
			Timeframe:  timeframe,
			OpenTime:   time.UnixMilli(bar.Timestamp),
			Open:       bar.Open,
			High:       bar.High,
			Low:        bar.Low,
			Close:      bar.Close,
			Volume:     bar.Volume,
			CreatedAt:  time.Now(),
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
