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
type Indicators struct {
	RSI6   *float64 `bson:"rsi6,omitempty" json:"rsi6,omitempty"`
	RSI14  *float64 `bson:"rsi14,omitempty" json:"rsi14,omitempty"`
	RSI24  *float64 `bson:"rsi24,omitempty" json:"rsi24,omitempty"`

	EMA12  *float64 `bson:"ema12,omitempty" json:"ema12,omitempty"`
	EMA26  *float64 `bson:"ema26,omitempty" json:"ema26,omitempty"`

	MACD        *float64 `bson:"macd,omitempty" json:"macd,omitempty"`
	MACDSignal  *float64 `bson:"macd_signal,omitempty" json:"macd_signal,omitempty"`
	MACDHist    *float64 `bson:"macd_hist,omitempty" json:"macd_hist,omitempty"`
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
