import { useState } from 'react'
import axios from 'axios'
import {
  ArrowPathIcon,
  PauseIcon,
  PlayIcon,
  TrashIcon,
  PlusIcon,
  PencilIcon
} from '@heroicons/react/24/outline'
import ConnectorWizard from './ConnectorWizard'

const API_BASE = '/api/v1'

function ConnectorList({ connectors, onRefresh, loading }) {
  const [showCreateWizard, setShowCreateWizard] = useState(false)
  const [showEditRateLimit, setShowEditRateLimit] = useState(false)
  const [selectedConnector, setSelectedConnector] = useState(null)
  const [recalculating, setRecalculating] = useState(new Set())
  const [editRateLimit, setEditRateLimit] = useState(1200)
  const [submitting, setSubmitting] = useState(false)

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
    setEditRateLimit(connector.rate_limit?.limit || 1200)
    setShowEditRateLimit(true)
  }

  const saveRateLimit = async () => {
    if (!selectedConnector) return

    setSubmitting(true)
    try {
      await axios.put(`${API_BASE}/connectors/${selectedConnector.id}`, {
        ...selectedConnector,
        rate_limit: {
          ...selectedConnector.rate_limit,
          limit: editRateLimit
        }
      })
      setShowEditRateLimit(false)
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
        <button
          onClick={() => setShowCreateWizard(true)}
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
            <h3 className="text-xl font-bold mb-4">Edit Rate Limit</h3>
            <div className="mb-4">
              <p className="text-sm text-gray-600 mb-2">
                Connector: <span className="font-medium">{selectedConnector.display_name}</span>
              </p>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Rate Limit (requests per minute)
              </label>
              <input
                type="number"
                value={editRateLimit}
                onChange={(e) => setEditRateLimit(parseInt(e.target.value))}
                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                min="1"
                max="10000"
              />
              <p className="text-xs text-gray-500 mt-1">
                Current: {selectedConnector.rate_limit?.limit || 'Not set'}
              </p>
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
