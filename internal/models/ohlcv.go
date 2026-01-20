package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OHLCVDocument represents a single document containing all candles for a job
// Unique identifier: (exchange_id, symbol, timeframe)
type OHLCVDocument struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ExchangeID   string             `bson:"exchange_id" json:"exchange_id"`
	Symbol       string             `bson:"symbol" json:"symbol"`
	Timeframe    string             `bson:"timeframe" json:"timeframe"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
	CandlesCount int                `bson:"candles_count" json:"candles_count"`
	Candles      []Candle           `bson:"candles" json:"candles"`
}

// Candle represents a single OHLCV candle with indicators
// Note: Newest candles are at index 0, oldest at the end
type Candle struct {
	Timestamp  int64      `bson:"timestamp" json:"timestamp"`   // Unix milliseconds
	Open       float64    `bson:"open" json:"open"`
	High       float64    `bson:"high" json:"high"`
	Low        float64    `bson:"low" json:"low"`
	Close      float64    `bson:"close" json:"close"`
	Volume     float64    `bson:"volume" json:"volume"`
	Indicators Indicators `bson:"indicators,omitempty" json:"indicators,omitempty"`
}

// Indicators holds computed technical indicators
// All fields are pointers to support null values (omitempty)
type Indicators struct {
	// ==================== Trend Indicators ====================

	// Simple Moving Average
	SMA20  *float64 `bson:"sma20,omitempty" json:"sma20,omitempty"`
	SMA50  *float64 `bson:"sma50,omitempty" json:"sma50,omitempty"`
	SMA200 *float64 `bson:"sma200,omitempty" json:"sma200,omitempty"`

	// Exponential Moving Average
	EMA12 *float64 `bson:"ema12,omitempty" json:"ema12,omitempty"`
	EMA26 *float64 `bson:"ema26,omitempty" json:"ema26,omitempty"`
	EMA50 *float64 `bson:"ema50,omitempty" json:"ema50,omitempty"`

	// Double Exponential Moving Average
	DEMA *float64 `bson:"dema,omitempty" json:"dema,omitempty"`

	// Triple Exponential Moving Average
	TEMA *float64 `bson:"tema,omitempty" json:"tema,omitempty"`

	// Weighted Moving Average
	WMA *float64 `bson:"wma,omitempty" json:"wma,omitempty"`

	// Hull Moving Average
	HMA *float64 `bson:"hma,omitempty" json:"hma,omitempty"`

	// Volume Weighted Moving Average
	VWMA *float64 `bson:"vwma,omitempty" json:"vwma,omitempty"`

	// Ichimoku Cloud
	IchimokuTenkan  *float64 `bson:"ichimoku_tenkan,omitempty" json:"ichimoku_tenkan,omitempty"`     // Conversion Line
	IchimokuKijun   *float64 `bson:"ichimoku_kijun,omitempty" json:"ichimoku_kijun,omitempty"`       // Base Line
	IchimokuSenkouA *float64 `bson:"ichimoku_senkou_a,omitempty" json:"ichimoku_senkou_a,omitempty"` // Leading Span A
	IchimokuSenkouB *float64 `bson:"ichimoku_senkou_b,omitempty" json:"ichimoku_senkou_b,omitempty"` // Leading Span B
	IchimokuChikou  *float64 `bson:"ichimoku_chikou,omitempty" json:"ichimoku_chikou,omitempty"`     // Lagging Span

	// ADX/DMI (Average Directional Index / Directional Movement Index)
	ADX     *float64 `bson:"adx,omitempty" json:"adx,omitempty"`
	PlusDI  *float64 `bson:"plus_di,omitempty" json:"plus_di,omitempty"`   // +DI
	MinusDI *float64 `bson:"minus_di,omitempty" json:"minus_di,omitempty"` // -DI

	// SuperTrend
	SuperTrend       *float64 `bson:"supertrend,omitempty" json:"supertrend,omitempty"`
	SuperTrendSignal *int     `bson:"supertrend_signal,omitempty" json:"supertrend_signal,omitempty"` // 1=buy, -1=sell, 0=neutral

	// ==================== Momentum Indicators ====================

	// RSI (Relative Strength Index)
	RSI6  *float64 `bson:"rsi6,omitempty" json:"rsi6,omitempty"`
	RSI14 *float64 `bson:"rsi14,omitempty" json:"rsi14,omitempty"`
	RSI24 *float64 `bson:"rsi24,omitempty" json:"rsi24,omitempty"`

	// Stochastic Oscillator
	StochK *float64 `bson:"stoch_k,omitempty" json:"stoch_k,omitempty"` // %K
	StochD *float64 `bson:"stoch_d,omitempty" json:"stoch_d,omitempty"` // %D

	// MACD (Moving Average Convergence Divergence)
	MACD       *float64 `bson:"macd,omitempty" json:"macd,omitempty"`
	MACDSignal *float64 `bson:"macd_signal,omitempty" json:"macd_signal,omitempty"`
	MACDHist   *float64 `bson:"macd_hist,omitempty" json:"macd_hist,omitempty"`

	// ROC (Rate of Change)
	ROC *float64 `bson:"roc,omitempty" json:"roc,omitempty"`

	// CCI (Commodity Channel Index)
	CCI *float64 `bson:"cci,omitempty" json:"cci,omitempty"`

	// Williams %R
	WilliamsR *float64 `bson:"williams_r,omitempty" json:"williams_r,omitempty"`

	// Momentum
	Momentum *float64 `bson:"momentum,omitempty" json:"momentum,omitempty"`

	// ==================== Volatility Indicators ====================

	// Bollinger Bands
	BollingerUpper     *float64 `bson:"bb_upper,omitempty" json:"bb_upper,omitempty"`
	BollingerMiddle    *float64 `bson:"bb_middle,omitempty" json:"bb_middle,omitempty"`
	BollingerLower     *float64 `bson:"bb_lower,omitempty" json:"bb_lower,omitempty"`
	BollingerBandwidth *float64 `bson:"bb_bandwidth,omitempty" json:"bb_bandwidth,omitempty"` // (Upper - Lower) / Middle * 100
	BollingerPercentB  *float64 `bson:"bb_percent_b,omitempty" json:"bb_percent_b,omitempty"` // (Price - Lower) / (Upper - Lower)

	// ATR (Average True Range)
	ATR *float64 `bson:"atr,omitempty" json:"atr,omitempty"`

	// Keltner Channels
	KeltnerUpper  *float64 `bson:"keltner_upper,omitempty" json:"keltner_upper,omitempty"`
	KeltnerMiddle *float64 `bson:"keltner_middle,omitempty" json:"keltner_middle,omitempty"`
	KeltnerLower  *float64 `bson:"keltner_lower,omitempty" json:"keltner_lower,omitempty"`

	// Donchian Channels
	DonchianUpper  *float64 `bson:"donchian_upper,omitempty" json:"donchian_upper,omitempty"`
	DonchianMiddle *float64 `bson:"donchian_middle,omitempty" json:"donchian_middle,omitempty"`
	DonchianLower  *float64 `bson:"donchian_lower,omitempty" json:"donchian_lower,omitempty"`

	// Standard Deviation
	StdDev *float64 `bson:"stddev,omitempty" json:"stddev,omitempty"`

	// ==================== Volume Indicators ====================

	// OBV (On-Balance Volume)
	OBV *float64 `bson:"obv,omitempty" json:"obv,omitempty"`

	// VWAP (Volume Weighted Average Price)
	VWAP *float64 `bson:"vwap,omitempty" json:"vwap,omitempty"`

	// MFI (Money Flow Index)
	MFI *float64 `bson:"mfi,omitempty" json:"mfi,omitempty"`

	// CMF (Chaikin Money Flow)
	CMF *float64 `bson:"cmf,omitempty" json:"cmf,omitempty"`

	// Volume SMA
	VolumeSMA *float64 `bson:"volume_sma,omitempty" json:"volume_sma,omitempty"`
}

// JobExecutionResult represents the result of a job execution
type JobExecutionResult struct {
	Success         bool      `json:"success"`
	Message         string    `json:"message"`
	RecordsFetched  int       `json:"records_fetched"`
	ExecutionTimeMs int64     `json:"execution_time_ms"`
	NextRunTime     time.Time `json:"next_run_time"`
	Error           *string   `json:"error,omitempty"`
}
