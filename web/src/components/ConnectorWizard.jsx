import { useState, useEffect, useMemo } from 'react'
import axios from 'axios'
import { CheckIcon, MagnifyingGlassIcon } from '@heroicons/react/24/outline'

const API_BASE = '/api/v1'

function ConnectorWizard({ onClose, onSave }) {
  const [exchangesMetadata, setExchangesMetadata] = useState([])
  const [loadingExchanges, setLoadingExchanges] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [formData, setFormData] = useState({
    exchange_id: '',
    display_name: '',
    rate_limit: {
      limit: 1200,
      period_ms: 60000
    }
  })
  const [selectedExchange, setSelectedExchange] = useState(null)
  const [saving, setSaving] = useState(false)

  // Fetch exchange metadata from API
  useEffect(() => {
    const fetchExchangesMetadata = async () => {
      try {
        const response = await axios.get(`${API_BASE}/exchanges/metadata`)
        const exchanges = response.data.exchanges || []
        setExchangesMetadata(exchanges)
      } catch (err) {
        console.error('Failed to fetch exchange metadata:', err)
        // Fallback to basic list if API fails
        try {
          const fallbackResponse = await axios.get(`${API_BASE}/exchanges`)
          const exchanges = fallbackResponse.data.exchanges || []
          setExchangesMetadata(exchanges.map(id => ({
            id: id,
            name: id.charAt(0).toUpperCase() + id.slice(1),
            rate_limit: 600,
            has_ohlcv: true,
            timeframes: {}
          })))
        } catch (e) {
          console.error('Fallback also failed:', e)
          setExchangesMetadata([])
        }
      } finally {
        setLoadingExchanges(false)
      }
    }

    fetchExchangesMetadata()
  }, [])

  // Filter exchanges based on search term
  const filteredExchanges = useMemo(() => {
    if (!searchTerm) return exchangesMetadata
    const term = searchTerm.toLowerCase()
    return exchangesMetadata.filter(exchange =>
      exchange.id.toLowerCase().includes(term) ||
      exchange.name.toLowerCase().includes(term) ||
      (exchange.countries && exchange.countries.some(c => c.toLowerCase().includes(term)))
    )
  }, [exchangesMetadata, searchTerm])

  const handleExchangeChange = (exchange) => {
    setSelectedExchange(exchange)
    setFormData({
      ...formData,
      exchange_id: exchange.id,
      display_name: exchange.name || exchange.id,
      rate_limit: {
        ...formData.rate_limit,
        limit: exchange.rate_limit || 600
      }
    })
  }

  const handleSave = async () => {
    // Validate form
    if (!formData.exchange_id || !formData.display_name) {
      alert('Please select an exchange and provide a display name')
      return
    }

    setSaving(true)
    try {
      await onSave(formData)
      onClose()
    } catch (err) {
      alert('Failed to save connector: ' + err.message)
    } finally {
      setSaving(false)
    }
  }

  const formatTimeframes = (timeframes) => {
    if (!timeframes || Object.keys(timeframes).length === 0) return 'N/A'
    const keys = Object.keys(timeframes)
    if (keys.length <= 5) return keys.join(', ')
    return `${keys.slice(0, 5).join(', ')} +${keys.length - 5} more`
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg w-full max-w-4xl max-h-[95vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="p-6 border-b bg-gradient-to-r from-blue-500 to-indigo-600">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-2xl font-bold text-white">Create New Connector</h3>
              <p className="text-blue-100 mt-1">Configure exchange connection</p>
            </div>
            <button
              onClick={onClose}
              className="text-white hover:text-blue-100 text-2xl"
              disabled={saving}
            >
              ×
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          <div className="space-y-6">
            {/* Search Bar */}
            <div className="relative">
              <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
              <input
                type="text"
                placeholder="Search exchanges by name or country..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-10 pr-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            {/* Exchange Selection */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Select Exchange * {loadingExchanges && <span className="text-xs text-gray-500">(Loading metadata...)</span>}
              </label>
              {filteredExchanges.length === 0 && !loadingExchanges && (
                <div className="text-sm text-gray-500 mb-2 p-4 bg-gray-50 rounded-lg text-center">
                  {searchTerm ? 'No exchanges match your search' : 'No exchanges available'}
                </div>
              )}
              <div className="max-h-64 overflow-y-auto border border-gray-200 rounded-lg">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-2 p-2">
                  {filteredExchanges.map(exchange => (
                    <button
                      key={exchange.id}
                      onClick={() => handleExchangeChange(exchange)}
                      className={`p-3 border-2 rounded-lg text-left transition ${
                        formData.exchange_id === exchange.id
                          ? 'border-blue-500 bg-blue-50'
                          : 'border-gray-200 hover:border-blue-300'
                      }`}
                    >
                      <div className="flex items-center justify-between">
                        <div className="font-semibold text-gray-900">{exchange.name}</div>
                        {exchange.has_ohlcv && (
                          <span className="text-xs bg-green-100 text-green-700 px-2 py-0.5 rounded">OHLCV</span>
                        )}
                      </div>
                      <div className="text-xs text-gray-500 mt-1">
                        <span className="mr-3">Rate: {exchange.rate_limit || 600}/min</span>
                        {exchange.ohlcv_limit && <span>Limit: {exchange.ohlcv_limit} candles</span>}
                      </div>
                      {exchange.countries && exchange.countries.length > 0 && (
                        <div className="text-xs text-gray-400 mt-1">
                          {exchange.countries.slice(0, 3).join(', ')}
                          {exchange.countries.length > 3 && ` +${exchange.countries.length - 3}`}
                        </div>
                      )}
                    </button>
                  ))}
                </div>
              </div>
            </div>

            {/* Selected Exchange Details */}
            {selectedExchange && (
              <div className="p-4 bg-blue-50 border border-blue-200 rounded-lg">
                <h5 className="font-semibold text-blue-900 mb-2">{selectedExchange.name} Details</h5>
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span className="text-blue-700 font-medium">Rate Limit:</span>
                    <span className="ml-2 text-blue-900">{selectedExchange.rate_limit || 600} req/min</span>
                  </div>
                  <div>
                    <span className="text-blue-700 font-medium">OHLCV Limit:</span>
                    <span className="ml-2 text-blue-900">{selectedExchange.ohlcv_limit || 500} candles/request</span>
                  </div>
                  <div className="col-span-2">
                    <span className="text-blue-700 font-medium">Timeframes:</span>
                    <span className="ml-2 text-blue-900">{formatTimeframes(selectedExchange.timeframes)}</span>
                  </div>
                </div>
              </div>
            )}

            {/* Display Name */}
            <div>
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
            <div>
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
                Auto-populated from exchange metadata. Adjust if needed.
              </p>
            </div>

            {/* Info about indicators */}
            <div className="p-4 bg-green-50 border border-green-200 rounded-lg">
              <p className="text-sm text-green-800">
                <strong>Note:</strong> All technical indicators will be automatically calculated for collected data.
              </p>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="p-6 border-t bg-gray-50">
          <div className="flex justify-end space-x-3">
            <button
              onClick={onClose}
              className="px-6 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-100 transition"
              disabled={saving}
            >
              Cancel
            </button>
            <button
              onClick={handleSave}
              className="px-6 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition flex items-center"
              disabled={saving || !formData.exchange_id}
            >
              <CheckIcon className="w-5 h-5 mr-2" />
              {saving ? 'Creating...' : 'Create Connector'}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

export default ConnectorWizard
