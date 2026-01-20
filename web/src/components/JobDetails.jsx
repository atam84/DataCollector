import { useState, useEffect } from 'react'
import axios from 'axios'
import {
  ArrowDownTrayIcon,
  ChartBarIcon,
  TableCellsIcon,
  InformationCircleIcon
} from '@heroicons/react/24/outline'
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer
} from 'recharts'

const API_BASE = '/api/v1'

function JobDetails({ job, connector }) {
  const [activeTab, setActiveTab] = useState('overview')
  const [ohlcvData, setOhlcvData] = useState([])
  const [loading, setLoading] = useState(false)
  const [pagination, setPagination] = useState({
    page: 1,
    limit: 50,
    total: 0,
    totalPages: 0
  })
  const [exporting, setExporting] = useState(false)

  useEffect(() => {
    if (activeTab === 'data' || activeTab === 'charts') {
      fetchOHLCVData()
    }
  }, [activeTab, pagination.page, job.id])

  const fetchOHLCVData = async () => {
    setLoading(true)
    try {
      const response = await axios.get(`${API_BASE}/jobs/${job.id}/ohlcv`, {
        params: {
          page: pagination.page,
          limit: pagination.limit
        }
      })
      setOhlcvData(response.data.data || [])
      setPagination(prev => ({
        ...prev,
        total: response.data.pagination?.total || 0,
        totalPages: response.data.pagination?.total_pages || 0
      }))
    } catch (err) {
      console.error('Failed to fetch OHLCV data:', err)
      setOhlcvData([])
    } finally {
      setLoading(false)
    }
  }

  const exportData = async (format = 'csv') => {
    setExporting(true)
    try {
      const response = await axios.get(`${API_BASE}/jobs/${job.id}/export`, {
        params: { format },
        responseType: 'blob'
      })

      const url = window.URL.createObjectURL(new Blob([response.data]))
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', `${job.symbol.replace('/', '-')}_${job.timeframe}_${format}.${format}`)
      document.body.appendChild(link)
      link.click()
      link.remove()
      window.URL.revokeObjectURL(url)
    } catch (err) {
      alert('Export failed: ' + (err.response?.data?.error || err.message))
    } finally {
      setExporting(false)
    }
  }

  const exportForML = async () => {
    setExporting(true)
    try {
      const response = await axios.get(`${API_BASE}/jobs/${job.id}/export/ml`, {
        responseType: 'blob'
      })

      const url = window.URL.createObjectURL(new Blob([response.data]))
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', `${job.symbol.replace('/', '-')}_${job.timeframe}_ml.csv`)
      document.body.appendChild(link)
      link.click()
      link.remove()
      window.URL.revokeObjectURL(url)
    } catch (err) {
      alert('ML Export failed: ' + (err.response?.data?.error || err.message))
    } finally {
      setExporting(false)
    }
  }

  const formatTimestamp = (timestamp) => {
    return new Date(timestamp).toLocaleString()
  }

  const formatNumber = (num) => {
    if (num === null || num === undefined) return 'N/A'
    return typeof num === 'number' ? num.toFixed(8) : num
  }

  // Prepare data for charts
  const chartData = ohlcvData.map(item => ({
    time: new Date(item.timestamp).toLocaleDateString(),
    open: item.open,
    high: item.high,
    low: item.low,
    close: item.close,
    volume: item.volume
  }))

  return (
    <div className="p-6">
      {/* Tab Navigation */}
      <div className="border-b border-gray-200 mb-6">
        <nav className="-mb-px flex space-x-8">
          <button
            onClick={() => setActiveTab('overview')}
            className={`${
              activeTab === 'overview'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm flex items-center`}
          >
            <InformationCircleIcon className="w-5 h-5 mr-2" />
            Overview
          </button>
          <button
            onClick={() => setActiveTab('data')}
            className={`${
              activeTab === 'data'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm flex items-center`}
          >
            <TableCellsIcon className="w-5 h-5 mr-2" />
            Raw Data
          </button>
          <button
            onClick={() => setActiveTab('charts')}
            className={`${
              activeTab === 'charts'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm flex items-center`}
          >
            <ChartBarIcon className="w-5 h-5 mr-2" />
            Charts
          </button>
        </nav>
      </div>

      {/* Overview Tab */}
      {activeTab === 'overview' && (
        <div className="space-y-6">
          {/* Job Information */}
          <div className="bg-gray-50 rounded-lg p-6">
            <h3 className="text-lg font-bold text-gray-900 mb-4">Job Information</h3>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <span className="text-sm font-medium text-gray-500">Symbol</span>
                <p className="text-base text-gray-900 mt-1">{job.symbol}</p>
              </div>
              <div>
                <span className="text-sm font-medium text-gray-500">Timeframe</span>
                <p className="text-base text-gray-900 mt-1">{job.timeframe}</p>
              </div>
              <div>
                <span className="text-sm font-medium text-gray-500">Status</span>
                <p className="text-base text-gray-900 mt-1">
                  <span className={`px-2 py-1 text-xs font-medium rounded ${
                    job.status === 'active'
                      ? 'bg-green-100 text-green-800'
                      : job.status === 'paused'
                      ? 'bg-yellow-100 text-yellow-800'
                      : 'bg-red-100 text-red-800'
                  }`}>
                    {job.status}
                  </span>
                </p>
              </div>
              <div>
                <span className="text-sm font-medium text-gray-500">Job ID</span>
                <p className="text-base text-gray-900 mt-1 font-mono text-sm">{job.id}</p>
              </div>
            </div>
          </div>

          {/* Connector Information */}
          {connector && (
            <div className="bg-gray-50 rounded-lg p-6">
              <h3 className="text-lg font-bold text-gray-900 mb-4">Connector Information</h3>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <span className="text-sm font-medium text-gray-500">Connector</span>
                  <p className="text-base text-gray-900 mt-1">{connector.display_name}</p>
                </div>
                <div>
                  <span className="text-sm font-medium text-gray-500">Exchange</span>
                  <p className="text-base text-gray-900 mt-1">{connector.exchange_id}</p>
                </div>
                <div>
                  <span className="text-sm font-medium text-gray-500">Mode</span>
                  <p className="text-base text-gray-900 mt-1">
                    {connector.sandbox_mode ? 'Sandbox (Testnet)' : 'Production'}
                  </p>
                </div>
                <div>
                  <span className="text-sm font-medium text-gray-500">Connector Status</span>
                  <p className="text-base text-gray-900 mt-1">
                    <span className={`px-2 py-1 text-xs font-medium rounded ${
                      connector.status === 'active'
                        ? 'bg-green-100 text-green-800'
                        : 'bg-gray-100 text-gray-800'
                    }`}>
                      {connector.status}
                    </span>
                  </p>
                </div>
              </div>
            </div>
          )}

          {/* Run Statistics */}
          {job.run_state && (
            <div className="bg-gray-50 rounded-lg p-6">
              <h3 className="text-lg font-bold text-gray-900 mb-4">Run Statistics</h3>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <span className="text-sm font-medium text-gray-500">Total Runs</span>
                  <p className="text-base text-gray-900 mt-1">{job.run_state.runs_total || 0}</p>
                </div>
                <div>
                  <span className="text-sm font-medium text-gray-500">Last Run</span>
                  <p className="text-base text-gray-900 mt-1">
                    {job.run_state.last_run_time
                      ? formatTimestamp(job.run_state.last_run_time)
                      : 'Never'}
                  </p>
                </div>
                <div>
                  <span className="text-sm font-medium text-gray-500">Next Run</span>
                  <p className="text-base text-gray-900 mt-1">
                    {job.run_state.next_run_time
                      ? formatTimestamp(job.run_state.next_run_time)
                      : 'Not scheduled'}
                  </p>
                </div>
                {job.run_state.last_error && (
                  <div className="col-span-2">
                    <span className="text-sm font-medium text-red-500">Last Error</span>
                    <p className="text-sm text-red-600 mt-1 p-2 bg-red-50 rounded">
                      {job.run_state.last_error}
                    </p>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Export Section */}
          <div className="bg-blue-50 rounded-lg p-6 border-2 border-blue-200">
            <h3 className="text-lg font-bold text-blue-900 mb-4">Export Data</h3>
            <p className="text-sm text-blue-700 mb-4">
              Download historical OHLCV data and calculated indicators for offline analysis or machine learning.
            </p>
            <div className="flex space-x-3">
              <button
                onClick={() => exportData('csv')}
                disabled={exporting}
                className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition disabled:opacity-50 flex items-center"
                title="Export as CSV"
              >
                <ArrowDownTrayIcon className="w-5 h-5 mr-2" />
                Export CSV
              </button>
              <button
                onClick={() => exportData('json')}
                disabled={exporting}
                className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 transition disabled:opacity-50 flex items-center"
                title="Export as JSON"
              >
                <ArrowDownTrayIcon className="w-5 h-5 mr-2" />
                Export JSON
              </button>
              <button
                onClick={exportForML}
                disabled={exporting}
                className="px-4 py-2 bg-purple-600 text-white rounded hover:bg-purple-700 transition disabled:opacity-50 flex items-center"
                title="Export optimized for ML (normalized, feature-engineered)"
              >
                <ArrowDownTrayIcon className="w-5 h-5 mr-2" />
                Export for ML
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Raw Data Tab */}
      {activeTab === 'data' && (
        <div>
          <div className="mb-4 flex justify-between items-center">
            <div className="text-sm text-gray-600">
              Showing {ohlcvData.length} records (Page {pagination.page} of {pagination.totalPages})
            </div>
            <div className="flex space-x-2">
              <button
                onClick={() => setPagination(prev => ({ ...prev, page: Math.max(1, prev.page - 1) }))}
                disabled={pagination.page === 1}
                className="px-3 py-1 border border-gray-300 rounded hover:bg-gray-50 disabled:opacity-50"
              >
                Previous
              </button>
              <button
                onClick={() => setPagination(prev => ({ ...prev, page: Math.min(prev.totalPages, prev.page + 1) }))}
                disabled={pagination.page === pagination.totalPages}
                className="px-3 py-1 border border-gray-300 rounded hover:bg-gray-50 disabled:opacity-50"
              >
                Next
              </button>
            </div>
          </div>

          {loading ? (
            <div className="text-center py-12">
              <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
              <p className="mt-2 text-gray-600">Loading data...</p>
            </div>
          ) : ohlcvData.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              No data available for this job
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200 border border-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Timestamp</th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Open</th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">High</th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Low</th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Close</th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Volume</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {ohlcvData.map((row, idx) => (
                    <tr key={idx} className="hover:bg-gray-50">
                      <td className="px-4 py-3 text-sm text-gray-900 whitespace-nowrap">
                        {formatTimestamp(row.timestamp)}
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-900 text-right font-mono">
                        {formatNumber(row.open)}
                      </td>
                      <td className="px-4 py-3 text-sm text-green-600 text-right font-mono">
                        {formatNumber(row.high)}
                      </td>
                      <td className="px-4 py-3 text-sm text-red-600 text-right font-mono">
                        {formatNumber(row.low)}
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-900 text-right font-mono">
                        {formatNumber(row.close)}
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-600 text-right font-mono">
                        {formatNumber(row.volume)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* Charts Tab */}
      {activeTab === 'charts' && (
        <div className="space-y-8">
          {loading ? (
            <div className="text-center py-12">
              <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
              <p className="mt-2 text-gray-600">Loading chart data...</p>
            </div>
          ) : chartData.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              No data available for charts
            </div>
          ) : (
            <>
              {/* Price Chart */}
              <div className="bg-white p-6 rounded-lg border border-gray-200">
                <h3 className="text-lg font-bold text-gray-900 mb-4">Price Chart (OHLC)</h3>
                <ResponsiveContainer width="100%" height={400}>
                  <LineChart data={chartData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="time" />
                    <YAxis domain={['auto', 'auto']} />
                    <Tooltip />
                    <Legend />
                    <Line type="monotone" dataKey="open" stroke="#8884d8" name="Open" dot={false} />
                    <Line type="monotone" dataKey="high" stroke="#82ca9d" name="High" dot={false} />
                    <Line type="monotone" dataKey="low" stroke="#ff7c7c" name="Low" dot={false} />
                    <Line type="monotone" dataKey="close" stroke="#ffc658" name="Close" strokeWidth={2} dot={false} />
                  </LineChart>
                </ResponsiveContainer>
              </div>

              {/* Volume Chart */}
              <div className="bg-white p-6 rounded-lg border border-gray-200">
                <h3 className="text-lg font-bold text-gray-900 mb-4">Volume Chart</h3>
                <ResponsiveContainer width="100%" height={300}>
                  <BarChart data={chartData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="time" />
                    <YAxis />
                    <Tooltip />
                    <Legend />
                    <Bar dataKey="volume" fill="#8884d8" name="Volume" />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </>
          )}
        </div>
      )}
    </div>
  )
}

export default JobDetails
