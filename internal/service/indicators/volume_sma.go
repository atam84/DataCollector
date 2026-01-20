package indicators

import "github.com/yourusername/datacollector/internal/models"

// CalculateVolumeSMA calculates Simple Moving Average of Volume
// Helps identify periods of high/low volume activity
func CalculateVolumeSMA(candles []models.Candle, period int) []float64 {
	n := len(candles)
	if n < period {
		return nanSlice(n)
	}

	result := nanSlice(n)

	// Calculate initial SMA
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += candles[i].Volume
	}
	result[period-1] = sum / float64(period)

	// Rolling calculation
	for i := period; i < n; i++ {
		sum = sum - candles[i-period].Volume + candles[i].Volume
		result[i] = sum / float64(period)
	}

	return result
}
