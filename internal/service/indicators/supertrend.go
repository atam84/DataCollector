package indicators

import "github.com/yourusername/datacollector/internal/models"

// SuperTrendResult holds SuperTrend values and signals
type SuperTrendResult struct {
	SuperTrend []float64 // SuperTrend line value
	Signal     []int     // 1 = buy (uptrend), -1 = sell (downtrend), 0 = neutral
}

// CalculateSuperTrend calculates SuperTrend indicator
// SuperTrend is a trend-following indicator based on ATR
// It provides dynamic support/resistance levels
func CalculateSuperTrend(candles []models.Candle, period int, multiplier float64) *SuperTrendResult {
	n := len(candles)
	if n < period+1 {
		return &SuperTrendResult{
			SuperTrend: nanSlice(n),
			Signal:     make([]int, n),
		}
	}

	// Calculate ATR
	atr := CalculateATR(candles, period)

	result := &SuperTrendResult{
		SuperTrend: nanSlice(n),
		Signal:     make([]int, n),
	}

	// Calculate basic bands
	upperBand := make([]float64, n)
	lowerBand := make([]float64, n)
	finalUpperBand := nanSlice(n)
	finalLowerBand := nanSlice(n)

	for i := period; i < n; i++ {
		if !isValidValue(atr[i]) {
			continue
		}

		// HL/2 (typical price)
		hl2 := (candles[i].High + candles[i].Low) / 2

		// Basic bands
		upperBand[i] = hl2 + (multiplier * atr[i])
		lowerBand[i] = hl2 - (multiplier * atr[i])

		// Final bands (apply conditions)
		if i == period {
			finalUpperBand[i] = upperBand[i]
			finalLowerBand[i] = lowerBand[i]
		} else {
			// Upper band
			if upperBand[i] < finalUpperBand[i-1] || candles[i-1].Close > finalUpperBand[i-1] {
				finalUpperBand[i] = upperBand[i]
			} else {
				finalUpperBand[i] = finalUpperBand[i-1]
			}

			// Lower band
			if lowerBand[i] > finalLowerBand[i-1] || candles[i-1].Close < finalLowerBand[i-1] {
				finalLowerBand[i] = lowerBand[i]
			} else {
				finalLowerBand[i] = finalLowerBand[i-1]
			}
		}

		// Determine SuperTrend value and signal
		if i == period {
			// Initial trend
			if candles[i].Close <= finalUpperBand[i] {
				result.SuperTrend[i] = finalUpperBand[i]
				result.Signal[i] = -1 // Downtrend
			} else {
				result.SuperTrend[i] = finalLowerBand[i]
				result.Signal[i] = 1 // Uptrend
			}
		} else {
			// Continue previous trend or flip
			prevSignal := result.Signal[i-1]

			if prevSignal == 1 {
				// Was in uptrend
				if candles[i].Close <= finalLowerBand[i] {
					// Flip to downtrend
					result.SuperTrend[i] = finalUpperBand[i]
					result.Signal[i] = -1
				} else {
					// Continue uptrend
					result.SuperTrend[i] = finalLowerBand[i]
					result.Signal[i] = 1
				}
			} else {
				// Was in downtrend
				if candles[i].Close >= finalUpperBand[i] {
					// Flip to uptrend
					result.SuperTrend[i] = finalLowerBand[i]
					result.Signal[i] = 1
				} else {
					// Continue downtrend
					result.SuperTrend[i] = finalUpperBand[i]
					result.Signal[i] = -1
				}
			}
		}
	}

	return result
}
