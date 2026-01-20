package indicators

import "github.com/yourusername/datacollector/internal/models"

// KeltnerResult holds Keltner Channel components
type KeltnerResult struct {
	Upper  []float64 // Upper channel
	Middle []float64 // Middle line (EMA)
	Lower  []float64 // Lower channel
}

// CalculateKeltner calculates Keltner Channels
// Middle = EMA(close, period)
// Upper = Middle + (multiplier * ATR)
// Lower = Middle - (multiplier * ATR)
func CalculateKeltner(candles []models.Candle, period, atrPeriod int, multiplier float64) *KeltnerResult {
	n := len(candles)
	maxPeriod := period
	if atrPeriod > maxPeriod {
		maxPeriod = atrPeriod
	}

	if n < maxPeriod+1 {
		return &KeltnerResult{
			Upper:  nanSlice(n),
			Middle: nanSlice(n),
			Lower:  nanSlice(n),
		}
	}

	result := &KeltnerResult{
		Upper:  nanSlice(n),
		Middle: nanSlice(n),
		Lower:  nanSlice(n),
	}

	// Calculate middle line (EMA of close)
	middle := CalculateEMA(candles, period, "close")
	copy(result.Middle, middle)

	// Calculate ATR
	atr := CalculateATR(candles, atrPeriod)

	// Calculate upper and lower bands
	for i := maxPeriod; i < n; i++ {
		if isValidValue(middle[i]) && isValidValue(atr[i]) {
			result.Upper[i] = middle[i] + (multiplier * atr[i])
			result.Lower[i] = middle[i] - (multiplier * atr[i])
		}
	}

	return result
}
