package indicators

import "github.com/yourusername/datacollector/internal/models"

// CalculateTEMA calculates Triple Exponential Moving Average
// TEMA = 3 * EMA - 3 * EMA(EMA) + EMA(EMA(EMA))
// Even more responsive than DEMA, minimal lag
func CalculateTEMA(candles []models.Candle, period int, source string) []float64 {
	n := len(candles)
	if n < period*3 {
		return nanSlice(n)
	}

	result := nanSlice(n)

	// Calculate first EMA
	ema1 := CalculateEMA(candles, period, source)

	// Create temporary candles for second EMA
	tempCandles1 := make([]models.Candle, n)
	for i := range candles {
		tempCandles1[i] = candles[i]
		if isValidValue(ema1[i]) {
			tempCandles1[i].Close = ema1[i]
		}
	}

	// Calculate EMA of EMA
	ema2 := CalculateEMA(tempCandles1, period, "close")

	// Create temporary candles for third EMA
	tempCandles2 := make([]models.Candle, n)
	for i := range candles {
		tempCandles2[i] = candles[i]
		if isValidValue(ema2[i]) {
			tempCandles2[i].Close = ema2[i]
		}
	}

	// Calculate EMA of EMA of EMA
	ema3 := CalculateEMA(tempCandles2, period, "close")

	// TEMA = 3 * EMA - 3 * EMA(EMA) + EMA(EMA(EMA))
	for i := range result {
		if isValidValue(ema1[i]) && isValidValue(ema2[i]) && isValidValue(ema3[i]) {
			result[i] = 3*ema1[i] - 3*ema2[i] + ema3[i]
		}
	}

	return result
}
