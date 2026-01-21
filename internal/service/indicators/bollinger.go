package indicators

import (
	"math"

	"github.com/yourusername/datacollector/internal/models"
)

// BollingerResult holds the Bollinger Bands indicator values
type BollingerResult struct {
	Upper     []float64 // Upper band
	Middle    []float64 // Middle band (SMA)
	Lower     []float64 // Lower band
	Bandwidth []float64 // (Upper - Lower) / Middle * 100
	PercentB  []float64 // (Price - Lower) / (Upper - Lower)
}

// CalculateBollingerBands calculates Bollinger Bands
// Returns upper, middle (SMA), lower bands, bandwidth, and %B
func CalculateBollingerBands(candles []models.Candle, period int, stdDev float64, source string) *BollingerResult {
	n := len(candles)
	if period <= 0 || n < period {
		return &BollingerResult{
			Upper:     nanSlice(n),
			Middle:    nanSlice(n),
			Lower:     nanSlice(n),
			Bandwidth: nanSlice(n),
			PercentB:  nanSlice(n),
		}
	}

	prices := extractPrices(candles, source)
	sma := CalculateSMA(candles, period, source)

	result := &BollingerResult{
		Upper:     nanSlice(n),
		Middle:    sma,
		Lower:     nanSlice(n),
		Bandwidth: nanSlice(n),
		PercentB:  nanSlice(n),
	}

	for i := period - 1; i < n; i++ {
		if !isValidValue(sma[i]) {
			continue
		}

		// Calculate standard deviation
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			diff := prices[j] - sma[i]
			sum += diff * diff
		}
		sd := math.Sqrt(sum / float64(period))

		result.Upper[i] = sma[i] + stdDev*sd
		result.Lower[i] = sma[i] - stdDev*sd

		// Bandwidth
		if sma[i] != 0 {
			result.Bandwidth[i] = (result.Upper[i] - result.Lower[i]) / sma[i] * 100
		}

		// Percent B
		if result.Upper[i] != result.Lower[i] {
			result.PercentB[i] = (prices[i] - result.Lower[i]) / (result.Upper[i] - result.Lower[i])
		}
	}

	return result
}
