package indicators

import (
	"math"

	"github.com/yourusername/datacollector/internal/models"
)

// CalculateCCI calculates Commodity Channel Index
// CCI = (Typical Price - SMA of Typical Price) / (0.015 * Mean Deviation)
// Identifies cyclical trends, typically oscillates between -100 and +100
func CalculateCCI(candles []models.Candle, period int) []float64 {
	n := len(candles)
	if period <= 0 || n < period {
		return nanSlice(n)
	}

	result := nanSlice(n)

	// Calculate typical price for each candle
	typicalPrice := make([]float64, n)
	for i := range candles {
		typicalPrice[i] = (candles[i].High + candles[i].Low + candles[i].Close) / 3
	}

	// Calculate SMA of typical price
	smaTP := make([]float64, n)
	for i := period - 1; i < n; i++ {
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += typicalPrice[j]
		}
		smaTP[i] = sum / float64(period)
	}

	// Calculate CCI
	for i := period - 1; i < n; i++ {
		// Calculate mean deviation
		meanDev := 0.0
		for j := i - period + 1; j <= i; j++ {
			meanDev += math.Abs(typicalPrice[j] - smaTP[i])
		}
		meanDev /= float64(period)

		// CCI formula
		if meanDev != 0 {
			result[i] = (typicalPrice[i] - smaTP[i]) / (0.015 * meanDev)
		}
	}

	return result
}
