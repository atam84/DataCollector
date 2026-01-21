package indicators

import (
	"github.com/yourusername/datacollector/internal/models"
)

// CalculateSMA calculates Simple Moving Average
// Returns a slice where each value is the SMA at that index
// Returns NaN for positions where there's insufficient data
func CalculateSMA(candles []models.Candle, period int, source string) []float64 {
	n := len(candles)
	if period <= 0 || n < period {
		return nanSlice(n)
	}

	result := nanSlice(n)
	prices := extractPrices(candles, source)

	// Calculate initial SMA
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	result[period-1] = sum / float64(period)

	// Rolling calculation
	for i := period; i < n; i++ {
		sum = sum - prices[i-period] + prices[i]
		result[i] = sum / float64(period)
	}

	return result
}
