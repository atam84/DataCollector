import { useEffect, useRef, useState, useMemo, useCallback } from 'react'
import { createChart, ColorType, CrosshairMode, CandlestickSeries, HistogramSeries, LineSeries } from 'lightweight-charts'
import { ChevronDownIcon, ChevronUpIcon, MagnifyingGlassPlusIcon, MagnifyingGlassMinusIcon, ArrowsPointingOutIcon } from '@heroicons/react/24/outline'

// Indicator definitions with metadata
const INDICATOR_GROUPS = {
  trend: {
    label: 'Trend',
    indicators: {
      sma20: { label: 'SMA 20', color: '#2196F3', field: 'sma20', overlay: true },
      sma50: { label: 'SMA 50', color: '#FF9800', field: 'sma50', overlay: true },
      sma200: { label: 'SMA 200', color: '#9C27B0', field: 'sma200', overlay: true },
      ema12: { label: 'EMA 12', color: '#00BCD4', field: 'ema12', overlay: true },
      ema26: { label: 'EMA 26', color: '#E91E63', field: 'ema26', overlay: true },
      ema50: { label: 'EMA 50', color: '#4CAF50', field: 'ema50', overlay: true },
      supertrend: { label: 'SuperTrend', color: '#FF5722', field: 'supertrend', overlay: true },
    }
  },
  momentum: {
    label: 'Momentum',
    indicators: {
      rsi14: { label: 'RSI 14', color: '#9C27B0', field: 'rsi14', overlay: false, pane: 'rsi', range: [0, 100] },
      macd: { label: 'MACD', color: '#2196F3', field: 'macd', overlay: false, pane: 'macd' },
      macd_signal: { label: 'MACD Signal', color: '#FF9800', field: 'macd_signal', overlay: false, pane: 'macd' },
      stoch_k: { label: 'Stoch %K', color: '#4CAF50', field: 'stoch_k', overlay: false, pane: 'stoch', range: [0, 100] },
      stoch_d: { label: 'Stoch %D', color: '#F44336', field: 'stoch_d', overlay: false, pane: 'stoch', range: [0, 100] },
    }
  },
  volatility: {
    label: 'Volatility',
    indicators: {
      bb_upper: { label: 'BB Upper', color: '#9E9E9E', field: 'bb_upper', overlay: true },
      bb_middle: { label: 'BB Middle', color: '#607D8B', field: 'bb_middle', overlay: true },
      bb_lower: { label: 'BB Lower', color: '#9E9E9E', field: 'bb_lower', overlay: true },
      atr: { label: 'ATR', color: '#FF5722', field: 'atr', overlay: false, pane: 'atr' },
    }
  },
  volume: {
    label: 'Volume',
    indicators: {
      obv: { label: 'OBV', color: '#3F51B5', field: 'obv', overlay: false, pane: 'obv' },
      vwap: { label: 'VWAP', color: '#009688', field: 'vwap', overlay: true },
      mfi: { label: 'MFI', color: '#673AB7', field: 'mfi', overlay: false, pane: 'mfi', range: [0, 100] },
    }
  }
}

// Time period definitions in milliseconds
const TIME_PERIODS = [
  { label: '1D', ms: 24 * 60 * 60 * 1000 },
  { label: '1W', ms: 7 * 24 * 60 * 60 * 1000 },
  { label: '1M', ms: 30 * 24 * 60 * 60 * 1000 },
  { label: '3M', ms: 90 * 24 * 60 * 60 * 1000 },
  { label: '6M', ms: 180 * 24 * 60 * 60 * 1000 },
  { label: '1Y', ms: 365 * 24 * 60 * 60 * 1000 },
  { label: 'All', ms: null },
]

