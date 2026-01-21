package indicators

import (
	"log"

	"github.com/yourusername/datacollector/internal/models"
)

// Service handles indicator calculations for candles
type Service struct{}

// NewService creates a new indicator service
func NewService() *Service {
	return &Service{}
}

// CalculateAll calculates ALL indicators and updates the candles
// Note: Candles should be ordered with newest first (index 0)
// The function handles reversal automatically
func (s *Service) CalculateAll(candles []models.Candle) ([]models.Candle, error) {
	if len(candles) == 0 {
		return candles, nil
	}

	// Reverse candles for calculation (indicators expect oldest first)
	reversed := make([]models.Candle, len(candles))
	for i := range candles {
		reversed[i] = candles[len(candles)-1-i]
	}

	log.Printf("[INDICATORS] Calculating all indicators for %d candles", len(reversed))

	// Initialize indicators map for each candle
	for i := range reversed {
		if reversed[i].Indicators == (models.Indicators{}) {
			reversed[i].Indicators = models.Indicators{}
		}
	}

	// Calculate all indicators with default periods
	s.calculateTrendIndicators(reversed)
	s.calculateMomentumIndicators(reversed)
	s.calculateVolatilityIndicators(reversed)
	s.calculateVolumeIndicators(reversed)

	// Reverse back to newest-first order
	for i := range candles {
		candles[i] = reversed[len(reversed)-1-i]
	}

	log.Printf("[INDICATORS] All indicator calculations complete")
	return candles, nil
}

// calculateTrendIndicators calculates all trend-following indicators
func (s *Service) calculateTrendIndicators(candles []models.Candle) {
	// SMA (Simple Moving Average) - periods: 20, 50, 200
	for _, period := range []int{20, 50, 200} {
		sma := CalculateSMA(candles, period, "close")
		for i, val := range sma {
			if isValidValue(val) {
				switch period {
				case 20:
					candles[i].Indicators.SMA20 = &val
				case 50:
					candles[i].Indicators.SMA50 = &val
				case 200:
					candles[i].Indicators.SMA200 = &val
				}
			}
		}
	}
	log.Printf("[INDICATORS] Calculated SMA(20, 50, 200)")

	// EMA (Exponential Moving Average) - periods: 12, 26, 50
	for _, period := range []int{12, 26, 50} {
		ema := CalculateEMA(candles, period, "close")
		for i, val := range ema {
			if isValidValue(val) {
				switch period {
				case 12:
					candles[i].Indicators.EMA12 = &val
				case 26:
					candles[i].Indicators.EMA26 = &val
				case 50:
					candles[i].Indicators.EMA50 = &val
				}
			}
		}
	}
	log.Printf("[INDICATORS] Calculated EMA(12, 26, 50)")

	// DEMA (Double Exponential Moving Average) - period: 20
	dema := CalculateDEMA(candles, 20, "close")
	for i, val := range dema {
		if isValidValue(val) {
			candles[i].Indicators.DEMA = &val
		}
	}
	log.Printf("[INDICATORS] Calculated DEMA(20)")

	// TEMA (Triple Exponential Moving Average) - period: 20
	tema := CalculateTEMA(candles, 20, "close")
	for i, val := range tema {
		if isValidValue(val) {
			candles[i].Indicators.TEMA = &val
		}
	}
	log.Printf("[INDICATORS] Calculated TEMA(20)")

	// WMA (Weighted Moving Average) - period: 20
	wma := CalculateWMA(candles, 20, "close")
	for i, val := range wma {
		if isValidValue(val) {
			candles[i].Indicators.WMA = &val
		}
	}
	log.Printf("[INDICATORS] Calculated WMA(20)")

	// HMA (Hull Moving Average) - period: 9
	hma := CalculateHMA(candles, 9, "close")
	for i, val := range hma {
		if isValidValue(val) {
			candles[i].Indicators.HMA = &val
		}
	}
	log.Printf("[INDICATORS] Calculated HMA(9)")

	// VWMA (Volume Weighted Moving Average) - period: 20
	vwma := CalculateVWMA(candles, 20, "close")
	for i, val := range vwma {
		if isValidValue(val) {
			candles[i].Indicators.VWMA = &val
		}
	}
	log.Printf("[INDICATORS] Calculated VWMA(20)")

	// Ichimoku Cloud - tenkan:9, kijun:26, senkouB:52, displacement:26
	ichimoku := CalculateIchimoku(candles, 9, 26, 52, 26, 26)
	for i := range candles {
		if isValidValue(ichimoku.Tenkan[i]) {
			candles[i].Indicators.IchimokuTenkan = &ichimoku.Tenkan[i]
		}
		if isValidValue(ichimoku.Kijun[i]) {
			candles[i].Indicators.IchimokuKijun = &ichimoku.Kijun[i]
		}
		if isValidValue(ichimoku.SenkouA[i]) {
			candles[i].Indicators.IchimokuSenkouA = &ichimoku.SenkouA[i]
		}
		if isValidValue(ichimoku.SenkouB[i]) {
			candles[i].Indicators.IchimokuSenkouB = &ichimoku.SenkouB[i]
		}
		if isValidValue(ichimoku.Chikou[i]) {
			candles[i].Indicators.IchimokuChikou = &ichimoku.Chikou[i]
		}
	}
	log.Printf("[INDICATORS] Calculated Ichimoku(9, 26, 52)")

	// ADX/DMI - period: 14
	adx := CalculateADX(candles, 14)
	for i := range candles {
		if isValidValue(adx.ADX[i]) {
			candles[i].Indicators.ADX = &adx.ADX[i]
		}
		if isValidValue(adx.PlusDI[i]) {
			candles[i].Indicators.PlusDI = &adx.PlusDI[i]
		}
		if isValidValue(adx.MinusDI[i]) {
			candles[i].Indicators.MinusDI = &adx.MinusDI[i]
		}
	}
	log.Printf("[INDICATORS] Calculated ADX(14)")

	// SuperTrend - period: 10, multiplier: 3.0
	supertrend := CalculateSuperTrend(candles, 10, 3.0)
	for i := range candles {
		if isValidValue(supertrend.SuperTrend[i]) {
			candles[i].Indicators.SuperTrend = &supertrend.SuperTrend[i]
		}
		if supertrend.Signal[i] != 0 {
			candles[i].Indicators.SuperTrendSignal = &supertrend.Signal[i]
		}
	}
	log.Printf("[INDICATORS] Calculated SuperTrend(10, 3.0)")
}

