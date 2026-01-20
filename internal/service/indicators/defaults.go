package indicators

import "github.com/yourusername/datacollector/internal/models"

// DefaultConfig returns the default indicator configuration
// This is used when no configuration is specified at connector or job level
func DefaultConfig() *models.IndicatorConfig {
	return &models.IndicatorConfig{
		// Trend Indicators - Enable most common ones by default
		SMA: &models.SMAConfig{
			Enabled: true,
			Periods: []int{20, 50, 200},
		},
		EMA: &models.EMAConfig{
			Enabled: true,
			Periods: []int{12, 26, 50},
		},
		DEMA: &models.DEMAConfig{
			Enabled: false,
			Period:  20,
		},
		TEMA: &models.TEMAConfig{
			Enabled: false,
			Period:  20,
		},
		WMA: &models.WMAConfig{
			Enabled: false,
			Period:  20,
		},
		HMA: &models.HMAConfig{
			Enabled: false,
			Period:  9,
		},
		VWMA: &models.VWMAConfig{
			Enabled: false,
			Period:  20,
		},
		Ichimoku: &models.IchimokuConfig{
			Enabled:         false,
			TenkanPeriod:    9,
			KijunPeriod:     26,
			SenkouBPeriod:   52,
			DisplacementFwd: 26,
			DisplacementBck: 26,
		},
		ADX: &models.ADXConfig{
			Enabled: false,
			Period:  14,
		},
		SuperTrend: &models.SuperTrendConfig{
			Enabled:    false,
			Period:     10,
			Multiplier: 3.0,
		},

		// Momentum Indicators - Enable most popular ones
		RSI: &models.RSIConfig{
			Enabled: true,
			Periods: []int{6, 14, 24},
		},
		Stochastic: &models.StochasticConfig{
			Enabled: false,
			KPeriod: 14,
			DPeriod: 3,
			Smooth:  3,
		},
		MACD: &models.MACDConfig{
			Enabled:      true,
			FastPeriod:   12,
			SlowPeriod:   26,
			SignalPeriod: 9,
		},
		ROC: &models.ROCConfig{
			Enabled: false,
			Period:  12,
		},
		CCI: &models.CCIConfig{
			Enabled: false,
			Period:  20,
		},
		WilliamsR: &models.WilliamsRConfig{
			Enabled: false,
			Period:  14,
		},
		Momentum: &models.MomentumConfig{
			Enabled: false,
			Period:  10,
		},

		// Volatility Indicators
		Bollinger: &models.BollingerConfig{
			Enabled: true,
			Period:  20,
			StdDev:  2.0,
		},
		ATR: &models.ATRConfig{
			Enabled: true,
			Period:  14,
		},
		Keltner: &models.KeltnerConfig{
			Enabled:    false,
			Period:     20,
			ATRPeriod:  10,
			Multiplier: 2.0,
		},
		Donchian: &models.DonchianConfig{
			Enabled: false,
			Period:  20,
		},
		StdDev: &models.StdDevConfig{
			Enabled: false,
			Period:  20,
		},

		// Volume Indicators
		OBV: &models.OBVConfig{
			Enabled: false,
		},
		VWAP: &models.VWAPConfig{
			Enabled: false,
		},
		MFI: &models.MFIConfig{
			Enabled: false,
			Period:  14,
		},
		CMF: &models.CMFConfig{
			Enabled: false,
			Period:  20,
		},
		VolumeSMA: &models.VolumeSMAConfig{
			Enabled: false,
			Period:  20,
		},
	}
}

// MinimalConfig returns a minimal configuration with only RSI, EMA, and MACD enabled
// Useful for low-resource environments or when minimal indicators are needed
func MinimalConfig() *models.IndicatorConfig {
	return &models.IndicatorConfig{
		RSI: &models.RSIConfig{
			Enabled: true,
			Periods: []int{14},
		},
		EMA: &models.EMAConfig{
			Enabled: true,
			Periods: []int{12, 26},
		},
		MACD: &models.MACDConfig{
			Enabled:      true,
			FastPeriod:   12,
			SlowPeriod:   26,
			SignalPeriod: 9,
		},
	}
}

// ComprehensiveConfig returns a configuration with all indicators enabled
// Warning: This will significantly increase computation time and storage
func ComprehensiveConfig() *models.IndicatorConfig {
	config := DefaultConfig()

	// Enable all trend indicators
	if config.DEMA != nil {
		config.DEMA.Enabled = true
	}
	if config.TEMA != nil {
		config.TEMA.Enabled = true
	}
	if config.WMA != nil {
		config.WMA.Enabled = true
	}
	if config.HMA != nil {
		config.HMA.Enabled = true
	}
	if config.VWMA != nil {
		config.VWMA.Enabled = true
	}
	if config.Ichimoku != nil {
		config.Ichimoku.Enabled = true
	}
	if config.ADX != nil {
		config.ADX.Enabled = true
	}
	if config.SuperTrend != nil {
		config.SuperTrend.Enabled = true
	}

	// Enable all momentum indicators
	if config.Stochastic != nil {
		config.Stochastic.Enabled = true
	}
	if config.ROC != nil {
		config.ROC.Enabled = true
	}
	if config.CCI != nil {
		config.CCI.Enabled = true
	}
	if config.WilliamsR != nil {
		config.WilliamsR.Enabled = true
	}
	if config.Momentum != nil {
		config.Momentum.Enabled = true
	}

	// Enable all volatility indicators
	if config.Keltner != nil {
		config.Keltner.Enabled = true
	}
	if config.Donchian != nil {
		config.Donchian.Enabled = true
	}
	if config.StdDev != nil {
		config.StdDev.Enabled = true
	}

	// Enable all volume indicators
	if config.OBV != nil {
		config.OBV.Enabled = true
	}
	if config.VWAP != nil {
		config.VWAP.Enabled = true
	}
	if config.MFI != nil {
		config.MFI.Enabled = true
	}
	if config.CMF != nil {
		config.CMF.Enabled = true
	}
	if config.VolumeSMA != nil {
		config.VolumeSMA.Enabled = true
	}

	return config
}
