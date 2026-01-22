import { useState, useEffect, useMemo } from 'react'
import axios from 'axios'
import {
  ArrowLeftIcon,
  ArrowRightIcon,
  CheckIcon,
  MagnifyingGlassIcon,
  ClockIcon,
  ExclamationTriangleIcon
} from '@heroicons/react/24/outline'

const API_BASE = '/api/v1'

// Fallback popular pairs (used if API fails)
const FALLBACK_POPULAR_PAIRS = [
  'BTC/USDT', 'ETH/USDT', 'BNB/USDT', 'SOL/USDT', 'XRP/USDT',
  'ADA/USDT', 'DOGE/USDT', 'DOT/USDT', 'MATIC/USDT', 'AVAX/USDT',
  'LINK/USDT', 'UNI/USDT', 'ATOM/USDT', 'LTC/USDT', 'ETC/USDT',
  'BCH/USDT', 'APT/USDT', 'ARB/USDT', 'OP/USDT', 'INJ/USDT'
]

// Default timeframes if exchange doesn't provide them
const DEFAULT_TIMEFRAMES = [
  { id: '1m', label: '1 Minute', description: 'Very short-term' },
  { id: '5m', label: '5 Minutes', description: 'Short-term' },
  { id: '15m', label: '15 Minutes', description: 'Intraday' },
  { id: '30m', label: '30 Minutes', description: 'Intraday' },
  { id: '1h', label: '1 Hour', description: 'Short-term' },
  { id: '4h', label: '4 Hours', description: 'Medium-term' },
  { id: '1d', label: '1 Day', description: 'Long-term' },
  { id: '1w', label: '1 Week', description: 'Very long-term' }
]

// Timeframe labels and descriptions
const TIMEFRAME_INFO = {
  '1m': { label: '1 Minute', description: 'Very short-term' },
  '3m': { label: '3 Minutes', description: 'Very short-term' },
  '5m': { label: '5 Minutes', description: 'Short-term' },
  '15m': { label: '15 Minutes', description: 'Intraday' },
  '30m': { label: '30 Minutes', description: 'Intraday' },
  '1h': { label: '1 Hour', description: 'Short-term' },
  '2h': { label: '2 Hours', description: 'Medium-term' },
  '4h': { label: '4 Hours', description: 'Medium-term' },
  '6h': { label: '6 Hours', description: 'Medium-term' },
  '8h': { label: '8 Hours', description: 'Medium-term' },
  '12h': { label: '12 Hours', description: 'Medium-term' },
  '1d': { label: '1 Day', description: 'Long-term' },
  '3d': { label: '3 Days', description: 'Long-term' },
  '1w': { label: '1 Week', description: 'Very long-term' },
  '1M': { label: '1 Month', description: 'Very long-term' }
}

