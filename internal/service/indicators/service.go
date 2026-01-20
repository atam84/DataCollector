package indicators

import (
	"fmt"
	"log"

	"github.com/yourusername/datacollector/internal/models"
)

// Service handles indicator calculations for candles
type Service struct{}

// NewService creates a new indicator service
func NewService() *Service {
	return &Service{}
}

// CalculateAll calculates all enabled indicators and updates the candles
// Note: Candles should be ordered with newest first (index 0)
// The function handles reversal automatically
func (s *Service) CalculateAll(candles []models.Candle, config *models.IndicatorConfig) ([]models.Candle, error) {
	if len(candles) == 0 {
		return candles, nil
	}

	if config == nil {
		config = DefaultConfig()
	}

	// Reverse candles for calculation (indicators expect oldest first)
	reversed := make([]models.Candle, len(candles))
	for i := range candles {
		reversed[i] = candles[len(candles)-1-i]
	}

	log.Printf("[INDICATORS] Calculating indicators for %d candles", len(reversed))

	// Initialize indicators map for each candle
	for i := range reversed {
		if reversed[i].Indicators == (models.Indicators{}) {
			reversed[i].Indicators = models.Indicators{}
		}
	}

	// Calculate all enabled indicators
	s.calculateTrendIndicators(reversed, config)
	s.calculateMomentumIndicators(reversed, config)
	s.calculateVolatilityIndicators(reversed, config)
	s.calculateVolumeIndicators(reversed, config)

	// Reverse back to newest-first order
	for i := range candles {
		candles[i] = reversed[len(reversed)-1-i]
	}

	log.Printf("[INDICATORS] Indicator calculation complete")
	return candles, nil
}

// calculateTrendIndicators calculates all trend-following indicators
func (s *Service) calculateTrendIndicators(candles []models.Candle, config *models.IndicatorConfig) {
	n := len(candles)

	// SMA (Simple Moving Average)
	if config.SMA != nil && config.SMA.Enabled {
		for _, period := range config.SMA.Periods {
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
			log.Printf("[INDICATORS] Calculated SMA(%d)", period)
		}
	}

	// EMA (Exponential Moving Average)
	if config.EMA != nil && config.EMA.Enabled {
		for _, period := range config.EMA.Periods {
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
			log.Printf("[INDICATORS] Calculated EMA(%d)", period)
		}
	}

	// DEMA (Double Exponential Moving Average)
	if config.DEMA != nil && config.DEMA.Enabled {
		dema := CalculateDEMA(candles, config.DEMA.Period, "close")
		for i, val := range dema {
			if isValidValue(val) {
				candles[i].Indicators.DEMA = &val
			}
		}
		log.Printf("[INDICATORS] Calculated DEMA(%d)", config.DEMA.Period)
	}

	// TEMA (Triple Exponential Moving Average)
	if config.TEMA != nil && config.TEMA.Enabled {
		tema := CalculateTEMA(candles, config.TEMA.Period, "close")
		for i, val := range tema {
			if isValidValue(val) {
				candles[i].Indicators.TEMA = &val
			}
		}
		log.Printf("[INDICATORS] Calculated TEMA(%d)", config.TEMA.Period)
	}

	// WMA (Weighted Moving Average)
	if config.WMA != nil && config.WMA.Enabled {
		wma := CalculateWMA(candles, config.WMA.Period, "close")
		for i, val := range wma {
			if isValidValue(val) {
				candles[i].Indicators.WMA = &val
			}
		}
		log.Printf("[INDICATORS] Calculated WMA(%d)", config.WMA.Period)
	}

	// HMA (Hull Moving Average)
	if config.HMA != nil && config.HMA.Enabled {
		hma := CalculateHMA(candles, config.HMA.Period, "close")
		for i, val := range hma {
			if isValidValue(val) {
				candles[i].Indicators.HMA = &val
			}
		}
		log.Printf("[INDICATORS] Calculated HMA(%d)", config.HMA.Period)
	}

	// VWMA (Volume Weighted Moving Average)
	if config.VWMA != nil && config.VWMA.Enabled {
		vwma := CalculateVWMA(candles, config.VWMA.Period, "close")
		for i, val := range vwma {
			if isValidValue(val) {
				candles[i].Indicators.VWMA = &val
			}
		}
		log.Printf("[INDICATORS] Calculated VWMA(%d)", config.VWMA.Period)
	}

	// Ichimoku Cloud
	if config.Ichimoku != nil && config.Ichimoku.Enabled {
		ichimoku := CalculateIchimoku(candles,
			config.Ichimoku.TenkanPeriod,
			config.Ichimoku.KijunPeriod,
			config.Ichimoku.SenkouBPeriod,
			config.Ichimoku.DisplacementFwd,
			config.Ichimoku.DisplacementBck)

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
		log.Printf("[INDICATORS] Calculated Ichimoku(%d,%d,%d)",
			config.Ichimoku.TenkanPeriod, config.Ichimoku.KijunPeriod, config.Ichimoku.SenkouBPeriod)
	}

	// ADX/DMI
	if config.ADX != nil && config.ADX.Enabled {
		adx := CalculateADX(candles, config.ADX.Period)
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
		log.Printf("[INDICATORS] Calculated ADX(%d)", config.ADX.Period)
	}

	// SuperTrend
	if config.SuperTrend != nil && config.SuperTrend.Enabled {
		supertrend := CalculateSuperTrend(candles, config.SuperTrend.Period, config.SuperTrend.Multiplier)
		for i := range candles {
			if isValidValue(supertrend.SuperTrend[i]) {
				candles[i].Indicators.SuperTrend = &supertrend.SuperTrend[i]
			}
			if supertrend.Signal[i] != 0 {
				candles[i].Indicators.SuperTrendSignal = &supertrend.Signal[i]
			}
		}
		log.Printf("[INDICATORS] Calculated SuperTrend(%d, %.1f)", config.SuperTrend.Period, config.SuperTrend.Multiplier)
	}

	_ = n // unused variable
}