// calculateMomentumIndicators calculates all momentum/oscillator indicators
func (s *Service) calculateMomentumIndicators(candles []models.Candle) {
	// RSI (Relative Strength Index) - periods: 6, 14, 24
	for _, period := range []int{6, 14, 24} {
		rsi := CalculateRSI(candles, period, "close")
		for i, val := range rsi {
			if isValidValue(val) {
				switch period {
				case 6:
					candles[i].Indicators.RSI6 = &val
				case 14:
					candles[i].Indicators.RSI14 = &val
				case 24:
					candles[i].Indicators.RSI24 = &val
				}
			}
		}
	}
	log.Printf("[INDICATORS] Calculated RSI(6, 14, 24)")

	// Stochastic Oscillator - K:14, D:3, smooth:3
	stoch := CalculateStochastic(candles, 14, 3, 3)
	for i := range candles {
		if isValidValue(stoch.K[i]) {
			candles[i].Indicators.StochK = &stoch.K[i]
		}
		if isValidValue(stoch.D[i]) {
			candles[i].Indicators.StochD = &stoch.D[i]
		}
	}
	log.Printf("[INDICATORS] Calculated Stochastic(14, 3, 3)")

	// MACD - fast:12, slow:26, signal:9
	macd := CalculateMACD(candles, 12, 26, 9, "close")
	for i := range candles {
		if isValidValue(macd.MACD[i]) {
			candles[i].Indicators.MACD = &macd.MACD[i]
		}
		if isValidValue(macd.Signal[i]) {
			candles[i].Indicators.MACDSignal = &macd.Signal[i]
		}
		if isValidValue(macd.Histogram[i]) {
			candles[i].Indicators.MACDHist = &macd.Histogram[i]
		}
	}
	log.Printf("[INDICATORS] Calculated MACD(12, 26, 9)")

	// ROC (Rate of Change) - period: 12
	roc := CalculateROC(candles, 12, "close")
	for i, val := range roc {
		if isValidValue(val) {
			candles[i].Indicators.ROC = &val
		}
	}
	log.Printf("[INDICATORS] Calculated ROC(12)")

	// CCI (Commodity Channel Index) - period: 20
	cci := CalculateCCI(candles, 20)
	for i, val := range cci {
		if isValidValue(val) {
			candles[i].Indicators.CCI = &val
		}
	}
	log.Printf("[INDICATORS] Calculated CCI(20)")

	// Williams %R - period: 14
	williamsR := CalculateWilliamsR(candles, 14)
	for i, val := range williamsR {
		if isValidValue(val) {
			candles[i].Indicators.WilliamsR = &val
		}
	}
	log.Printf("[INDICATORS] Calculated Williams %%R(14)")

	// Momentum - period: 10
	momentum := CalculateMomentum(candles, 10, "close")
	for i, val := range momentum {
		if isValidValue(val) {
			candles[i].Indicators.Momentum = &val
		}
	}
	log.Printf("[INDICATORS] Calculated Momentum(10)")
}

