package indicators

import (
	"math"

	"github.com/yourusername/datacollector/internal/models"
)

// CalculateHMA calculates Hull Moving Average
// HMA = WMA(2 * WMA(n/2) - WMA(n)), period = sqrt(n)
// Extremely responsive with minimal lag
func CalculateHMA(candles []models.Candle, period int, source string) []float64 {
	n := len(candles)
	if period <= 0 || n < period {
		return nanSlice(n)
	}

	result := nanSlice(n)

	halfPeriod := period / 2
	sqrtPeriod := int(math.Sqrt(float64(period)))

	// Calculate WMA with half period
	wmaHalf := CalculateWMA(candles, halfPeriod, source)

	// Calculate WMA with full period
	wmaFull := CalculateWMA(candles, period, source)

	// Calculate 2 * WMA(n/2) - WMA(n)
	rawHMA := make([]float64, n)
	for i := range rawHMA {
		if isValidValue(wmaHalf[i]) && isValidValue(wmaFull[i]) {
			rawHMA[i] = 2*wmaHalf[i] - wmaFull[i]
		} else {
			rawHMA[i] = math.NaN()
		}
	}

	// Create temporary candles for final WMA
	tempCandles := make([]models.Candle, n)
	for i := range tempCandles {
		tempCandles[i] = candles[i]
		tempCandles[i].Close = rawHMA[i]
	}

	// Calculate WMA of the raw HMA with sqrt(period)
	result = CalculateWMA(tempCandles, sqrtPeriod, "close")

	return result
}
