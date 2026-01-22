import { useState } from 'react'
import axios from 'axios'
import {
  ArrowPathIcon,
  PauseIcon,
  PlayIcon,
  TrashIcon,
  PlusIcon,
  BoltIcon,
  MagnifyingGlassIcon,
  FunnelIcon,
  ExclamationTriangleIcon,
  CheckCircleIcon,
  XCircleIcon,
  ClockIcon,
  DocumentDuplicateIcon
} from '@heroicons/react/24/outline'
import JobDetails from './JobDetails'
import JobWizard from './JobWizard'

const API_BASE = '/api/v1'

// Truncate text with ellipsis
const truncateText = (text, maxLength = 50) => {
  if (!text) return ''
  if (text.length <= maxLength) return text
  return text.substring(0, maxLength) + '...'
}

// Extract the main error message from a verbose error string
const extractMainError = (error) => {
  if (!error) return ''

  // Try to extract the core error message
  // Look for common patterns like [ccxtError]::[Type]::[exchange {...}]
  const ccxtMatch = error.match(/\[ccxtError\]::\[(\w+)\]::\[(\w+)\s*({.*?})\]/)
  if (ccxtMatch) {
    try {
      const jsonPart = JSON.parse(ccxtMatch[3])
      return `${ccxtMatch[1]}: ${jsonPart.msg || jsonPart.message || ccxtMatch[2]}`
    } catch {
      return `${ccxtMatch[1]}: ${ccxtMatch[2]}`
    }
  }

  // Look for "Error:" prefix
  const errorMatch = error.match(/^Error:\s*(.+?)(?:\s*Stack:|$)/i)
  if (errorMatch) {
    return truncateText(errorMatch[1], 80)
  }

  // Just return truncated version
  return truncateText(error, 80)
}

