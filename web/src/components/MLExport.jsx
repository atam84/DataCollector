import { useState, useEffect, useCallback } from 'react'
import axios from 'axios'

const API_BASE = '/api/v1'

// Feature category definitions
const FEATURE_CATEGORIES = {
  ohlcv: { label: 'OHLCV', features: ['open', 'high', 'low', 'close', 'volume'] },
  price_features: {
    label: 'Price Features',
    features: ['returns', 'log_returns', 'price_change', 'volatility', 'gaps', 'body_ratio', 'range_pct']
  },
  temporal: {
    label: 'Temporal Features',
    features: ['hour', 'hour_sin', 'hour_cos', 'day_of_week', 'dow_sin', 'dow_cos', 'month', 'is_weekend']
  },
  cross: {
    label: 'Cross Features',
    features: ['bb_position', 'price_vs_sma20', 'price_vs_sma50', 'ma_crossover', 'rsi_oversold', 'rsi_overbought']
  }
}

const INDICATOR_CATEGORIES = {
  trend: ['sma20', 'sma50', 'sma200', 'ema12', 'ema26', 'ema50', 'adx', 'supertrend'],
  momentum: ['rsi14', 'stoch_k', 'stoch_d', 'macd', 'macd_signal', 'macd_hist', 'cci', 'williams_r'],
  volatility: ['bb_upper', 'bb_middle', 'bb_lower', 'atr', 'keltner_upper', 'keltner_lower', 'stddev'],
  volume: ['obv', 'vwap', 'mfi', 'cmf', 'volume_sma']
}

