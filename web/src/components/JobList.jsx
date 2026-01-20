import { useState } from 'react'
import axios from 'axios'

const API_BASE = '/api/v1'

function JobList({ jobs, connectors, onRefresh, loading }) {
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [formData, setFormData] = useState({
    connector_exchange_id: '',
    symbol: 'BTC/USDT',
    timeframe: '1h',
    status: 'active'
  })
  const [submitting, setSubmitting] = useState(false)
  const [executingJobs, setExecutingJobs] = useState(new Set())

  const handleCreate = async (e) => {
    e.preventDefault()
    setSubmitting(true)
    try {
      await axios.post(`${API_BASE}/jobs`, formData)
      setShowCreateModal(false)
      setFormData({
        connector_exchange_id: '',
        symbol: 'BTC/USDT',
        timeframe: '1h',
        status: 'active'
      })
      onRefresh()
    } catch (err) {
      alert('Failed to create job: ' + (err.response?.data?.error || err.message))
    } finally {
      setSubmitting(false)
    }
  }

  const pauseJob = async (id) => {
    try {
      await axios.post(`${API_BASE}/jobs/${id}/pause`)
      onRefresh()
    } catch (err) {
      alert('Failed to pause job: ' + (err.response?.data?.error || err.message))
    }
  }

  const resumeJob = async (id) => {
    try {
      await axios.post(`${API_BASE}/jobs/${id}/resume`)
      onRefresh()
    } catch (err) {
      alert('Failed to resume job: ' + (err.response?.data?.error || err.message))
    }
  }

  const deleteJob = async (id) => {
    if (!confirm('Are you sure you want to delete this job?')) return

    try {
      await axios.delete(`${API_BASE}/jobs/${id}`)
      onRefresh()
    } catch (err) {
      alert('Failed to delete job: ' + (err.response?.data?.error || err.message))
    }
  }

  const executeJob = async (id) => {
    setExecutingJobs(prev => new Set(prev).add(id))
    try {
      const response = await axios.post(`${API_BASE}/jobs/${id}/execute`)
      const result = response.data

      if (result.success) {
        alert(`Job executed successfully!\nRecords fetched: ${result.records_fetched}\nExecution time: ${result.execution_time_ms}ms`)
      } else {
        alert(`Job execution failed: ${result.error || result.message}`)
      }

      onRefresh()
    } catch (err) {
      alert('Failed to execute job: ' + (err.response?.data?.error || err.message))
    } finally {
      setExecutingJobs(prev => {
        const newSet = new Set(prev)
        newSet.delete(id)
        return newSet
      })
    }
  }

  const getConnectorName = (exchangeId) => {
    const connector = connectors.find(c => c.exchange_id === exchangeId)
    return connector ? connector.display_name : exchangeId
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold text-gray-900">Jobs</h2>
        <button
          onClick={() => setShowCreateModal(true)}
          className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition"
          disabled={connectors.length === 0}
        >
          + New Job
        </button>
      </div>

      {/* Jobs Table */}
      {loading ? (
        <div className="text-center py-12">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
          <p className="mt-2 text-gray-600">Loading jobs...</p>
        </div>
      ) : connectors.length === 0 ? (
        <div className="bg-white rounded-lg shadow p-12 text-center">
          <p className="text-gray-500">Please create a connector first before creating jobs</p>
        </div>
      ) : jobs.length === 0 ? (
        <div className="bg-white rounded-lg shadow p-12 text-center">
          <p className="text-gray-500 mb-4">No jobs configured yet</p>
          <button
            onClick={() => setShowCreateModal(true)}
            className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition"
          >
            Create Your First Job
          </button>
        </div>
      ) : (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Symbol
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Timeframe
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Connector
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Last Run
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {jobs.map(job => (
                <tr key={job.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm font-medium text-gray-900">{job.symbol}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-gray-900">{job.timeframe}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-gray-900">
                      {getConnectorName(job.connector_exchange_id)}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`px-2 py-1 text-xs font-medium rounded ${
                      job.status === 'active'
                        ? 'bg-green-100 text-green-800'
                        : job.status === 'paused'
                        ? 'bg-yellow-100 text-yellow-800'
                        : 'bg-red-100 text-red-800'
                    }`}>
                      {job.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-gray-500">
                      {job.run_state?.last_run_time
                        ? new Date(job.run_state.last_run_time).toLocaleString()
                        : 'Never'}
                    </div>
                    {job.run_state?.last_error && (
                      <div className="text-xs text-red-500 mt-1">
                        Error: {job.run_state.last_error}
                      </div>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">
                    <div className="flex space-x-2">
                      <button
                        onClick={() => executeJob(job.id)}
                        className="px-3 py-1 bg-blue-500 text-white rounded hover:bg-blue-600 transition disabled:opacity-50 disabled:cursor-not-allowed"
                        disabled={executingJobs.has(job.id)}
                      >
                        {executingJobs.has(job.id) ? 'Running...' : 'Run Now'}
                      </button>
                      {job.status === 'active' ? (
                        <button
                          onClick={() => pauseJob(job.id)}
                          className="text-yellow-600 hover:text-yellow-900"
                        >
                          Pause
                        </button>
                      ) : (
                        <button
                          onClick={() => resumeJob(job.id)}
                          className="text-green-600 hover:text-green-900"
                        >
                          Resume
                        </button>
                      )}
                      <button
                        onClick={() => deleteJob(job.id)}
                        className="text-red-600 hover:text-red-900"
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
      )}

      {/* Create Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h3 className="text-xl font-bold mb-4">Create New Job</h3>
            <form onSubmit={handleCreate}>
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Connector
                </label>
                <select
                  value={formData.connector_exchange_id}
                  onChange={(e) => setFormData({ ...formData, connector_exchange_id: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                  required
                >
                  <option value="">Select a connector</option>
                  {connectors.map(connector => (
                    <option key={connector.id} value={connector.exchange_id}>
                      {connector.display_name} ({connector.exchange_id})
                    </option>
                  ))}
                </select>
              </div>

              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Symbol
                </label>
                <input
                  type="text"
                  value={formData.symbol}
                  onChange={(e) => setFormData({ ...formData, symbol: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="e.g., BTC/USDT"
                  required
                />
              </div>

              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Timeframe
                </label>
                <select
                  value={formData.timeframe}
                  onChange={(e) => setFormData({ ...formData, timeframe: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                  required
                >
                  <option value="1m">1 minute</option>
                  <option value="5m">5 minutes</option>
                  <option value="15m">15 minutes</option>
                  <option value="30m">30 minutes</option>
                  <option value="1h">1 hour</option>
                  <option value="4h">4 hours</option>
                  <option value="1d">1 day</option>
                </select>
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
    </div>
  )
}

export default JobList
