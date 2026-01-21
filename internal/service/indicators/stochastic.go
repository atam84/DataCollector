package indicators

import "github.com/yourusername/datacollector/internal/models"

// StochasticResult holds %K and %D components
type StochasticResult struct {
	K []float64 // %K (fast stochastic)
	D []float64 // %D (slow stochastic, SMA of %K)
}

// CalculateStochastic calculates Stochastic Oscillator
// Compares closing price to price range over a period
// Values range from 0 to 100
func CalculateStochastic(candles []models.Candle, kPeriod, dPeriod, smooth int) *StochasticResult {
	n := len(candles)
	if kPeriod <= 0 || dPeriod <= 0 || n < kPeriod+dPeriod {
		return &StochasticResult{
			K: nanSlice(n),
			D: nanSlice(n),
		}
	}

	result := &StochasticResult{
		K: nanSlice(n),
		D: nanSlice(n),
	}

	// Calculate raw %K
	rawK := make([]float64, n)
	for i := kPeriod - 1; i < n; i++ {
		// Find highest high and lowest low in period
		highest := candles[i].High
		lowest := candles[i].Low

		for j := i - kPeriod + 1; j < i; j++ {
			if candles[j].High > highest {
				highest = candles[j].High
			}
			if candles[j].Low < lowest {
				lowest = candles[j].Low
			}
		}

		// Calculate %K
		if highest != lowest {
			rawK[i] = ((candles[i].Close - lowest) / (highest - lowest)) * 100
		} else {
			rawK[i] = 50 // Neutral if no range
		}
	}

	// Smooth %K if smooth > 1
	if smooth > 1 {
		for i := kPeriod + smooth - 2; i < n; i++ {
			sum := 0.0
			for j := 0; j < smooth; j++ {
				sum += rawK[i-j]
			}
			result.K[i] = sum / float64(smooth)
		}
	} else {
		copy(result.K, rawK)
	}

	// Calculate %D (SMA of %K)
	for i := kPeriod + dPeriod - 1; i < n; i++ {
		sum := 0.0
		count := 0
		for j := 0; j < dPeriod; j++ {
			if isValidValue(result.K[i-j]) {
				sum += result.K[i-j]
				count++
			}
		}
		if count > 0 {
			result.D[i] = sum / float64(count)
		}
	}

	return result
}
