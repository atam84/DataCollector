package indicators

import "github.com/yourusername/datacollector/internal/models"

// CalculateMomentum calculates Momentum indicator
// Momentum = Current Price - Price n periods ago
// Measures rate of change, positive values indicate upward momentum
func CalculateMomentum(candles []models.Candle, period int, source string) []float64 {
	n := len(candles)
	if period <= 0 || n < period+1 {
		return nanSlice(n)
	}

	result := nanSlice(n)
	prices := extractPrices(candles, source)

	for i := period; i < n; i++ {
		result[i] = prices[i] - prices[i-period]
	}

	return result
}
