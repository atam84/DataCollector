import { useState } from 'react'
import axios from 'axios'
import {
  Cog6ToothIcon,
  ArrowPathIcon,
  PauseIcon,
  PlayIcon,
  TrashIcon,
  PlusIcon
} from '@heroicons/react/24/outline'
import IndicatorConfig from './IndicatorConfig'

const API_BASE = '/api/v1'

function ConnectorList({ connectors, onRefresh, loading }) {
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showConfigModal, setShowConfigModal] = useState(false)
  const [selectedConnector, setSelectedConnector] = useState(null)
  const [indicatorConfig, setIndicatorConfig] = useState(null)
  const [loadingConfig, setLoadingConfig] = useState(false)
  const [recalculating, setRecalculating] = useState(new Set())
  const [formData, setFormData] = useState({
    exchange_id: 'binance',
    display_name: '',
    sandbox_mode: true,
    rate_limit: {
      limit: 1200,
      period_ms: 60000
    }
  })
  const [submitting, setSubmitting] = useState(false)

  const handleCreate = async (e) => {
    e.preventDefault()
    setSubmitting(true)
    try {
      await axios.post(`${API_BASE}/connectors`, formData)
      setShowCreateModal(false)
      setFormData({
        exchange_id: 'binance',
        display_name: '',
        sandbox_mode: true,
        rate_limit: { limit: 1200, period_ms: 60000 }
      })
      onRefresh()
    } catch (err) {
      alert('Failed to create connector: ' + (err.response?.data?.error || err.message))
    } finally {
      setSubmitting(false)
    }
  }

  const toggleSandboxMode = async (connector) => {
    try {
      await axios.patch(`${API_BASE}/connectors/${connector.id}/sandbox`, {
        sandbox_mode: !connector.sandbox_mode
      })
      onRefresh()
    } catch (err) {
      alert('Failed to toggle sandbox mode: ' + (err.response?.data?.error || err.message))
    }
  }

  const deleteConnector = async (id) => {
    if (!confirm('Are you sure you want to delete this connector?')) return

    try {
      await axios.delete(`${API_BASE}/connectors/${id}`)
      onRefresh()
    } catch (err) {
      alert('Failed to delete connector: ' + (err.response?.data?.error || err.message))
    }
  }

  const suspendConnector = async (id) => {
    try {
      await axios.post(`${API_BASE}/connectors/${id}/suspend`)
      onRefresh()
    } catch (err) {
      alert('Failed to suspend connector: ' + (err.response?.data?.error || err.message))
    }
  }

  const resumeConnector = async (id) => {
    try {
      await axios.post(`${API_BASE}/connectors/${id}/resume`)
      onRefresh()
    } catch (err) {
      alert('Failed to resume connector: ' + (err.response?.data?.error || err.message))
    }
  }

  const calculateRateLimitUsage = (rateLimit) => {
    if (!rateLimit || !rateLimit.limit) return 0
    const available = rateLimit.available_tokens || rateLimit.limit - (rateLimit.usage || 0)
    const used = rateLimit.limit - available
    return (used / rateLimit.limit) * 100
  }

  const openConfigModal = async (connector) => {
    setSelectedConnector(connector)
    setShowConfigModal(true)
    setLoadingConfig(true)

    try {
      const response = await axios.get(`${API_BASE}/connectors/${connector.id}/indicators/config`)
      setIndicatorConfig(response.data.config || {})
    } catch (err) {
      alert('Failed to load indicator config: ' + (err.response?.data?.error || err.message))
      setShowConfigModal(false)
    } finally {
      setLoadingConfig(false)
    }
  }

  const saveIndicatorConfig = async (config) => {
    try {
      await axios.put(`${API_BASE}/connectors/${selectedConnector.id}/indicators/config`, config)
      alert('Indicator configuration saved successfully!')
      setShowConfigModal(false)
    } catch (err) {
      alert('Failed to save indicator config: ' + (err.response?.data?.error || err.message))
    }
  }

  const recalculateConnector = async (connectorId) => {
    if (!confirm('Recalculate indicators for all jobs on this connector? This may take a while.')) return

    setRecalculating(prev => new Set(prev).add(connectorId))
    try {
      await axios.post(`${API_BASE}/connectors/${connectorId}/indicators/recalculate`)
      alert('Indicators recalculated successfully for all jobs!')
    } catch (err) {
      alert('Failed to recalculate: ' + (err.response?.data?.error || err.message))
    } finally {
      setRecalculating(prev => {
        const newSet = new Set(prev)
        newSet.delete(connectorId)
        return newSet
      })
    }
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold text-gray-900">Connectors</h2>
        <button
          onClick={() => setShowCreateModal(true)}
          className="p-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition"
          title="New Connector"
        >
          <PlusIcon className="w-5 h-5" />
        </button>
      </div>

      {/* Connectors Grid */}
      {loading ? (
        <div className="text-center py-12">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
          <p className="mt-2 text-gray-600">Loading connectors...</p>
        </div>
      ) : connectors.length === 0 ? (
        <div className="bg-white rounded-lg shadow p-12 text-center">
          <p className="text-gray-500 mb-4">No connectors configured yet</p>
          <button
            onClick={() => setShowCreateModal(true)}
            className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition"
          >
            Create Your First Connector
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {connectors.map(connector => (
            <div key={connector.id} className="bg-white rounded-lg shadow-md p-6 hover:shadow-lg transition">
              {/* Header */}
              <div className="flex items-start justify-between mb-4">
                <div>
                  <h3 className="text-lg font-semibold text-gray-900">{connector.display_name}</h3>
                  <p className="text-sm text-gray-500">{connector.exchange_id}</p>
                </div>
                <span className={`px-2 py-1 text-xs font-medium rounded ${
                  connector.status === 'active'
                    ? 'bg-green-100 text-green-800'
                    : 'bg-gray-100 text-gray-800'
                }`}>
                  {connector.status}
                </span>
              </div>

              {/* Sandbox Mode Toggle - KEY FEATURE */}
              <div className="mb-4 p-3 bg-gray-50 rounded-lg">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-700">Sandbox Mode</p>
                    <p className="text-xs text-gray-500">
                      {connector.sandbox_mode ? 'Using testnet' : 'Using production'}
                    </p>
                  </div>
                  <button
                    onClick={() => toggleSandboxMode(connector)}
                    className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                      connector.sandbox_mode ? 'bg-yellow-500' : 'bg-green-500'
                    }`}
                  >
                    <span
                      className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                        connector.sandbox_mode ? 'translate-x-6' : 'translate-x-1'
                      }`}
                    />
                  </button>
                </div>
              </div>

              {/* Rate Limit Progress Bar */}
              <div className="mb-4">
                <div className="flex justify-between items-center mb-1">
                  <p className="text-xs font-medium text-gray-700">Rate Limit Usage</p>
                  <p className="text-xs text-gray-600">
                    {connector.rate_limit?.limit - (connector.rate_limit?.usage || 0)} / {connector.rate_limit?.limit || 0}
                  </p>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div
                    className={`h-2 rounded-full transition-all ${
                      calculateRateLimitUsage(connector.rate_limit) > 80
                        ? 'bg-red-500'
                        : calculateRateLimitUsage(connector.rate_limit) > 50
                        ? 'bg-yellow-500'
                        : 'bg-green-500'
                    }`}
                    style={{ width: `${calculateRateLimitUsage(connector.rate_limit)}%` }}
                  />
                </div>
                <p className="text-xs text-gray-500 mt-1">
                  {calculateRateLimitUsage(connector.rate_limit).toFixed(1)}% used
                </p>
              </div>

              {/* Job Count */}
              <div className="mb-4 flex items-center justify-between p-2 bg-blue-50 rounded">
                <span className="text-xs font-medium text-gray-700">Jobs Attached</span>
                <span className="px-2 py-1 text-xs font-bold bg-blue-500 text-white rounded">
                  {connector.job_count || 0}
                </span>
              </div>

              {/* Indicator Actions */}
              <div className="flex space-x-2 mb-2">
                <button
                  onClick={() => openConfigModal(connector)}
                  className="flex-1 py-2 bg-purple-500 text-white text-sm rounded hover:bg-purple-600 transition flex items-center justify-center"
                  title="Configure indicators"
                >
                  <Cog6ToothIcon className="w-5 h-5" />
                </button>
                <button
                  onClick={() => recalculateConnector(connector.id)}
                  className="flex-1 py-2 bg-indigo-500 text-white text-sm rounded hover:bg-indigo-600 transition flex items-center justify-center disabled:opacity-50"
                  disabled={recalculating.has(connector.id)}
                  title="Recalculate all indicators"
                >
                  <ArrowPathIcon className={`w-5 h-5 ${recalculating.has(connector.id) ? 'animate-spin' : ''}`} />
                </button>
              </div>

              {/* Actions */}
              <div className="flex space-x-2">
                {connector.status === 'active' ? (
                  <button
                    onClick={() => suspendConnector(connector.id)}
                    className="flex-1 py-2 bg-yellow-500 text-white text-sm rounded hover:bg-yellow-600 transition flex items-center justify-center"
                    title="Suspend connector"
                  >
                    <PauseIcon className="w-5 h-5" />
                  </button>
                ) : (
                  <button
                    onClick={() => resumeConnector(connector.id)}
                    className="flex-1 py-2 bg-green-500 text-white text-sm rounded hover:bg-green-600 transition flex items-center justify-center"
                    title="Resume connector"
                  >
                    <PlayIcon className="w-5 h-5" />
                  </button>
                )}
                <button
                  onClick={() => deleteConnector(connector.id)}
                  className="flex-1 py-2 bg-red-500 text-white text-sm rounded hover:bg-red-600 transition flex items-center justify-center"
                  title="Delete connector"
                >
                  <TrashIcon className="w-5 h-5" />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h3 className="text-xl font-bold mb-4">Create New Connector</h3>
            <form onSubmit={handleCreate}>
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Exchange
                </label>
                <select
                  value={formData.exchange_id}
                  onChange={(e) => setFormData({ ...formData, exchange_id: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                  required
                >
                  <option value="binance">Binance</option>
                  <option value="bybit">Bybit</option>
                  <option value="coinbase">Coinbase</option>
                  <option value="kraken">Kraken</option>
                  <option value="kucoin">KuCoin</option>
                </select>
              </div>

              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Display Name
                </label>
                <input
                  type="text"
                  value={formData.display_name}
                  onChange={(e) => setFormData({ ...formData, display_name: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="e.g., Binance Production"
                  required
                />
              </div>

              <div className="mb-4">
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={formData.sandbox_mode}
                    onChange={(e) => setFormData({ ...formData, sandbox_mode: e.target.checked })}
                    className="mr-2"
                  />
                  <span className="text-sm font-medium text-gray-700">Enable Sandbox Mode (Testnet)</span>
                </label>
              </div>

              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Rate Limit (requests per minute)
                </label>
                <input
                  type="number"
                  value={formData.rate_limit.limit}
                  onChange={(e) => setFormData({
                    ...formData,
                    rate_limit: { ...formData.rate_limit, limit: parseInt(e.target.value) }
                  })}
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                  min="1"
                  required
                />
              </div>

              <div className="flex space-x-3">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded hover:bg-gray-50 transition"
                  disabled={submitting}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition disabled:opacity-50"
                  disabled={submitting}
                >
                  {submitting ? 'Creating...' : 'Create'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Indicator Config Modal */}
      {showConfigModal && selectedConnector && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col">
            <div className="p-6 border-b">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-xl font-bold text-gray-900">Indicator Configuration</h3>
                  <p className="text-sm text-gray-500 mt-1">
                    {selectedConnector.display_name} ({selectedConnector.exchange_id})
                  </p>
                </div>
                <button
                  onClick={() => setShowConfigModal(false)}
                  className="text-gray-400 hover:text-gray-600 text-2xl"
                >
                  ×
                </button>
              </div>
            </div>

            <div className="flex-1 overflow-y-auto p-6">
              {loadingConfig ? (
                <div className="flex items-center justify-center py-12">
                  <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
                  <p className="ml-3 text-gray-600">Loading configuration...</p>
                </div>
              ) : (
                <IndicatorConfig
                  config={indicatorConfig}
                  onChange={setIndicatorConfig}
                  onSave={saveIndicatorConfig}
                  isJobLevel={false}
                />
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default ConnectorList
