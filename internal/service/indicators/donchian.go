package indicators

import "github.com/yourusername/datacollector/internal/models"

// DonchianResult holds Donchian Channel components
type DonchianResult struct {
	Upper  []float64 // Highest high over period
	Middle []float64 // (Upper + Lower) / 2
	Lower  []float64 // Lowest low over period
}

// CalculateDonchian calculates Donchian Channels
// Upper = Highest high over period
// Lower = Lowest low over period
// Middle = (Upper + Lower) / 2
func CalculateDonchian(candles []models.Candle, period int) *DonchianResult {
	n := len(candles)
	if n < period {
		return &DonchianResult{
			Upper:  nanSlice(n),
			Middle: nanSlice(n),
			Lower:  nanSlice(n),
		}
	}

	result := &DonchianResult{
		Upper:  nanSlice(n),
		Middle: nanSlice(n),
		Lower:  nanSlice(n),
	}

	for i := period - 1; i < n; i++ {
		// Find highest high and lowest low in period
		highest := candles[i].High
		lowest := candles[i].Low

		for j := i - period + 1; j < i; j++ {
			if candles[j].High > highest {
				highest = candles[j].High
			}
			if candles[j].Low < lowest {
				lowest = candles[j].Low
			}
		}

		result.Upper[i] = highest
		result.Lower[i] = lowest
		result.Middle[i] = (highest + lowest) / 2
	}

	return result
}
