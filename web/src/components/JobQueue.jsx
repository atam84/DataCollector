import { useState, useEffect } from 'react'
import axios from 'axios'

const API_BASE = '/api/v1'

function JobQueue({ connectors }) {
  const [queue, setQueue] = useState([])
  const [loading, setLoading] = useState(true)

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
        <span className="text-sm text-gray-500">{queue.length} jobs scheduled</span>
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
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {queue.map(job => (
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
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}

export default JobQueue
