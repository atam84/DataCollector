package indicators

import "github.com/yourusername/datacollector/internal/models"

// MergeConfigs merges job-level and connector-level configurations
// Priority: Job config > Connector config > System defaults
//
// Rules:
// 1. If job config exists for an indicator, use it
// 2. If job config doesn't exist but connector config does, use connector config
// 3. If neither exists, use system defaults ONLY if no configs are provided at all
// 4. If an indicator is explicitly disabled at job level, it stays disabled even if connector enables it
func MergeConfigs(connectorConfig, jobConfig *models.IndicatorConfig) *models.IndicatorConfig {
	// If both configs are nil/empty, use defaults
	if (connectorConfig == nil || isEmptyConfig(connectorConfig)) &&
		(jobConfig == nil || isEmptyConfig(jobConfig)) {
		return DefaultConfig()
	}

	// Start with an empty config (no indicators enabled by default)
	merged := &models.IndicatorConfig{}

	// Apply connector config first (base level)
	if connectorConfig != nil {
		applyConnectorConfig(merged, connectorConfig)
	}

	// Apply job config (overrides connector)
	if jobConfig != nil {
		applyJobConfig(merged, jobConfig)
	}

	return merged
}

// isEmptyConfig checks if a config has any indicators configured
func isEmptyConfig(config *models.IndicatorConfig) bool {
	if config == nil {
		return true
	}
	// Check if any indicator field is set
	return config.SMA == nil && config.EMA == nil && config.DEMA == nil &&
		config.TEMA == nil && config.WMA == nil && config.HMA == nil &&
		config.VWMA == nil && config.Ichimoku == nil && config.ADX == nil &&
		config.SuperTrend == nil && config.RSI == nil && config.Stochastic == nil &&
		config.MACD == nil && config.ROC == nil && config.CCI == nil &&
		config.WilliamsR == nil && config.Momentum == nil && config.Bollinger == nil &&
		config.ATR == nil && config.Keltner == nil && config.Donchian == nil &&
		config.StdDev == nil && config.OBV == nil && config.VWAP == nil &&
		config.MFI == nil && config.CMF == nil && config.VolumeSMA == nil
}

// applyConnectorConfig applies connector-level configuration to the merged config
func applyConnectorConfig(merged, connector *models.IndicatorConfig) {
	// Trend Indicators
	if connector.SMA != nil {
		merged.SMA = connector.SMA
	}
	if connector.EMA != nil {
		merged.EMA = connector.EMA
	}
	if connector.DEMA != nil {
		merged.DEMA = connector.DEMA
	}
	if connector.TEMA != nil {
		merged.TEMA = connector.TEMA
	}
	if connector.WMA != nil {
		merged.WMA = connector.WMA
	}
	if connector.HMA != nil {
		merged.HMA = connector.HMA
	}
	if connector.VWMA != nil {
		merged.VWMA = connector.VWMA
	}
	if connector.Ichimoku != nil {
		merged.Ichimoku = connector.Ichimoku
	}
	if connector.ADX != nil {
		merged.ADX = connector.ADX
	}
	if connector.SuperTrend != nil {
		merged.SuperTrend = connector.SuperTrend
	}

	// Momentum Indicators
	if connector.RSI != nil {
		merged.RSI = connector.RSI
	}
	if connector.Stochastic != nil {
		merged.Stochastic = connector.Stochastic
	}
	if connector.MACD != nil {
		merged.MACD = connector.MACD
	}
	if connector.ROC != nil {
		merged.ROC = connector.ROC
	}
	if connector.CCI != nil {
		merged.CCI = connector.CCI
	}
	if connector.WilliamsR != nil {
		merged.WilliamsR = connector.WilliamsR
	}
	if connector.Momentum != nil {
		merged.Momentum = connector.Momentum
	}

	// Volatility Indicators
	if connector.Bollinger != nil {
		merged.Bollinger = connector.Bollinger
	}
	if connector.ATR != nil {
		merged.ATR = connector.ATR
	}
	if connector.Keltner != nil {
		merged.Keltner = connector.Keltner
	}
	if connector.Donchian != nil {
		merged.Donchian = connector.Donchian
	}
	if connector.StdDev != nil {
		merged.StdDev = connector.StdDev
	}

	// Volume Indicators
	if connector.OBV != nil {
		merged.OBV = connector.OBV
	}
	if connector.VWAP != nil {
		merged.VWAP = connector.VWAP
	}
	if connector.MFI != nil {
		merged.MFI = connector.MFI
	}
	if connector.CMF != nil {
		merged.CMF = connector.CMF
	}
	if connector.VolumeSMA != nil {
		merged.VolumeSMA = connector.VolumeSMA
	}
}

