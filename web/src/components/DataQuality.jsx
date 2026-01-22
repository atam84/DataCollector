import { useState, useEffect, useCallback, useRef } from 'react'
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
  WrenchScrewdriverIcon,
  EyeIcon,
  ArrowDownIcon,
  ChevronDownIcon
} from '@heroicons/react/24/outline'
import JobDetails from './JobDetails'

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

// Format date range nicely
const formatDateRange = (start, end) => {
  if (!start || !end) return 'N/A'
  const startDate = new Date(start)
  const endDate = new Date(end)
  const opts = { month: 'short', day: 'numeric', year: 'numeric' }
  return `${startDate.toLocaleDateString('en-US', opts)} - ${endDate.toLocaleDateString('en-US', opts)}`
}

function DataQuality({ jobs, connectors }) {
  const [summary, setSummary] = useState(null)
  const [qualities, setQualities] = useState([])
  const [initialLoading, setInitialLoading] = useState(true)
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
  const [showCheckDropdown, setShowCheckDropdown] = useState(false)

  // Quality modal state
  const [selectedQuality, setSelectedQuality] = useState(null)
  const [showModal, setShowModal] = useState(false)
  const [backfillLoading, setBackfillLoading] = useState(false)

  // Job details modal state
  const [selectedJob, setSelectedJob] = useState(null)
  const [showJobModal, setShowJobModal] = useState(false)

  // Ref to track if initial load is done
  const initialLoadDone = useRef(false)

  // Get unique exchanges from jobs
  const exchanges = [...new Set(jobs.map(j => j.connector_exchange_id))].filter(Boolean)

  // Get job from quality data
  const getJobFromQuality = (quality) => {
    return jobs.find(j =>
      j.connector_exchange_id === quality.exchange_id &&
      j.symbol === quality.symbol &&
      j.timeframe === quality.timeframe
    )
  }

  // Get connector from exchange ID
  const getConnector = (exchangeId) => {
    return connectors?.find(c => c.exchange_id === exchangeId)
  }

  // Open job details modal
  const openJobDetails = (quality) => {
    const job = getJobFromQuality(quality)
    if (job) {
      setSelectedJob(job)
      setShowJobModal(true)
    } else {
      alert('Job not found')
    }
  }

  useEffect(() => {
    const loadData = async () => {
      if (!initialLoadDone.current) {
        setInitialLoading(true)
      }
      await Promise.all([
        fetchSummary(false),
        fetchQualities(false),
        fetchActiveChecks(),
        fetchRecentChecks()
      ])
      if (!initialLoadDone.current) {
        setInitialLoading(false)
        initialLoadDone.current = true
      }
    }
    loadData()
  }, [selectedExchange])

  // Poll for active checks - don't trigger loading states
  useEffect(() => {
    if (activeChecks.length > 0) {
      const interval = setInterval(() => {
        fetchActiveChecks()
        fetchRecentChecks()
        fetchQualities(false)
        fetchSummary(false)
      }, 2000) // Poll every 2 seconds for active checks
      return () => clearInterval(interval)
    }
  }, [activeChecks.length])

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (e) => {
      if (showCheckDropdown && !e.target.closest('.check-dropdown-container')) {
        setShowCheckDropdown(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [showCheckDropdown])

  const fetchSummary = async (showLoading = true) => {
    try {
      const params = selectedExchange ? { exchange_id: selectedExchange } : {}
      const response = await axios.get(`${API_BASE}/quality/summary`, { params })
      setSummary(response.data.data)
    } catch (err) {
      console.error('Failed to fetch quality summary:', err)
    }
  }

  const fetchQualities = async (showLoading = true) => {
    try {
      const params = {}
      if (selectedExchange) params.exchange_id = selectedExchange
      if (selectedStatus) params.quality_status = selectedStatus
      const response = await axios.get(`${API_BASE}/quality`, { params })
      setQualities(response.data.data || [])
    } catch (err) {
      console.error('Failed to fetch qualities:', err)
      setQualities([])
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

  const handleRefresh = async () => {
    setInitialLoading(true)
    await Promise.all([
      fetchSummary(false),
      fetchQualities(false),
      fetchActiveChecks(),
      fetchRecentChecks()
    ])
    setInitialLoading(false)
  }

  const openModal = (quality) => {
    setSelectedQuality(quality)
    setShowModal(true)
  }

  const closeModal = () => {
    setSelectedQuality(null)
    setShowModal(false)
  }

  const startGapFill = async (jobId, fillAll = true) => {
    try {
      await axios.post(`${API_BASE}/jobs/${jobId}/quality/fill-gaps`, { fill_all: fillAll })
      alert('Gap fill started. Check job details for progress.')
      closeModal()
    } catch (err) {
      alert('Failed to start gap fill: ' + (err.response?.data?.error || err.message))
    }
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

  // Check if data is insufficient (less than 100 candles is suspicious)
  const isInsufficientData = (q) => q.total_candles < 100

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold text-gray-900 flex items-center">
          <ChartBarSquareIcon className="w-7 h-7 mr-2 text-indigo-600" />
          Data Quality Analysis
        </h2>
        <div className="flex items-center space-x-2">
          {/* Start Check Dropdown */}
          <div className="relative inline-block check-dropdown-container">
            <div className="flex">
              <button
                onClick={() => startQualityCheck('all')}
                disabled={startingCheck || activeChecks.length > 0}
                className="px-4 py-2 bg-indigo-600 text-white rounded-l hover:bg-indigo-700 transition disabled:opacity-50 flex items-center"
                title="Run quality check for all jobs"
              >
                <PlayIcon className="w-5 h-5 mr-1" />
                Check All
              </button>
              <button
                onClick={() => setShowCheckDropdown(!showCheckDropdown)}
                disabled={startingCheck || activeChecks.length > 0}
                className="px-2 py-2 bg-indigo-500 text-white rounded-r border-l border-indigo-400 hover:bg-indigo-600 disabled:opacity-50 flex items-center"
              >
                <ChevronDownIcon className="w-4 h-4" />
              </button>
            </div>
            {showCheckDropdown && (
              <div className="absolute right-0 mt-1 w-56 bg-white rounded-lg shadow-lg border border-gray-200 z-20">
                <div className="py-1">
                  <button
                    onClick={() => {
                      startQualityCheck('all')
                      setShowCheckDropdown(false)
                    }}
                    className="w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 flex items-center"
                  >
                    <PlayIcon className="w-4 h-4 mr-2 text-indigo-600" />
                    Check All Jobs
                  </button>
                  <div className="border-t border-gray-100 my-1"></div>
                  <p className="px-4 py-1 text-xs text-gray-500 font-medium">By Exchange</p>
                  {exchanges.map(ex => (
                    <button
                      key={ex}
                      onClick={() => {
                        startQualityCheck('exchange', ex)
                        setShowCheckDropdown(false)
                      }}
                      className="w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100"
                    >
                      Check {ex}
                    </button>
                  ))}
                </div>
              </div>
            )}
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
            disabled={initialLoading}
          >
            <ArrowPathIcon className={`w-5 h-5 ${initialLoading ? 'animate-spin' : ''}`} />
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
              <h4 className="text-sm font-medium text-blue-700 mb-2 flex items-center">
                <ArrowPathIcon className="w-4 h-4 mr-1 animate-spin" />
                Running Checks ({activeChecks.length})
              </h4>
              {activeChecks.map(check => (
                <div key={check.id} className="bg-blue-50 rounded-lg p-4 mb-2 border border-blue-200">
                  <div className="flex justify-between items-center mb-2">
                    <div>
                      <span className="font-medium text-blue-900">
                        {check.type === 'all' ? 'All Jobs Quality Check' :
                         check.type === 'exchange' ? `Exchange: ${check.exchange_id}` :
                         check.type === 'scheduled' ? 'Scheduled Quality Check' :
                         `${check.symbol} ${check.timeframe}`}
                      </span>
                      {check.started_at && (
                        <span className="text-xs text-blue-600 ml-2">
                          Started {formatTimeAgo(check.started_at)}
                        </span>
                      )}
                    </div>
                    <span className={`px-2 py-0.5 text-xs font-medium rounded flex items-center ${getCheckStatusColor(check.status)}`}>
                      {check.status === 'running' && <ArrowPathIcon className="w-3 h-3 mr-1 animate-spin" />}
                      {check.status}
                    </span>
                  </div>
                  <div className="w-full bg-blue-200 rounded-full h-3 mb-2">
                    <div
                      className="bg-blue-600 h-3 rounded-full transition-all duration-300 flex items-center justify-center"
                      style={{ width: `${Math.max(check.progress || 0, 5)}%` }}
                    >
                      {(check.progress || 0) > 15 && (
                        <span className="text-[10px] text-white font-medium">{(check.progress || 0).toFixed(0)}%</span>
                      )}
                    </div>
                  </div>
                  <div className="flex justify-between text-xs text-blue-700">
                    <span>
                      <strong>{check.completed_jobs || 0}</strong> of <strong>{check.total_jobs || 0}</strong> jobs analyzed
                    </span>
                    <span>
                      {check.current_job && (
                        <span className="text-blue-600">
                          Currently: {check.current_job}
                        </span>
                      )}
                    </span>
                  </div>
                  {/* Show intermediate results */}
                  {(check.excellent_count > 0 || check.good_count > 0 || check.fair_count > 0 || check.poor_count > 0) && (
                    <div className="mt-2 pt-2 border-t border-blue-200 flex space-x-4 text-xs">
                      <span className="text-green-600">{check.excellent_count || 0} excellent</span>
                      <span className="text-blue-600">{check.good_count || 0} good</span>
                      <span className="text-yellow-600">{check.fair_count || 0} fair</span>
                      <span className="text-red-600">{check.poor_count || 0} poor</span>
                    </div>
                  )}
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
            <option value="total_candles-asc">Sort: Candles (Low-High)</option>
            <option value="total_candles-desc">Sort: Candles (High-Low)</option>
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

        {initialLoading ? (
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
                  <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase">Actions</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {filteredQualities.map((q, idx) => (
                  <tr
                    key={`${q.exchange_id}-${q.symbol}-${q.timeframe}-${idx}`}
                    className={`hover:bg-gray-50 ${isInsufficientData(q) ? 'bg-orange-50' : ''}`}
                  >
                    <td className="px-4 py-3 text-sm text-gray-900">{q.exchange_id}</td>
                    <td className="px-4 py-3 text-sm font-medium">
                      <button
                        onClick={() => openJobDetails(q)}
                        className="text-blue-600 hover:text-blue-800 hover:underline cursor-pointer"
                      >
                        {q.symbol}
                      </button>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-600">{q.timeframe}</td>
                    <td className="px-4 py-3 text-center">
                      <div className="flex items-center justify-center space-x-1">
                        <span className={`px-2 py-1 text-xs font-medium rounded border ${getStatusColor(q.quality_status)}`}>
                          {q.quality_status || 'unknown'}
                        </span>
                        {isInsufficientData(q) && (
                          <span className="px-1.5 py-0.5 text-xs bg-orange-200 text-orange-800 rounded" title="Insufficient data">
                            !
                          </span>
                        )}
                      </div>
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
                    <td className="px-4 py-3 text-sm text-right font-mono">
                      <span className={isInsufficientData(q) ? 'text-orange-600 font-semibold' : ''}>
                        {formatNumber(q.total_candles || 0)}
                      </span>
                    </td>
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
                          <div className="font-medium text-gray-700">
                            {formatDateRange(q.data_period_start, q.data_period_end)}
                          </div>
                          <div className="text-gray-400">
                            {q.data_period_days || 0} days coverage
                          </div>
                        </div>
                      ) : (
                        'N/A'
                      )}
                    </td>
                    <td className="px-4 py-3 text-center">
                      <button
                        onClick={() => openModal(q)}
                        className="p-1.5 bg-indigo-100 text-indigo-700 rounded hover:bg-indigo-200 transition"
                        title="View details"
                      >
                        <EyeIcon className="w-4 h-4" />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Quality Detail Modal */}
      {showModal && selectedQuality && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <div className="px-6 py-4 border-b border-gray-200 flex justify-between items-center sticky top-0 bg-white">
              <div>
                <h3 className="text-lg font-semibold text-gray-900">
                  <button
                    onClick={() => {
                      closeModal()
                      openJobDetails(selectedQuality)
                    }}
                    className="text-blue-600 hover:text-blue-800 hover:underline"
                  >
                    {selectedQuality.symbol}
                  </button>
                  <span className="text-gray-500"> ({selectedQuality.timeframe}) - {selectedQuality.exchange_id}</span>
                </h3>
                <p className="text-xs text-gray-500 mt-1">Click symbol to view full job details</p>
              </div>
              <button onClick={closeModal} className="text-gray-400 hover:text-gray-600">
                <XCircleIcon className="w-6 h-6" />
              </button>
            </div>

            <div className="p-6 space-y-6">
              {/* Status Banner */}
              <div className={`p-4 rounded-lg border-2 ${
                selectedQuality.quality_status === 'excellent' ? 'bg-green-50 border-green-300' :
                selectedQuality.quality_status === 'good' ? 'bg-blue-50 border-blue-300' :
                selectedQuality.quality_status === 'fair' ? 'bg-yellow-50 border-yellow-300' :
                'bg-red-50 border-red-300'
              }`}>
                <div className="flex items-center justify-between">
                  <div className="flex items-center">
                    {selectedQuality.quality_status === 'excellent' && <CheckCircleIcon className="w-6 h-6 text-green-600 mr-2" />}
                    {selectedQuality.quality_status === 'good' && <CheckCircleIcon className="w-6 h-6 text-blue-600 mr-2" />}
                    {selectedQuality.quality_status === 'fair' && <ExclamationTriangleIcon className="w-6 h-6 text-yellow-600 mr-2" />}
                    {selectedQuality.quality_status === 'poor' && <ExclamationTriangleIcon className="w-6 h-6 text-red-600 mr-2" />}
                    <div>
                      <span className="text-lg font-bold capitalize">{selectedQuality.quality_status} Quality</span>
                      {isInsufficientData(selectedQuality) && (
                        <span className="ml-2 px-2 py-0.5 text-xs bg-orange-200 text-orange-800 rounded">
                          Insufficient Data
                        </span>
                      )}
                    </div>
                  </div>
                  <span className={`text-2xl font-bold ${
                    (selectedQuality.completeness_score || 0) >= 95 ? 'text-green-600' :
                    (selectedQuality.completeness_score || 0) >= 80 ? 'text-yellow-600' : 'text-red-600'
                  }`}>
                    {(selectedQuality.completeness_score || 0).toFixed(1)}%
                  </span>
                </div>
              </div>

              {/* Metrics Grid */}
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <div className="bg-gray-50 rounded-lg p-3 text-center">
                  <p className="text-2xl font-bold text-indigo-600">{selectedQuality.total_candles?.toLocaleString() || 0}</p>
                  <p className="text-xs text-gray-600">Total Candles</p>
                </div>
                <div className="bg-gray-50 rounded-lg p-3 text-center">
                  <p className="text-2xl font-bold text-blue-600">{selectedQuality.expected_candles?.toLocaleString() || 0}</p>
                  <p className="text-xs text-gray-600">Expected</p>
                </div>
                <div className="bg-gray-50 rounded-lg p-3 text-center">
                  <p className={`text-2xl font-bold ${(selectedQuality.missing_candles || 0) > 0 ? 'text-red-600' : 'text-green-600'}`}>
                    {selectedQuality.missing_candles?.toLocaleString() || 0}
                  </p>
                  <p className="text-xs text-gray-600">Missing</p>
                </div>
                <div className="bg-gray-50 rounded-lg p-3 text-center">
                  <p className={`text-2xl font-bold ${(selectedQuality.gaps_detected || 0) > 0 ? 'text-orange-600' : 'text-green-600'}`}>
                    {selectedQuality.gaps_detected || 0}
                  </p>
                  <p className="text-xs text-gray-600">Gaps</p>
                </div>
              </div>

              {/* Data Period */}
              <div className="bg-indigo-50 rounded-lg p-4 border border-indigo-200">
                <h4 className="text-sm font-semibold text-indigo-900 mb-3 flex items-center">
                  <CalendarIcon className="w-4 h-4 mr-2" />
                  Data Period
                </h4>
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span className="text-indigo-700 font-medium">From:</span>
                    <p className="text-indigo-900">
                      {selectedQuality.data_period_start ? new Date(selectedQuality.data_period_start).toLocaleString() : 'N/A'}
                    </p>
                  </div>
                  <div>
                    <span className="text-indigo-700 font-medium">To:</span>
                    <p className="text-indigo-900">
                      {selectedQuality.data_period_end ? new Date(selectedQuality.data_period_end).toLocaleString() : 'N/A'}
                    </p>
                  </div>
                  <div>
                    <span className="text-indigo-700 font-medium">Coverage:</span>
                    <p className="text-indigo-900">{selectedQuality.data_period_days || 0} days</p>
                  </div>
                  <div>
                    <span className="text-indigo-700 font-medium">Data Freshness:</span>
                    <p className={`font-medium ${getFreshnessColor(selectedQuality.data_freshness)}`}>
                      {selectedQuality.data_freshness || 'unknown'}
                      {selectedQuality.freshness_minutes && (
                        <span className="text-gray-500 font-normal ml-1">
                          ({selectedQuality.freshness_minutes < 60 ? `${selectedQuality.freshness_minutes}m` :
                            selectedQuality.freshness_minutes < 1440 ? `${Math.round(selectedQuality.freshness_minutes / 60)}h` :
                            `${Math.round(selectedQuality.freshness_minutes / 1440)}d`} ago)
                        </span>
                      )}
                    </p>
                  </div>
                </div>
              </div>

              {/* Gaps List */}
              {selectedQuality.gaps && selectedQuality.gaps.length > 0 && (
                <div className="bg-orange-50 rounded-lg p-4 border border-orange-200">
                  <h4 className="text-sm font-semibold text-orange-900 mb-3 flex items-center">
                    <ExclamationTriangleIcon className="w-4 h-4 mr-2" />
                    Detected Gaps ({selectedQuality.gaps.length})
                  </h4>
                  <div className="max-h-40 overflow-y-auto">
                    <table className="min-w-full text-xs">
                      <thead>
                        <tr className="text-orange-700">
                          <th className="text-left py-1">Start</th>
                          <th className="text-left py-1">End</th>
                          <th className="text-right py-1">Missing</th>
                        </tr>
                      </thead>
                      <tbody className="text-gray-700">
                        {selectedQuality.gaps.slice(0, 10).map((gap, idx) => (
                          <tr key={idx} className="border-t border-orange-200">
                            <td className="py-1">{new Date(gap.start_time).toLocaleString()}</td>
                            <td className="py-1">{new Date(gap.end_time).toLocaleString()}</td>
                            <td className="py-1 text-right font-medium text-orange-700">{gap.missing_candles}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                    {selectedQuality.gaps.length > 10 && (
                      <p className="text-xs text-orange-600 mt-2">+ {selectedQuality.gaps.length - 10} more gaps</p>
                    )}
                  </div>
                </div>
              )}

              {/* Actions */}
              <div className="flex justify-between items-center pt-4 border-t border-gray-200">
                <div className="text-xs text-gray-500">
                  Last checked: {selectedQuality.checked_at ? formatTimeAgo(selectedQuality.checked_at) : 'Never'}
                </div>
                <div className="flex space-x-2">
                  {selectedQuality.gaps && selectedQuality.gaps.length > 0 && (
                    <button
                      onClick={() => startGapFill(selectedQuality.job_id, true)}
                      className="px-4 py-2 bg-orange-600 text-white rounded hover:bg-orange-700 transition flex items-center"
                    >
                      <WrenchScrewdriverIcon className="w-4 h-4 mr-2" />
                      Fill Gaps
                    </button>
                  )}
                  <button
                    onClick={closeModal}
                    className="px-4 py-2 bg-gray-200 text-gray-700 rounded hover:bg-gray-300 transition"
                  >
                    Close
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Job Details Modal */}
      {showJobModal && selectedJob && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg w-full max-w-7xl max-h-[95vh] overflow-hidden flex flex-col">
            <div className="p-6 border-b">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-2xl font-bold text-gray-900">{selectedJob.symbol}</h3>
                  <p className="text-sm text-gray-500 mt-1">
                    {selectedJob.timeframe} • {selectedJob.connector_exchange_id}
                  </p>
                </div>
                <button
                  onClick={() => setShowJobModal(false)}
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

export default DataQuality
