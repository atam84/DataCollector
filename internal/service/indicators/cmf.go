package indicators

import "github.com/yourusername/datacollector/internal/models"

// CalculateCMF calculates Chaikin Money Flow
// CMF measures buying and selling pressure over a period
// Positive values indicate buying pressure, negative indicate selling pressure
func CalculateCMF(candles []models.Candle, period int) []float64 {
	n := len(candles)
	if n < period {
		return nanSlice(n)
	}

	result := nanSlice(n)

	// Calculate Money Flow Multiplier and Money Flow Volume for each candle
	mfMultiplier := make([]float64, n)
	mfVolume := make([]float64, n)

	for i := 0; i < n; i++ {
		highLowDiff := candles[i].High - candles[i].Low
		if highLowDiff != 0 {
			// Money Flow Multiplier = ((Close - Low) - (High - Close)) / (High - Low)
			mfMultiplier[i] = ((candles[i].Close - candles[i].Low) - (candles[i].High - candles[i].Close)) / highLowDiff
		} else {
			mfMultiplier[i] = 0
		}

		// Money Flow Volume = MF Multiplier * Volume
		mfVolume[i] = mfMultiplier[i] * candles[i].Volume
	}

	// Calculate CMF
	for i := period - 1; i < n; i++ {
		sumMFVolume := 0.0
		sumVolume := 0.0

		for j := i - period + 1; j <= i; j++ {
			sumMFVolume += mfVolume[j]
			sumVolume += candles[j].Volume
		}

		if sumVolume > 0 {
			result[i] = sumMFVolume / sumVolume
		}
	}

	return result
}
