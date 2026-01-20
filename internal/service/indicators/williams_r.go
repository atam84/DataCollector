package indicators

import "github.com/yourusername/datacollector/internal/models"

// CalculateWilliamsR calculates Williams %R
// %R = (Highest High - Close) / (Highest High - Lowest Low) * -100
// Momentum indicator, ranges from 0 to -100
// Values above -20 indicate overbought, below -80 indicate oversold
func CalculateWilliamsR(candles []models.Candle, period int) []float64 {
	n := len(candles)
	if n < period {
		return nanSlice(n)
	}

	result := nanSlice(n)

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

		// Calculate Williams %R
		if highest != lowest {
			result[i] = ((highest - candles[i].Close) / (highest - lowest)) * -100
		} else {
			result[i] = -50 // Neutral if no range
		}
	}

	return result
}
