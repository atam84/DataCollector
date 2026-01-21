package indicators

import "github.com/yourusername/datacollector/internal/models"

// CalculateDEMA calculates Double Exponential Moving Average
// DEMA = 2 * EMA - EMA(EMA)
// More responsive than EMA, less lag
func CalculateDEMA(candles []models.Candle, period int, source string) []float64 {
	n := len(candles)
	if period <= 0 || n < period*2 {
		return nanSlice(n)
	}

	result := nanSlice(n)

	// Calculate first EMA
	ema1 := CalculateEMA(candles, period, source)

	// Create temporary candles for second EMA calculation
	// We need to convert EMA values back to candle format
	tempCandles := make([]models.Candle, n)
	for i := range candles {
		tempCandles[i] = candles[i]
		if isValidValue(ema1[i]) {
			tempCandles[i].Close = ema1[i]
		}
	}

	// Calculate EMA of EMA
	ema2 := CalculateEMA(tempCandles, period, "close")

	// DEMA = 2 * EMA - EMA(EMA)
	for i := range result {
		if isValidValue(ema1[i]) && isValidValue(ema2[i]) {
			result[i] = 2*ema1[i] - ema2[i]
		}
	}

	return result
}