// calculateMomentumIndicators calculates all momentum/oscillator indicators
func (s *Service) calculateMomentumIndicators(candles []models.Candle, config *models.IndicatorConfig) {
	// RSI (Relative Strength Index)
	if config.RSI != nil && config.RSI.Enabled {
		for _, period := range config.RSI.Periods {
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
			log.Printf("[INDICATORS] Calculated RSI(%d)", period)
		}
	}

	// Stochastic Oscillator
	if config.Stochastic != nil && config.Stochastic.Enabled {
		stoch := CalculateStochastic(candles,
			config.Stochastic.KPeriod,
			config.Stochastic.DPeriod,
			config.Stochastic.Smooth)

		for i := range candles {
			if isValidValue(stoch.K[i]) {
				candles[i].Indicators.StochK = &stoch.K[i]
			}
			if isValidValue(stoch.D[i]) {
				candles[i].Indicators.StochD = &stoch.D[i]
			}
		}
		log.Printf("[INDICATORS] Calculated Stochastic(%d,%d,%d)",
			config.Stochastic.KPeriod, config.Stochastic.DPeriod, config.Stochastic.Smooth)
	}

	// MACD
	if config.MACD != nil && config.MACD.Enabled {
		macd := CalculateMACD(candles,
			config.MACD.FastPeriod,
			config.MACD.SlowPeriod,
			config.MACD.SignalPeriod,
			"close")

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
		log.Printf("[INDICATORS] Calculated MACD(%d,%d,%d)",
			config.MACD.FastPeriod, config.MACD.SlowPeriod, config.MACD.SignalPeriod)
	}

	// ROC (Rate of Change)
	if config.ROC != nil && config.ROC.Enabled {
		roc := CalculateROC(candles, config.ROC.Period, "close")
		for i, val := range roc {
			if isValidValue(val) {
				candles[i].Indicators.ROC = &val
			}
		}
		log.Printf("[INDICATORS] Calculated ROC(%d)", config.ROC.Period)
	}

	// CCI (Commodity Channel Index)
	if config.CCI != nil && config.CCI.Enabled {
		cci := CalculateCCI(candles, config.CCI.Period)
		for i, val := range cci {
			if isValidValue(val) {
				candles[i].Indicators.CCI = &val
			}
		}
		log.Printf("[INDICATORS] Calculated CCI(%d)", config.CCI.Period)
	}

	// Williams %R
	if config.WilliamsR != nil && config.WilliamsR.Enabled {
		williamsR := CalculateWilliamsR(candles, config.WilliamsR.Period)
		for i, val := range williamsR {
			if isValidValue(val) {
				candles[i].Indicators.WilliamsR = &val
			}
		}
		log.Printf("[INDICATORS] Calculated Williams %%R(%d)", config.WilliamsR.Period)
	}

	// Momentum
	if config.Momentum != nil && config.Momentum.Enabled {
		momentum := CalculateMomentum(candles, config.Momentum.Period, "close")
		for i, val := range momentum {
			if isValidValue(val) {
				candles[i].Indicators.Momentum = &val
			}
		}
		log.Printf("[INDICATORS] Calculated Momentum(%d)", config.Momentum.Period)
	}
}

