package indicators

import "github.com/yourusername/datacollector/internal/models"

// CalculateMFI calculates Money Flow Index
// MFI is a volume-weighted RSI, ranges from 0 to 100
// Above 80 indicates overbought, below 20 indicates oversold
func CalculateMFI(candles []models.Candle, period int) []float64 {
	n := len(candles)
	if period <= 0 || n < period+1 {
		return nanSlice(n)
	}

	result := nanSlice(n)

	// Calculate typical price and money flow for each candle
	typicalPrice := make([]float64, n)
	moneyFlow := make([]float64, n)

	for i := 0; i < n; i++ {
		typicalPrice[i] = (candles[i].High + candles[i].Low + candles[i].Close) / 3
		moneyFlow[i] = typicalPrice[i] * candles[i].Volume
	}

	// Separate positive and negative money flows
	positiveFlow := make([]float64, n)
	negativeFlow := make([]float64, n)

	for i := 1; i < n; i++ {
		if typicalPrice[i] > typicalPrice[i-1] {
			positiveFlow[i] = moneyFlow[i]
		} else if typicalPrice[i] < typicalPrice[i-1] {
			negativeFlow[i] = moneyFlow[i]
		}
	}

	// Calculate MFI
	for i := period; i < n; i++ {
		sumPositive := 0.0
		sumNegative := 0.0

		for j := i - period + 1; j <= i; j++ {
			sumPositive += positiveFlow[j]
			sumNegative += negativeFlow[j]
		}

		if sumNegative == 0 {
			result[i] = 100
		} else {
			moneyRatio := sumPositive / sumNegative
			result[i] = 100 - (100 / (1 + moneyRatio))
		}
	}

	return result
}
