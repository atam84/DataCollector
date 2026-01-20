package indicators

import (
	"math"

	"github.com/yourusername/datacollector/internal/models"
)

// extractPrices extracts price values from candles based on source type
func extractPrices(candles []models.Candle, source string) []float64 {
	result := make([]float64, len(candles))
	for i, c := range candles {
		switch source {
		case "open":
			result[i] = c.Open
		case "high":
			result[i] = c.High
		case "low":
			result[i] = c.Low
		case "close":
			result[i] = c.Close
		case "hl2":
			result[i] = (c.High + c.Low) / 2
		case "hlc3":
			result[i] = (c.High + c.Low + c.Close) / 3
		case "ohlc4":
			result[i] = (c.Open + c.High + c.Low + c.Close) / 4
		default:
			result[i] = c.Close
		}
	}
	return result
}

// max returns the maximum of two float64 values
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// nanSlice creates a slice filled with NaN values
func nanSlice(n int) []float64 {
	result := make([]float64, n)
	for i := range result {
		result[i] = math.NaN()
	}
	return result
}

// isValidValue checks if a float64 value is valid (not NaN or Inf)
func isValidValue(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}
