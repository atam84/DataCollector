package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OHLCV represents a candlestick with indicators
type OHLCV struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ExchangeID string             `bson:"exchange_id" json:"exchange_id"`
	Symbol     string             `bson:"symbol" json:"symbol"`
	Timeframe  string             `bson:"timeframe" json:"timeframe"`
	OpenTime   time.Time          `bson:"open_time" json:"open_time"`

	Open   float64 `bson:"open" json:"open"`
	High   float64 `bson:"high" json:"high"`
	Low    float64 `bson:"low" json:"low"`
	Close  float64 `bson:"close" json:"close"`
	Volume float64 `bson:"volume" json:"volume"`

	Indicators Indicators `bson:"indicators,omitempty" json:"indicators,omitempty"`
	CreatedAt  time.Time  `bson:"created_at" json:"created_at"`
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
