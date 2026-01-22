package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OHLCVDocument represents a single document containing all candles for a job
// DEPRECATED: Use OHLCVChunk instead for chunked storage to avoid 16MB limit
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

// OHLCVChunk represents a monthly chunk of OHLCV data
// Unique identifier: (exchange_id, symbol, timeframe, year_month)
// This solves the MongoDB 16MB document size limit by splitting data into monthly chunks
type OHLCVChunk struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ExchangeID   string             `bson:"exchange_id" json:"exchange_id"`
	Symbol       string             `bson:"symbol" json:"symbol"`
	Timeframe    string             `bson:"timeframe" json:"timeframe"`
	YearMonth    string             `bson:"year_month" json:"year_month"` // Format: "YYYY-MM" (e.g., "2024-01")
	StartTime    time.Time          `bson:"start_time" json:"start_time"` // First candle timestamp in chunk
	EndTime      time.Time          `bson:"end_time" json:"end_time"`     // Last candle timestamp in chunk
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
	CandlesCount int                `bson:"candles_count" json:"candles_count"`
	Candles      []Candle           `bson:"candles" json:"candles"` // Sorted by timestamp descending (newest first)
}

// GetYearMonthFromTimestamp extracts the year-month string from a Unix millisecond timestamp
func GetYearMonthFromTimestamp(timestampMs int64) string {
	t := time.UnixMilli(timestampMs)
	return t.Format("2006-01")
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

// OHLCVStats represents aggregate statistics for OHLCV data
type OHLCVStats struct {
	ExchangeID       string    `json:"exchange_id,omitempty"`
	TotalCandles     int64     `json:"total_candles"`
	TotalChunks      int       `json:"total_chunks"`
	LegacyDocuments  int       `json:"legacy_documents"`
	UniqueExchanges  int       `json:"unique_exchanges,omitempty"`
	UniqueSymbols    int       `json:"unique_symbols"`
	UniqueTimeframes int       `json:"unique_timeframes"`
	OldestData       time.Time `json:"oldest_data,omitempty"`
	NewestData       time.Time `json:"newest_data,omitempty"`
}

// DataQuality represents data quality metrics for OHLCV data
type DataQuality struct {
	ExchangeID        string    `json:"exchange_id"`
	Symbol            string    `json:"symbol"`
	Timeframe         string    `json:"timeframe"`
	TotalCandles      int64     `json:"total_candles"`
	ExpectedCandles   int64     `json:"expected_candles"`
	MissingCandles    int64     `json:"missing_candles"`
	CompletenessScore float64   `json:"completeness_score"` // 0-100 percentage
	OldestCandle      time.Time `json:"oldest_candle,omitempty"`
	NewestCandle      time.Time `json:"newest_candle,omitempty"`
	DataFreshness     string    `json:"data_freshness"`   // "fresh", "stale", "very_stale"
	FreshnessMinutes  int64     `json:"freshness_minutes"` // Minutes since last candle
	GapsDetected      int       `json:"gaps_detected"`     // Number of gap periods
	Gaps              []DataGap `json:"gaps,omitempty"`    // Details of detected gaps
	QualityStatus     string    `json:"quality_status"`    // "excellent", "good", "fair", "poor"
}

// DataGap represents a gap in the time series data
type DataGap struct {
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	MissingCandles   int       `json:"missing_candles"`
	DurationMinutes  int64     `json:"duration_minutes"`
}

// DataQualitySummary represents aggregated data quality metrics
type DataQualitySummary struct {
	TotalJobs            int     `json:"total_jobs"`
	ExcellentQuality     int     `json:"excellent_quality"`
	GoodQuality          int     `json:"good_quality"`
	FairQuality          int     `json:"fair_quality"`
	PoorQuality          int     `json:"poor_quality"`
	AverageCompleteness  float64 `json:"average_completeness"`
	TotalMissingCandles  int64   `json:"total_missing_candles"`
	TotalGaps            int     `json:"total_gaps"`
	FreshDataJobs        int     `json:"fresh_data_jobs"`
	StaleDataJobs        int     `json:"stale_data_jobs"`
}

// GetTimeframeDurationMinutes returns the duration of a timeframe in minutes
func GetTimeframeDurationMinutes(timeframe string) int64 {
	switch timeframe {
	case "1m":
		return 1
	case "3m":
		return 3
	case "5m":
		return 5
	case "15m":
		return 15
	case "30m":
		return 30
	case "1h":
		return 60
	case "2h":
		return 120
	case "4h":
		return 240
	case "6h":
		return 360
	case "8h":
		return 480
	case "12h":
		return 720
	case "1d":
		return 1440
	case "3d":
		return 4320
	case "1w":
		return 10080
	case "1M":
		return 43200 // Approximate month
	default:
		return 60 // Default to 1 hour
	}
}