// calculateVolatilityIndicators calculates all volatility indicators
func (s *Service) calculateVolatilityIndicators(candles []models.Candle, config *models.IndicatorConfig) {
	// Bollinger Bands
	if config.Bollinger != nil && config.Bollinger.Enabled {
		bb := CalculateBollingerBands(candles,
			config.Bollinger.Period,
			config.Bollinger.StdDev,
			"close")

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
		log.Printf("[INDICATORS] Calculated Bollinger Bands(%d, %.1f)",
			config.Bollinger.Period, config.Bollinger.StdDev)
	}

	// ATR (Average True Range)
	if config.ATR != nil && config.ATR.Enabled {
		atr := CalculateATR(candles, config.ATR.Period)
		for i, val := range atr {
			if isValidValue(val) {
				candles[i].Indicators.ATR = &val
			}
		}
		log.Printf("[INDICATORS] Calculated ATR(%d)", config.ATR.Period)
	}

	// Keltner Channels
	if config.Keltner != nil && config.Keltner.Enabled {
		keltner := CalculateKeltner(candles,
			config.Keltner.Period,
			config.Keltner.ATRPeriod,
			config.Keltner.Multiplier)

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
		log.Printf("[INDICATORS] Calculated Keltner Channels(%d, %d, %.1f)",
			config.Keltner.Period, config.Keltner.ATRPeriod, config.Keltner.Multiplier)
	}

	// Donchian Channels
	if config.Donchian != nil && config.Donchian.Enabled {
		donchian := CalculateDonchian(candles, config.Donchian.Period)
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
		log.Printf("[INDICATORS] Calculated Donchian Channels(%d)", config.Donchian.Period)
	}

	// Standard Deviation
	if config.StdDev != nil && config.StdDev.Enabled {
		stddev := CalculateStdDev(candles, config.StdDev.Period, "close")
		for i, val := range stddev {
			if isValidValue(val) {
				candles[i].Indicators.StdDev = &val
			}
		}
		log.Printf("[INDICATORS] Calculated StdDev(%d)", config.StdDev.Period)
	}
}

// calculateVolumeIndicators calculates all volume-based indicators
func (s *Service) calculateVolumeIndicators(candles []models.Candle, config *models.IndicatorConfig) {
	// OBV (On-Balance Volume)
	if config.OBV != nil && config.OBV.Enabled {
		obv := CalculateOBV(candles)
		for i, val := range obv {
			if isValidValue(val) {
				candles[i].Indicators.OBV = &val
			}
		}
		log.Printf("[INDICATORS] Calculated OBV")
	}

	// VWAP (Volume Weighted Average Price)
	if config.VWAP != nil && config.VWAP.Enabled {
		vwap := CalculateVWAP(candles)
		for i, val := range vwap {
			if isValidValue(val) {
				candles[i].Indicators.VWAP = &val
			}
		}
		log.Printf("[INDICATORS] Calculated VWAP")
	}

	// MFI (Money Flow Index)
	if config.MFI != nil && config.MFI.Enabled {
		mfi := CalculateMFI(candles, config.MFI.Period)
		for i, val := range mfi {
			if isValidValue(val) {
				candles[i].Indicators.MFI = &val
			}
		}
		log.Printf("[INDICATORS] Calculated MFI(%d)", config.MFI.Period)
	}

	// CMF (Chaikin Money Flow)
	if config.CMF != nil && config.CMF.Enabled {
		cmf := CalculateCMF(candles, config.CMF.Period)
		for i, val := range cmf {
			if isValidValue(val) {
				candles[i].Indicators.CMF = &val
			}
		}
		log.Printf("[INDICATORS] Calculated CMF(%d)", config.CMF.Period)
	}

	// Volume SMA
	if config.VolumeSMA != nil && config.VolumeSMA.Enabled {
		volumeSMA := CalculateVolumeSMA(candles, config.VolumeSMA.Period)
		for i, val := range volumeSMA {
			if isValidValue(val) {
				candles[i].Indicators.VolumeSMA = &val
			}
		}
		log.Printf("[INDICATORS] Calculated Volume SMA(%d)", config.VolumeSMA.Period)
	}
}

// ValidateConfig validates the indicator configuration
func (s *Service) ValidateConfig(config *models.IndicatorConfig) error {
	if config == nil {
		return nil
	}

	// Validate RSI
	if config.RSI != nil && config.RSI.Enabled {
		for _, period := range config.RSI.Periods {
			if period < 2 || period > 100 {
				return fmt.Errorf("RSI period %d out of range (2-100)", period)
			}
		}
	}

	// Validate MACD
	if config.MACD != nil && config.MACD.Enabled {
		if config.MACD.FastPeriod >= config.MACD.SlowPeriod {
			return fmt.Errorf("MACD fast period (%d) must be less than slow period (%d)",
				config.MACD.FastPeriod, config.MACD.SlowPeriod)
		}
		if config.MACD.SignalPeriod < 2 {
			return fmt.Errorf("MACD signal period must be at least 2")
		}
	}

	// Validate Bollinger Bands
	if config.Bollinger != nil && config.Bollinger.Enabled {
		if config.Bollinger.Period < 2 {
			return fmt.Errorf("Bollinger period must be at least 2")
		}
		if config.Bollinger.StdDev <= 0 {
			return fmt.Errorf("Bollinger stdDev must be positive")
		}
	}

	// Validate Stochastic
	if config.Stochastic != nil && config.Stochastic.Enabled {
		if config.Stochastic.KPeriod < 1 {
			return fmt.Errorf("Stochastic K period must be at least 1")
		}
		if config.Stochastic.DPeriod < 1 {
			return fmt.Errorf("Stochastic D period must be at least 1")
		}
	}

	return nil
}
