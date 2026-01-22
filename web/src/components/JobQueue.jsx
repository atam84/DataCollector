import { useState, useEffect } from 'react'
import axios from 'axios'
import { BoltIcon, ArrowPathIcon } from '@heroicons/react/24/outline'
import JobDetails from './JobDetails'

const API_BASE = '/api/v1'

function JobQueue({ connectors }) {
  const [queue, setQueue] = useState([])
  const [loading, setLoading] = useState(true)
  const [executingJobs, setExecutingJobs] = useState(new Set())
  const [selectedJob, setSelectedJob] = useState(null)
  const [showDetailsModal, setShowDetailsModal] = useState(false)

  useEffect(() => {
    fetchQueue()
    const interval = setInterval(fetchQueue, 5000) // Refresh every 5 seconds
    return () => clearInterval(interval)
  }, [])

  const fetchQueue = async () => {
    try {
      const response = await axios.get(`${API_BASE}/jobs/queue`)
      setQueue(response.data.data || [])
    } catch (err) {
      console.error('Failed to fetch job queue:', err)
    } finally {
      setLoading(false)
    }
  }

  const getConnectorName = (exchangeId) => {
    const connector = connectors.find(c => c.exchange_id === exchangeId)
    return connector ? connector.display_name : exchangeId
  }

  const getConnector = (exchangeId) => {
    return connectors.find(c => c.exchange_id === exchangeId)
  }

  const openJobDetails = (job) => {
    setSelectedJob(job)
    setShowDetailsModal(true)
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

      fetchQueue()
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

  const getCountdown = (nextRunTime) => {
    if (!nextRunTime) return 'N/A'

    const now = new Date()
    const target = new Date(nextRunTime)
    const diff = target - now

    if (diff <= 0) return 'Now'

    const minutes = Math.floor(diff / 60000)
    const seconds = Math.floor((diff % 60000) / 1000)

    if (minutes > 60) {
      const hours = Math.floor(minutes / 60)
      const remainingMinutes = minutes % 60
      return `${hours}h ${remainingMinutes}m`
    }

    return `${minutes}m ${seconds}s`
  }

  const [countdowns, setCountdowns] = useState({})

  useEffect(() => {
    const interval = setInterval(() => {
      const newCountdowns = {}
      queue.forEach(job => {
        if (job.run_state?.next_run_time) {
          newCountdowns[job.id] = getCountdown(job.run_state.next_run_time)
        }
      })
      setCountdowns(newCountdowns)
    }, 1000)

    return () => clearInterval(interval)
  }, [queue])

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold text-gray-900">Upcoming Executions</h2>
        <div className="flex items-center space-x-4">
          <span className="text-sm text-gray-500">{queue.length} jobs scheduled</span>
          <button
            onClick={fetchQueue}
            className="p-2 bg-gray-100 text-gray-700 rounded hover:bg-gray-200 transition"
            title="Refresh"
            disabled={loading}
          >
            <ArrowPathIcon className={`w-5 h-5 ${loading ? 'animate-spin' : ''}`} />
          </button>
        </div>
      </div>

      {loading ? (
        <div className="text-center py-12">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
          <p className="mt-2 text-gray-600">Loading queue...</p>
        </div>
      ) : queue.length === 0 ? (
        <div className="bg-white rounded-lg shadow p-12 text-center">
          <p className="text-gray-500">No jobs scheduled for execution</p>
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
                  Next Run
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Countdown
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Total Runs
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {queue.map(job => (
                <tr key={job.id} className="hover:bg-gray-50">
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
                    <div className="text-sm text-gray-500">
                      {job.run_state?.next_run_time
                        ? new Date(job.run_state.next_run_time).toLocaleString()
                        : 'Not scheduled'}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`px-2 py-1 text-xs font-medium rounded ${
                      countdowns[job.id] === 'Now'
                        ? 'bg-green-100 text-green-800'
                        : 'bg-blue-100 text-blue-800'
                    }`}>
                      {countdowns[job.id] || 'Calculating...'}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-gray-500">
                      {job.run_state?.runs_total || 0}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <button
                      onClick={() => executeJob(job.id)}
                      className="p-1 bg-blue-500 text-white rounded hover:bg-blue-600 transition disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center"
                      disabled={executingJobs.has(job.id)}
                      title={executingJobs.has(job.id) ? 'Running...' : 'Run now'}
                    >
                      <BoltIcon className={`w-4 h-4 ${executingJobs.has(job.id) ? 'animate-spin' : ''}`} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
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
    </div>
  )
}

export default JobQueue
