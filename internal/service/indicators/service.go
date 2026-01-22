package indicators

import (
	"log"

	"github.com/yourusername/datacollector/internal/models"
)

// Service handles indicator calculations for candles
type Service struct {
	config *models.IndicatorConfig
}

// NewService creates a new indicator service with default config
func NewService() *Service {
	return &Service{
		config: models.DefaultIndicatorConfig(),
	}
}

// NewServiceWithConfig creates a new indicator service with a specific config
func NewServiceWithConfig(config *models.IndicatorConfig) *Service {
	if config == nil {
		config = models.DefaultIndicatorConfig()
	}
	return &Service{
		config: config,
	}
}

// SetConfig updates the indicator configuration
func (s *Service) SetConfig(config *models.IndicatorConfig) {
	if config != nil {
		s.config = config
	}
}

// GetConfig returns the current indicator configuration
func (s *Service) GetConfig() *models.IndicatorConfig {
	return s.config
}

// CalculateAll calculates ALL indicators and updates the candles
// Note: Candles should be ordered with newest first (index 0)
// The function handles reversal automatically
func (s *Service) CalculateAll(candles []models.Candle) ([]models.Candle, error) {
	return s.CalculateWithConfig(candles, s.config)
}

// CalculateWithConfig calculates indicators using a specific configuration
func (s *Service) CalculateWithConfig(candles []models.Candle, config *models.IndicatorConfig) ([]models.Candle, error) {
	if len(candles) == 0 {
		return candles, nil
	}

	if config == nil {
		config = models.DefaultIndicatorConfig()
	}

	// Reverse candles for calculation (indicators expect oldest first)
	reversed := make([]models.Candle, len(candles))
	for i := range candles {
		reversed[i] = candles[len(candles)-1-i]
	}

	log.Printf("[INDICATORS] Calculating indicators for %d candles using config '%s'", len(reversed), config.Name)

	// Initialize indicators map for each candle
	for i := range reversed {
		if reversed[i].Indicators == (models.Indicators{}) {
			reversed[i].Indicators = models.Indicators{}
		}
	}

	// Calculate indicators based on config
	if config.EnableTrend {
		s.calculateTrendIndicatorsWithConfig(reversed, config.Trend)
	}
	if config.EnableMomentum {
		s.calculateMomentumIndicatorsWithConfig(reversed, config.Momentum)
	}
	if config.EnableVolatility {
		s.calculateVolatilityIndicatorsWithConfig(reversed, config.Volatility)
	}
	if config.EnableVolume {
		s.calculateVolumeIndicatorsWithConfig(reversed, config.Volume)
	}

	// Reverse back to newest-first order
	for i := range candles {
		candles[i] = reversed[len(reversed)-1-i]
	}

	log.Printf("[INDICATORS] Indicator calculations complete")
	return candles, nil
}

// calculateTrendIndicators calculates all trend-following indicators with default config
func (s *Service) calculateTrendIndicators(candles []models.Candle) {
	s.calculateTrendIndicatorsWithConfig(candles, models.DefaultTrendConfig())
}

