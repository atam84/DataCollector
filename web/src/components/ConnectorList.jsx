import { useState, useEffect, useCallback } from 'react'
import axios from 'axios'
import {
  ArrowPathIcon,
  PauseIcon,
  PlayIcon,
  TrashIcon,
  PlusIcon,
  PencilIcon,
  ClockIcon,
  BoltIcon,
  ExclamationTriangleIcon
} from '@heroicons/react/24/outline'
import ConnectorWizard from './ConnectorWizard'

const API_BASE = '/api/v1'

function ConnectorList({ connectors, onRefresh, loading }) {
  const [showCreateWizard, setShowCreateWizard] = useState(false)
  const [showEditRateLimit, setShowEditRateLimit] = useState(false)
  const [showRateLimitDetails, setShowRateLimitDetails] = useState(null)
  const [selectedConnector, setSelectedConnector] = useState(null)
  const [recalculating, setRecalculating] = useState(new Set())
  const [editRateLimit, setEditRateLimit] = useState(1200)
  const [editMinDelay, setEditMinDelay] = useState(3000)
  const [submitting, setSubmitting] = useState(false)
  const [rateLimitStatus, setRateLimitStatus] = useState({})
  const [loadingRateLimit, setLoadingRateLimit] = useState(new Set())

  // Fetch rate limit status for all connectors
  const fetchRateLimitStatus = useCallback(async (connectorId) => {
    setLoadingRateLimit(prev => new Set(prev).add(connectorId))
    try {
      const response = await axios.get(`${API_BASE}/connectors/${connectorId}/rate-limit`)
      setRateLimitStatus(prev => ({
        ...prev,
        [connectorId]: response.data
      }))
    } catch (err) {
      console.error('Failed to fetch rate limit status:', err)
    } finally {
      setLoadingRateLimit(prev => {
        const newSet = new Set(prev)
        newSet.delete(connectorId)
        return newSet
      })
    }
  }, [])

  // Fetch rate limit status for all connectors on mount and periodically
  useEffect(() => {
    if (connectors.length === 0) return

    // Initial fetch
    connectors.forEach(c => fetchRateLimitStatus(c.id))

    // Refresh every 10 seconds
    const interval = setInterval(() => {
      connectors.forEach(c => fetchRateLimitStatus(c.id))
    }, 10000)

    return () => clearInterval(interval)
  }, [connectors, fetchRateLimitStatus])

  // Reset rate limit usage
  const resetRateLimitUsage = async (connectorId) => {
    try {
      await axios.post(`${API_BASE}/connectors/${connectorId}/rate-limit/reset`)
      fetchRateLimitStatus(connectorId)
      onRefresh()
    } catch (err) {
      alert('Failed to reset rate limit: ' + (err.response?.data?.error || err.message))
    }
  }

  const handleWizardSave = async (connectorData) => {
    try {
      await axios.post(`${API_BASE}/connectors`, connectorData)
      onRefresh()
    } catch (err) {
      throw new Error(err.response?.data?.error || err.message)
    }
  }

  const openEditRateLimit = (connector) => {
    setSelectedConnector(connector)
    setEditRateLimit(connector.rate_limit?.limit || 20)
    setEditMinDelay(connector.rate_limit?.min_delay_ms || 3000)
    setShowEditRateLimit(true)
  }

  const saveRateLimit = async () => {
    if (!selectedConnector) return

    setSubmitting(true)
    try {
      await axios.put(`${API_BASE}/connectors/${selectedConnector.id}`, {
        rate_limit: {
          limit: editRateLimit,
          min_delay_ms: editMinDelay
        }
      })
      setShowEditRateLimit(false)
      fetchRateLimitStatus(selectedConnector.id)
      onRefresh()
    } catch (err) {
      alert('Failed to update rate limit: ' + (err.response?.data?.error || err.message))
    } finally {
      setSubmitting(false)
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
        <div className="flex space-x-2">
          <button
            onClick={onRefresh}
            className="p-2 bg-gray-100 text-gray-700 rounded hover:bg-gray-200 transition"
            title="Refresh"
            disabled={loading}
          >
            <ArrowPathIcon className={`w-5 h-5 ${loading ? 'animate-spin' : ''}`} />
          </button>
          <button
            onClick={() => setShowCreateWizard(true)}
            className="p-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition"
            title="New Connector"
          >
            <PlusIcon className="w-5 h-5" />
          </button>
        </div>
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
            onClick={() => setShowCreateWizard(true)}
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

              {/* Rate Limit Section */}
              <div className="mb-4 p-3 bg-gray-50 rounded-lg">
                <div className="flex justify-between items-center mb-2">
                  <p className="text-xs font-medium text-gray-700 flex items-center">
                    <BoltIcon className="w-4 h-4 mr-1" />
                    Rate Limit
                  </p>
                  <button
                    onClick={() => setShowRateLimitDetails(showRateLimitDetails === connector.id ? null : connector.id)}
                    className="text-xs text-blue-600 hover:text-blue-800"
                  >
                    {showRateLimitDetails === connector.id ? 'Hide' : 'Details'}
                  </button>
                </div>

                {/* Usage Bar */}
                <div className="mb-2">
                  <div className="flex justify-between items-center mb-1">
                    <p className="text-xs text-gray-600">
                      {rateLimitStatus[connector.id]?.usage || connector.rate_limit?.usage || 0} / {connector.rate_limit?.limit || 0} calls
                    </p>
                    <p className="text-xs text-gray-500">
                      {((rateLimitStatus[connector.id]?.period_remaining_ms || 0) / 1000).toFixed(0)}s left
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
                </div>

                {/* Status Indicator */}
                <div className="flex items-center justify-between">
                  {rateLimitStatus[connector.id]?.can_call_now ? (
                    <span className="text-xs text-green-600 flex items-center">
                      <span className="w-2 h-2 bg-green-500 rounded-full mr-1 animate-pulse"></span>
                      Ready
                    </span>
                  ) : (
                    <span className="text-xs text-orange-600 flex items-center">
                      <ClockIcon className="w-3 h-3 mr-1" />
                      Cooling down
                    </span>
                  )}
                  <span className="text-xs text-gray-500">
                    Min: {(connector.rate_limit?.min_delay_ms || 3000) / 1000}s delay
                  </span>
                </div>

                {/* Expanded Details */}
                {showRateLimitDetails === connector.id && (
                  <div className="mt-3 pt-3 border-t border-gray-200 space-y-2">
                    <div className="flex justify-between text-xs">
                      <span className="text-gray-600">Period:</span>
                      <span className="font-medium">{(connector.rate_limit?.period_ms || 60000) / 1000}s</span>
                    </div>
                    <div className="flex justify-between text-xs">
                      <span className="text-gray-600">Min Delay:</span>
                      <span className="font-medium">{connector.rate_limit?.min_delay_ms || 3000}ms</span>
                    </div>
                    {rateLimitStatus[connector.id]?.last_api_call_at && (
                      <div className="flex justify-between text-xs">
                        <span className="text-gray-600">Last Call:</span>
                        <span className="font-medium">
                          {new Date(rateLimitStatus[connector.id].last_api_call_at).toLocaleTimeString()}
                        </span>
                      </div>
                    )}
                    <button
                      onClick={() => resetRateLimitUsage(connector.id)}
                      className="w-full mt-2 px-2 py-1 text-xs bg-blue-100 text-blue-700 rounded hover:bg-blue-200 transition"
                    >
                      Reset Usage Counter
                    </button>
                  </div>
                )}
              </div>

              {/* Job Count */}
              <div className="mb-4 flex items-center justify-between p-2 bg-blue-50 rounded">
                <span className="text-xs font-medium text-gray-700">Jobs Attached</span>
                <span className="px-2 py-1 text-xs font-bold bg-blue-500 text-white rounded">
                  {connector.job_count || 0}
                </span>
              </div>

              {/* Actions Row 1 */}
              <div className="flex space-x-2 mb-2">
                <button
                  onClick={() => recalculateConnector(connector.id)}
                  className="flex-1 py-2 bg-indigo-500 text-white text-sm rounded hover:bg-indigo-600 transition flex items-center justify-center disabled:opacity-50"
                  disabled={recalculating.has(connector.id)}
                  title="Recalculate all indicators"
                >
                  <ArrowPathIcon className={`w-5 h-5 ${recalculating.has(connector.id) ? 'animate-spin' : ''}`} />
                </button>
                <button
                  onClick={() => openEditRateLimit(connector)}
                  className="flex-1 py-2 bg-cyan-500 text-white text-sm rounded hover:bg-cyan-600 transition flex items-center justify-center"
                  title="Edit rate limit"
                >
                  <PencilIcon className="w-5 h-5" />
                </button>
              </div>

              {/* Actions Row 2 */}
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

      {/* Connector Creation Wizard */}
      {showCreateWizard && (
        <ConnectorWizard
          onClose={() => setShowCreateWizard(false)}
          onSave={handleWizardSave}
        />
      )}

      {/* Edit Rate Limit Modal */}
      {showEditRateLimit && selectedConnector && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h3 className="text-xl font-bold mb-4">Edit Rate Limit Settings</h3>
            <p className="text-sm text-gray-600 mb-4">
              Connector: <span className="font-medium">{selectedConnector.display_name}</span>
              <span className="text-gray-400 ml-2">({selectedConnector.exchange_id})</span>
            </p>

            {/* Rate Limit */}
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Rate Limit (requests per period)
              </label>
              <input
                type="number"
                value={editRateLimit}
                onChange={(e) => setEditRateLimit(parseInt(e.target.value) || 0)}
                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                min="1"
                max="1000"
              />
              <p className="text-xs text-gray-500 mt-1">
                Current: {selectedConnector.rate_limit?.limit || 'Not set'} requests per {(selectedConnector.rate_limit?.period_ms || 60000) / 1000}s
              </p>
            </div>

            {/* Min Delay */}
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Minimum Delay Between Calls (ms)
              </label>
              <input
                type="number"
                value={editMinDelay}
                onChange={(e) => setEditMinDelay(parseInt(e.target.value) || 0)}
                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                min="1000"
                max="60000"
                step="100"
              />
              <p className="text-xs text-gray-500 mt-1">
                Current: {selectedConnector.rate_limit?.min_delay_ms || 3000}ms ({((selectedConnector.rate_limit?.min_delay_ms || 3000) / 1000).toFixed(1)}s)
              </p>
            </div>

            {/* Recommended Settings */}
            <div className="mb-4 p-3 bg-amber-50 border border-amber-200 rounded-lg">
              <div className="flex items-start">
                <ExclamationTriangleIcon className="w-5 h-5 text-amber-600 mr-2 flex-shrink-0 mt-0.5" />
                <div className="text-xs text-amber-800">
                  <p className="font-medium mb-1">Recommended for most exchanges:</p>
                  <ul className="list-disc list-inside space-y-1">
                    <li>Rate limit: 20 requests/minute</li>
                    <li>Min delay: 3000-5000ms (3-5 seconds)</li>
                  </ul>
                </div>
              </div>
            </div>

            <div className="flex space-x-3">
              <button
                onClick={() => setShowEditRateLimit(false)}
                className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded hover:bg-gray-50 transition"
                disabled={submitting}
              >
                Cancel
              </button>
              <button
                onClick={saveRateLimit}
                className="flex-1 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition disabled:opacity-50"
                disabled={submitting}
              >
                {submitting ? 'Saving...' : 'Save'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default ConnectorList
