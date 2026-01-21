package indicators

import "github.com/yourusername/datacollector/internal/models"

// CalculateROC calculates Rate of Change
// ROC = ((Price - Price n periods ago) / Price n periods ago) * 100
// Measures percentage change in price over period
func CalculateROC(candles []models.Candle, period int, source string) []float64 {
	n := len(candles)
	if period <= 0 || n < period+1 {
		return nanSlice(n)
	}

	result := nanSlice(n)
	prices := extractPrices(candles, source)

	for i := period; i < n; i++ {
		if prices[i-period] != 0 {
			result[i] = ((prices[i] - prices[i-period]) / prices[i-period]) * 100
		}
	}

	return result
}
