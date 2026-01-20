import { useState, useEffect } from 'react'
import axios from 'axios'
import {
  ArrowLeftIcon,
  ArrowRightIcon,
  CheckIcon
} from '@heroicons/react/24/outline'
import IndicatorConfig from './IndicatorConfig'

const API_BASE = '/api/v1'

// Default rate limits for known exchanges
const DEFAULT_RATE_LIMITS = {
  'binance': 1200,
  'bybit': 600,
  'coinbase': 300,
  'kraken': 300,
  'kucoin': 600,
  'okx': 600,
  'gateio': 400,
  'gate': 400,
  'huobi': 500
}

// Exchange display names
const EXCHANGE_NAMES = {
  'binance': 'Binance',
  'bybit': 'Bybit',
  'coinbase': 'Coinbase',
  'kraken': 'Kraken',
  'kucoin': 'KuCoin',
  'okx': 'OKX',
  'gateio': 'Gate.io',
  'gate': 'Gate.io',
  'huobi': 'Huobi'
}

function ConnectorWizard({ onClose, onSave }) {
  const [currentStep, setCurrentStep] = useState(1)
  const [supportedExchanges, setSupportedExchanges] = useState([])
  const [loadingExchanges, setLoadingExchanges] = useState(true)
  const [formData, setFormData] = useState({
    exchange_id: '',
    display_name: '',
    sandbox_mode: true,
    rate_limit: {
      limit: 1200,
      period_ms: 60000
    }
  })
  const [indicatorConfig, setIndicatorConfig] = useState({})
  const [saving, setSaving] = useState(false)

  // Fetch supported exchanges from API
  useEffect(() => {
    const fetchSupportedExchanges = async () => {
      try {
        const response = await axios.get(`${API_BASE}/exchanges`)
        const exchanges = response.data.exchanges || []

        // Map to our format
        const mapped = exchanges.map(id => ({
          id: id,
          name: EXCHANGE_NAMES[id] || id.charAt(0).toUpperCase() + id.slice(1),
          defaultRateLimit: DEFAULT_RATE_LIMITS[id] || 600
        }))

        setSupportedExchanges(mapped)
      } catch (err) {
        console.error('Failed to fetch supported exchanges:', err)
        // Fallback to common exchanges if API fails
        setSupportedExchanges([
          { id: 'binance', name: 'Binance', defaultRateLimit: 1200 },
          { id: 'bybit', name: 'Bybit', defaultRateLimit: 600 },
          { id: 'coinbase', name: 'Coinbase', defaultRateLimit: 300 },
          { id: 'kraken', name: 'Kraken', defaultRateLimit: 300 }
        ])
      } finally {
        setLoadingExchanges(false)
      }
    }

    fetchSupportedExchanges()
  }, [])

  const handleExchangeChange = (exchangeId) => {
    const exchange = supportedExchanges.find(e => e.id === exchangeId)
    setFormData({
      ...formData,
      exchange_id: exchangeId,
      display_name: exchange ? exchange.name : exchangeId,
      rate_limit: {
        ...formData.rate_limit,
        limit: exchange ? exchange.defaultRateLimit : 1200
      }
    })
  }

  const handleNext = () => {
    if (currentStep === 1) {
      // Validate step 1
      if (!formData.exchange_id || !formData.display_name) {
        alert('Please select an exchange and provide a display name')
        return
      }
    }
    setCurrentStep(currentStep + 1)
  }

  const handleBack = () => {
    setCurrentStep(currentStep - 1)
  }

  const handleSaveAndClose = async () => {
    setSaving(true)
    try {
      await onSave(formData, null) // Save without indicators
      onClose()
    } catch (err) {
      alert('Failed to save connector: ' + err.message)
    } finally {
      setSaving(false)
    }
  }

  const handleSaveWithIndicators = async () => {
    setSaving(true)
    try {
      await onSave(formData, indicatorConfig)
      onClose()
    } catch (err) {
      alert('Failed to save connector: ' + err.message)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg w-full max-w-5xl max-h-[95vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="p-6 border-b bg-gradient-to-r from-blue-500 to-indigo-600">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-2xl font-bold text-white">Create New Connector</h3>
              <p className="text-blue-100 mt-1">Step {currentStep} of 2</p>
            </div>
            <button
              onClick={onClose}
              className="text-white hover:text-blue-100 text-2xl"
              disabled={saving}
            >
              ×
            </button>
          </div>

          {/* Progress Bar */}
          <div className="mt-4 flex items-center space-x-2">
            <div className={`flex-1 h-2 rounded ${currentStep >= 1 ? 'bg-white' : 'bg-blue-300'}`} />
            <div className={`flex-1 h-2 rounded ${currentStep >= 2 ? 'bg-white' : 'bg-blue-300'}`} />
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          {/* Step 1: Exchange & Rate Limit */}
          {currentStep === 1 && (
            <div className="space-y-6">
              <div>
                <h4 className="text-lg font-bold text-gray-900 mb-4">Exchange Configuration</h4>

                {/* Exchange Selection */}
                <div className="mb-6">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Select Exchange * {loadingExchanges && <span className="text-xs text-gray-500">(Loading...)</span>}
                  </label>
                  {supportedExchanges.length === 0 && !loadingExchanges && (
                    <div className="text-sm text-red-600 mb-2">
                      No exchanges available. Please check your CCXT installation.
                    </div>
                  )}
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
                    {supportedExchanges.map(exchange => (
                      <button
                        key={exchange.id}
                        onClick={() => handleExchangeChange(exchange.id)}
                        className={`p-4 border-2 rounded-lg text-center transition ${
                          formData.exchange_id === exchange.id
                            ? 'border-blue-500 bg-blue-50 text-blue-700'
                            : 'border-gray-200 hover:border-blue-300 text-gray-700'
                        }`}
                      >
                        <div className="font-semibold">{exchange.name}</div>
                        <div className="text-xs text-gray-500 mt-1">
                          {exchange.defaultRateLimit} req/min
                        </div>
                      </button>
                    ))}
                  </div>
                </div>

                {/* Display Name */}
                <div className="mb-6">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Display Name *
                  </label>
                  <input
                    type="text"
                    value={formData.display_name}
                    onChange={(e) => setFormData({ ...formData, display_name: e.target.value })}
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="e.g., Binance Production"
                    required
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    A friendly name to identify this connector
                  </p>
                </div>

                {/* Rate Limit */}
                <div className="mb-6">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Rate Limit (requests per minute) *
                  </label>
                  <input
                    type="number"
                    value={formData.rate_limit.limit}
                    onChange={(e) => setFormData({
                      ...formData,
                      rate_limit: { ...formData.rate_limit, limit: parseInt(e.target.value) }
                    })}
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    min="1"
                    max="10000"
                    required
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    Maximum API requests allowed per minute. Default varies by exchange.
                  </p>
                </div>

                {/* Sandbox Mode */}
                <div className="mb-6">
                  <label className="flex items-center space-x-3 p-4 border border-gray-200 rounded-lg hover:bg-gray-50 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={formData.sandbox_mode}
                      onChange={(e) => setFormData({ ...formData, sandbox_mode: e.target.checked })}
                      className="w-5 h-5 text-blue-600 rounded focus:ring-blue-500"
                    />
                    <div>
                      <span className="font-medium text-gray-900">Enable Sandbox Mode (Testnet)</span>
                      <p className="text-sm text-gray-500">
                        Use test environment for development without real funds
                      </p>
                    </div>
                  </label>
                </div>
              </div>
            </div>
          )}

          {/* Step 2: Indicators */}
          {currentStep === 2 && (
            <div>
              <div className="mb-4">
                <h4 className="text-lg font-bold text-gray-900">Indicator Configuration</h4>
                <p className="text-sm text-gray-600 mt-1">
                  Configure default indicators for this connector. Jobs can override these settings.
                </p>
              </div>

              <IndicatorConfig
                config={indicatorConfig}
                onChange={setIndicatorConfig}
                isJobLevel={false}
              />
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
              {currentStep === 1 && (
                <>
                  <button
                    onClick={handleSaveAndClose}
                    className="px-6 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-100 transition"
                    disabled={saving}
                  >
                    Save & Close
                  </button>
                  <button
                    onClick={handleNext}
                    className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition flex items-center"
                    disabled={saving}
                  >
                    Next: Configure Indicators
                    <ArrowRightIcon className="w-4 h-4 ml-2" />
                  </button>
                </>
              )}

              {currentStep === 2 && (
                <button
                  onClick={handleSaveWithIndicators}
                  className="px-6 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition flex items-center"
                  disabled={saving}
                >
                  <CheckIcon className="w-5 h-5 mr-2" />
                  {saving ? 'Creating...' : 'Create Connector'}
                </button>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default ConnectorWizard
