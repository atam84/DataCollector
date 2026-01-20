package indicators

import "github.com/yourusername/datacollector/internal/models"

// CalculateVWAP calculates Volume Weighted Average Price
// VWAP = Sum(Typical Price * Volume) / Sum(Volume)
// Cumulative from start of data or can reset daily
func CalculateVWAP(candles []models.Candle) []float64 {
	n := len(candles)
	if n == 0 {
		return nanSlice(n)
	}

	result := make([]float64, n)

	cumulativeTPV := 0.0 // Cumulative (Typical Price * Volume)
	cumulativeVolume := 0.0

	for i := 0; i < n; i++ {
		// Typical Price = (High + Low + Close) / 3
		typicalPrice := (candles[i].High + candles[i].Low + candles[i].Close) / 3

		// Add to cumulative values
		cumulativeTPV += typicalPrice * candles[i].Volume
		cumulativeVolume += candles[i].Volume

		// Calculate VWAP
		if cumulativeVolume > 0 {
			result[i] = cumulativeTPV / cumulativeVolume
		}
	}

	return result
}

// CalculateVWAPWithReset calculates VWAP that resets at specified intervals
// This is useful for daily VWAP calculations
func CalculateVWAPWithReset(candles []models.Candle, resetPeriod int) []float64 {
	n := len(candles)
	if n == 0 {
		return nanSlice(n)
	}

	result := make([]float64, n)

	cumulativeTPV := 0.0
	cumulativeVolume := 0.0

	for i := 0; i < n; i++ {
		// Reset cumulative values at intervals
		if i%resetPeriod == 0 {
			cumulativeTPV = 0.0
			cumulativeVolume = 0.0
		}

		// Typical Price = (High + Low + Close) / 3
		typicalPrice := (candles[i].High + candles[i].Low + candles[i].Close) / 3

		// Add to cumulative values
		cumulativeTPV += typicalPrice * candles[i].Volume
		cumulativeVolume += candles[i].Volume

		// Calculate VWAP
		if cumulativeVolume > 0 {
			result[i] = cumulativeTPV / cumulativeVolume
		}
	}

	return result
}
