package indicators

import "github.com/yourusername/datacollector/internal/models"

// CalculateOBV calculates On-Balance Volume
// OBV is a cumulative indicator that adds/subtracts volume based on price direction
// Rising OBV indicates buying pressure, falling OBV indicates selling pressure
func CalculateOBV(candles []models.Candle) []float64 {
	n := len(candles)
	if n < 2 {
		return nanSlice(n)
	}

	result := make([]float64, n)

	// First value is just the volume
	result[0] = candles[0].Volume

	// Calculate cumulative OBV
	for i := 1; i < n; i++ {
		if candles[i].Close > candles[i-1].Close {
			// Price up: add volume
			result[i] = result[i-1] + candles[i].Volume
		} else if candles[i].Close < candles[i-1].Close {
			// Price down: subtract volume
			result[i] = result[i-1] - candles[i].Volume
		} else {
			// Price unchanged: keep same
			result[i] = result[i-1]
		}
	}

	return result
}
