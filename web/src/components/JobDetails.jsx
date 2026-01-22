import { useState, useEffect } from 'react'
import axios from 'axios'
import {
  ArrowDownTrayIcon,
  ChartBarIcon,
  TableCellsIcon,
  InformationCircleIcon,
  ChartBarSquareIcon,
  CheckCircleIcon,
  ExclamationTriangleIcon,
  ClockIcon
} from '@heroicons/react/24/outline'
import CandlestickChart from './CandlestickChart'

const API_BASE = '/api/v1'

function JobDetails({ job, connector }) {
  const [activeTab, setActiveTab] = useState('overview')
  const [ohlcvData, setOhlcvData] = useState([])
  const [chartData, setChartData] = useState([])
  const [qualityData, setQualityData] = useState(null)
  const [loading, setLoading] = useState(false)
  const [chartLoading, setChartLoading] = useState(false)
  const [qualityLoading, setQualityLoading] = useState(false)
  const [pagination, setPagination] = useState({
    page: 1,
    limit: 50,
    total: 0,
    totalPages: 0
  })
  const [exporting, setExporting] = useState(false)

  useEffect(() => {
    if (activeTab === 'data') {
      fetchOHLCVData()
    } else if (activeTab === 'charts') {
      fetchChartData()
    } else if (activeTab === 'quality') {
      fetchQualityData()
    }
  }, [activeTab, pagination.page, job.id])

  const fetchQualityData = async () => {
    setQualityLoading(true)
    try {
      const response = await axios.get(`${API_BASE}/jobs/${job.id}/quality`)
      setQualityData(response.data.data)
    } catch (err) {
      console.error('Failed to fetch quality data:', err)
      setQualityData(null)
    } finally {
      setQualityLoading(false)
    }
  }

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

  const fetchChartData = async () => {
    setChartLoading(true)
    try {
      // Fetch more data for charts (up to 500 candles)
      const response = await axios.get(`${API_BASE}/jobs/${job.id}/ohlcv`, {
        params: {
          page: 1,
          limit: 500
        }
      })
      setChartData(response.data.data || [])
    } catch (err) {
      console.error('Failed to fetch chart data:', err)
      setChartData([])
    } finally {
      setChartLoading(false)
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
          <button
            onClick={() => setActiveTab('quality')}
            className={`${
              activeTab === 'quality'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm flex items-center`}
          >
            <ChartBarSquareIcon className="w-5 h-5 mr-2" />
            Quality
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
        <div>
          {chartLoading ? (
            <div className="text-center py-12">
              <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
              <p className="mt-2 text-gray-600">Loading chart data...</p>
            </div>
          ) : chartData.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              No data available for charts. Execute the job to collect data.
            </div>
          ) : (
            <div className="bg-white p-6 rounded-lg border border-gray-200">
              <CandlestickChart
                data={chartData}
                symbol={job.symbol}
                timeframe={job.timeframe}
                height={550}
              />
              <div className="mt-4 text-sm text-gray-500">
                Showing {chartData.length} candles. Technical indicators are displayed when available.
              </div>
            </div>
          )}
        </div>
      )}

      {/* Quality Tab */}
      {activeTab === 'quality' && (
        <div>
          {qualityLoading ? (
            <div className="text-center py-12">
              <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
              <p className="mt-2 text-gray-600">Analyzing data quality...</p>
            </div>
          ) : !qualityData ? (
            <div className="text-center py-12 text-gray-500">
              No quality data available. Execute the job to collect data first.
            </div>
          ) : (
            <div className="space-y-6">
              {/* Quality Status Banner */}
              <div className={`p-6 rounded-lg border-2 ${
                qualityData.quality_status === 'excellent' ? 'bg-green-50 border-green-300' :
                qualityData.quality_status === 'good' ? 'bg-blue-50 border-blue-300' :
                qualityData.quality_status === 'fair' ? 'bg-yellow-50 border-yellow-300' :
                'bg-red-50 border-red-300'
              }`}>
                <div className="flex items-center justify-between">
                  <div className="flex items-center">
                    {qualityData.quality_status === 'excellent' && <CheckCircleIcon className="w-8 h-8 text-green-600 mr-3" />}
                    {qualityData.quality_status === 'good' && <CheckCircleIcon className="w-8 h-8 text-blue-600 mr-3" />}
                    {qualityData.quality_status === 'fair' && <ExclamationTriangleIcon className="w-8 h-8 text-yellow-600 mr-3" />}
                    {qualityData.quality_status === 'poor' && <ExclamationTriangleIcon className="w-8 h-8 text-red-600 mr-3" />}
                    <div>
                      <h3 className="text-xl font-bold capitalize">{qualityData.quality_status} Quality</h3>
                      <p className="text-sm text-gray-600">
                        {qualityData.completeness_score?.toFixed(1)}% complete with {qualityData.gaps_detected || 0} gaps detected
                      </p>
                    </div>
                  </div>
                  <div className={`px-4 py-2 rounded-lg font-semibold ${
                    qualityData.data_freshness === 'fresh' ? 'bg-green-200 text-green-800' :
                    qualityData.data_freshness === 'stale' ? 'bg-yellow-200 text-yellow-800' :
                    'bg-red-200 text-red-800'
                  }`}>
                    <ClockIcon className="w-4 h-4 inline mr-1" />
                    {qualityData.data_freshness === 'fresh' ? 'Fresh Data' :
                     qualityData.data_freshness === 'stale' ? 'Stale Data' : 'Very Stale'}
                  </div>
                </div>
              </div>

              {/* Metrics Grid */}
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <div className="bg-gray-50 rounded-lg p-4 text-center">
                  <p className="text-3xl font-bold text-indigo-600">{qualityData.total_candles?.toLocaleString() || 0}</p>
                  <p className="text-sm text-gray-600">Total Candles</p>
                </div>
                <div className="bg-gray-50 rounded-lg p-4 text-center">
                  <p className="text-3xl font-bold text-blue-600">{qualityData.expected_candles?.toLocaleString() || 0}</p>
                  <p className="text-sm text-gray-600">Expected Candles</p>
                </div>
                <div className="bg-gray-50 rounded-lg p-4 text-center">
                  <p className={`text-3xl font-bold ${(qualityData.missing_candles || 0) > 0 ? 'text-red-600' : 'text-green-600'}`}>
                    {qualityData.missing_candles?.toLocaleString() || 0}
                  </p>
                  <p className="text-sm text-gray-600">Missing Candles</p>
                </div>
                <div className="bg-gray-50 rounded-lg p-4 text-center">
                  <p className={`text-3xl font-bold ${(qualityData.gaps_detected || 0) > 0 ? 'text-orange-600' : 'text-green-600'}`}>
                    {qualityData.gaps_detected || 0}
                  </p>
                  <p className="text-sm text-gray-600">Gaps Detected</p>
                </div>
              </div>

              {/* Completeness Progress */}
              <div className="bg-gray-50 rounded-lg p-6">
                <h4 className="text-lg font-semibold text-gray-900 mb-4">Data Completeness</h4>
                <div className="relative pt-1">
                  <div className="flex mb-2 items-center justify-between">
                    <div>
                      <span className={`text-xs font-semibold inline-block py-1 px-2 uppercase rounded-full ${
                        (qualityData.completeness_score || 0) >= 95 ? 'text-green-600 bg-green-200' :
                        (qualityData.completeness_score || 0) >= 80 ? 'text-yellow-600 bg-yellow-200' :
                        'text-red-600 bg-red-200'
                      }`}>
                        {(qualityData.completeness_score || 0).toFixed(1)}%
                      </span>
                    </div>
                  </div>
                  <div className="overflow-hidden h-4 mb-4 text-xs flex rounded-full bg-gray-200">
                    <div
                      style={{ width: `${Math.min(qualityData.completeness_score || 0, 100)}%` }}
                      className={`shadow-none flex flex-col text-center whitespace-nowrap text-white justify-center transition-all duration-500 ${
                        (qualityData.completeness_score || 0) >= 95 ? 'bg-green-500' :
                        (qualityData.completeness_score || 0) >= 80 ? 'bg-yellow-500' :
                        'bg-red-500'
                      }`}
                    />
                  </div>
                </div>
              </div>

              {/* Date Range */}
              <div className="bg-gray-50 rounded-lg p-6">
                <h4 className="text-lg font-semibold text-gray-900 mb-4">Data Range</h4>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <span className="text-sm font-medium text-gray-500">Oldest Data</span>
                    <p className="text-base text-gray-900 mt-1">
                      {qualityData.oldest_candle ? new Date(qualityData.oldest_candle).toLocaleString() : 'N/A'}
                    </p>
                  </div>
                  <div>
                    <span className="text-sm font-medium text-gray-500">Newest Data</span>
                    <p className="text-base text-gray-900 mt-1">
                      {qualityData.newest_candle ? new Date(qualityData.newest_candle).toLocaleString() : 'N/A'}
                    </p>
                  </div>
                  <div>
                    <span className="text-sm font-medium text-gray-500">Data Age</span>
                    <p className="text-base text-gray-900 mt-1">
                      {qualityData.freshness_minutes ?
                        qualityData.freshness_minutes < 60 ? `${qualityData.freshness_minutes} minutes ago` :
                        qualityData.freshness_minutes < 1440 ? `${Math.round(qualityData.freshness_minutes / 60)} hours ago` :
                        `${Math.round(qualityData.freshness_minutes / 1440)} days ago`
                        : 'N/A'}
                    </p>
                  </div>
                  <div>
                    <span className="text-sm font-medium text-gray-500">Timeframe</span>
                    <p className="text-base text-gray-900 mt-1">{job.timeframe}</p>
                  </div>
                </div>
              </div>

              {/* Gaps Detail */}
              {qualityData.gaps && qualityData.gaps.length > 0 && (
                <div className="bg-orange-50 rounded-lg p-6 border border-orange-200">
                  <h4 className="text-lg font-semibold text-orange-900 mb-4">
                    <ExclamationTriangleIcon className="w-5 h-5 inline mr-2" />
                    Detected Gaps ({qualityData.gaps.length})
                  </h4>
                  <div className="overflow-x-auto">
                    <table className="min-w-full divide-y divide-orange-200">
                      <thead>
                        <tr>
                          <th className="px-4 py-2 text-left text-xs font-medium text-orange-700 uppercase">Start</th>
                          <th className="px-4 py-2 text-left text-xs font-medium text-orange-700 uppercase">End</th>
                          <th className="px-4 py-2 text-right text-xs font-medium text-orange-700 uppercase">Missing Candles</th>
                          <th className="px-4 py-2 text-right text-xs font-medium text-orange-700 uppercase">Duration</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-orange-200">
                        {qualityData.gaps.slice(0, 10).map((gap, idx) => (
                          <tr key={idx}>
                            <td className="px-4 py-2 text-sm text-gray-900">
                              {new Date(gap.start_time).toLocaleString()}
                            </td>
                            <td className="px-4 py-2 text-sm text-gray-900">
                              {new Date(gap.end_time).toLocaleString()}
                            </td>
                            <td className="px-4 py-2 text-sm text-right font-semibold text-orange-700">
                              {gap.missing_candles}
                            </td>
                            <td className="px-4 py-2 text-sm text-right text-gray-600">
                              {gap.duration_minutes < 60 ? `${gap.duration_minutes}m` :
                               gap.duration_minutes < 1440 ? `${Math.round(gap.duration_minutes / 60)}h` :
                               `${Math.round(gap.duration_minutes / 1440)}d`}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                    {qualityData.gaps.length > 10 && (
                      <p className="text-sm text-orange-600 mt-2">
                        Showing first 10 gaps of {qualityData.gaps.length} total
                      </p>
                    )}
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default JobDetails
