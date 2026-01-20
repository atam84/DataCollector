import { useState } from 'react'
import {
  ArrowLeftIcon,
  ArrowRightIcon,
  CheckIcon
} from '@heroicons/react/24/outline'
import IndicatorConfig from './IndicatorConfig'

const POPULAR_PAIRS = [
  'BTC/USDT', 'ETH/USDT', 'BNB/USDT', 'SOL/USDT', 'XRP/USDT',
  'ADA/USDT', 'DOGE/USDT', 'DOT/USDT', 'MATIC/USDT', 'AVAX/USDT',
  'LINK/USDT', 'UNI/USDT', 'ATOM/USDT', 'LTC/USDT', 'ETC/USDT',
  'BCH/USDT', 'APT/USDT', 'ARB/USDT', 'OP/USDT', 'INJ/USDT'
]

const TIMEFRAMES = [
  { id: '1m', label: '1 Minute', description: 'Very short-term' },
  { id: '5m', label: '5 Minutes', description: 'Short-term' },
  { id: '15m', label: '15 Minutes', description: 'Intraday' },
  { id: '30m', label: '30 Minutes', description: 'Intraday' },
  { id: '1h', label: '1 Hour', description: 'Short-term' },
  { id: '4h', label: '4 Hours', description: 'Medium-term' },
  { id: '1d', label: '1 Day', description: 'Long-term' },
  { id: '1w', label: '1 Week', description: 'Very long-term' }
]

function JobWizard({ connectors, onClose, onSave }) {
  const [currentStep, setCurrentStep] = useState(1)
  const [selectedConnector, setSelectedConnector] = useState('')
  const [selectedPairs, setSelectedPairs] = useState([])
  const [customPair, setCustomPair] = useState('')
  const [selectedTimeframes, setSelectedTimeframes] = useState([])
  const [indicatorConfig, setIndicatorConfig] = useState({})
  const [useConnectorIndicators, setUseConnectorIndicators] = useState(true)
  const [saving, setSaving] = useState(false)

  const handleNext = () => {
    if (currentStep === 1 && !selectedConnector) {
      alert('Please select a connector')
      return
    }
    if (currentStep === 2 && selectedPairs.length === 0) {
      alert('Please select at least one cryptocurrency pair')
      return
    }
    if (currentStep === 3 && selectedTimeframes.length === 0) {
      alert('Please select at least one timeframe')
      return
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

  const handleCreateJobs = async () => {
    setSaving(true)
    try {
      // Calculate total jobs to create
      const totalJobs = selectedPairs.length * selectedTimeframes.length

      // Prepare jobs data
      const jobs = []
      for (const pair of selectedPairs) {
        for (const timeframe of selectedTimeframes) {
          jobs.push({
            connector_exchange_id: selectedConnector,
            symbol: pair,
            timeframe: timeframe,
            status: 'active',
            indicator_config: useConnectorIndicators ? null : indicatorConfig
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
              <p className="text-green-100 mt-1">Step {currentStep} of 4</p>
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
            <div className={`flex-1 h-2 rounded ${currentStep >= 4 ? 'bg-white' : 'bg-green-300'}`} />
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
                    onClick={() => setSelectedConnector(connector.exchange_id)}
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
                      <div>Mode: {connector.sandbox_mode ? 'Sandbox' : 'Production'}</div>
                      <div>Jobs: {connector.job_count || 0}</div>
                    </div>
                  </button>
                ))}
              </div>

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
              </p>

              {/* Popular Pairs */}
              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-3">
                  Popular Pairs ({selectedPairs.length} selected)
                </label>
                <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-2">
                  {POPULAR_PAIRS.map(pair => (
                    <button
                      key={pair}
                      onClick={() => togglePair(pair)}
                      className={`px-3 py-2 text-sm border-2 rounded-lg transition ${
                        selectedPairs.includes(pair)
                          ? 'border-green-500 bg-green-50 text-green-700 font-medium'
                          : 'border-gray-200 hover:border-green-300 text-gray-700'
                      }`}
                    >
                      {pair}
                    </button>
                  ))}
                </div>
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

          {/* Step 3: Timeframe Selection */}
          {currentStep === 3 && (
            <div>
              <h4 className="text-lg font-bold text-gray-900 mb-4">Select Timeframes</h4>
              <p className="text-sm text-gray-600 mb-6">
                Choose timeframes to collect data for each pair
              </p>

              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                {TIMEFRAMES.map(timeframe => (
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
                      const timeframe = TIMEFRAMES.find(t => t.id === tf)
                      return (
                        <span key={tf} className="px-3 py-1 bg-green-600 text-white text-sm rounded-full">
                          {timeframe.label}
                        </span>
                      )
                    })}
                  </div>
                </div>
              )}
            </div>
          )}

          {/* Step 4: Indicator Configuration */}
          {currentStep === 4 && (
            <div>
              <h4 className="text-lg font-bold text-gray-900 mb-4">Indicator Configuration</h4>
              <p className="text-sm text-gray-600 mb-6">
                Configure indicators for all jobs or inherit from connector
              </p>

              {/* Option: Use Connector Settings */}
              <div className="mb-6">
                <label className="flex items-center space-x-3 p-4 border-2 border-gray-200 rounded-lg hover:bg-gray-50 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={useConnectorIndicators}
                    onChange={(e) => setUseConnectorIndicators(e.target.checked)}
                    className="w-5 h-5 text-green-600 rounded focus:ring-green-500"
                  />
                  <div>
                    <span className="font-medium text-gray-900">Use Connector's Indicator Configuration</span>
                    <p className="text-sm text-gray-500">
                      Jobs will inherit indicator settings from {getConnectorName()}
                    </p>
                  </div>
                </label>
              </div>

              {/* Custom Indicators */}
              {!useConnectorIndicators && (
                <div>
                  <IndicatorConfig
                    config={indicatorConfig}
                    onChange={setIndicatorConfig}
                    isJobLevel={true}
                  />
                </div>
              )}

              {/* Summary */}
              <div className="mt-6 bg-blue-50 border border-blue-200 rounded-lg p-6">
                <h5 className="font-bold text-blue-900 mb-3">Summary</h5>
                <div className="space-y-2 text-sm text-blue-800">
                  <div><strong>Connector:</strong> {getConnectorName()}</div>
                  <div><strong>Pairs:</strong> {selectedPairs.join(', ')}</div>
                  <div><strong>Timeframes:</strong> {selectedTimeframes.map(tf => TIMEFRAMES.find(t => t.id === tf)?.label).join(', ')}</div>
                  <div><strong>Total Jobs:</strong> {totalJobs}</div>
                  <div><strong>Indicators:</strong> {useConnectorIndicators ? 'Inherit from connector' : 'Custom configuration'}</div>
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
              {currentStep < 4 && (
                <button
                  onClick={handleNext}
                  className="px-6 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition flex items-center"
                  disabled={saving}
                >
                  Next
                  <ArrowRightIcon className="w-4 h-4 ml-2" />
                </button>
              )}

              {currentStep === 4 && (
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
