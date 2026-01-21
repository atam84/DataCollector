package indicators

import "github.com/yourusername/datacollector/internal/models"

// CalculateVWMA calculates Volume Weighted Moving Average
// VWMA = Sum(Price * Volume) / Sum(Volume) over period
func CalculateVWMA(candles []models.Candle, period int, source string) []float64 {
	n := len(candles)
	if period <= 0 || n < period {
		return nanSlice(n)
	}

	result := nanSlice(n)
	prices := extractPrices(candles, source)

	for i := period - 1; i < n; i++ {
		priceVolumeSum := 0.0
		volumeSum := 0.0

		for j := i - period + 1; j <= i; j++ {
			priceVolumeSum += prices[j] * candles[j].Volume
			volumeSum += candles[j].Volume
		}

		if volumeSum > 0 {
			result[i] = priceVolumeSum / volumeSum
		}
	}

	return result
}