function CandlestickChart({ data, symbol, timeframe, height = 500 }) {
  const chartContainerRef = useRef(null)
  const chartRef = useRef(null)
  const seriesRef = useRef({})
  const [selectedIndicators, setSelectedIndicators] = useState(['sma20', 'sma50'])
  const [showVolume, setShowVolume] = useState(true)
  const [expandedGroups, setExpandedGroups] = useState({ trend: true, momentum: false, volatility: false, volume: false })
  const [selectedPeriod, setSelectedPeriod] = useState('All')

  // Transform data for lightweight-charts
  const chartData = useMemo(() => {
    if (!data || data.length === 0) return { candles: [], volume: [], indicators: {} }

    // Sort by timestamp ascending
    const sorted = [...data].sort((a, b) => a.timestamp - b.timestamp)

    const candles = sorted.map(item => ({
      time: Math.floor(item.timestamp / 1000), // Convert to seconds
      open: item.open,
      high: item.high,
      low: item.low,
      close: item.close,
    }))

    const volume = sorted.map(item => ({
      time: Math.floor(item.timestamp / 1000),
      value: item.volume,
      color: item.close >= item.open ? '#26a69a80' : '#ef535080'
    }))

    // Extract indicators
    const indicators = {}
    Object.values(INDICATOR_GROUPS).forEach(group => {
      Object.entries(group.indicators).forEach(([key, config]) => {
        indicators[key] = sorted
          .filter(item => item.indicators && item.indicators[config.field] != null)
          .map(item => ({
            time: Math.floor(item.timestamp / 1000),
            value: item.indicators[config.field]
          }))
      })
    })

    return { candles, volume, indicators }
  }, [data])

  // Initialize chart
  useEffect(() => {
    if (!chartContainerRef.current) return

    const chart = createChart(chartContainerRef.current, {
      layout: {
        background: { type: ColorType.Solid, color: '#ffffff' },
        textColor: '#333',
      },
      grid: {
        vertLines: { color: '#e0e0e0' },
        horzLines: { color: '#e0e0e0' },
      },
      crosshair: {
        mode: CrosshairMode.Normal,
      },
      rightPriceScale: {
        borderColor: '#cccccc',
      },
      timeScale: {
        borderColor: '#cccccc',
        timeVisible: true,
        secondsVisible: false,
      },
      width: chartContainerRef.current.clientWidth,
      height: height,
    })

    // Add candlestick series (v5 API)
    const candlestickSeriesInstance = chart.addSeries(CandlestickSeries, {
      upColor: '#26a69a',
      downColor: '#ef5350',
      borderDownColor: '#ef5350',
      borderUpColor: '#26a69a',
      wickDownColor: '#ef5350',
      wickUpColor: '#26a69a',
    })
    seriesRef.current.candlestick = candlestickSeriesInstance

    // Add volume series (v5 API)
    const volumeSeriesInstance = chart.addSeries(HistogramSeries, {
      priceFormat: { type: 'volume' },
      priceScaleId: 'volume',
    })
    chart.priceScale('volume').applyOptions({
      scaleMargins: { top: 0.8, bottom: 0 },
    })
    seriesRef.current.volume = volumeSeriesInstance

    chartRef.current = chart

    // Handle resize
    const handleResize = () => {
      if (chartContainerRef.current) {
        chart.applyOptions({ width: chartContainerRef.current.clientWidth })
      }
    }
    window.addEventListener('resize', handleResize)

    return () => {
      window.removeEventListener('resize', handleResize)
      chart.remove()
      chartRef.current = null
      seriesRef.current = {}
    }
  }, [height])

  // Update chart data
  useEffect(() => {
    if (!chartRef.current || !chartData.candles.length) return

    // Update candlestick data
    if (seriesRef.current.candlestick) {
      seriesRef.current.candlestick.setData(chartData.candles)
    }

    // Update volume data
    if (seriesRef.current.volume && showVolume) {
      seriesRef.current.volume.setData(chartData.volume)
    } else if (seriesRef.current.volume) {
      seriesRef.current.volume.setData([])
    }

    // Fit content
    chartRef.current.timeScale().fitContent()
  }, [chartData, showVolume])

  // Update indicator series
  useEffect(() => {
    if (!chartRef.current) return

    // Remove old indicator series
    Object.keys(seriesRef.current).forEach(key => {
      if (key !== 'candlestick' && key !== 'volume') {
        try {
          chartRef.current.removeSeries(seriesRef.current[key])
        } catch (e) {
          // Series may already be removed
        }
        delete seriesRef.current[key]
      }
    })

    // Add selected indicators
    selectedIndicators.forEach(indicatorKey => {
      // Find indicator config
      let config = null
      Object.values(INDICATOR_GROUPS).forEach(group => {
        if (group.indicators[indicatorKey]) {
          config = group.indicators[indicatorKey]
        }
      })

      if (!config || !chartData.indicators[indicatorKey]?.length) return

      if (config.overlay) {
        // Add as line on main price scale (v5 API)
        const series = chartRef.current.addSeries(LineSeries, {
          color: config.color,
          lineWidth: 2,
          title: config.label,
          priceScaleId: 'right',
        })
        series.setData(chartData.indicators[indicatorKey])
        seriesRef.current[indicatorKey] = series
      } else {
        // Add to separate pane (v5 API)
        const series = chartRef.current.addSeries(LineSeries, {
          color: config.color,
          lineWidth: 2,
          title: config.label,
          priceScaleId: config.pane,
          lastValueVisible: true,
          priceLineVisible: false,
        })

        // Configure price scale for the pane
        chartRef.current.priceScale(config.pane).applyOptions({
          scaleMargins: { top: 0.1, bottom: 0.1 },
          autoScale: true,
        })

        series.setData(chartData.indicators[indicatorKey])
        seriesRef.current[indicatorKey] = series
      }
    })
  }, [selectedIndicators, chartData.indicators])

  const toggleIndicator = (key) => {
    setSelectedIndicators(prev =>
      prev.includes(key) ? prev.filter(k => k !== key) : [...prev, key]
    )
  }

  const toggleGroup = (group) => {
    setExpandedGroups(prev => ({ ...prev, [group]: !prev[group] }))
  }

  // Zoom functions
  const zoomIn = useCallback(() => {
    if (!chartRef.current) return
    const timeScale = chartRef.current.timeScale()
    const currentRange = timeScale.getVisibleLogicalRange()
    if (!currentRange) return

    const rangeSize = currentRange.to - currentRange.from
    const center = (currentRange.from + currentRange.to) / 2
    const newSize = rangeSize * 0.7 // Zoom in by 30%

    timeScale.setVisibleLogicalRange({
      from: center - newSize / 2,
      to: center + newSize / 2
    })
  }, [])

  const zoomOut = useCallback(() => {
    if (!chartRef.current) return
    const timeScale = chartRef.current.timeScale()
    const currentRange = timeScale.getVisibleLogicalRange()
    if (!currentRange) return

    const rangeSize = currentRange.to - currentRange.from
    const center = (currentRange.from + currentRange.to) / 2
    const newSize = rangeSize * 1.4 // Zoom out by 40%

    timeScale.setVisibleLogicalRange({
      from: center - newSize / 2,
      to: center + newSize / 2
    })
  }, [])

  const resetZoom = useCallback(() => {
    if (!chartRef.current) return
    chartRef.current.timeScale().fitContent()
    setSelectedPeriod('All')
  }, [])

  // Set visible range based on selected time period
  const setTimePeriod = useCallback((periodLabel) => {
    if (!chartRef.current || !chartData.candles.length) return

    setSelectedPeriod(periodLabel)
    const period = TIME_PERIODS.find(p => p.label === periodLabel)

    if (!period || period.ms === null) {
      // Show all data
      chartRef.current.timeScale().fitContent()
      return
    }

    // Get the latest timestamp
    const latestCandle = chartData.candles[chartData.candles.length - 1]
    const latestTime = latestCandle.time
    const earliestTime = latestTime - (period.ms / 1000) // Convert ms to seconds

    // Find the index of the first candle in range
    const fromIndex = chartData.candles.findIndex(c => c.time >= earliestTime)
    const toIndex = chartData.candles.length - 1

    if (fromIndex >= 0) {
      chartRef.current.timeScale().setVisibleLogicalRange({
        from: fromIndex,
        to: toIndex
      })
    }
  }, [chartData.candles])

  if (!data || data.length === 0) {
    return (
      <div className="flex items-center justify-center h-64 text-gray-500">
        No data available for chart
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Chart Header */}
      <div className="flex items-center justify-between flex-wrap gap-2">
        <div>
          <h3 className="text-lg font-bold text-gray-900">{symbol}</h3>
          <p className="text-sm text-gray-500">{timeframe} timeframe</p>
        </div>

        {/* Period Selection Buttons */}
        <div className="flex items-center space-x-1 bg-gray-100 rounded-lg p-1">
          {TIME_PERIODS.map(period => (
            <button
              key={period.label}
              onClick={() => setTimePeriod(period.label)}
              className={`px-3 py-1 text-xs font-medium rounded transition ${
                selectedPeriod === period.label
                  ? 'bg-blue-500 text-white'
                  : 'text-gray-600 hover:bg-gray-200'
              }`}
            >
              {period.label}
            </button>
          ))}
        </div>

        {/* Zoom Controls and Volume Toggle */}
        <div className="flex items-center space-x-3">
          {/* Zoom Buttons */}
          <div className="flex items-center space-x-1 bg-gray-100 rounded-lg p-1">
            <button
              onClick={zoomIn}
              className="p-1.5 text-gray-600 hover:bg-gray-200 rounded transition"
              title="Zoom In"
            >
              <MagnifyingGlassPlusIcon className="w-4 h-4" />
            </button>
            <button
              onClick={zoomOut}
              className="p-1.5 text-gray-600 hover:bg-gray-200 rounded transition"
              title="Zoom Out"
            >
              <MagnifyingGlassMinusIcon className="w-4 h-4" />
            </button>
            <button
              onClick={resetZoom}
              className="p-1.5 text-gray-600 hover:bg-gray-200 rounded transition"
              title="Reset Zoom"
            >
              <ArrowsPointingOutIcon className="w-4 h-4" />
            </button>
          </div>

          <label className="flex items-center space-x-2 text-sm">
            <input
              type="checkbox"
              checked={showVolume}
              onChange={(e) => setShowVolume(e.target.checked)}
              className="h-4 w-4 text-blue-600 rounded"
            />
            <span>Volume</span>
          </label>
        </div>
      </div>

      <div className="flex gap-4">
        {/* Chart Container */}
        <div className="flex-1">
          <div ref={chartContainerRef} className="border border-gray-200 rounded-lg" />
          <p className="text-xs text-gray-400 mt-1 text-center">
            Use mouse wheel to zoom, drag to pan. Double-click to reset.
          </p>
        </div>

        {/* Indicator Panel */}
        <div className="w-64 border border-gray-200 rounded-lg p-3 max-h-[500px] overflow-y-auto">
          <h4 className="font-semibold text-gray-900 mb-3">Indicators</h4>

          {Object.entries(INDICATOR_GROUPS).map(([groupKey, group]) => (
            <div key={groupKey} className="mb-3">
              <button
                onClick={() => toggleGroup(groupKey)}
                className="flex items-center justify-between w-full py-2 px-2 bg-gray-100 rounded hover:bg-gray-200 transition"
              >
                <span className="text-sm font-medium text-gray-700">{group.label}</span>
                {expandedGroups[groupKey] ? (
                  <ChevronUpIcon className="w-4 h-4 text-gray-500" />
                ) : (
                  <ChevronDownIcon className="w-4 h-4 text-gray-500" />
                )}
              </button>

              {expandedGroups[groupKey] && (
                <div className="mt-2 space-y-1 pl-2">
                  {Object.entries(group.indicators).map(([key, config]) => {
                    const hasData = chartData.indicators[key]?.length > 0
                    return (
                      <label
                        key={key}
                        className={`flex items-center space-x-2 text-sm py-1 ${!hasData ? 'opacity-50' : 'cursor-pointer'}`}
                      >
                        <input
                          type="checkbox"
                          checked={selectedIndicators.includes(key)}
                          onChange={() => toggleIndicator(key)}
                          disabled={!hasData}
                          className="h-3 w-3 rounded"
                          style={{ accentColor: config.color }}
                        />
                        <span
                          className="w-3 h-3 rounded-full"
                          style={{ backgroundColor: config.color }}
                        />
                        <span className={hasData ? 'text-gray-700' : 'text-gray-400'}>
                          {config.label}
                        </span>
                        {!hasData && (
                          <span className="text-xs text-gray-400">(no data)</span>
                        )}
                      </label>
                    )
                  })}
                </div>
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Legend */}
      {selectedIndicators.length > 0 && (
        <div className="flex flex-wrap gap-3 text-sm">
          {selectedIndicators.map(key => {
            let config = null
            Object.values(INDICATOR_GROUPS).forEach(group => {
              if (group.indicators[key]) config = group.indicators[key]
            })
            if (!config) return null
            return (
              <div key={key} className="flex items-center space-x-1">
                <span
                  className="w-4 h-1 rounded"
                  style={{ backgroundColor: config.color }}
                />
                <span className="text-gray-600">{config.label}</span>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}

export default CandlestickChart
