import { useState } from 'react'
import axios from 'axios'
import {
  Cog6ToothIcon,
  ArrowPathIcon,
  PauseIcon,
  PlayIcon,
  TrashIcon,
  PlusIcon,
  BoltIcon
} from '@heroicons/react/24/outline'
import IndicatorConfig from './IndicatorConfig'

const API_BASE = '/api/v1'

function JobList({ jobs, connectors, onRefresh, loading }) {
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showConfigModal, setShowConfigModal] = useState(false)
  const [selectedJob, setSelectedJob] = useState(null)
  const [indicatorConfig, setIndicatorConfig] = useState(null)
  const [connectorConfig, setConnectorConfig] = useState(null)
  const [loadingConfig, setLoadingConfig] = useState(false)
  const [recalculatingJobs, setRecalculatingJobs] = useState(new Set())
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

  const getConnector = (exchangeId) => {
    return connectors.find(c => c.exchange_id === exchangeId)
  }

  const openConfigModal = async (job) => {
    setSelectedJob(job)
    setShowConfigModal(true)
    setLoadingConfig(true)

    try {
      // Fetch job config
      const jobResponse = await axios.get(`${API_BASE}/jobs/${job.id}/indicators/config`)
      setIndicatorConfig(jobResponse.data.config || {})

      // Fetch connector config for inheritance display
      const connector = getConnector(job.connector_exchange_id)
      if (connector) {
        const connectorResponse = await axios.get(`${API_BASE}/connectors/${connector.id}/indicators/config`)
        setConnectorConfig(connectorResponse.data.config || {})
      }
    } catch (err) {
      alert('Failed to load indicator config: ' + (err.response?.data?.error || err.message))
      setShowConfigModal(false)
    } finally {
      setLoadingConfig(false)
    }
  }

  const saveIndicatorConfig = async (config) => {
    try {
      await axios.put(`${API_BASE}/jobs/${selectedJob.id}/indicators/config`, config)
      alert('Indicator configuration saved successfully!')
      setShowConfigModal(false)
    } catch (err) {
      alert('Failed to save indicator config: ' + (err.response?.data?.error || err.message))
    }
  }

  const recalculateJob = async (jobId) => {
    if (!confirm('Recalculate all indicators for this job? This may take a while.')) return

    setRecalculatingJobs(prev => new Set(prev).add(jobId))
    try {
      await axios.post(`${API_BASE}/jobs/${jobId}/indicators/recalculate`)
      alert('Indicators recalculated successfully!')
    } catch (err) {
      alert('Failed to recalculate: ' + (err.response?.data?.error || err.message))
    } finally {
      setRecalculatingJobs(prev => {
        const newSet = new Set(prev)
        newSet.delete(jobId)
        return newSet
      })
    }
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold text-gray-900">Jobs</h2>
        <button
          onClick={() => setShowCreateModal(true)}
          className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition flex items-center space-x-2"
          disabled={connectors.length === 0}
        >
          <PlusIcon className="w-5 h-5" />
          <span>New Job</span>
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
                    <div className="flex flex-col space-y-1">
                      <div className="flex space-x-2">
                        <button
                          onClick={() => executeJob(job.id)}
                          className="px-3 py-1 bg-blue-500 text-white rounded hover:bg-blue-600 transition disabled:opacity-50 disabled:cursor-not-allowed text-xs flex items-center"
                          disabled={executingJobs.has(job.id)}
                        >
                          <BoltIcon className="w-3 h-3 mr-1" />
                          <span>{executingJobs.has(job.id) ? 'Running...' : 'Run'}</span>
                        </button>
                        <button
                          onClick={() => openConfigModal(job)}
                          className="px-3 py-1 bg-purple-500 text-white rounded hover:bg-purple-600 transition text-xs flex items-center"
                        >
                          <Cog6ToothIcon className="w-3 h-3 mr-1" />
                          <span>Config</span>
                        </button>
                        <button
                          onClick={() => recalculateJob(job.id)}
                          className="px-3 py-1 bg-indigo-500 text-white rounded hover:bg-indigo-600 transition disabled:opacity-50 text-xs flex items-center"
                          disabled={recalculatingJobs.has(job.id)}
                          title="Recalculate indicators"
                        >
                          <ArrowPathIcon className={`w-3 h-3 ${recalculatingJobs.has(job.id) ? 'animate-spin' : ''}`} />
                        </button>
                      </div>
                      <div className="flex space-x-2">
                        {job.status === 'active' ? (
                          <button
                            onClick={() => pauseJob(job.id)}
                            className="text-yellow-600 hover:text-yellow-900 text-xs flex items-center"
                          >
                            <PauseIcon className="w-3 h-3 mr-1" />
                            <span>Pause</span>
                          </button>
                        ) : (
                          <button
                            onClick={() => resumeJob(job.id)}
                            className="text-green-600 hover:text-green-900 text-xs flex items-center"
                          >
                            <PlayIcon className="w-3 h-3 mr-1" />
                            <span>Resume</span>
                          </button>
                        )}
                        <button
                          onClick={() => deleteJob(job.id)}
                          className="text-red-600 hover:text-red-900 text-xs flex items-center"
                        >
                          <TrashIcon className="w-3 h-3 mr-1" />
                          <span>Delete</span>
                        </button>
                      </div>
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

      {/* Indicator Config Modal */}
      {showConfigModal && selectedJob && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col">
            <div className="p-6 border-b">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-xl font-bold text-gray-900">Indicator Configuration</h3>
                  <p className="text-sm text-gray-500 mt-1">
                    {selectedJob.symbol} - {selectedJob.timeframe} ({getConnectorName(selectedJob.connector_exchange_id)})
                  </p>
                  <p className="text-xs text-blue-600 mt-1">
                    💡 Job-level config overrides connector defaults. Disabled items inherit from connector.
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
                  isJobLevel={true}
                  connectorConfig={connectorConfig}
                />
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default JobList
