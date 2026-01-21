package indicators

import (
	"github.com/yourusername/datacollector/internal/models"
)

// CalculateEMA calculates Exponential Moving Average
// Returns a slice where each value is the EMA at that index
// Returns NaN for positions where there's insufficient data
func CalculateEMA(candles []models.Candle, period int, source string) []float64 {
	n := len(candles)
	if period <= 0 || n < period {
		return nanSlice(n)
	}

	result := nanSlice(n)
	prices := extractPrices(candles, source)
	multiplier := 2.0 / float64(period+1)

	// Initial SMA for first EMA value
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	result[period-1] = sum / float64(period)

	// EMA calculation
	for i := period; i < n; i++ {
		result[i] = (prices[i]-result[i-1])*multiplier + result[i-1]
	}

	return result
}
