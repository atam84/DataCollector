package indicators

import (
	"math"

	"github.com/yourusername/datacollector/internal/models"
)

// CalculateATR calculates Average True Range using Wilder's smoothing
// Returns a slice where each value is the ATR at that index
// Returns NaN for positions where there's insufficient data
func CalculateATR(candles []models.Candle, period int) []float64 {
	n := len(candles)
	if n < period {
		return nanSlice(n)
	}

	result := nanSlice(n)
	tr := make([]float64, n)

	// First TR is just High - Low
	tr[0] = candles[0].High - candles[0].Low

	// Calculate True Range
	for i := 1; i < n; i++ {
		hl := candles[i].High - candles[i].Low
		hpc := math.Abs(candles[i].High - candles[i-1].Close)
		lpc := math.Abs(candles[i].Low - candles[i-1].Close)
		tr[i] = max(hl, max(hpc, lpc))
	}

	// Initial ATR (SMA of first period TRs)
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += tr[i]
	}
	result[period-1] = sum / float64(period)

	// Wilder's smoothing
	for i := period; i < n; i++ {
		result[i] = (result[i-1]*float64(period-1) + tr[i]) / float64(period)
	}

	return result
}