function JobList({ jobs, connectors, onRefresh, loading }) {
  const [showCreateWizard, setShowCreateWizard] = useState(false)
  const [showDetailsModal, setShowDetailsModal] = useState(false)
  const [selectedJob, setSelectedJob] = useState(null)
  const [showErrorModal, setShowErrorModal] = useState(false)
  const [selectedError, setSelectedError] = useState({ job: null, error: '' })
  const [recalculatingJobs, setRecalculatingJobs] = useState(new Set())
  const [searchTerm, setSearchTerm] = useState('')
  const [selectedConnectors, setSelectedConnectors] = useState([])
  const [showConnectorFilter, setShowConnectorFilter] = useState(false)
  const [executingJobs, setExecutingJobs] = useState(new Set())
  const [selectedJobs, setSelectedJobs] = useState(new Set())
  const [deletingSelected, setDeletingSelected] = useState(false)
  const [showExecutionModal, setShowExecutionModal] = useState(false)
  const [executionResult, setExecutionResult] = useState(null)
  const [copiedToClipboard, setCopiedToClipboard] = useState(false)

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text)
    setCopiedToClipboard(true)
    setTimeout(() => setCopiedToClipboard(false), 2000)
  }

  const handleWizardSave = async (jobs) => {
    try {
      // Create all jobs via batch endpoint
      await axios.post(`${API_BASE}/jobs/batch`, { jobs })
      onRefresh()
    } catch (err) {
      throw new Error(err.response?.data?.error || err.message)
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

  const deleteSelectedJobs = async () => {
    if (selectedJobs.size === 0) return
    if (!confirm(`Are you sure you want to delete ${selectedJobs.size} selected job(s)?`)) return

    setDeletingSelected(true)
    let successCount = 0
    let failCount = 0

    for (const jobId of selectedJobs) {
      try {
        await axios.delete(`${API_BASE}/jobs/${jobId}`)
        successCount++
      } catch (err) {
        failCount++
        console.error(`Failed to delete job ${jobId}:`, err)
      }
    }

    setDeletingSelected(false)
    setSelectedJobs(new Set())

    if (failCount > 0) {
      alert(`Deleted ${successCount} jobs. Failed to delete ${failCount} jobs.`)
    }

    onRefresh()
  }

  const executeJob = async (id) => {
    const job = jobs.find(j => j.id === id)
    setExecutingJobs(prev => new Set(prev).add(id))
    try {
      const startTime = Date.now()
      const response = await axios.post(`${API_BASE}/jobs/${id}/execute`)
      const result = response.data

      // Show execution result modal with detailed information
      setExecutionResult({
        job: job,
        success: result.success && result.data?.success,
        data: result.data,
        error: result.data?.error || result.error || result.message,
        clientElapsedMs: Date.now() - startTime
      })
      setShowExecutionModal(true)

      onRefresh()
    } catch (err) {
      // Show error in modal instead of alert
      setExecutionResult({
        job: job,
        success: false,
        error: err.response?.data?.error || err.response?.data?.message || err.message,
        data: null
      })
      setShowExecutionModal(true)
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

  const recalculateJob = async (jobId) => {
    if (!confirm('Recalculate all indicators for this job? This may take a while.')) return

    const job = jobs.find(j => j.id === jobId)
    setRecalculatingJobs(prev => new Set(prev).add(jobId))
    try {
      const response = await axios.post(`${API_BASE}/jobs/${jobId}/indicators/recalculate`)
      setExecutionResult({
        job: job,
        success: true,
        data: {
          message: response.data?.message || 'Indicators recalculated successfully',
          records_fetched: response.data?.records_updated || response.data?.count,
          execution_time_ms: response.data?.execution_time_ms
        }
      })
      setShowExecutionModal(true)
    } catch (err) {
      setExecutionResult({
        job: job,
        success: false,
        error: err.response?.data?.error || err.response?.data?.message || err.message
      })
      setShowExecutionModal(true)
    } finally {
      setRecalculatingJobs(prev => {
        const newSet = new Set(prev)
        newSet.delete(jobId)
        return newSet
      })
    }
  }

  const openJobDetails = (job) => {
    setSelectedJob(job)
    setShowDetailsModal(true)
  }

  const toggleConnectorFilter = (exchangeId) => {
    setSelectedConnectors(prev => {
      if (prev.includes(exchangeId)) {
        return prev.filter(id => id !== exchangeId)
      } else {
        return [...prev, exchangeId]
      }
    })
  }

  const toggleJobSelection = (jobId) => {
    setSelectedJobs(prev => {
      const newSet = new Set(prev)
      if (newSet.has(jobId)) {
        newSet.delete(jobId)
      } else {
        newSet.add(jobId)
      }
      return newSet
    })
  }

  const toggleSelectAll = () => {
    if (selectedJobs.size === filteredJobs.length) {
      setSelectedJobs(new Set())
    } else {
      setSelectedJobs(new Set(filteredJobs.map(job => job.id)))
    }
  }

  // Filter jobs based on search and connector selection
  const filteredJobs = jobs.filter(job => {
    const matchesSearch = job.symbol.toLowerCase().includes(searchTerm.toLowerCase())
    const matchesConnector = selectedConnectors.length === 0 ||
                            selectedConnectors.includes(job.connector_exchange_id)
    return matchesSearch && matchesConnector
  })

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold text-gray-900">Jobs</h2>
        <div className="flex space-x-2">
          {selectedJobs.size > 0 && (
            <button
              onClick={deleteSelectedJobs}
              className="px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600 transition disabled:opacity-50 flex items-center"
              disabled={deletingSelected}
              title={`Delete ${selectedJobs.size} selected job(s)`}
            >
              <TrashIcon className="w-4 h-4 mr-2" />
              {deletingSelected ? 'Deleting...' : `Delete (${selectedJobs.size})`}
            </button>
          )}
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
            className="p-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition disabled:opacity-50"
            disabled={connectors.length === 0}
            title="Create Multiple Jobs"
          >
            <PlusIcon className="w-5 h-5" />
          </button>
        </div>
      </div>

      {/* Search and Filter Bar */}
      <div className="bg-white rounded-lg shadow p-4 mb-6">
        <div className="flex flex-col md:flex-row gap-4">
          {/* Search Input */}
          <div className="flex-1 relative">
            <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
            <input
              type="text"
              placeholder="Search by symbol..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          {/* Connector Filter */}
          <div className="relative">
            <button
              onClick={() => setShowConnectorFilter(!showConnectorFilter)}
              className={`px-4 py-2 border rounded-lg flex items-center space-x-2 ${
                selectedConnectors.length > 0
                  ? 'border-blue-500 bg-blue-50 text-blue-700'
                  : 'border-gray-300 text-gray-700'
              }`}
            >
              <FunnelIcon className="w-5 h-5" />
              <span>
                Connectors {selectedConnectors.length > 0 && `(${selectedConnectors.length})`}
              </span>
            </button>

            {/* Connector Dropdown */}
            {showConnectorFilter && (
              <div className="absolute right-0 mt-2 w-64 bg-white rounded-lg shadow-lg border border-gray-200 z-10">
                <div className="p-3">
                  <div className="flex justify-between items-center mb-2">
                    <span className="text-sm font-medium text-gray-700">Filter by Connector</span>
                    {selectedConnectors.length > 0 && (
                      <button
                        onClick={() => setSelectedConnectors([])}
                        className="text-xs text-blue-600 hover:text-blue-800"
                      >
                        Clear all
                      </button>
                    )}
                  </div>
                  <div className="max-h-64 overflow-y-auto">
                    {connectors.map(connector => (
                      <label
                        key={connector.id}
                        className="flex items-center space-x-2 p-2 hover:bg-gray-50 rounded cursor-pointer"
                      >
                        <input
                          type="checkbox"
                          checked={selectedConnectors.includes(connector.exchange_id)}
                          onChange={() => toggleConnectorFilter(connector.exchange_id)}
                          className="rounded text-blue-600 focus:ring-blue-500"
                        />
                        <span className="text-sm text-gray-700">{connector.display_name}</span>
                      </label>
                    ))}
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Active Filters Display */}
        {(searchTerm || selectedConnectors.length > 0) && (
          <div className="mt-3 flex items-center space-x-2 text-sm text-gray-600">
            <span>Active filters:</span>
            {searchTerm && (
              <span className="px-2 py-1 bg-blue-100 text-blue-700 rounded">
                Symbol: "{searchTerm}"
              </span>
            )}
            {selectedConnectors.length > 0 && (
              <span className="px-2 py-1 bg-purple-100 text-purple-700 rounded">
                {selectedConnectors.length} connector(s)
              </span>
            )}
            <button
              onClick={() => {
                setSearchTerm('')
                setSelectedConnectors([])
              }}
              className="text-blue-600 hover:text-blue-800"
            >
              Clear all
            </button>
          </div>
        )}
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
      ) : filteredJobs.length === 0 ? (
        <div className="bg-white rounded-lg shadow p-12 text-center">
          <p className="text-gray-500 mb-4">
            {jobs.length === 0 ? 'No jobs configured yet' : 'No jobs match your filters'}
          </p>
          <button
            onClick={() => setShowCreateWizard(true)}
            className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition"
          >
            Create Your First Job
          </button>
        </div>
      ) : (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <div className="px-6 py-3 bg-gray-50 border-b border-gray-200 flex justify-between items-center">
            <span className="text-sm text-gray-600">
              Showing {filteredJobs.length} of {jobs.length} jobs
              {selectedJobs.size > 0 && ` (${selectedJobs.size} selected)`}
            </span>
            <div className="text-sm text-blue-600">
              All indicators are calculated automatically
            </div>
          </div>
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left">
                  <input
                    type="checkbox"
                    checked={selectedJobs.size === filteredJobs.length && filteredJobs.length > 0}
                    onChange={toggleSelectAll}
                    className="rounded text-blue-600 focus:ring-blue-500"
                    title="Select all"
                  />
                </th>
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
              {filteredJobs.map(job => (
                <tr key={job.id} className={`hover:bg-gray-50 ${selectedJobs.has(job.id) ? 'bg-blue-50' : ''}`}>
                  <td className="px-4 py-4">
                    <input
                      type="checkbox"
                      checked={selectedJobs.has(job.id)}
                      onChange={() => toggleJobSelection(job.id)}
                      className="rounded text-blue-600 focus:ring-blue-500"
                    />
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <button
                      onClick={() => openJobDetails(job)}
                      className="text-sm font-medium text-blue-600 hover:text-blue-800 hover:underline cursor-pointer"
                    >
                      {job.symbol}
                    </button>
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
                  <td className="px-6 py-4">
                    <div className="text-sm text-gray-500">
                      {job.run_state?.last_run_time
                        ? new Date(job.run_state.last_run_time).toLocaleString()
                        : 'Never'}
                    </div>
                    {job.run_state?.last_error && (
                      <div className="flex items-start mt-1 max-w-xs">
                        <ExclamationTriangleIcon className="w-4 h-4 text-red-500 flex-shrink-0 mr-1 mt-0.5" />
                        <div>
                          <span className="text-xs text-red-500">
                            {extractMainError(job.run_state.last_error)}
                          </span>
                          <button
                            onClick={() => {
                              setSelectedError({ job: job, error: job.run_state.last_error })
                              setShowErrorModal(true)
                            }}
                            className="block text-xs text-blue-500 hover:text-blue-700 hover:underline mt-0.5"
                          >
                            View full error
                          </button>
                        </div>
                      </div>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">
                    <div className="flex space-x-1">
                      <button
                        onClick={() => executeJob(job.id)}
                        className="p-1 bg-blue-500 text-white rounded hover:bg-blue-600 transition disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center"
                        disabled={executingJobs.has(job.id)}
                        title={executingJobs.has(job.id) ? 'Running...' : 'Run now'}
                      >
                        <BoltIcon className={`w-4 h-4 ${executingJobs.has(job.id) ? 'animate-spin' : ''}`} />
                      </button>
                      <button
                        onClick={() => recalculateJob(job.id)}
                        className="p-1 bg-indigo-500 text-white rounded hover:bg-indigo-600 transition disabled:opacity-50 flex items-center justify-center"
                        disabled={recalculatingJobs.has(job.id)}
                        title="Recalculate indicators"
                      >
                        <ArrowPathIcon className={`w-4 h-4 ${recalculatingJobs.has(job.id) ? 'animate-spin' : ''}`} />
                      </button>
                      {job.status === 'active' ? (
                        <button
                          onClick={() => pauseJob(job.id)}
                          className="p-1 bg-yellow-500 text-white rounded hover:bg-yellow-600 transition flex items-center justify-center"
                          title="Pause job"
                        >
                          <PauseIcon className="w-4 h-4" />
                        </button>
                      ) : (
                        <button
                          onClick={() => resumeJob(job.id)}
                          className="p-1 bg-green-500 text-white rounded hover:bg-green-600 transition flex items-center justify-center"
                          title="Resume job"
                        >
                          <PlayIcon className="w-4 h-4" />
                        </button>
                      )}
                      <button
                        onClick={() => deleteJob(job.id)}
                        className="p-1 bg-red-500 text-white rounded hover:bg-red-600 transition flex items-center justify-center"
                        title="Delete job"
                      >
                        <TrashIcon className="w-4 h-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Job Creation Wizard */}
      {showCreateWizard && (
        <JobWizard
          connectors={connectors}
          onClose={() => setShowCreateWizard(false)}
          onSave={handleWizardSave}
        />
      )}

      {/* Job Details Modal */}
      {showDetailsModal && selectedJob && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg w-full max-w-7xl max-h-[95vh] overflow-hidden flex flex-col">
            <div className="p-6 border-b">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-2xl font-bold text-gray-900">{selectedJob.symbol}</h3>
                  <p className="text-sm text-gray-500 mt-1">
                    {selectedJob.timeframe} • {getConnectorName(selectedJob.connector_exchange_id)}
                  </p>
                </div>
                <button
                  onClick={() => setShowDetailsModal(false)}
                  className="text-gray-400 hover:text-gray-600 text-2xl"
                >
                  ×
                </button>
              </div>
            </div>

            <div className="flex-1 overflow-y-auto">
              <JobDetails job={selectedJob} connector={getConnector(selectedJob.connector_exchange_id)} />
            </div>
          </div>
        </div>
      )}

      {/* Error Details Modal */}
      {showErrorModal && selectedError.job && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg w-full max-w-3xl max-h-[80vh] overflow-hidden flex flex-col">
            <div className="p-4 border-b bg-red-50">
              <div className="flex items-center justify-between">
                <div className="flex items-center">
                  <ExclamationTriangleIcon className="w-6 h-6 text-red-500 mr-2" />
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900">Job Error Details</h3>
                    <p className="text-sm text-gray-500">
                      {selectedError.job.symbol} • {selectedError.job.timeframe}
                    </p>
                  </div>
                </div>
                <button
                  onClick={() => setShowErrorModal(false)}
                  className="text-gray-400 hover:text-gray-600 text-2xl"
                >
                  ×
                </button>
              </div>
            </div>

            <div className="p-4 overflow-y-auto flex-1">
              <div className="mb-4">
                <h4 className="text-sm font-medium text-gray-700 mb-1">Last Run Time</h4>
                <p className="text-sm text-gray-600">
                  {selectedError.job.run_state?.last_run_time
                    ? new Date(selectedError.job.run_state.last_run_time).toLocaleString()
                    : 'Unknown'}
                </p>
              </div>

              <div>
                <h4 className="text-sm font-medium text-gray-700 mb-1">Full Error Message</h4>
                <div className="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto">
                  <pre className="text-xs whitespace-pre-wrap break-words font-mono">
                    {selectedError.error}
                  </pre>
                </div>
              </div>
            </div>

            <div className="p-4 border-t bg-gray-50 flex justify-end space-x-2">
              <button
                onClick={() => copyToClipboard(selectedError.error)}
                className={`px-4 py-2 rounded transition flex items-center ${
                  copiedToClipboard
                    ? 'bg-green-500 text-white'
                    : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
                }`}
              >
                <DocumentDuplicateIcon className="w-4 h-4 mr-2" />
                {copiedToClipboard ? 'Copied!' : 'Copy to Clipboard'}
              </button>
              <button
                onClick={() => setShowErrorModal(false)}
                className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Job Execution Result Modal */}
      {showExecutionModal && executionResult && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg w-full max-w-lg overflow-hidden">
            {/* Header */}
            <div className={`p-4 border-b ${executionResult.success ? 'bg-green-50' : 'bg-red-50'}`}>
              <div className="flex items-center justify-between">
                <div className="flex items-center">
                  {executionResult.success ? (
                    <CheckCircleIcon className="w-6 h-6 text-green-500 mr-2" />
                  ) : (
                    <XCircleIcon className="w-6 h-6 text-red-500 mr-2" />
                  )}
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900">
                      {executionResult.success ? 'Job Executed Successfully' : 'Job Execution Failed'}
                    </h3>
                    {executionResult.job && (
                      <p className="text-sm text-gray-500">
                        {executionResult.job.symbol} • {executionResult.job.timeframe}
                      </p>
                    )}
                  </div>
                </div>
                <button
                  onClick={() => setShowExecutionModal(false)}
                  className="text-gray-400 hover:text-gray-600 text-2xl"
                >
                  ×
                </button>
              </div>
            </div>

            {/* Content */}
            <div className="p-4">
              {executionResult.success && executionResult.data ? (
                <div className="space-y-4">
                  {/* Stats Grid */}
                  <div className="grid grid-cols-2 gap-4">
                    <div className="bg-blue-50 rounded-lg p-4">
                      <div className="flex items-center text-blue-600 mb-1">
                        <DocumentDuplicateIcon className="w-5 h-5 mr-2" />
                        <span className="text-sm font-medium">Records Fetched</span>
                      </div>
                      <p className="text-2xl font-bold text-blue-700">
                        {executionResult.data.records_fetched?.toLocaleString() || 0}
                      </p>
                      <p className="text-xs text-blue-500 mt-1">candles collected</p>
                    </div>
                    <div className="bg-purple-50 rounded-lg p-4">
                      <div className="flex items-center text-purple-600 mb-1">
                        <ClockIcon className="w-5 h-5 mr-2" />
                        <span className="text-sm font-medium">Execution Time</span>
                      </div>
                      <p className="text-2xl font-bold text-purple-700">
                        {executionResult.data.execution_time_ms?.toLocaleString() || 0}
                        <span className="text-sm font-normal ml-1">ms</span>
                      </p>
                      <p className="text-xs text-purple-500 mt-1">server processing</p>
                    </div>
                  </div>

                  {/* Additional Info */}
                  <div className="bg-gray-50 rounded-lg p-4">
                    <h4 className="text-sm font-medium text-gray-700 mb-2">Additional Details</h4>
                    <div className="space-y-2 text-sm">
                      {executionResult.data.message && (
                        <div className="flex justify-between">
                          <span className="text-gray-500">Message:</span>
                          <span className="text-gray-900 font-medium">{executionResult.data.message}</span>
                        </div>
                      )}
                      {executionResult.data.next_run_time && (
                        <div className="flex justify-between">
                          <span className="text-gray-500">Next Scheduled Run:</span>
                          <span className="text-gray-900 font-medium">
                            {new Date(executionResult.data.next_run_time).toLocaleString()}
                          </span>
                        </div>
                      )}
                      {executionResult.clientElapsedMs && (
                        <div className="flex justify-between">
                          <span className="text-gray-500">Total Round-Trip:</span>
                          <span className="text-gray-900 font-medium">
                            {(executionResult.clientElapsedMs / 1000).toFixed(2)}s
                          </span>
                        </div>
                      )}
                      {executionResult.job && (
                        <div className="flex justify-between">
                          <span className="text-gray-500">Connector:</span>
                          <span className="text-gray-900 font-medium">
                            {getConnectorName(executionResult.job.connector_exchange_id)}
                          </span>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              ) : (
                /* Error Display */
                <div className="space-y-4">
                  <div className="bg-red-50 border border-red-200 rounded-lg p-4">
                    <h4 className="text-sm font-medium text-red-700 mb-2">Error Details</h4>
                    <div className="bg-gray-900 text-gray-100 p-3 rounded overflow-x-auto">
                      <pre className="text-xs whitespace-pre-wrap break-words font-mono">
                        {executionResult.error || 'Unknown error occurred'}
                      </pre>
                    </div>
                  </div>
                  <p className="text-sm text-gray-500">
                    Check the job configuration and connector status, then try again.
                  </p>
                </div>
              )}
            </div>

            {/* Footer */}
            <div className="p-4 border-t bg-gray-50 flex justify-end space-x-2">
              {!executionResult.success && executionResult.error && (
                <button
                  onClick={() => copyToClipboard(executionResult.error)}
                  className={`px-4 py-2 rounded transition flex items-center ${
                    copiedToClipboard
                      ? 'bg-green-500 text-white'
                      : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
                  }`}
                >
                  <DocumentDuplicateIcon className="w-4 h-4 mr-2" />
                  {copiedToClipboard ? 'Copied!' : 'Copy Error'}
                </button>
              )}
              <button
                onClick={() => setShowExecutionModal(false)}
                className={`px-4 py-2 rounded transition ${
                  executionResult.success
                    ? 'bg-green-500 text-white hover:bg-green-600'
                    : 'bg-blue-500 text-white hover:bg-blue-600'
                }`}
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default JobList