// calculateVolatilityIndicators calculates all volatility indicators
func (s *Service) calculateVolatilityIndicators(candles []models.Candle) {
	// Bollinger Bands - period: 20, stdDev: 2.0
	bb := CalculateBollingerBands(candles, 20, 2.0, "close")
	for i := range candles {
		if isValidValue(bb.Upper[i]) {
			candles[i].Indicators.BollingerUpper = &bb.Upper[i]
		}
		if isValidValue(bb.Middle[i]) {
			candles[i].Indicators.BollingerMiddle = &bb.Middle[i]
		}
		if isValidValue(bb.Lower[i]) {
			candles[i].Indicators.BollingerLower = &bb.Lower[i]
		}
		if isValidValue(bb.Bandwidth[i]) {
			candles[i].Indicators.BollingerBandwidth = &bb.Bandwidth[i]
		}
		if isValidValue(bb.PercentB[i]) {
			candles[i].Indicators.BollingerPercentB = &bb.PercentB[i]
		}
	}
	log.Printf("[INDICATORS] Calculated Bollinger Bands(20, 2.0)")

	// ATR (Average True Range) - period: 14
	atr := CalculateATR(candles, 14)
	for i, val := range atr {
		if isValidValue(val) {
			candles[i].Indicators.ATR = &val
		}
	}
	log.Printf("[INDICATORS] Calculated ATR(14)")

	// Keltner Channels - period: 20, atrPeriod: 10, multiplier: 2.0
	keltner := CalculateKeltner(candles, 20, 10, 2.0)
	for i := range candles {
		if isValidValue(keltner.Upper[i]) {
			candles[i].Indicators.KeltnerUpper = &keltner.Upper[i]
		}
		if isValidValue(keltner.Middle[i]) {
			candles[i].Indicators.KeltnerMiddle = &keltner.Middle[i]
		}
		if isValidValue(keltner.Lower[i]) {
			candles[i].Indicators.KeltnerLower = &keltner.Lower[i]
		}
	}
	log.Printf("[INDICATORS] Calculated Keltner Channels(20, 10, 2.0)")

	// Donchian Channels - period: 20
	donchian := CalculateDonchian(candles, 20)
	for i := range candles {
		if isValidValue(donchian.Upper[i]) {
			candles[i].Indicators.DonchianUpper = &donchian.Upper[i]
		}
		if isValidValue(donchian.Middle[i]) {
			candles[i].Indicators.DonchianMiddle = &donchian.Middle[i]
		}
		if isValidValue(donchian.Lower[i]) {
			candles[i].Indicators.DonchianLower = &donchian.Lower[i]
		}
	}
	log.Printf("[INDICATORS] Calculated Donchian Channels(20)")

	// Standard Deviation - period: 20
	stddev := CalculateStdDev(candles, 20, "close")
	for i, val := range stddev {
		if isValidValue(val) {
			candles[i].Indicators.StdDev = &val
		}
	}
	log.Printf("[INDICATORS] Calculated StdDev(20)")
}

// calculateVolumeIndicators calculates all volume-based indicators
func (s *Service) calculateVolumeIndicators(candles []models.Candle) {
	// OBV (On-Balance Volume)
	obv := CalculateOBV(candles)
	for i, val := range obv {
		if isValidValue(val) {
			candles[i].Indicators.OBV = &val
		}
	}
	log.Printf("[INDICATORS] Calculated OBV")

	// VWAP (Volume Weighted Average Price)
	vwap := CalculateVWAP(candles)
	for i, val := range vwap {
		if isValidValue(val) {
			candles[i].Indicators.VWAP = &val
		}
	}
	log.Printf("[INDICATORS] Calculated VWAP")

	// MFI (Money Flow Index) - period: 14
	mfi := CalculateMFI(candles, 14)
	for i, val := range mfi {
		if isValidValue(val) {
			candles[i].Indicators.MFI = &val
		}
	}
	log.Printf("[INDICATORS] Calculated MFI(14)")

	// CMF (Chaikin Money Flow) - period: 20
	cmf := CalculateCMF(candles, 20)
	for i, val := range cmf {
		if isValidValue(val) {
			candles[i].Indicators.CMF = &val
		}
	}
	log.Printf("[INDICATORS] Calculated CMF(20)")

	// Volume SMA - period: 20
	volumeSMA := CalculateVolumeSMA(candles, 20)
	for i, val := range volumeSMA {
		if isValidValue(val) {
			candles[i].Indicators.VolumeSMA = &val
		}
	}
	log.Printf("[INDICATORS] Calculated Volume SMA(20)")
}
