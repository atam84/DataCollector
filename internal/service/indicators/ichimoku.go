package indicators

import "github.com/yourusername/datacollector/internal/models"

// IchimokuResult holds all Ichimoku Cloud components
type IchimokuResult struct {
	Tenkan  []float64 // Conversion Line (Tenkan-sen)
	Kijun   []float64 // Base Line (Kijun-sen)
	SenkouA []float64 // Leading Span A (Senkou Span A)
	SenkouB []float64 // Leading Span B (Senkou Span B)
	Chikou  []float64 // Lagging Span (Chikou Span)
}

// CalculateIchimoku calculates the Ichimoku Cloud indicator
// Returns all 5 components of the Ichimoku system
func CalculateIchimoku(candles []models.Candle, tenkanPeriod, kijunPeriod, senkouBPeriod, displacementFwd, displacementBck int) *IchimokuResult {
	n := len(candles)
	maxPeriod := senkouBPeriod
	if kijunPeriod > maxPeriod {
		maxPeriod = kijunPeriod
	}

	if n < maxPeriod+displacementFwd {
		return &IchimokuResult{
			Tenkan:  nanSlice(n),
			Kijun:   nanSlice(n),
			SenkouA: nanSlice(n),
			SenkouB: nanSlice(n),
			Chikou:  nanSlice(n),
		}
	}

	result := &IchimokuResult{
		Tenkan:  nanSlice(n),
		Kijun:   nanSlice(n),
		SenkouA: nanSlice(n),
		SenkouB: nanSlice(n),
		Chikou:  nanSlice(n),
	}

	// Helper function to calculate midpoint of high-low range
	midpoint := func(start, end int) float64 {
		highest := candles[start].High
		lowest := candles[start].Low
		for i := start + 1; i <= end; i++ {
			if candles[i].High > highest {
				highest = candles[i].High
			}
			if candles[i].Low < lowest {
				lowest = candles[i].Low
			}
		}
		return (highest + lowest) / 2
	}

	// Calculate Tenkan-sen (Conversion Line)
	for i := tenkanPeriod - 1; i < n; i++ {
		result.Tenkan[i] = midpoint(i-tenkanPeriod+1, i)
	}

	// Calculate Kijun-sen (Base Line)
	for i := kijunPeriod - 1; i < n; i++ {
		result.Kijun[i] = midpoint(i-kijunPeriod+1, i)
	}

	// Calculate Senkou Span A (Leading Span A) - displaced forward
	// Senkou A = (Tenkan + Kijun) / 2, displaced forward
	for i := kijunPeriod - 1; i < n; i++ {
		if isValidValue(result.Tenkan[i]) && isValidValue(result.Kijun[i]) {
			futureIndex := i + displacementFwd
			if futureIndex < n {
				result.SenkouA[futureIndex] = (result.Tenkan[i] + result.Kijun[i]) / 2
			}
		}
	}

	// Calculate Senkou Span B (Leading Span B) - displaced forward
	// Senkou B = midpoint of 52-period high-low, displaced forward
	for i := senkouBPeriod - 1; i < n; i++ {
		futureIndex := i + displacementFwd
		if futureIndex < n {
			result.SenkouB[futureIndex] = midpoint(i-senkouBPeriod+1, i)
		}
	}

	// Calculate Chikou Span (Lagging Span) - displaced backward
	// Chikou = current close, displaced backward
	for i := displacementBck; i < n; i++ {
		result.Chikou[i-displacementBck] = candles[i].Close
	}

	return result
}
