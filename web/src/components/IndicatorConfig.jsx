import { useState, useEffect } from 'react'

/**
 * IndicatorConfig Component
 *
 * A comprehensive configuration UI for all 29 technical indicators.
 * Supports both connector-level and job-level configuration.
 *
 * Props:
 * - config: Current indicator configuration object
 * - onChange: Callback when configuration changes
 * - onSave: Callback to save configuration
 * - onReset: Optional callback to reset to defaults
 * - isJobLevel: Boolean indicating if this is job-level config (shows inheritance)
 * - connectorConfig: Optional connector config to show inherited values
 */
function IndicatorConfig({ config, onChange, onSave, onReset, isJobLevel = false, connectorConfig = null }) {
  const [localConfig, setLocalConfig] = useState(config || {})
  const [hasChanges, setHasChanges] = useState(false)

  useEffect(() => {
    setLocalConfig(config || {})
  }, [config])

  const handleToggle = (category, indicator) => {
    const newConfig = { ...localConfig }
    if (!newConfig[indicator]) {
      newConfig[indicator] = { enabled: true }
    } else {
      newConfig[indicator] = { ...newConfig[indicator], enabled: !newConfig[indicator].enabled }
    }
    setLocalConfig(newConfig)
    setHasChanges(true)
    onChange?.(newConfig)
  }

  const handleFieldChange = (indicator, field, value) => {
    const newConfig = { ...localConfig }
    if (!newConfig[indicator]) {
      newConfig[indicator] = {}
    }
    newConfig[indicator] = { ...newConfig[indicator], [field]: value }
    setLocalConfig(newConfig)
    setHasChanges(true)
    onChange?.(newConfig)
  }

  const handleArrayChange = (indicator, field, value) => {
    const arrayValue = value.split(',').map(v => parseInt(v.trim())).filter(v => !isNaN(v))
    handleFieldChange(indicator, field, arrayValue)
  }

  const handleSave = () => {
    onSave?.(localConfig)
    setHasChanges(false)
  }

  const handleReset = () => {
    onReset?.()
    setHasChanges(false)
  }

  const isEnabled = (indicator) => {
    return localConfig[indicator]?.enabled || false
  }

  const getValue = (indicator, field, defaultValue = '') => {
    const value = localConfig[indicator]?.[field]
    if (value === undefined || value === null) return defaultValue
    if (Array.isArray(value)) return value.join(', ')
    return value
  }

  const isInherited = (indicator) => {
    if (!isJobLevel || !connectorConfig) return false
    return !localConfig[indicator] && connectorConfig[indicator]?.enabled
  }

  const IndicatorRow = ({ indicator, label, children, category }) => (
    <div className={`p-3 rounded-lg ${isEnabled(indicator) ? 'bg-blue-50' : 'bg-gray-50'} ${isInherited(indicator) ? 'border-2 border-dashed border-blue-300' : ''}`}>
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center space-x-2">
          <button
            onClick={() => handleToggle(category, indicator)}
            className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
              isEnabled(indicator) ? 'bg-blue-500' : 'bg-gray-300'
            }`}
          >
            <span
              className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                isEnabled(indicator) ? 'translate-x-6' : 'translate-x-1'
              }`}
            />
          </button>
          <span className="text-sm font-medium text-gray-900">{label}</span>
          {isInherited(indicator) && (
            <span className="text-xs bg-blue-200 text-blue-800 px-2 py-0.5 rounded">Inherited</span>
          )}
        </div>
      </div>
      {isEnabled(indicator) && children && (
        <div className="ml-12 space-y-2">
          {children}
        </div>
      )}
    </div>
  )

  const NumberInput = ({ indicator, field, label, min = 1, max = 1000 }) => (
    <div className="flex items-center space-x-2">
      <label className="text-xs text-gray-600 w-24">{label}:</label>
      <input
        type="number"
        value={getValue(indicator, field, min)}
        onChange={(e) => handleFieldChange(indicator, field, parseInt(e.target.value))}
        min={min}
        max={max}
        className="w-20 px-2 py-1 text-sm border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
      />
    </div>
  )

  const PeriodsInput = ({ indicator }) => (
    <div className="flex items-center space-x-2">
      <label className="text-xs text-gray-600 w-24">Periods:</label>
      <input
        type="text"
        value={getValue(indicator, 'periods', '')}
        onChange={(e) => handleArrayChange(indicator, 'periods', e.target.value)}
        placeholder="e.g., 20, 50, 200"
        className="flex-1 px-2 py-1 text-sm border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
      />
    </div>
  )

  return (
    <div className="space-y-6">
      {/* Trend Indicators */}
      <div>
        <h3 className="text-lg font-bold text-gray-900 mb-3 flex items-center">
          <span className="mr-2">📈</span> Trend Indicators
        </h3>
        <div className="space-y-2">
          <IndicatorRow indicator="sma" label="SMA (Simple Moving Average)" category="trend">
            <PeriodsInput indicator="sma" />
          </IndicatorRow>

          <IndicatorRow indicator="ema" label="EMA (Exponential Moving Average)" category="trend">
            <PeriodsInput indicator="ema" />
          </IndicatorRow>

          <IndicatorRow indicator="dema" label="DEMA (Double Exponential MA)" category="trend">
            <NumberInput indicator="dema" field="period" label="Period" />
          </IndicatorRow>

          <IndicatorRow indicator="tema" label="TEMA (Triple Exponential MA)" category="trend">
            <NumberInput indicator="tema" field="period" label="Period" />
          </IndicatorRow>

          <IndicatorRow indicator="wma" label="WMA (Weighted Moving Average)" category="trend">
            <NumberInput indicator="wma" field="period" label="Period" />
          </IndicatorRow>

          <IndicatorRow indicator="hma" label="HMA (Hull Moving Average)" category="trend">
            <NumberInput indicator="hma" field="period" label="Period" min={2} />
          </IndicatorRow>

          <IndicatorRow indicator="vwma" label="VWMA (Volume Weighted MA)" category="trend">
            <NumberInput indicator="vwma" field="period" label="Period" />
          </IndicatorRow>

          <IndicatorRow indicator="ichimoku" label="Ichimoku Cloud" category="trend">
            <NumberInput indicator="ichimoku" field="tenkan_period" label="Tenkan" />
            <NumberInput indicator="ichimoku" field="kijun_period" label="Kijun" />
            <NumberInput indicator="ichimoku" field="senkou_b_period" label="Senkou B" />
          </IndicatorRow>

          <IndicatorRow indicator="adx" label="ADX (Average Directional Index)" category="trend">
            <NumberInput indicator="adx" field="period" label="Period" />
          </IndicatorRow>

          <IndicatorRow indicator="supertrend" label="SuperTrend" category="trend">
            <NumberInput indicator="supertrend" field="period" label="Period" />
            <div className="flex items-center space-x-2">
              <label className="text-xs text-gray-600 w-24">Multiplier:</label>
              <input
                type="number"
                value={getValue('supertrend', 'multiplier', 3.0)}
                onChange={(e) => handleFieldChange('supertrend', 'multiplier', parseFloat(e.target.value))}
                step="0.1"
                min="1"
                max="10"
                className="w-20 px-2 py-1 text-sm border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
          </IndicatorRow>
        </div>
      </div>

      {/* Momentum Indicators */}
      <div>
        <h3 className="text-lg font-bold text-gray-900 mb-3 flex items-center">
          <span className="mr-2">⚡</span> Momentum Indicators
        </h3>
        <div className="space-y-2">
          <IndicatorRow indicator="rsi" label="RSI (Relative Strength Index)" category="momentum">
            <PeriodsInput indicator="rsi" />
          </IndicatorRow>

          <IndicatorRow indicator="stochastic" label="Stochastic Oscillator" category="momentum">
            <NumberInput indicator="stochastic" field="k_period" label="K Period" />
            <NumberInput indicator="stochastic" field="d_period" label="D Period" />
            <NumberInput indicator="stochastic" field="smooth" label="Smooth" />
          </IndicatorRow>

          <IndicatorRow indicator="macd" label="MACD" category="momentum">
            <NumberInput indicator="macd" field="fast_period" label="Fast Period" />
            <NumberInput indicator="macd" field="slow_period" label="Slow Period" />
            <NumberInput indicator="macd" field="signal_period" label="Signal Period" />
          </IndicatorRow>

          <IndicatorRow indicator="roc" label="ROC (Rate of Change)" category="momentum">
            <NumberInput indicator="roc" field="period" label="Period" />
          </IndicatorRow>

          <IndicatorRow indicator="cci" label="CCI (Commodity Channel Index)" category="momentum">
            <NumberInput indicator="cci" field="period" label="Period" />
          </IndicatorRow>

          <IndicatorRow indicator="williams_r" label="Williams %R" category="momentum">
            <NumberInput indicator="williams_r" field="period" label="Period" />
          </IndicatorRow>

          <IndicatorRow indicator="momentum" label="Momentum" category="momentum">
            <NumberInput indicator="momentum" field="period" label="Period" />
          </IndicatorRow>
        </div>
      </div>

      {/* Volatility Indicators */}
      <div>
        <h3 className="text-lg font-bold text-gray-900 mb-3 flex items-center">
          <span className="mr-2">📊</span> Volatility Indicators
        </h3>
        <div className="space-y-2">
          <IndicatorRow indicator="bollinger" label="Bollinger Bands" category="volatility">
            <NumberInput indicator="bollinger" field="period" label="Period" />
            <div className="flex items-center space-x-2">
              <label className="text-xs text-gray-600 w-24">Std Dev:</label>
              <input
                type="number"
                value={getValue('bollinger', 'std_dev', 2.0)}
                onChange={(e) => handleFieldChange('bollinger', 'std_dev', parseFloat(e.target.value))}
                step="0.1"
                min="1"
                max="5"
                className="w-20 px-2 py-1 text-sm border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
          </IndicatorRow>

          <IndicatorRow indicator="atr" label="ATR (Average True Range)" category="volatility">
            <NumberInput indicator="atr" field="period" label="Period" />
          </IndicatorRow>

          <IndicatorRow indicator="keltner" label="Keltner Channels" category="volatility">
            <NumberInput indicator="keltner" field="period" label="EMA Period" />
            <NumberInput indicator="keltner" field="atr_period" label="ATR Period" />
            <div className="flex items-center space-x-2">
              <label className="text-xs text-gray-600 w-24">Multiplier:</label>
              <input
                type="number"
                value={getValue('keltner', 'multiplier', 2.0)}
                onChange={(e) => handleFieldChange('keltner', 'multiplier', parseFloat(e.target.value))}
                step="0.1"
                min="1"
                max="5"
                className="w-20 px-2 py-1 text-sm border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
          </IndicatorRow>

          <IndicatorRow indicator="donchian" label="Donchian Channels" category="volatility">
            <NumberInput indicator="donchian" field="period" label="Period" />
          </IndicatorRow>

          <IndicatorRow indicator="stddev" label="Standard Deviation" category="volatility">
            <NumberInput indicator="stddev" field="period" label="Period" />
          </IndicatorRow>
        </div>
      </div>

      {/* Volume Indicators */}
      <div>
        <h3 className="text-lg font-bold text-gray-900 mb-3 flex items-center">
          <span className="mr-2">📦</span> Volume Indicators
        </h3>
        <div className="space-y-2">
          <IndicatorRow indicator="obv" label="OBV (On-Balance Volume)" category="volume" />

          <IndicatorRow indicator="vwap" label="VWAP (Volume Weighted Average Price)" category="volume" />

          <IndicatorRow indicator="mfi" label="MFI (Money Flow Index)" category="volume">
            <NumberInput indicator="mfi" field="period" label="Period" />
          </IndicatorRow>

          <IndicatorRow indicator="cmf" label="CMF (Chaikin Money Flow)" category="volume">
            <NumberInput indicator="cmf" field="period" label="Period" />
          </IndicatorRow>

          <IndicatorRow indicator="volume_sma" label="Volume SMA" category="volume">
            <NumberInput indicator="volume_sma" field="period" label="Period" />
          </IndicatorRow>
        </div>
      </div>

      {/* Action Buttons */}
      <div className="flex space-x-3 pt-4 border-t">
        {onReset && (
          <button
            onClick={handleReset}
            className="px-4 py-2 border border-gray-300 text-gray-700 rounded hover:bg-gray-50 transition"
          >
            Reset to Defaults
          </button>
        )}
        <button
          onClick={handleSave}
          className={`flex-1 px-4 py-2 rounded transition ${
            hasChanges
              ? 'bg-blue-500 text-white hover:bg-blue-600'
              : 'bg-gray-300 text-gray-500 cursor-not-allowed'
          }`}
          disabled={!hasChanges}
        >
          {hasChanges ? 'Save Configuration' : 'No Changes'}
        </button>
      </div>
    </div>
  )
}

export default IndicatorConfig
