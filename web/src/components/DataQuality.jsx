import { useState, useEffect, useCallback } from 'react'
import axios from 'axios'
import {
  ArrowPathIcon,
  ChartBarSquareIcon,
  CheckCircleIcon,
  ExclamationTriangleIcon,
  XCircleIcon,
  ClockIcon,
  MagnifyingGlassIcon,
  PlayIcon,
  CalendarIcon,
  WrenchScrewdriverIcon
} from '@heroicons/react/24/outline'

const API_BASE = '/api/v1'

// Format large numbers with K, M suffixes
const formatNumber = (num) => {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M'
  }
  if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'K'
  }
  return num?.toString() || '0'
}

// Format relative time
const formatTimeAgo = (dateString) => {
  if (!dateString) return 'Never'
  const date = new Date(dateString)
  const now = new Date()
  const diffMs = now - date
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMins / 60)
  const diffDays = Math.floor(diffHours / 24)

  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  if (diffDays < 7) return `${diffDays}d ago`
  return date.toLocaleDateString()
}

function DataQuality({ jobs }) {
  const [summary, setSummary] = useState(null)
  const [qualities, setQualities] = useState([])
  const [loading, setLoading] = useState(false)
  const [loadingQualities, setLoadingQualities] = useState(false)
  const [selectedExchange, setSelectedExchange] = useState('')
  const [selectedStatus, setSelectedStatus] = useState('')
  const [searchTerm, setSearchTerm] = useState('')
  const [sortBy, setSortBy] = useState('quality_status')
  const [sortDir, setSortDir] = useState('asc')

  // Check jobs state
  const [activeChecks, setActiveChecks] = useState([])
  const [recentChecks, setRecentChecks] = useState([])
  const [startingCheck, setStartingCheck] = useState(false)
  const [showChecks, setShowChecks] = useState(false)

  // Get unique exchanges from jobs
  const exchanges = [...new Set(jobs.map(j => j.connector_exchange_id))].filter(Boolean)

  useEffect(() => {
    fetchSummary()
    fetchQualities()
    fetchActiveChecks()
    fetchRecentChecks()
  }, [selectedExchange])

  // Poll for active checks
  useEffect(() => {
    if (activeChecks.length > 0) {
      const interval = setInterval(() => {
        fetchActiveChecks()
        fetchQualities()
        fetchSummary()
      }, 5000) // Poll every 5 seconds
      return () => clearInterval(interval)
    }
  }, [activeChecks.length])

  const fetchSummary = async () => {
    setLoading(true)
    try {
      const params = selectedExchange ? { exchange_id: selectedExchange } : {}
      const response = await axios.get(`${API_BASE}/quality/summary`, { params })
      setSummary(response.data.data)
    } catch (err) {
      console.error('Failed to fetch quality summary:', err)
    } finally {
      setLoading(false)
    }
  }

  const fetchQualities = async () => {
    setLoadingQualities(true)
    try {
      const params = {}
      if (selectedExchange) params.exchange_id = selectedExchange
      if (selectedStatus) params.quality_status = selectedStatus
      const response = await axios.get(`${API_BASE}/quality`, { params })
      setQualities(response.data.data || [])
    } catch (err) {
      console.error('Failed to fetch qualities:', err)
      setQualities([])
    } finally {
      setLoadingQualities(false)
    }
  }

  const fetchActiveChecks = async () => {
    try {
      const response = await axios.get(`${API_BASE}/quality/checks/active`)
      setActiveChecks(response.data.data || [])
    } catch (err) {
      console.error('Failed to fetch active checks:', err)
    }
  }

  const fetchRecentChecks = async () => {
    try {
      const response = await axios.get(`${API_BASE}/quality/checks`, { params: { limit: 10 } })
      setRecentChecks(response.data.data || [])
    } catch (err) {
      console.error('Failed to fetch recent checks:', err)
    }
  }

  const startQualityCheck = async (type, exchangeId = '', symbol = '', timeframe = '') => {
    setStartingCheck(true)
    try {
      await axios.post(`${API_BASE}/quality/check`, {
        type,
        exchange_id: exchangeId,
        symbol,
        timeframe
      })
      // Refresh checks
      await fetchActiveChecks()
      await fetchRecentChecks()
      setShowChecks(true)
    } catch (err) {
      alert('Failed to start quality check: ' + (err.response?.data?.error || err.message))
    } finally {
      setStartingCheck(false)
    }
  }

  const handleRefresh = () => {
    fetchSummary()
    fetchQualities()
    fetchActiveChecks()
    fetchRecentChecks()
  }

  // Filter and sort qualities
  const filteredQualities = qualities
    .filter(q => {
      if (searchTerm) {
        const search = searchTerm.toLowerCase()
        return q.symbol?.toLowerCase().includes(search) ||
               q.exchange_id?.toLowerCase().includes(search) ||
               q.timeframe?.toLowerCase().includes(search)
      }
      return true
    })
    .filter(q => !selectedStatus || q.quality_status === selectedStatus)
    .sort((a, b) => {
      let aVal, bVal
      switch (sortBy) {
        case 'completeness_score':
          aVal = a.completeness_score || 0
          bVal = b.completeness_score || 0
          break
        case 'total_candles':
          aVal = a.total_candles || 0
          bVal = b.total_candles || 0
          break
        case 'gaps_detected':
          aVal = a.gaps_detected || 0
          bVal = b.gaps_detected || 0
          break
        case 'data_age_days':
          aVal = a.data_age_days || 0
          bVal = b.data_age_days || 0
          break
        case 'symbol':
          aVal = a.symbol || ''
          bVal = b.symbol || ''
          return sortDir === 'asc' ? aVal.localeCompare(bVal) : bVal.localeCompare(aVal)
        case 'checked_at':
          aVal = a.checked_at ? new Date(a.checked_at).getTime() : 0
          bVal = b.checked_at ? new Date(b.checked_at).getTime() : 0
          break
        case 'quality_status':
        default:
          const statusOrder = { poor: 0, fair: 1, good: 2, excellent: 3 }
          aVal = statusOrder[a.quality_status] || 0
          bVal = statusOrder[b.quality_status] || 0
      }
      return sortDir === 'asc' ? aVal - bVal : bVal - aVal
    })

  const getStatusColor = (status) => {
    switch (status) {
      case 'excellent': return 'bg-green-100 text-green-800 border-green-300'
      case 'good': return 'bg-blue-100 text-blue-800 border-blue-300'
      case 'fair': return 'bg-yellow-100 text-yellow-800 border-yellow-300'
      case 'poor': return 'bg-red-100 text-red-800 border-red-300'
      default: return 'bg-gray-100 text-gray-800 border-gray-300'
    }
  }

  const getFreshnessColor = (freshness) => {
    switch (freshness) {
      case 'fresh': return 'text-green-600'
      case 'stale': return 'text-yellow-600'
      case 'very_stale': return 'text-red-600'
      default: return 'text-gray-600'
    }
  }

  const getCheckStatusColor = (status) => {
    switch (status) {
      case 'completed': return 'bg-green-100 text-green-800'
      case 'running': return 'bg-blue-100 text-blue-800'
      case 'pending': return 'bg-yellow-100 text-yellow-800'
      case 'failed': return 'bg-red-100 text-red-800'
      default: return 'bg-gray-100 text-gray-800'
    }
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold text-gray-900 flex items-center">
          <ChartBarSquareIcon className="w-7 h-7 mr-2 text-indigo-600" />
          Data Quality Analysis
        </h2>
        <div className="flex items-center space-x-2">
          {/* Start Check Dropdown */}
          <div className="relative inline-block">
            <button
              onClick={() => startQualityCheck('all')}
              disabled={startingCheck || activeChecks.length > 0}
              className="px-4 py-2 bg-indigo-600 text-white rounded-l hover:bg-indigo-700 transition disabled:opacity-50 flex items-center"
              title="Run quality check for all jobs"
            >
              <PlayIcon className="w-5 h-5 mr-1" />
              Check All
            </button>
            <select
              onChange={(e) => {
                if (e.target.value === 'exchange' && selectedExchange) {
                  startQualityCheck('exchange', selectedExchange)
                }
                e.target.value = ''
              }}
              disabled={startingCheck || activeChecks.length > 0}
              className="px-2 py-2 bg-indigo-500 text-white rounded-r border-l border-indigo-400 hover:bg-indigo-600 disabled:opacity-50 cursor-pointer"
            >
              <option value="">...</option>
              {selectedExchange && (
                <option value="exchange">Check {selectedExchange}</option>
              )}
            </select>
          </div>

          {/* Show Checks Button */}
          <button
            onClick={() => setShowChecks(!showChecks)}
            className={`px-3 py-2 rounded transition flex items-center ${
              showChecks ? 'bg-gray-200 text-gray-700' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            } ${activeChecks.length > 0 ? 'ring-2 ring-blue-400' : ''}`}
            title="View quality checks"
          >
            <WrenchScrewdriverIcon className="w-5 h-5 mr-1" />
            Checks
            {activeChecks.length > 0 && (
              <span className="ml-1 px-1.5 py-0.5 text-xs bg-blue-500 text-white rounded-full">
                {activeChecks.length}
              </span>
            )}
          </button>

          <button
            onClick={handleRefresh}
            className="p-2 bg-gray-100 text-gray-700 rounded hover:bg-gray-200 transition"
            title="Refresh"
            disabled={loading}
          >
            <ArrowPathIcon className={`w-5 h-5 ${loading ? 'animate-spin' : ''}`} />
          </button>
        </div>
      </div>

      {/* Active/Recent Checks Panel */}
      {showChecks && (
        <div className="bg-white rounded-lg shadow p-4 mb-6">
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-lg font-semibold text-gray-900 flex items-center">
              <WrenchScrewdriverIcon className="w-5 h-5 mr-2 text-gray-600" />
              Quality Checks
            </h3>
            <button
              onClick={() => setShowChecks(false)}
              className="text-gray-400 hover:text-gray-600"
            >
              <XCircleIcon className="w-5 h-5" />
            </button>
          </div>

          {/* Active Checks */}
          {activeChecks.length > 0 && (
            <div className="mb-4">
              <h4 className="text-sm font-medium text-blue-700 mb-2">Running Checks</h4>
              {activeChecks.map(check => (
                <div key={check.id} className="bg-blue-50 rounded-lg p-3 mb-2 border border-blue-200">
                  <div className="flex justify-between items-center mb-2">
                    <span className="font-medium text-blue-900">
                      {check.type === 'all' ? 'All Jobs' :
                       check.type === 'exchange' ? `Exchange: ${check.exchange_id}` :
                       check.type === 'scheduled' ? 'Scheduled Check' :
                       `${check.symbol} ${check.timeframe}`}
                    </span>
                    <span className={`px-2 py-0.5 text-xs font-medium rounded ${getCheckStatusColor(check.status)}`}>
                      {check.status}
                    </span>
                  </div>
                  <div className="w-full bg-blue-200 rounded-full h-2 mb-1">
                    <div
                      className="bg-blue-600 h-2 rounded-full transition-all duration-500"
                      style={{ width: `${check.progress || 0}%` }}
                    />
                  </div>
                  <div className="flex justify-between text-xs text-blue-700">
                    <span>{check.completed_jobs || 0} / {check.total_jobs} jobs</span>
                    <span>{(check.progress || 0).toFixed(1)}%</span>
                  </div>
                </div>
              ))}
            </div>
          )}

          {/* Recent Checks */}
          <div>
            <h4 className="text-sm font-medium text-gray-700 mb-2">Recent Checks</h4>
            {recentChecks.length === 0 ? (
              <p className="text-sm text-gray-500">No recent quality checks</p>
            ) : (
              <div className="space-y-2 max-h-48 overflow-y-auto">
                {recentChecks.filter(c => c.status !== 'running' && c.status !== 'pending').slice(0, 5).map(check => (
                  <div key={check.id} className="flex justify-between items-center py-2 border-b border-gray-100 last:border-0">
                    <div>
                      <span className="text-sm font-medium text-gray-900">
                        {check.type === 'all' ? 'All Jobs' :
                         check.type === 'exchange' ? `Exchange: ${check.exchange_id}` :
                         check.type === 'scheduled' ? 'Scheduled Check' :
                         `${check.symbol} ${check.timeframe}`}
                      </span>
                      <span className="text-xs text-gray-500 ml-2">
                        {formatTimeAgo(check.completed_at)}
                      </span>
                    </div>
                    <div className="flex items-center space-x-3 text-xs">
                      <span className="text-green-600">{check.excellent_count || 0} excellent</span>
                      <span className="text-blue-600">{check.good_count || 0} good</span>
                      <span className="text-yellow-600">{check.fair_count || 0} fair</span>
                      <span className="text-red-600">{check.poor_count || 0} poor</span>
                      <span className={`px-2 py-0.5 font-medium rounded ${getCheckStatusColor(check.status)}`}>
                        {check.status}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}

      {/* Summary Cards */}
      {summary && (
        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-7 gap-4 mb-6">
          <div className="bg-white rounded-lg shadow p-4 text-center">
            <p className="text-2xl font-bold text-indigo-600">
              {summary.average_completeness?.toFixed(1) || 0}%
            </p>
            <p className="text-xs text-gray-600">Avg Completeness</p>
          </div>
          <div className="bg-white rounded-lg shadow p-4 text-center">
            <p className="text-2xl font-bold text-green-600">{summary.excellent_quality || 0}</p>
            <p className="text-xs text-gray-600">Excellent</p>
          </div>
          <div className="bg-white rounded-lg shadow p-4 text-center">
            <p className="text-2xl font-bold text-blue-600">{summary.good_quality || 0}</p>
            <p className="text-xs text-gray-600">Good</p>
          </div>
          <div className="bg-white rounded-lg shadow p-4 text-center">
            <p className="text-2xl font-bold text-yellow-600">{summary.fair_quality || 0}</p>
            <p className="text-xs text-gray-600">Fair</p>
          </div>
          <div className="bg-white rounded-lg shadow p-4 text-center">
            <p className="text-2xl font-bold text-red-600">{summary.poor_quality || 0}</p>
            <p className="text-xs text-gray-600">Poor</p>
          </div>
          <div className="bg-white rounded-lg shadow p-4 text-center">
            <p className="text-2xl font-bold text-orange-600">{formatNumber(summary.total_gaps || 0)}</p>
            <p className="text-xs text-gray-600">Total Gaps</p>
          </div>
          <div className="bg-white rounded-lg shadow p-4 text-center">
            <p className="text-xs font-medium text-gray-600 flex items-center justify-center">
              <ClockIcon className="w-4 h-4 mr-1" />
              {summary.updated_at ? formatTimeAgo(summary.updated_at) : 'Never'}
            </p>
            <p className="text-xs text-gray-500">Last Updated</p>
          </div>
        </div>
      )}

      {/* Filters */}
      <div className="bg-white rounded-lg shadow p-4 mb-6">
        <div className="flex flex-wrap gap-4 items-center">
          <div className="flex-1 min-w-[200px]">
            <div className="relative">
              <MagnifyingGlassIcon className="w-5 h-5 absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" />
              <input
                type="text"
                placeholder="Search symbol, exchange, timeframe..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
              />
            </div>
          </div>
          <select
            value={selectedExchange}
            onChange={(e) => setSelectedExchange(e.target.value)}
            className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500"
          >
            <option value="">All Exchanges</option>
            {exchanges.map(ex => (
              <option key={ex} value={ex}>{ex}</option>
            ))}
          </select>
          <select
            value={selectedStatus}
            onChange={(e) => setSelectedStatus(e.target.value)}
            className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500"
          >
            <option value="">All Status</option>
            <option value="excellent">Excellent</option>
            <option value="good">Good</option>
            <option value="fair">Fair</option>
            <option value="poor">Poor</option>
          </select>
          <select
            value={`${sortBy}-${sortDir}`}
            onChange={(e) => {
              const [by, dir] = e.target.value.split('-')
              setSortBy(by)
              setSortDir(dir)
            }}
            className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500"
          >
            <option value="quality_status-asc">Sort: Quality (Worst First)</option>
            <option value="quality_status-desc">Sort: Quality (Best First)</option>
            <option value="completeness_score-asc">Sort: Completeness (Low-High)</option>
            <option value="completeness_score-desc">Sort: Completeness (High-Low)</option>
            <option value="gaps_detected-desc">Sort: Gaps (Most First)</option>
            <option value="data_age_days-desc">Sort: Data Age (Oldest First)</option>
            <option value="checked_at-asc">Sort: Last Checked (Oldest First)</option>
            <option value="symbol-asc">Sort: Symbol (A-Z)</option>
          </select>
        </div>
      </div>

      {/* Quality Table */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-semibold text-gray-900">
            Job Data Quality ({filteredQualities.length} jobs)
          </h3>
        </div>

        {loadingQualities ? (
          <div className="text-center py-12">
            <ArrowPathIcon className="w-8 h-8 animate-spin mx-auto text-gray-400" />
            <p className="text-sm text-gray-500 mt-2">Loading quality data...</p>
          </div>
        ) : filteredQualities.length === 0 ? (
          <div className="text-center py-12 text-gray-500">
            <ChartBarSquareIcon className="w-12 h-12 mx-auto text-gray-300 mb-3" />
            <p>No quality data available.</p>
            <p className="text-sm mt-1">Click "Check All" to analyze all jobs.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Exchange</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Symbol</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">TF</th>
                  <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase">Status</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Completeness</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Candles</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Gaps</th>
                  <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase">Freshness</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Data Period</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Checked</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {filteredQualities.map((q, idx) => (
                  <tr key={`${q.exchange_id}-${q.symbol}-${q.timeframe}-${idx}`} className="hover:bg-gray-50">
                    <td className="px-4 py-3 text-sm text-gray-900">{q.exchange_id}</td>
                    <td className="px-4 py-3 text-sm font-medium text-gray-900">{q.symbol}</td>
                    <td className="px-4 py-3 text-sm text-gray-600">{q.timeframe}</td>
                    <td className="px-4 py-3 text-center">
                      <span className={`px-2 py-1 text-xs font-medium rounded border ${getStatusColor(q.quality_status)}`}>
                        {q.quality_status || 'unknown'}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm text-right">
                      <div className="flex items-center justify-end">
                        <div className="w-16 bg-gray-200 rounded-full h-2 mr-2">
                          <div
                            className={`h-2 rounded-full ${
                              (q.completeness_score || 0) >= 95 ? 'bg-green-500' :
                              (q.completeness_score || 0) >= 80 ? 'bg-yellow-500' : 'bg-red-500'
                            }`}
                            style={{ width: `${Math.min(q.completeness_score || 0, 100)}%` }}
                          />
                        </div>
                        <span className="font-mono text-xs">{(q.completeness_score || 0).toFixed(1)}%</span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-sm text-right font-mono">{formatNumber(q.total_candles || 0)}</td>
                    <td className="px-4 py-3 text-sm text-right">
                      {(q.gaps_detected || 0) > 0 ? (
                        <span className="text-orange-600 font-medium">{q.gaps_detected}</span>
                      ) : (
                        <span className="text-green-600">0</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-center">
                      <span className={`text-xs font-medium ${getFreshnessColor(q.data_freshness)}`}>
                        {q.data_freshness === 'fresh' && <CheckCircleIcon className="w-4 h-4 inline mr-1" />}
                        {q.data_freshness === 'stale' && <ClockIcon className="w-4 h-4 inline mr-1" />}
                        {q.data_freshness === 'very_stale' && <ExclamationTriangleIcon className="w-4 h-4 inline mr-1" />}
                        {q.data_freshness || 'unknown'}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-xs text-gray-500">
                      {q.data_period_start && q.data_period_end ? (
                        <div>
                          <div>{new Date(q.data_period_start).toLocaleDateString()}</div>
                          <div>to {new Date(q.data_period_end).toLocaleDateString()}</div>
                          <div className="text-gray-400">
                            ({q.data_period_days || 0}d, {q.data_age_days || 0}d old)
                          </div>
                        </div>
                      ) : (
                        'N/A'
                      )}
                    </td>
                    <td className="px-4 py-3 text-xs text-gray-500">
                      {q.checked_at ? formatTimeAgo(q.checked_at) : 'Never'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}

export default DataQuality
