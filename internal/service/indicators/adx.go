package indicators

import (
	"math"

	"github.com/yourusername/datacollector/internal/models"
)

// ADXResult holds ADX and Directional Movement components
type ADXResult struct {
	ADX     []float64 // Average Directional Index
	PlusDI  []float64 // +DI (Positive Directional Indicator)
	MinusDI []float64 // -DI (Negative Directional Indicator)
}

// CalculateADX calculates Average Directional Index and directional indicators
// ADX measures trend strength (0-100), not direction
// +DI and -DI show trend direction
func CalculateADX(candles []models.Candle, period int) *ADXResult {
	n := len(candles)
	if period <= 0 || n < period*2 {
		return &ADXResult{
			ADX:     nanSlice(n),
			PlusDI:  nanSlice(n),
			MinusDI: nanSlice(n),
		}
	}

	result := &ADXResult{
		ADX:     nanSlice(n),
		PlusDI:  nanSlice(n),
		MinusDI: nanSlice(n),
	}

	// Calculate True Range and Directional Movements
	tr := make([]float64, n)
	plusDM := make([]float64, n)
	minusDM := make([]float64, n)

	for i := 1; i < n; i++ {
		// True Range
		highLow := candles[i].High - candles[i].Low
		highClosePrev := math.Abs(candles[i].High - candles[i-1].Close)
		lowClosePrev := math.Abs(candles[i].Low - candles[i-1].Close)
		tr[i] = math.Max(highLow, math.Max(highClosePrev, lowClosePrev))

		// Directional Movements
		upMove := candles[i].High - candles[i-1].High
		downMove := candles[i-1].Low - candles[i].Low

		if upMove > downMove && upMove > 0 {
			plusDM[i] = upMove
		}
		if downMove > upMove && downMove > 0 {
			minusDM[i] = downMove
		}
	}

	// Smooth TR, +DM, and -DM using Wilder's smoothing
	smoothTR := make([]float64, n)
	smoothPlusDM := make([]float64, n)
	smoothMinusDM := make([]float64, n)

	// Initial values (sum of first period)
	for i := 1; i <= period; i++ {
		smoothTR[period] += tr[i]
		smoothPlusDM[period] += plusDM[i]
		smoothMinusDM[period] += minusDM[i]
	}

	// Wilder's smoothing
	for i := period + 1; i < n; i++ {
		smoothTR[i] = smoothTR[i-1] - (smoothTR[i-1] / float64(period)) + tr[i]
		smoothPlusDM[i] = smoothPlusDM[i-1] - (smoothPlusDM[i-1] / float64(period)) + plusDM[i]
		smoothMinusDM[i] = smoothMinusDM[i-1] - (smoothMinusDM[i-1] / float64(period)) + minusDM[i]
	}

	// Calculate +DI and -DI
	dx := make([]float64, n)
	for i := period; i < n; i++ {
		if smoothTR[i] > 0 {
			result.PlusDI[i] = (smoothPlusDM[i] / smoothTR[i]) * 100
			result.MinusDI[i] = (smoothMinusDM[i] / smoothTR[i]) * 100

			// Calculate DX (Directional Index)
			diDiff := math.Abs(result.PlusDI[i] - result.MinusDI[i])
			diSum := result.PlusDI[i] + result.MinusDI[i]
			if diSum > 0 {
				dx[i] = (diDiff / diSum) * 100
			}
		}
	}

	// Calculate ADX (smoothed DX)
	// Initial ADX (average of first period DX values)
	adxStart := period * 2 - 1
	if adxStart >= n {
		return result
	}

	sum := 0.0
	for i := period; i < adxStart; i++ {
		sum += dx[i]
	}
	result.ADX[adxStart] = sum / float64(period)

	// Wilder's smoothing for ADX
	for i := adxStart + 1; i < n; i++ {
		result.ADX[i] = ((result.ADX[i-1] * float64(period-1)) + dx[i]) / float64(period)
	}

	return result
}
