package indicators

import (
	"github.com/yourusername/datacollector/internal/models"
)

// MACDResult holds the MACD indicator values
type MACDResult struct {
	MACD      []float64 // MACD line
	Signal    []float64 // Signal line
	Histogram []float64 // MACD - Signal
}

// CalculateMACD calculates Moving Average Convergence Divergence
// Returns MACD line, Signal line, and Histogram
func CalculateMACD(candles []models.Candle, fastPeriod, slowPeriod, signalPeriod int, source string) *MACDResult {
	n := len(candles)
	if n < slowPeriod+signalPeriod-1 {
		return &MACDResult{
			MACD:      nanSlice(n),
			Signal:    nanSlice(n),
			Histogram: nanSlice(n),
		}
	}

	// Calculate EMAs
	fastEMA := CalculateEMA(candles, fastPeriod, source)
	slowEMA := CalculateEMA(candles, slowPeriod, source)

	result := &MACDResult{
		MACD:      nanSlice(n),
		Signal:    nanSlice(n),
		Histogram: nanSlice(n),
	}

	// Calculate MACD line
	for i := slowPeriod - 1; i < n; i++ {
		if isValidValue(fastEMA[i]) && isValidValue(slowEMA[i]) {
			result.MACD[i] = fastEMA[i] - slowEMA[i]
		}
	}

	// Calculate Signal line (EMA of MACD)
	signalMultiplier := 2.0 / float64(signalPeriod+1)

	// Initial signal (SMA of first 'signalPeriod' MACD values)
	sum := 0.0
	firstSignalIndex := slowPeriod + signalPeriod - 2
	for i := slowPeriod - 1; i < firstSignalIndex; i++ {
		if isValidValue(result.MACD[i]) {
			sum += result.MACD[i]
		}
	}
	if firstSignalIndex < n {
		result.Signal[firstSignalIndex] = sum / float64(signalPeriod)

		// EMA for subsequent signal values
		for i := firstSignalIndex + 1; i < n; i++ {
			if isValidValue(result.MACD[i]) && isValidValue(result.Signal[i-1]) {
				result.Signal[i] = (result.MACD[i]-result.Signal[i-1])*signalMultiplier + result.Signal[i-1]
			}
		}
	}

	// Calculate Histogram
	for i := 0; i < n; i++ {
		if isValidValue(result.MACD[i]) && isValidValue(result.Signal[i]) {
			result.Histogram[i] = result.MACD[i] - result.Signal[i]
		}
	}

	return result
}