// calculateTrendIndicatorsWithConfig calculates trend indicators using config
func (s *Service) calculateTrendIndicatorsWithConfig(candles []models.Candle, cfg models.TrendConfig) {
	// SMA (Simple Moving Average)
	if cfg.SMAEnabled && len(cfg.SMAPeriods) > 0 {
		for _, period := range cfg.SMAPeriods {
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
		log.Printf("[INDICATORS] Calculated SMA(%v)", cfg.SMAPeriods)
	}

	// EMA (Exponential Moving Average)
	if cfg.EMAEnabled && len(cfg.EMAPeriods) > 0 {
		for _, period := range cfg.EMAPeriods {
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
		log.Printf("[INDICATORS] Calculated EMA(%v)", cfg.EMAPeriods)
	}

	// DEMA (Double Exponential Moving Average)
	if cfg.DEMAEnabled && cfg.DEMAPeriod > 0 {
		dema := CalculateDEMA(candles, cfg.DEMAPeriod, "close")
		for i, val := range dema {
			if isValidValue(val) {
				candles[i].Indicators.DEMA = &val
			}
		}
		log.Printf("[INDICATORS] Calculated DEMA(%d)", cfg.DEMAPeriod)
	}

	// TEMA (Triple Exponential Moving Average)
	if cfg.TEMAEnabled && cfg.TEMAPeriod > 0 {
		tema := CalculateTEMA(candles, cfg.TEMAPeriod, "close")
		for i, val := range tema {
			if isValidValue(val) {
				candles[i].Indicators.TEMA = &val
			}
		}
		log.Printf("[INDICATORS] Calculated TEMA(%d)", cfg.TEMAPeriod)
	}

	// WMA (Weighted Moving Average)
	if cfg.WMAEnabled && cfg.WMAPeriod > 0 {
		wma := CalculateWMA(candles, cfg.WMAPeriod, "close")
		for i, val := range wma {
			if isValidValue(val) {
				candles[i].Indicators.WMA = &val
			}
		}
		log.Printf("[INDICATORS] Calculated WMA(%d)", cfg.WMAPeriod)
	}

	// HMA (Hull Moving Average)
	if cfg.HMAEnabled && cfg.HMAPeriod > 0 {
		hma := CalculateHMA(candles, cfg.HMAPeriod, "close")
		for i, val := range hma {
			if isValidValue(val) {
				candles[i].Indicators.HMA = &val
			}
		}
		log.Printf("[INDICATORS] Calculated HMA(%d)", cfg.HMAPeriod)
	}

	// VWMA (Volume Weighted Moving Average)
	if cfg.VWMAEnabled && cfg.VWMAPeriod > 0 {
		vwma := CalculateVWMA(candles, cfg.VWMAPeriod, "close")
		for i, val := range vwma {
			if isValidValue(val) {
				candles[i].Indicators.VWMA = &val
			}
		}
		log.Printf("[INDICATORS] Calculated VWMA(%d)", cfg.VWMAPeriod)
	}

	// Ichimoku Cloud
	if cfg.IchimokuEnabled {
		ichimoku := CalculateIchimoku(candles, cfg.IchimokuTenkan, cfg.IchimokuKijun, cfg.IchimokuSenkouB, cfg.IchimokuDisplacement, cfg.IchimokuDisplacement)
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
		log.Printf("[INDICATORS] Calculated Ichimoku(%d, %d, %d)", cfg.IchimokuTenkan, cfg.IchimokuKijun, cfg.IchimokuSenkouB)
	}

	// ADX/DMI
	if cfg.ADXEnabled && cfg.ADXPeriod > 0 {
		adx := CalculateADX(candles, cfg.ADXPeriod)
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
		log.Printf("[INDICATORS] Calculated ADX(%d)", cfg.ADXPeriod)
	}

	// SuperTrend
	if cfg.SuperTrendEnabled && cfg.SuperTrendPeriod > 0 {
		supertrend := CalculateSuperTrend(candles, cfg.SuperTrendPeriod, cfg.SuperTrendMultiplier)
		for i := range candles {
			if isValidValue(supertrend.SuperTrend[i]) {
				candles[i].Indicators.SuperTrend = &supertrend.SuperTrend[i]
			}
			if supertrend.Signal[i] != 0 {
				candles[i].Indicators.SuperTrendSignal = &supertrend.Signal[i]
			}
		}
		log.Printf("[INDICATORS] Calculated SuperTrend(%d, %.1f)", cfg.SuperTrendPeriod, cfg.SuperTrendMultiplier)
	}
}

// calculateMomentumIndicators calculates all momentum/oscillator indicators with default config
func (s *Service) calculateMomentumIndicators(candles []models.Candle) {
	s.calculateMomentumIndicatorsWithConfig(candles, models.DefaultMomentumConfig())
}

// calculateMomentumIndicatorsWithConfig calculates momentum indicators using config
func (s *Service) calculateMomentumIndicatorsWithConfig(candles []models.Candle, cfg models.MomentumConfig) {
	// RSI (Relative Strength Index)
	if cfg.RSIEnabled && len(cfg.RSIPeriods) > 0 {
		for _, period := range cfg.RSIPeriods {
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
		log.Printf("[INDICATORS] Calculated RSI(%v)", cfg.RSIPeriods)
	}

	// Stochastic Oscillator
	if cfg.StochEnabled {
		stoch := CalculateStochastic(candles, cfg.StochK, cfg.StochD, cfg.StochSmooth)
		for i := range candles {
			if isValidValue(stoch.K[i]) {
				candles[i].Indicators.StochK = &stoch.K[i]
			}
			if isValidValue(stoch.D[i]) {
				candles[i].Indicators.StochD = &stoch.D[i]
			}
		}
		log.Printf("[INDICATORS] Calculated Stochastic(%d, %d, %d)", cfg.StochK, cfg.StochD, cfg.StochSmooth)
	}

	// MACD
	if cfg.MACDEnabled {
		macd := CalculateMACD(candles, cfg.MACDFast, cfg.MACDSlow, cfg.MACDSignal, "close")
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
		log.Printf("[INDICATORS] Calculated MACD(%d, %d, %d)", cfg.MACDFast, cfg.MACDSlow, cfg.MACDSignal)
	}

	// ROC (Rate of Change)
	if cfg.ROCEnabled && cfg.ROCPeriod > 0 {
		roc := CalculateROC(candles, cfg.ROCPeriod, "close")
		for i, val := range roc {
			if isValidValue(val) {
				candles[i].Indicators.ROC = &val
			}
		}
		log.Printf("[INDICATORS] Calculated ROC(%d)", cfg.ROCPeriod)
	}

	// CCI (Commodity Channel Index)
	if cfg.CCIEnabled && cfg.CCIPeriod > 0 {
		cci := CalculateCCI(candles, cfg.CCIPeriod)
		for i, val := range cci {
			if isValidValue(val) {
				candles[i].Indicators.CCI = &val
			}
		}
		log.Printf("[INDICATORS] Calculated CCI(%d)", cfg.CCIPeriod)
	}

	// Williams %R
	if cfg.WilliamsREnabled && cfg.WilliamsRPeriod > 0 {
		williamsR := CalculateWilliamsR(candles, cfg.WilliamsRPeriod)
		for i, val := range williamsR {
			if isValidValue(val) {
				candles[i].Indicators.WilliamsR = &val
			}
		}
		log.Printf("[INDICATORS] Calculated Williams %%R(%d)", cfg.WilliamsRPeriod)
	}

	// Momentum
	if cfg.MomentumEnabled && cfg.MomentumPeriod > 0 {
		momentum := CalculateMomentum(candles, cfg.MomentumPeriod, "close")
		for i, val := range momentum {
			if isValidValue(val) {
				candles[i].Indicators.Momentum = &val
			}
		}
		log.Printf("[INDICATORS] Calculated Momentum(%d)", cfg.MomentumPeriod)
	}
}

// calculateVolatilityIndicators calculates all volatility indicators with default config
func (s *Service) calculateVolatilityIndicators(candles []models.Candle) {
	s.calculateVolatilityIndicatorsWithConfig(candles, models.DefaultVolatilityConfig())
}

// calculateVolatilityIndicatorsWithConfig calculates volatility indicators using config
func (s *Service) calculateVolatilityIndicatorsWithConfig(candles []models.Candle, cfg models.VolatilityConfig) {
	// Bollinger Bands
	if cfg.BollingerEnabled && cfg.BollingerPeriod > 0 {
		bb := CalculateBollingerBands(candles, cfg.BollingerPeriod, cfg.BollingerStdDev, "close")
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
		log.Printf("[INDICATORS] Calculated Bollinger Bands(%d, %.1f)", cfg.BollingerPeriod, cfg.BollingerStdDev)
	}

	// ATR (Average True Range)
	if cfg.ATREnabled && cfg.ATRPeriod > 0 {
		atr := CalculateATR(candles, cfg.ATRPeriod)
		for i, val := range atr {
			if isValidValue(val) {
				candles[i].Indicators.ATR = &val
			}
		}
		log.Printf("[INDICATORS] Calculated ATR(%d)", cfg.ATRPeriod)
	}

	// Keltner Channels
	if cfg.KeltnerEnabled && cfg.KeltnerPeriod > 0 {
		keltner := CalculateKeltner(candles, cfg.KeltnerPeriod, cfg.KeltnerATRPeriod, cfg.KeltnerMultiplier)
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
		log.Printf("[INDICATORS] Calculated Keltner Channels(%d, %d, %.1f)", cfg.KeltnerPeriod, cfg.KeltnerATRPeriod, cfg.KeltnerMultiplier)
	}

	// Donchian Channels
	if cfg.DonchianEnabled && cfg.DonchianPeriod > 0 {
		donchian := CalculateDonchian(candles, cfg.DonchianPeriod)
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
		log.Printf("[INDICATORS] Calculated Donchian Channels(%d)", cfg.DonchianPeriod)
	}

	// Standard Deviation
	if cfg.StdDevEnabled && cfg.StdDevPeriod > 0 {
		stddev := CalculateStdDev(candles, cfg.StdDevPeriod, "close")
		for i, val := range stddev {
			if isValidValue(val) {
				candles[i].Indicators.StdDev = &val
			}
		}
		log.Printf("[INDICATORS] Calculated StdDev(%d)", cfg.StdDevPeriod)
	}
}

// calculateVolumeIndicators calculates all volume-based indicators with default config
func (s *Service) calculateVolumeIndicators(candles []models.Candle) {
	s.calculateVolumeIndicatorsWithConfig(candles, models.DefaultVolumeConfig())
}

// calculateVolumeIndicatorsWithConfig calculates volume indicators using config
func (s *Service) calculateVolumeIndicatorsWithConfig(candles []models.Candle, cfg models.VolumeConfig) {
	// OBV (On-Balance Volume)
	if cfg.OBVEnabled {
		obv := CalculateOBV(candles)
		for i, val := range obv {
			if isValidValue(val) {
				candles[i].Indicators.OBV = &val
			}
		}
		log.Printf("[INDICATORS] Calculated OBV")
	}

	// VWAP (Volume Weighted Average Price)
	if cfg.VWAPEnabled {
		vwap := CalculateVWAP(candles)
		for i, val := range vwap {
			if isValidValue(val) {
				candles[i].Indicators.VWAP = &val
			}
		}
		log.Printf("[INDICATORS] Calculated VWAP")
	}

	// MFI (Money Flow Index)
	if cfg.MFIEnabled && cfg.MFIPeriod > 0 {
		mfi := CalculateMFI(candles, cfg.MFIPeriod)
		for i, val := range mfi {
			if isValidValue(val) {
				candles[i].Indicators.MFI = &val
			}
		}
		log.Printf("[INDICATORS] Calculated MFI(%d)", cfg.MFIPeriod)
	}

	// CMF (Chaikin Money Flow)
	if cfg.CMFEnabled && cfg.CMFPeriod > 0 {
		cmf := CalculateCMF(candles, cfg.CMFPeriod)
		for i, val := range cmf {
			if isValidValue(val) {
				candles[i].Indicators.CMF = &val
			}
		}
		log.Printf("[INDICATORS] Calculated CMF(%d)", cfg.CMFPeriod)
	}

	// Volume SMA
	if cfg.VolumeSMAEnabled && cfg.VolumeSMAPeriod > 0 {
		volumeSMA := CalculateVolumeSMA(candles, cfg.VolumeSMAPeriod)
		for i, val := range volumeSMA {
			if isValidValue(val) {
				candles[i].Indicators.VolumeSMA = &val
			}
		}
		log.Printf("[INDICATORS] Calculated Volume SMA(%d)", cfg.VolumeSMAPeriod)
	}
}
