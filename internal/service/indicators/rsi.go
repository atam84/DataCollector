package indicators

import (
	"github.com/yourusername/datacollector/internal/models"
)

// CalculateRSI calculates Relative Strength Index using Wilder's smoothing
// Returns a slice where each value is the RSI at that index (0-100)
// Returns NaN for positions where there's insufficient data
func CalculateRSI(candles []models.Candle, period int, source string) []float64 {
	n := len(candles)
	if period <= 0 || n < period+1 {
		return nanSlice(n)
	}

	result := nanSlice(n)
	prices := extractPrices(candles, source)

	gains := make([]float64, n)
	losses := make([]float64, n)

	// Calculate gains and losses
	for i := 1; i < n; i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			gains[i] = change
		} else {
			losses[i] = -change
		}
	}

	// Initial averages (SMA)
	avgGain := 0.0
	avgLoss := 0.0
	for i := 1; i <= period; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	// First RSI
	if avgLoss == 0 {
		result[period] = 100
	} else {
		rs := avgGain / avgLoss
		result[period] = 100 - (100 / (1 + rs))
	}

	// Subsequent RSI values using Wilder's smoothing
	for i := period + 1; i < n; i++ {
		avgGain = (avgGain*float64(period-1) + gains[i]) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + losses[i]) / float64(period)

		if avgLoss == 0 {
			result[i] = 100
		} else {
			rs := avgGain / avgLoss
			result[i] = 100 - (100 / (1 + rs))
		}
	}

	return result
}
