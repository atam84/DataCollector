package indicators

import "github.com/yourusername/datacollector/internal/models"

// CalculateWMA calculates Weighted Moving Average
// Gives more weight to recent prices
func CalculateWMA(candles []models.Candle, period int, source string) []float64 {
	n := len(candles)
	if period <= 0 || n < period {
		return nanSlice(n)
	}

	result := nanSlice(n)
	prices := extractPrices(candles, source)

	// Calculate sum of weights: 1 + 2 + 3 + ... + period
	weightSum := float64(period * (period + 1) / 2)

	for i := period - 1; i < n; i++ {
		weightedSum := 0.0
		for j := 0; j < period; j++ {
			weight := float64(j + 1) // Weight increases for more recent prices
			weightedSum += prices[i-period+1+j] * weight
		}
		result[i] = weightedSum / weightSum
	}

	return result
}
