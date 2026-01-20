package indicators

import (
	"math"

	"github.com/yourusername/datacollector/internal/models"
)

// CalculateStdDev calculates Standard Deviation
// Measures price volatility/dispersion from the mean
func CalculateStdDev(candles []models.Candle, period int, source string) []float64 {
	n := len(candles)
	if n < period {
		return nanSlice(n)
	}

	result := nanSlice(n)
	prices := extractPrices(candles, source)

	// Calculate SMA first
	sma := CalculateSMA(candles, period, source)

	// Calculate standard deviation
	for i := period - 1; i < n; i++ {
		if !isValidValue(sma[i]) {
			continue
		}

		// Calculate variance
		variance := 0.0
		for j := i - period + 1; j <= i; j++ {
			diff := prices[j] - sma[i]
			variance += diff * diff
		}
		variance /= float64(period)

		// Standard deviation is square root of variance
		result[i] = math.Sqrt(variance)
	}

	return result
}