function JobWizard({ connectors, onClose, onSave }) {
  const [currentStep, setCurrentStep] = useState(1)
  const [selectedConnector, setSelectedConnector] = useState('')
  const [selectedPairs, setSelectedPairs] = useState([])
  const [customPair, setCustomPair] = useState('')
  const [selectedTimeframes, setSelectedTimeframes] = useState([])
  const [saving, setSaving] = useState(false)
  const [collectHistorical, setCollectHistorical] = useState(false)

  // Exchange-specific data
  const [exchangeMetadata, setExchangeMetadata] = useState(null)
  const [exchangeSymbols, setExchangeSymbols] = useState([])
  const [popularPairs, setPopularPairs] = useState([])
  const [loadingSymbols, setLoadingSymbols] = useState(false)
  const [loadingPopular, setLoadingPopular] = useState(false)
  const [symbolSearch, setSymbolSearch] = useState('')
  const [validationResults, setValidationResults] = useState({})
  const [validating, setValidating] = useState(false)

  // Fetch exchange metadata when connector changes
  useEffect(() => {
    if (!selectedConnector) {
      setExchangeMetadata(null)
      setExchangeSymbols([])
      return
    }

    const fetchExchangeData = async () => {
      try {
        // Fetch metadata for timeframes
        const metaResponse = await axios.get(`${API_BASE}/exchanges/${selectedConnector}/metadata`)
        setExchangeMetadata(metaResponse.data)
      } catch (err) {
        console.error('Failed to fetch exchange metadata:', err)
        setExchangeMetadata(null)
      }
    }

    fetchExchangeData()
  }, [selectedConnector])

  // Fetch symbols and popular pairs when entering step 2
  useEffect(() => {
    if (currentStep === 2 && selectedConnector) {
      // Fetch all symbols
      if (exchangeSymbols.length === 0) {
        const fetchSymbols = async () => {
          setLoadingSymbols(true)
          try {
            const response = await axios.get(`${API_BASE}/exchanges/${selectedConnector}/symbols`)
            setExchangeSymbols(response.data.symbols || [])
          } catch (err) {
            console.error('Failed to fetch exchange symbols:', err)
            setExchangeSymbols([])
          } finally {
            setLoadingSymbols(false)
          }
        }
        fetchSymbols()
      }

      // Fetch popular pairs for this exchange
      if (popularPairs.length === 0) {
        const fetchPopular = async () => {
          setLoadingPopular(true)
          try {
            const response = await axios.get(`${API_BASE}/exchanges/${selectedConnector}/symbols/popular`)
            setPopularPairs(response.data.popular || [])
          } catch (err) {
            console.error('Failed to fetch popular pairs:', err)
            // Use fallback and filter by available symbols
            setPopularPairs(FALLBACK_POPULAR_PAIRS)
          } finally {
            setLoadingPopular(false)
          }
        }
        fetchPopular()
      }
    }
  }, [currentStep, selectedConnector, exchangeSymbols.length, popularPairs.length])

  // Get available timeframes from exchange metadata or use defaults
  const availableTimeframes = useMemo(() => {
    if (exchangeMetadata?.timeframes && Object.keys(exchangeMetadata.timeframes).length > 0) {
      return Object.keys(exchangeMetadata.timeframes).map(tf => ({
        id: tf,
        label: TIMEFRAME_INFO[tf]?.label || tf,
        description: TIMEFRAME_INFO[tf]?.description || ''
      }))
    }
    return DEFAULT_TIMEFRAMES
  }, [exchangeMetadata])

  // Filter symbols based on search
  const filteredSymbols = useMemo(() => {
    if (!symbolSearch) return exchangeSymbols.slice(0, 100) // Show first 100 by default
    const term = symbolSearch.toUpperCase()
    return exchangeSymbols.filter(s => s.toUpperCase().includes(term)).slice(0, 100)
  }, [exchangeSymbols, symbolSearch])

  // Check if symbol is available on exchange
  const isSymbolAvailable = (symbol) => {
    if (exchangeSymbols.length === 0) return true // If no symbols loaded, assume available
    return exchangeSymbols.includes(symbol)
  }

  // Validate selected pairs before moving to step 3
  const validateSelectedPairs = async () => {
    if (selectedPairs.length === 0 || exchangeSymbols.length === 0) return true

    setValidating(true)
    try {
      const response = await axios.post(`${API_BASE}/exchanges/${selectedConnector}/symbols/validate`, {
        symbols: selectedPairs
      })

      setValidationResults(response.data.results || {})

      if (response.data.invalid_count > 0) {
        const invalidSymbols = selectedPairs.filter(p => !response.data.results[p])
        return { valid: false, invalid: invalidSymbols }
      }
      return { valid: true }
    } catch (err) {
      console.error('Failed to validate symbols:', err)
      // If validation fails, allow to proceed but warn
      return { valid: true, warning: 'Could not validate symbols' }
    } finally {
      setValidating(false)
    }
  }

  const handleNext = async () => {
    if (currentStep === 1 && !selectedConnector) {
      alert('Please select a connector')
      return
    }
    if (currentStep === 2 && selectedPairs.length === 0) {
      alert('Please select at least one cryptocurrency pair')
      return
    }

    // Validate symbols before moving to step 3
    if (currentStep === 2) {
      const validationResult = await validateSelectedPairs()
      if (!validationResult.valid) {
        const invalidList = validationResult.invalid.join(', ')
        const proceed = window.confirm(
          `The following pairs may not be available on ${getConnectorName()}:\n\n${invalidList}\n\nDo you want to continue anyway?`
        )
        if (!proceed) return
      }
    }

    setCurrentStep(currentStep + 1)
  }

  const handleBack = () => {
    setCurrentStep(currentStep - 1)
  }

  const togglePair = (pair) => {
    if (selectedPairs.includes(pair)) {
      setSelectedPairs(selectedPairs.filter(p => p !== pair))
    } else {
      setSelectedPairs([...selectedPairs, pair])
    }
  }

  const addCustomPair = () => {
    if (customPair && !selectedPairs.includes(customPair)) {
      setSelectedPairs([...selectedPairs, customPair])
      setCustomPair('')
    }
  }

  const toggleTimeframe = (timeframe) => {
    if (selectedTimeframes.includes(timeframe)) {
      setSelectedTimeframes(selectedTimeframes.filter(t => t !== timeframe))
    } else {
      setSelectedTimeframes([...selectedTimeframes, timeframe])
    }
  }

  const handleConnectorSelect = (exchangeId) => {
    setSelectedConnector(exchangeId)
    // Reset exchange-specific data when connector changes
    setExchangeSymbols([])
    setPopularPairs([])
    setSelectedTimeframes([])
    setSelectedPairs([])
    setValidationResults({})
  }

  const handleCreateJobs = async () => {
    if (selectedTimeframes.length === 0) {
      alert('Please select at least one timeframe')
      return
    }

    setSaving(true)
    try {
      // Prepare jobs data
      const jobs = []
      for (const pair of selectedPairs) {
        for (const timeframe of selectedTimeframes) {
          jobs.push({
            connector_exchange_id: selectedConnector,
            symbol: pair,
            timeframe: timeframe,
            status: 'active',
            collect_historical: collectHistorical
          })
        }
      }

      await onSave(jobs)
      onClose()
    } catch (err) {
      alert('Failed to create jobs: ' + err.message)
    } finally {
      setSaving(false)
    }
  }

  const getConnectorName = () => {
    const connector = connectors.find(c => c.exchange_id === selectedConnector)
    return connector ? connector.display_name : selectedConnector
  }

  const totalJobs = selectedPairs.length * selectedTimeframes.length

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg w-full max-w-6xl max-h-[95vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="p-6 border-b bg-gradient-to-r from-green-500 to-teal-600">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-2xl font-bold text-white">Create Multiple Jobs</h3>
              <p className="text-green-100 mt-1">Step {currentStep} of 3</p>
              {totalJobs > 0 && (
                <p className="text-white font-medium mt-2">
                  Will create {totalJobs} job{totalJobs !== 1 ? 's' : ''}
                </p>
              )}
            </div>
            <button
              onClick={onClose}
              className="text-white hover:text-green-100 text-2xl"
              disabled={saving}
            >
              ×
            </button>
          </div>

          {/* Progress Bar */}
          <div className="mt-4 flex items-center space-x-2">
            <div className={`flex-1 h-2 rounded ${currentStep >= 1 ? 'bg-white' : 'bg-green-300'}`} />
            <div className={`flex-1 h-2 rounded ${currentStep >= 2 ? 'bg-white' : 'bg-green-300'}`} />
            <div className={`flex-1 h-2 rounded ${currentStep >= 3 ? 'bg-white' : 'bg-green-300'}`} />
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          {/* Step 1: Connector Selection */}
          {currentStep === 1 && (
            <div>
              <h4 className="text-lg font-bold text-gray-900 mb-4">Select Connector</h4>
              <p className="text-sm text-gray-600 mb-6">
                Choose which connector will be used for all jobs
              </p>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {connectors.map(connector => (
                  <button
                    key={connector.id}
                    onClick={() => handleConnectorSelect(connector.exchange_id)}
                    className={`p-4 border-2 rounded-lg text-left transition ${
                      selectedConnector === connector.exchange_id
                        ? 'border-green-500 bg-green-50'
                        : 'border-gray-200 hover:border-green-300'
                    }`}
                  >
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-bold text-gray-900">{connector.display_name}</span>
                      <span className={`px-2 py-1 text-xs font-medium rounded ${
                        connector.status === 'active'
                          ? 'bg-green-100 text-green-800'
                          : 'bg-gray-100 text-gray-800'
                      }`}>
                        {connector.status}
                      </span>
                    </div>
                    <div className="text-sm text-gray-600">
                      <div>Exchange: {connector.exchange_id}</div>
                      <div>Jobs: {connector.job_count || 0}</div>
                    </div>
                  </button>
                ))}
              </div>

              {/* Exchange Info */}
              {exchangeMetadata && (
                <div className="mt-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
                  <h5 className="font-semibold text-blue-900 mb-2">{exchangeMetadata.name} Info</h5>
                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <span className="text-blue-700 font-medium">OHLCV Limit:</span>
                      <span className="ml-2 text-blue-900">{exchangeMetadata.ohlcv_limit || 500} candles/request</span>
                    </div>
                    <div>
                      <span className="text-blue-700 font-medium">Timeframes:</span>
                      <span className="ml-2 text-blue-900">
                        {exchangeMetadata.timeframes ? Object.keys(exchangeMetadata.timeframes).length : 0} available
                      </span>
                    </div>
                  </div>
                </div>
              )}

              {connectors.length === 0 && (
                <div className="text-center py-12 text-gray-500">
                  No connectors available. Please create a connector first.
                </div>
              )}
            </div>
          )}

          {/* Step 2: Cryptocurrency Selection */}
          {currentStep === 2 && (
            <div>
              <h4 className="text-lg font-bold text-gray-900 mb-4">Select Cryptocurrency Pairs</h4>
              <p className="text-sm text-gray-600 mb-6">
                Choose one or more trading pairs for {getConnectorName()}
                {exchangeSymbols.length > 0 && (
                  <span className="ml-2 text-green-600">({exchangeSymbols.length} symbols available)</span>
                )}
              </p>

              {/* Symbol Search */}
              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Search Exchange Symbols
                  {loadingSymbols && <span className="text-xs text-gray-500 ml-2">(Loading...)</span>}
                </label>
                <div className="relative">
                  <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
                  <input
                    type="text"
                    placeholder="Search symbols (e.g., BTC, ETH, USDT)..."
                    value={symbolSearch}
                    onChange={(e) => setSymbolSearch(e.target.value)}
                    className="w-full pl-10 pr-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-green-500"
                  />
                </div>
              </div>

              {/* Exchange Symbols */}
              {exchangeSymbols.length > 0 && (symbolSearch || filteredSymbols.length <= 100) && (
                <div className="mb-6">
                  <label className="block text-sm font-medium text-gray-700 mb-3">
                    {symbolSearch ? `Matching Symbols (${filteredSymbols.length})` : 'Exchange Symbols (showing first 100)'}
                  </label>
                  <div className="max-h-48 overflow-y-auto border border-gray-200 rounded-lg p-2">
                    <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-2">
                      {filteredSymbols.map(symbol => (
                        <button
                          key={symbol}
                          onClick={() => togglePair(symbol)}
                          className={`px-3 py-2 text-sm border-2 rounded-lg transition truncate ${
                            selectedPairs.includes(symbol)
                              ? 'border-green-500 bg-green-50 text-green-700 font-medium'
                              : 'border-gray-200 hover:border-green-300 text-gray-700'
                          }`}
                          title={symbol}
                        >
                          {symbol}
                        </button>
                      ))}
                    </div>
                  </div>
                  {filteredSymbols.length === 100 && (
                    <p className="text-xs text-gray-500 mt-2">Use search to find more symbols</p>
                  )}
                </div>
              )}

              {/* Popular Pairs - Dynamic based on exchange */}
              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-3">
                  Popular Pairs on {getConnectorName()} ({selectedPairs.length} selected)
                  {loadingPopular && <span className="text-xs text-gray-500 ml-2">(Loading...)</span>}
                </label>
                {popularPairs.length > 0 ? (
                  <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-2">
                    {popularPairs.map(pair => {
                      const isSelected = selectedPairs.includes(pair)
                      const isValidated = validationResults[pair]
                      const showWarning = validationResults[pair] === false
                      return (
                        <button
                          key={pair}
                          onClick={() => togglePair(pair)}
                          className={`px-3 py-2 text-sm border-2 rounded-lg transition ${
                            isSelected
                              ? showWarning
                                ? 'border-orange-500 bg-orange-50 text-orange-700 font-medium'
                                : 'border-green-500 bg-green-50 text-green-700 font-medium'
                              : 'border-gray-200 hover:border-green-300 text-gray-700'
                          }`}
                          title={pair}
                        >
                          {pair}
                          {showWarning && <ExclamationTriangleIcon className="w-3 h-3 inline ml-1 text-orange-500" />}
                        </button>
                      )
                    })}
                  </div>
                ) : !loadingPopular && (
                  <div className="text-sm text-gray-500 p-4 bg-gray-50 rounded-lg">
                    No popular pairs found. Use the search above to find available symbols.
                  </div>
                )}

                {/* Quick Actions */}
                {popularPairs.length > 0 && (
                  <div className="mt-3 flex space-x-2">
                    <button
                      onClick={() => {
                        // Select top 5 popular pairs
                        const top5 = popularPairs.slice(0, 5)
                        setSelectedPairs(prev => [...new Set([...prev, ...top5])])
                      }}
                      className="text-xs px-3 py-1 bg-blue-100 text-blue-700 rounded hover:bg-blue-200 transition"
                    >
                      + Top 5
                    </button>
                    <button
                      onClick={() => {
                        // Select all USDT pairs
                        const usdtPairs = popularPairs.filter(p => p.endsWith('/USDT'))
                        setSelectedPairs(prev => [...new Set([...prev, ...usdtPairs])])
                      }}
                      className="text-xs px-3 py-1 bg-blue-100 text-blue-700 rounded hover:bg-blue-200 transition"
                    >
                      + All USDT
                    </button>
                  </div>
                )}
              </div>

              {/* Custom Pair */}
              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Add Custom Pair
                </label>
                <div className="flex space-x-2">
                  <input
                    type="text"
                    value={customPair}
                    onChange={(e) => setCustomPair(e.target.value.toUpperCase())}
                    onKeyPress={(e) => e.key === 'Enter' && addCustomPair()}
                    placeholder="e.g., ALGO/USDT"
                    className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-green-500"
                  />
                  <button
                    onClick={addCustomPair}
                    className="px-6 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition"
                  >
                    Add
                  </button>
                </div>
              </div>

              {/* Selected Pairs */}
              {selectedPairs.length > 0 && (
                <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm font-medium text-green-900">
                      Selected Pairs ({selectedPairs.length})
                    </span>
                    <button
                      onClick={() => setSelectedPairs([])}
                      className="text-xs text-green-700 hover:text-green-900"
                    >
                      Clear all
                    </button>
                  </div>
                  <div className="flex flex-wrap gap-2">
                    {selectedPairs.map(pair => (
                      <span
                        key={pair}
                        className="px-3 py-1 bg-green-600 text-white text-sm rounded-full flex items-center"
                      >
                        {pair}
                        <button
                          onClick={() => togglePair(pair)}
                          className="ml-2 hover:text-green-200"
                        >
                          ×
                        </button>
                      </span>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}

          {/* Step 3: Timeframe Selection & Summary */}
          {currentStep === 3 && (
            <div>
              <h4 className="text-lg font-bold text-gray-900 mb-4">Select Timeframes</h4>
              <p className="text-sm text-gray-600 mb-6">
                Choose timeframes to collect data for each pair
                {exchangeMetadata?.timeframes && (
                  <span className="ml-2 text-green-600">
                    (from {exchangeMetadata.name})
                  </span>
                )}
              </p>

              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                {availableTimeframes.map(timeframe => (
                  <button
                    key={timeframe.id}
                    onClick={() => toggleTimeframe(timeframe.id)}
                    className={`p-4 border-2 rounded-lg text-left transition ${
                      selectedTimeframes.includes(timeframe.id)
                        ? 'border-green-500 bg-green-50'
                        : 'border-gray-200 hover:border-green-300'
                    }`}
                  >
                    <div className="font-bold text-gray-900">{timeframe.label}</div>
                    <div className="text-xs text-gray-500 mt-1">{timeframe.description}</div>
                  </button>
                ))}
              </div>

              {selectedTimeframes.length > 0 && (
                <div className="mt-6 bg-green-50 border border-green-200 rounded-lg p-4">
                  <div className="text-sm font-medium text-green-900 mb-2">
                    Selected Timeframes ({selectedTimeframes.length})
                  </div>
                  <div className="flex flex-wrap gap-2">
                    {selectedTimeframes.map(tf => {
                      const timeframe = availableTimeframes.find(t => t.id === tf)
                      return (
                        <span key={tf} className="px-3 py-1 bg-green-600 text-white text-sm rounded-full">
                          {timeframe?.label || tf}
                        </span>
                      )
                    })}
                  </div>
                </div>
              )}

              {/* Historical Data Collection Option */}
              <div className="mt-6 p-4 bg-amber-50 border border-amber-200 rounded-lg">
                <label className="flex items-start cursor-pointer">
                  <input
                    type="checkbox"
                    checked={collectHistorical}
                    onChange={(e) => setCollectHistorical(e.target.checked)}
                    className="mt-1 h-4 w-4 text-amber-600 focus:ring-amber-500 border-gray-300 rounded"
                  />
                  <div className="ml-3">
                    <div className="flex items-center">
                      <ClockIcon className="w-5 h-5 text-amber-600 mr-2" />
                      <span className="font-medium text-amber-900">Collect Historical Data</span>
                    </div>
                    <p className="text-sm text-amber-700 mt-1">
                      When enabled, the job will attempt to fetch as much historical data as possible
                      from the exchange. Each request retrieves up to {exchangeMetadata?.ohlcv_limit || 500} candles.
                    </p>
                  </div>
                </label>
              </div>

              {/* Summary */}
              <div className="mt-6 bg-blue-50 border border-blue-200 rounded-lg p-6">
                <h5 className="font-bold text-blue-900 mb-3">Summary</h5>
                <div className="space-y-2 text-sm text-blue-800">
                  <div><strong>Connector:</strong> {getConnectorName()}</div>
                  <div><strong>Pairs:</strong> {selectedPairs.join(', ')}</div>
                  <div>
                    <strong>Timeframes:</strong>{' '}
                    {selectedTimeframes.map(tf => availableTimeframes.find(t => t.id === tf)?.label || tf).join(', ') || 'None selected'}
                  </div>
                  <div><strong>Total Jobs:</strong> {totalJobs}</div>
                  <div><strong>Historical Data:</strong> {collectHistorical ? 'Yes' : 'No'}</div>
                </div>
                <div className="mt-4 p-3 bg-blue-100 rounded-lg">
                  <p className="text-sm text-blue-800">
                    <strong>Note:</strong> All technical indicators will be automatically calculated for collected data.
                  </p>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="p-6 border-t bg-gray-50">
          <div className="flex justify-between items-center">
            <div>
              {currentStep > 1 && (
                <button
                  onClick={handleBack}
                  className="px-4 py-2 text-gray-700 hover:bg-gray-200 rounded-lg transition flex items-center"
                  disabled={saving}
                >
                  <ArrowLeftIcon className="w-4 h-4 mr-2" />
                  Back
                </button>
              )}
            </div>

            <div className="flex space-x-3">
              {currentStep < 3 && (
                <button
                  onClick={handleNext}
                  className="px-6 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition flex items-center"
                  disabled={saving || validating}
                >
                  {validating ? 'Validating...' : 'Next'}
                  {!validating && <ArrowRightIcon className="w-4 h-4 ml-2" />}
                </button>
              )}

              {currentStep === 3 && (
                <button
                  onClick={handleCreateJobs}
                  className="px-6 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition flex items-center"
                  disabled={saving || totalJobs === 0}
                >
                  <CheckIcon className="w-5 h-5 mr-2" />
                  {saving ? `Creating ${totalJobs} jobs...` : `Create ${totalJobs} Job${totalJobs !== 1 ? 's' : ''}`}
                </button>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default JobWizard