// applyJobConfig applies job-level configuration to the merged config (highest priority)
func applyJobConfig(merged, job *models.IndicatorConfig) {
	// Trend Indicators
	if job.SMA != nil {
		merged.SMA = job.SMA
	}
	if job.EMA != nil {
		merged.EMA = job.EMA
	}
	if job.DEMA != nil {
		merged.DEMA = job.DEMA
	}
	if job.TEMA != nil {
		merged.TEMA = job.TEMA
	}
	if job.WMA != nil {
		merged.WMA = job.WMA
	}
	if job.HMA != nil {
		merged.HMA = job.HMA
	}
	if job.VWMA != nil {
		merged.VWMA = job.VWMA
	}
	if job.Ichimoku != nil {
		merged.Ichimoku = job.Ichimoku
	}
	if job.ADX != nil {
		merged.ADX = job.ADX
	}
	if job.SuperTrend != nil {
		merged.SuperTrend = job.SuperTrend
	}

	// Momentum Indicators
	if job.RSI != nil {
		merged.RSI = job.RSI
	}
	if job.Stochastic != nil {
		merged.Stochastic = job.Stochastic
	}
	if job.MACD != nil {
		merged.MACD = job.MACD
	}
	if job.ROC != nil {
		merged.ROC = job.ROC
	}
	if job.CCI != nil {
		merged.CCI = job.CCI
	}
	if job.WilliamsR != nil {
		merged.WilliamsR = job.WilliamsR
	}
	if job.Momentum != nil {
		merged.Momentum = job.Momentum
	}

	// Volatility Indicators
	if job.Bollinger != nil {
		merged.Bollinger = job.Bollinger
	}
	if job.ATR != nil {
		merged.ATR = job.ATR
	}
	if job.Keltner != nil {
		merged.Keltner = job.Keltner
	}
	if job.Donchian != nil {
		merged.Donchian = job.Donchian
	}
	if job.StdDev != nil {
		merged.StdDev = job.StdDev
	}

	// Volume Indicators
	if job.OBV != nil {
		merged.OBV = job.OBV
	}
	if job.VWAP != nil {
		merged.VWAP = job.VWAP
	}
	if job.MFI != nil {
		merged.MFI = job.MFI
	}
	if job.CMF != nil {
		merged.CMF = job.CMF
	}
	if job.VolumeSMA != nil {
		merged.VolumeSMA = job.VolumeSMA
	}
}

// GetEffectiveConfig determines the effective configuration for a job
// This is the main entry point for getting the configuration to use during indicator calculation
func GetEffectiveConfig(connectorConfig, jobConfig *models.IndicatorConfig) *models.IndicatorConfig {
	return MergeConfigs(connectorConfig, jobConfig)
}

// CalculateMinimumCandles calculates the minimum number of candles required
// for the given configuration to produce valid indicator values
func CalculateMinimumCandles(config *models.IndicatorConfig) int {
	minCandles := 0

	// Helper to update minimum if needed
	updateMin := func(required int) {
		if required > minCandles {
			minCandles = required
		}
	}

	// Trend Indicators
	if config.SMA != nil && config.SMA.Enabled {
		for _, period := range config.SMA.Periods {
			updateMin(period)
		}
	}
	if config.EMA != nil && config.EMA.Enabled {
		for _, period := range config.EMA.Periods {
			updateMin(period)
		}
	}
	if config.DEMA != nil && config.DEMA.Enabled {
		updateMin(config.DEMA.Period * 2) // DEMA needs more data
	}
	if config.TEMA != nil && config.TEMA.Enabled {
		updateMin(config.TEMA.Period * 3) // TEMA needs even more data
	}
	if config.WMA != nil && config.WMA.Enabled {
		updateMin(config.WMA.Period)
	}
	if config.HMA != nil && config.HMA.Enabled {
		updateMin(config.HMA.Period * 2) // HMA needs more data
	}
	if config.VWMA != nil && config.VWMA.Enabled {
		updateMin(config.VWMA.Period)
	}
	if config.Ichimoku != nil && config.Ichimoku.Enabled {
		// Ichimoku needs the longest period
		updateMin(config.Ichimoku.SenkouBPeriod + config.Ichimoku.DisplacementFwd)
	}
	if config.ADX != nil && config.ADX.Enabled {
		updateMin(config.ADX.Period * 2) // ADX needs smoothing
	}
	if config.SuperTrend != nil && config.SuperTrend.Enabled {
		updateMin(config.SuperTrend.Period)
	}

	// Momentum Indicators
	if config.RSI != nil && config.RSI.Enabled {
		for _, period := range config.RSI.Periods {
			updateMin(period + 1) // RSI needs period + 1
		}
	}
	if config.Stochastic != nil && config.Stochastic.Enabled {
		updateMin(config.Stochastic.KPeriod + config.Stochastic.DPeriod)
	}
	if config.MACD != nil && config.MACD.Enabled {
		updateMin(config.MACD.SlowPeriod + config.MACD.SignalPeriod)
	}
	if config.ROC != nil && config.ROC.Enabled {
		updateMin(config.ROC.Period + 1)
	}
	if config.CCI != nil && config.CCI.Enabled {
		updateMin(config.CCI.Period)
	}
	if config.WilliamsR != nil && config.WilliamsR.Enabled {
		updateMin(config.WilliamsR.Period)
	}
	if config.Momentum != nil && config.Momentum.Enabled {
		updateMin(config.Momentum.Period + 1)
	}

	// Volatility Indicators
	if config.Bollinger != nil && config.Bollinger.Enabled {
		updateMin(config.Bollinger.Period)
	}
	if config.ATR != nil && config.ATR.Enabled {
		updateMin(config.ATR.Period + 1)
	}
	if config.Keltner != nil && config.Keltner.Enabled {
		updateMin(config.Keltner.Period + config.Keltner.ATRPeriod)
	}
	if config.Donchian != nil && config.Donchian.Enabled {
		updateMin(config.Donchian.Period)
	}
	if config.StdDev != nil && config.StdDev.Enabled {
		updateMin(config.StdDev.Period)
	}

	// Volume Indicators
	if config.MFI != nil && config.MFI.Enabled {
		updateMin(config.MFI.Period + 1)
	}
	if config.CMF != nil && config.CMF.Enabled {
		updateMin(config.CMF.Period)
	}
	if config.VolumeSMA != nil && config.VolumeSMA.Enabled {
		updateMin(config.VolumeSMA.Period)
	}

	// If no indicators enabled, return a sensible default
	if minCandles == 0 {
		return 50
	}

	return minCandles
}