export default function MLExport({ jobs = [] }) {
  const [selectedJobs, setSelectedJobs] = useState([])
  const [exportJobs, setExportJobs] = useState([])
  const [presets, setPresets] = useState([])
  const [profiles, setProfiles] = useState([])
  const [activePreset, setActivePreset] = useState(null)
  const [loading, setLoading] = useState(false)
  const [exportLoading, setExportLoading] = useState(false)
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [pollInterval, setPollInterval] = useState(null)

  // Export configuration state
  const [config, setConfig] = useState({
    format: 'csv',
    features: {
      include_ohlcv: true,
      include_volume: true,
      include_timestamp: true,
      include_all_indicators: false,
      indicator_categories: ['trend', 'momentum'],
      specific_indicators: [],
      price_features: ['returns', 'log_returns'],
      lagged_features: { enabled: true, lag_periods: [1, 5, 10] },
      rolling_features: { enabled: true, windows: [5, 10, 20], stats: ['mean', 'std'] },
      temporal_features: ['hour_sin', 'hour_cos', 'day_of_week'],
      cross_features: []
    },
    target: {
      enabled: true,
      type: 'future_returns',
      lookahead_periods: [1, 5],
      classification_bins: []
    },
    preprocessing: {
      normalization: 'zscore',
      nan_handling: 'forward_fill',
      remove_nan_rows: true,
      clip_outliers: false,
      outlier_stddev: 5
    },
    split: {
      enabled: true,
      train_ratio: 0.7,
      validation_ratio: 0.15,
      test_ratio: 0.15,
      time_based: true
    },
    sequence: {
      enabled: false,
      length: 60,
      stride: 1,
      include_target: true
    }
  })

  // Fetch presets and profiles
  const fetchPresets = async () => {
    try {
      const response = await axios.get(`${API_BASE}/ml/profiles/presets`)
      setPresets(response.data.data || [])
    } catch (err) {
      console.error('Failed to fetch presets:', err)
    }
  }

  const fetchProfiles = async () => {
    try {
      const response = await axios.get(`${API_BASE}/ml/profiles`)
      setProfiles(response.data.data || [])
    } catch (err) {
      console.error('Failed to fetch profiles:', err)
    }
  }

  const fetchExportJobs = async () => {
    try {
      const response = await axios.get(`${API_BASE}/ml/export/jobs?limit=10`)
      setExportJobs(response.data.data || [])
    } catch (err) {
      console.error('Failed to fetch export jobs:', err)
    }
  }

  useEffect(() => {
    fetchPresets()
    fetchProfiles()
    fetchExportJobs()
  }, [])

  // Poll for active export jobs
  useEffect(() => {
    const hasActiveJobs = exportJobs.some(j => j.status === 'pending' || j.status === 'running')
    if (hasActiveJobs && !pollInterval) {
      const interval = setInterval(fetchExportJobs, 3000)
      setPollInterval(interval)
    } else if (!hasActiveJobs && pollInterval) {
      clearInterval(pollInterval)
      setPollInterval(null)
    }
    return () => {
      if (pollInterval) clearInterval(pollInterval)
    }
  }, [exportJobs, pollInterval])

  // Apply preset
  const applyPreset = (preset) => {
    setActivePreset(preset.name)
    setConfig({
      format: preset.format || 'csv',
      features: {
        include_ohlcv: preset.features?.include_ohlcv ?? true,
        include_volume: preset.features?.include_volume ?? true,
        include_timestamp: preset.features?.include_timestamp ?? true,
        include_all_indicators: preset.features?.include_all_indicators ?? false,
        indicator_categories: preset.features?.indicator_categories || [],
        specific_indicators: preset.features?.specific_indicators || [],
        price_features: preset.features?.price_features || [],
        lagged_features: preset.features?.lagged_features || { enabled: false },
        rolling_features: preset.features?.rolling_features || { enabled: false },
        temporal_features: preset.features?.temporal_features || [],
        cross_features: preset.features?.cross_features || []
      },
      target: preset.target || { enabled: false },
      preprocessing: preset.preprocessing || { normalization: 'none', nan_handling: 'drop' },
      split: preset.split || { enabled: false },
      sequence: preset.sequence || { enabled: false }
    })
  }

  // Toggle job selection
  const toggleJobSelection = (jobId) => {
    setSelectedJobs(prev =>
      prev.includes(jobId)
        ? prev.filter(id => id !== jobId)
        : [...prev, jobId]
    )
  }

  // Start export
  const startExport = async () => {
    if (selectedJobs.length === 0) {
      alert('Please select at least one job to export')
      return
    }

    setExportLoading(true)
    try {
      const payload = {
        job_ids: selectedJobs,
        config: config
      }
      await axios.post(`${API_BASE}/ml/export/start`, payload)
      fetchExportJobs()
      setSelectedJobs([])
    } catch (err) {
      console.error('Failed to start export:', err)
      alert('Failed to start export: ' + (err.response?.data?.message || err.message))
    } finally {
      setExportLoading(false)
    }
  }

  // Download export
  const downloadExport = (exportJob) => {
    window.open(`${API_BASE}/ml/export/jobs/${exportJob.id}/download`, '_blank')
  }

  // Delete export job
  const deleteExportJob = async (id) => {
    try {
      await axios.delete(`${API_BASE}/ml/export/jobs/${id}`)
      fetchExportJobs()
    } catch (err) {
      console.error('Failed to delete export job:', err)
    }
  }

  // Update config helper
  const updateConfig = (path, value) => {
    setConfig(prev => {
      const newConfig = { ...prev }
      const keys = path.split('.')
      let obj = newConfig
      for (let i = 0; i < keys.length - 1; i++) {
        obj = obj[keys[i]]
      }
      obj[keys[keys.length - 1]] = value
      return newConfig
    })
    setActivePreset(null)
  }

  // Toggle array item
  const toggleArrayItem = (path, item) => {
    setConfig(prev => {
      const newConfig = { ...prev }
      const keys = path.split('.')
      let obj = newConfig
      for (let i = 0; i < keys.length - 1; i++) {
        obj = obj[keys[i]]
      }
      const arr = obj[keys[keys.length - 1]] || []
      if (arr.includes(item)) {
        obj[keys[keys.length - 1]] = arr.filter(i => i !== item)
      } else {
        obj[keys[keys.length - 1]] = [...arr, item]
      }
      return newConfig
    })
    setActivePreset(null)
  }

  const getStatusColor = (status) => {
    switch (status) {
      case 'completed': return 'bg-green-100 text-green-800'
      case 'running': return 'bg-blue-100 text-blue-800'
      case 'pending': return 'bg-yellow-100 text-yellow-800'
      case 'failed': return 'bg-red-100 text-red-800'
      case 'cancelled': return 'bg-gray-100 text-gray-800'
      default: return 'bg-gray-100 text-gray-600'
    }
  }

  const formatBytes = (bytes) => {
    if (!bytes) return '-'
    if (bytes < 1024) return bytes + ' B'
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">ML Export</h2>
          <p className="text-sm text-gray-500 mt-1">Export data with feature engineering for machine learning</p>
        </div>
        <button
          onClick={fetchExportJobs}
          className="px-4 py-2 bg-gray-100 text-gray-700 rounded hover:bg-gray-200"
        >
          Refresh
        </button>
      </div>

      {/* Recent Export Jobs */}
      {exportJobs.length > 0 && (
        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold mb-4">Recent Exports</h3>
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead>
                <tr>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Progress</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Format</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Rows</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Features</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Size</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {exportJobs.map(job => (
                  <tr key={job.id}>
                    <td className="px-4 py-3">
                      <span className={`px-2 py-1 text-xs rounded-full ${getStatusColor(job.status)}`}>
                        {job.status}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      {job.status === 'running' || job.status === 'pending' ? (
                        <div className="flex items-center">
                          <div className="w-24 bg-gray-200 rounded-full h-2 mr-2">
                            <div
                              className="bg-blue-500 h-2 rounded-full transition-all"
                              style={{ width: `${job.progress}%` }}
                            />
                          </div>
                          <span className="text-xs text-gray-500">{job.progress?.toFixed(0)}%</span>
                        </div>
                      ) : job.status === 'completed' ? (
                        <span className="text-green-600">100%</span>
                      ) : '-'}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-600 uppercase">{job.format || '-'}</td>
                    <td className="px-4 py-3 text-sm text-gray-600">{job.row_count?.toLocaleString() || '-'}</td>
                    <td className="px-4 py-3 text-sm text-gray-600">{job.feature_count || '-'}</td>
                    <td className="px-4 py-3 text-sm text-gray-600">{formatBytes(job.file_size_bytes)}</td>
                    <td className="px-4 py-3">
                      <div className="flex space-x-2">
                        {job.status === 'completed' && (
                          <button
                            onClick={() => downloadExport(job)}
                            className="text-blue-600 hover:text-blue-800 text-sm"
                          >
                            Download
                          </button>
                        )}
                        <button
                          onClick={() => deleteExportJob(job.id)}
                          className="text-red-600 hover:text-red-800 text-sm"
                        >
                          Delete
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Job Selection */}
        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold mb-4">Select Jobs</h3>
          <div className="space-y-2 max-h-96 overflow-y-auto">
            {jobs.length === 0 ? (
              <p className="text-gray-500 text-sm">No jobs available</p>
            ) : (
              jobs.map(job => (
                <label
                  key={job.id}
                  className={`flex items-center p-3 rounded border cursor-pointer transition ${
                    selectedJobs.includes(job.id)
                      ? 'border-blue-500 bg-blue-50'
                      : 'border-gray-200 hover:border-gray-300'
                  }`}
                >
                  <input
                    type="checkbox"
                    checked={selectedJobs.includes(job.id)}
                    onChange={() => toggleJobSelection(job.id)}
                    className="h-4 w-4 text-blue-600 rounded"
                  />
                  <div className="ml-3">
                    <div className="font-medium text-sm">{job.symbol}</div>
                    <div className="text-xs text-gray-500">
                      {job.connector_exchange_id} | {job.timeframe} | {job.candle_count?.toLocaleString() || 0} candles
                    </div>
                  </div>
                </label>
              ))
            )}
          </div>
          {selectedJobs.length > 0 && (
            <div className="mt-4 p-3 bg-blue-50 rounded text-sm text-blue-700">
              {selectedJobs.length} job(s) selected
            </div>
          )}
        </div>

        {/* Configuration */}
        <div className="lg:col-span-2 space-y-6">
          {/* Presets */}
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-lg font-semibold mb-4">Quick Presets</h3>
            <div className="flex flex-wrap gap-2">
              {presets.map(preset => (
                <button
                  key={preset.name}
                  onClick={() => applyPreset(preset.config || preset)}
                  className={`px-4 py-2 rounded text-sm transition ${
                    activePreset === preset.name
                      ? 'bg-blue-500 text-white'
                      : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                  }`}
                  title={preset.description}
                >
                  {preset.name}
                </button>
              ))}
            </div>
          </div>

          {/* Format & Output */}
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-lg font-semibold mb-4">Output Format</h3>
            <div className="grid grid-cols-4 gap-3">
              {['csv', 'parquet', 'numpy', 'jsonl'].map(format => (
                <button
                  key={format}
                  onClick={() => updateConfig('format', format)}
                  className={`p-3 rounded border text-sm font-medium transition ${
                    config.format === format
                      ? 'border-blue-500 bg-blue-50 text-blue-700'
                      : 'border-gray-200 hover:border-gray-300'
                  }`}
                >
                  {format.toUpperCase()}
                </button>
              ))}
            </div>
            <div className="mt-3 text-xs text-gray-500">
              {config.format === 'csv' && 'Compatible with pandas, Excel, and most ML frameworks'}
              {config.format === 'parquet' && 'Efficient columnar format for large datasets'}
              {config.format === 'numpy' && 'Direct loading with np.load() for TensorFlow/PyTorch'}
              {config.format === 'jsonl' && 'Line-delimited JSON for streaming processing'}
            </div>
          </div>

          {/* Features */}
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-lg font-semibold mb-4">Features</h3>

            {/* Base Features */}
            <div className="mb-4">
              <h4 className="text-sm font-medium text-gray-700 mb-2">Base Data</h4>
              <div className="flex flex-wrap gap-2">
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={config.features.include_ohlcv}
                    onChange={e => updateConfig('features.include_ohlcv', e.target.checked)}
                    className="h-4 w-4 text-blue-600 rounded"
                  />
                  <span className="ml-2 text-sm">OHLCV</span>
                </label>
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={config.features.include_volume}
                    onChange={e => updateConfig('features.include_volume', e.target.checked)}
                    className="h-4 w-4 text-blue-600 rounded"
                  />
                  <span className="ml-2 text-sm">Volume</span>
                </label>
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={config.features.include_timestamp}
                    onChange={e => updateConfig('features.include_timestamp', e.target.checked)}
                    className="h-4 w-4 text-blue-600 rounded"
                  />
                  <span className="ml-2 text-sm">Timestamp</span>
                </label>
              </div>
            </div>

            {/* Indicator Categories */}
            <div className="mb-4">
              <h4 className="text-sm font-medium text-gray-700 mb-2">Indicator Categories</h4>
              <div className="flex flex-wrap gap-2">
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={config.features.include_all_indicators}
                    onChange={e => updateConfig('features.include_all_indicators', e.target.checked)}
                    className="h-4 w-4 text-blue-600 rounded"
                  />
                  <span className="ml-2 text-sm font-medium">All Indicators</span>
                </label>
                {!config.features.include_all_indicators && (
                  <>
                    {['trend', 'momentum', 'volatility', 'volume'].map(cat => (
                      <label key={cat} className="flex items-center">
                        <input
                          type="checkbox"
                          checked={config.features.indicator_categories.includes(cat)}
                          onChange={() => toggleArrayItem('features.indicator_categories', cat)}
                          className="h-4 w-4 text-blue-600 rounded"
                        />
                        <span className="ml-2 text-sm capitalize">{cat}</span>
                      </label>
                    ))}
                  </>
                )}
              </div>
            </div>

            {/* Price Features */}
            <div className="mb-4">
              <h4 className="text-sm font-medium text-gray-700 mb-2">Price Features</h4>
              <div className="flex flex-wrap gap-2">
                {FEATURE_CATEGORIES.price_features.features.map(feat => (
                  <label key={feat} className="flex items-center">
                    <input
                      type="checkbox"
                      checked={config.features.price_features.includes(feat)}
                      onChange={() => toggleArrayItem('features.price_features', feat)}
                      className="h-4 w-4 text-blue-600 rounded"
                    />
                    <span className="ml-2 text-sm">{feat.replace(/_/g, ' ')}</span>
                  </label>
                ))}
              </div>
            </div>

            {/* Temporal Features */}
            <div className="mb-4">
              <h4 className="text-sm font-medium text-gray-700 mb-2">Temporal Features</h4>
              <div className="flex flex-wrap gap-2">
                {FEATURE_CATEGORIES.temporal.features.map(feat => (
                  <label key={feat} className="flex items-center">
                    <input
                      type="checkbox"
                      checked={config.features.temporal_features.includes(feat)}
                      onChange={() => toggleArrayItem('features.temporal_features', feat)}
                      className="h-4 w-4 text-blue-600 rounded"
                    />
                    <span className="ml-2 text-sm">{feat.replace(/_/g, ' ')}</span>
                  </label>
                ))}
              </div>
            </div>

            {/* Lagged Features */}
            <div className="mb-4">
              <label className="flex items-center mb-2">
                <input
                  type="checkbox"
                  checked={config.features.lagged_features.enabled}
                  onChange={e => updateConfig('features.lagged_features.enabled', e.target.checked)}
                  className="h-4 w-4 text-blue-600 rounded"
                />
                <span className="ml-2 text-sm font-medium text-gray-700">Lagged Features</span>
              </label>
              {config.features.lagged_features.enabled && (
                <input
                  type="text"
                  value={(config.features.lagged_features.lag_periods || []).join(', ')}
                  onChange={e => updateConfig('features.lagged_features.lag_periods',
                    e.target.value.split(',').map(v => parseInt(v.trim())).filter(v => !isNaN(v))
                  )}
                  placeholder="1, 5, 10, 20"
                  className="w-full px-3 py-2 border rounded text-sm"
                />
              )}
            </div>

            {/* Rolling Features */}
            <div>
              <label className="flex items-center mb-2">
                <input
                  type="checkbox"
                  checked={config.features.rolling_features.enabled}
                  onChange={e => updateConfig('features.rolling_features.enabled', e.target.checked)}
                  className="h-4 w-4 text-blue-600 rounded"
                />
                <span className="ml-2 text-sm font-medium text-gray-700">Rolling Statistics</span>
              </label>
              {config.features.rolling_features.enabled && (
                <div className="grid grid-cols-2 gap-2">
                  <input
                    type="text"
                    value={(config.features.rolling_features.windows || []).join(', ')}
                    onChange={e => updateConfig('features.rolling_features.windows',
                      e.target.value.split(',').map(v => parseInt(v.trim())).filter(v => !isNaN(v))
                    )}
                    placeholder="Windows: 5, 10, 20"
                    className="px-3 py-2 border rounded text-sm"
                  />
                  <select
                    multiple
                    value={config.features.rolling_features.stats || []}
                    onChange={e => updateConfig('features.rolling_features.stats',
                      Array.from(e.target.selectedOptions, opt => opt.value)
                    )}
                    className="px-3 py-2 border rounded text-sm"
                  >
                    <option value="mean">Mean</option>
                    <option value="std">Std</option>
                    <option value="min">Min</option>
                    <option value="max">Max</option>
                    <option value="median">Median</option>
                  </select>
                </div>
              )}
            </div>
          </div>

          {/* Target */}
          <div className="bg-white rounded-lg shadow p-6">
            <label className="flex items-center mb-4">
              <input
                type="checkbox"
                checked={config.target.enabled}
                onChange={e => updateConfig('target.enabled', e.target.checked)}
                className="h-4 w-4 text-blue-600 rounded"
              />
              <h3 className="text-lg font-semibold ml-2">Target Variable</h3>
            </label>

            {config.target.enabled && (
              <div className="space-y-4">
                <div>
                  <label className="text-sm font-medium text-gray-700">Target Type</label>
                  <select
                    value={config.target.type}
                    onChange={e => updateConfig('target.type', e.target.value)}
                    className="mt-1 w-full px-3 py-2 border rounded"
                  >
                    <option value="future_returns">Future Returns (%)</option>
                    <option value="future_direction">Future Direction (0/1)</option>
                    <option value="future_class">Future Class (multi-class)</option>
                    <option value="future_volatility">Future Volatility</option>
                  </select>
                </div>
                <div>
                  <label className="text-sm font-medium text-gray-700">Lookahead Periods</label>
                  <input
                    type="text"
                    value={(config.target.lookahead_periods || []).join(', ')}
                    onChange={e => updateConfig('target.lookahead_periods',
                      e.target.value.split(',').map(v => parseInt(v.trim())).filter(v => !isNaN(v))
                    )}
                    placeholder="1, 5, 10"
                    className="mt-1 w-full px-3 py-2 border rounded"
                  />
                </div>
                {config.target.type === 'future_class' && (
                  <div>
                    <label className="text-sm font-medium text-gray-700">Classification Bins</label>
                    <input
                      type="text"
                      value={(config.target.classification_bins || []).join(', ')}
                      onChange={e => updateConfig('target.classification_bins',
                        e.target.value.split(',').map(v => parseFloat(v.trim())).filter(v => !isNaN(v))
                      )}
                      placeholder="-0.02, -0.01, 0.01, 0.02"
                      className="mt-1 w-full px-3 py-2 border rounded"
                    />
                    <p className="text-xs text-gray-500 mt-1">Returns below first bin = class 0, between bins = middle classes, above last bin = highest class</p>
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Advanced Settings */}
          <div className="bg-white rounded-lg shadow">
            <button
              onClick={() => setShowAdvanced(!showAdvanced)}
              className="w-full p-6 flex justify-between items-center"
            >
              <h3 className="text-lg font-semibold">Advanced Settings</h3>
              <svg
                className={`w-5 h-5 transform transition ${showAdvanced ? 'rotate-180' : ''}`}
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
              </svg>
            </button>

            {showAdvanced && (
              <div className="px-6 pb-6 space-y-6">
                {/* Preprocessing */}
                <div>
                  <h4 className="text-sm font-medium text-gray-700 mb-3">Preprocessing</h4>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="text-xs text-gray-500">Normalization</label>
                      <select
                        value={config.preprocessing.normalization}
                        onChange={e => updateConfig('preprocessing.normalization', e.target.value)}
                        className="mt-1 w-full px-3 py-2 border rounded text-sm"
                      >
                        <option value="none">None</option>
                        <option value="minmax">Min-Max [0,1]</option>
                        <option value="zscore">Z-Score</option>
                        <option value="robust">Robust (IQR)</option>
                      </select>
                    </div>
                    <div>
                      <label className="text-xs text-gray-500">NaN Handling</label>
                      <select
                        value={config.preprocessing.nan_handling}
                        onChange={e => updateConfig('preprocessing.nan_handling', e.target.value)}
                        className="mt-1 w-full px-3 py-2 border rounded text-sm"
                      >
                        <option value="drop">Drop rows</option>
                        <option value="forward_fill">Forward fill</option>
                        <option value="backward_fill">Backward fill</option>
                        <option value="interpolate">Interpolate</option>
                        <option value="zero">Replace with 0</option>
                      </select>
                    </div>
                  </div>
                  <div className="mt-3 flex items-center gap-4">
                    <label className="flex items-center">
                      <input
                        type="checkbox"
                        checked={config.preprocessing.remove_nan_rows}
                        onChange={e => updateConfig('preprocessing.remove_nan_rows', e.target.checked)}
                        className="h-4 w-4 text-blue-600 rounded"
                      />
                      <span className="ml-2 text-sm">Remove remaining NaN rows</span>
                    </label>
                    <label className="flex items-center">
                      <input
                        type="checkbox"
                        checked={config.preprocessing.clip_outliers}
                        onChange={e => updateConfig('preprocessing.clip_outliers', e.target.checked)}
                        className="h-4 w-4 text-blue-600 rounded"
                      />
                      <span className="ml-2 text-sm">Clip outliers</span>
                    </label>
                  </div>
                </div>

                {/* Split */}
                <div>
                  <label className="flex items-center mb-3">
                    <input
                      type="checkbox"
                      checked={config.split.enabled}
                      onChange={e => updateConfig('split.enabled', e.target.checked)}
                      className="h-4 w-4 text-blue-600 rounded"
                    />
                    <span className="ml-2 text-sm font-medium text-gray-700">Train/Val/Test Split</span>
                  </label>
                  {config.split.enabled && (
                    <div className="grid grid-cols-4 gap-3">
                      <div>
                        <label className="text-xs text-gray-500">Train</label>
                        <input
                          type="number"
                          step="0.05"
                          min="0"
                          max="1"
                          value={config.split.train_ratio}
                          onChange={e => updateConfig('split.train_ratio', parseFloat(e.target.value))}
                          className="mt-1 w-full px-3 py-2 border rounded text-sm"
                        />
                      </div>
                      <div>
                        <label className="text-xs text-gray-500">Validation</label>
                        <input
                          type="number"
                          step="0.05"
                          min="0"
                          max="1"
                          value={config.split.validation_ratio}
                          onChange={e => updateConfig('split.validation_ratio', parseFloat(e.target.value))}
                          className="mt-1 w-full px-3 py-2 border rounded text-sm"
                        />
                      </div>
                      <div>
                        <label className="text-xs text-gray-500">Test</label>
                        <input
                          type="number"
                          step="0.05"
                          min="0"
                          max="1"
                          value={config.split.test_ratio}
                          onChange={e => updateConfig('split.test_ratio', parseFloat(e.target.value))}
                          className="mt-1 w-full px-3 py-2 border rounded text-sm"
                        />
                      </div>
                      <div className="flex items-end">
                        <label className="flex items-center">
                          <input
                            type="checkbox"
                            checked={config.split.time_based}
                            onChange={e => updateConfig('split.time_based', e.target.checked)}
                            className="h-4 w-4 text-blue-600 rounded"
                          />
                          <span className="ml-2 text-sm">Time-based</span>
                        </label>
                      </div>
                    </div>
                  )}
                </div>

                {/* Sequence */}
                <div>
                  <label className="flex items-center mb-3">
                    <input
                      type="checkbox"
                      checked={config.sequence.enabled}
                      onChange={e => updateConfig('sequence.enabled', e.target.checked)}
                      className="h-4 w-4 text-blue-600 rounded"
                    />
                    <span className="ml-2 text-sm font-medium text-gray-700">Generate Sequences (LSTM/Transformer)</span>
                  </label>
                  {config.sequence.enabled && (
                    <div className="grid grid-cols-3 gap-3">
                      <div>
                        <label className="text-xs text-gray-500">Sequence Length</label>
                        <input
                          type="number"
                          min="1"
                          value={config.sequence.length}
                          onChange={e => updateConfig('sequence.length', parseInt(e.target.value))}
                          className="mt-1 w-full px-3 py-2 border rounded text-sm"
                        />
                      </div>
                      <div>
                        <label className="text-xs text-gray-500">Stride</label>
                        <input
                          type="number"
                          min="1"
                          value={config.sequence.stride}
                          onChange={e => updateConfig('sequence.stride', parseInt(e.target.value))}
                          className="mt-1 w-full px-3 py-2 border rounded text-sm"
                        />
                      </div>
                      <div className="flex items-end">
                        <label className="flex items-center">
                          <input
                            type="checkbox"
                            checked={config.sequence.include_target}
                            onChange={e => updateConfig('sequence.include_target', e.target.checked)}
                            className="h-4 w-4 text-blue-600 rounded"
                          />
                          <span className="ml-2 text-sm">Include target</span>
                        </label>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            )}
          </div>

          {/* Export Button */}
          <div className="flex justify-end">
            <button
              onClick={startExport}
              disabled={selectedJobs.length === 0 || exportLoading}
              className={`px-6 py-3 rounded-lg font-medium text-white transition ${
                selectedJobs.length === 0 || exportLoading
                  ? 'bg-gray-400 cursor-not-allowed'
                  : 'bg-blue-600 hover:bg-blue-700'
              }`}
            >
              {exportLoading ? (
                <span className="flex items-center">
                  <svg className="animate-spin -ml-1 mr-2 h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                  Starting Export...
                </span>
              ) : (
                `Export ${selectedJobs.length} Job${selectedJobs.length !== 1 ? 's' : ''}`
              )}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
