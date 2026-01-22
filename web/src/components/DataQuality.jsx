import { useState, useEffect } from 'react'
import axios from 'axios'
import {
  ArrowPathIcon,
  ChartBarSquareIcon,
  CheckCircleIcon,
  ExclamationTriangleIcon,
  XCircleIcon,
  ClockIcon,
  MagnifyingGlassIcon
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

  // Get unique exchanges from jobs
  const exchanges = [...new Set(jobs.map(j => j.connector_exchange_id))].filter(Boolean)

  useEffect(() => {
    fetchSummary()
    fetchQualities()
  }, [selectedExchange])

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

  const handleRefresh = () => {
    fetchSummary()
    fetchQualities()
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
        case 'symbol':
          aVal = a.symbol || ''
          bVal = b.symbol || ''
          return sortDir === 'asc' ? aVal.localeCompare(bVal) : bVal.localeCompare(aVal)
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

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold text-gray-900 flex items-center">
          <ChartBarSquareIcon className="w-7 h-7 mr-2 text-indigo-600" />
          Data Quality Analysis
        </h2>
        <button
          onClick={handleRefresh}
          className="p-2 bg-gray-100 text-gray-700 rounded hover:bg-gray-200 transition"
          title="Refresh"
          disabled={loading}
        >
          <ArrowPathIcon className={`w-5 h-5 ${loading ? 'animate-spin' : ''}`} />
        </button>
      </div>

      {/* Summary Cards */}
      {summary && (
        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4 mb-6">
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
            <p className="text-sm text-gray-500 mt-2">Analyzing data quality...</p>
          </div>
        ) : filteredQualities.length === 0 ? (
          <div className="text-center py-12 text-gray-500">
            No quality data available. Make sure jobs have collected data.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Exchange</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Symbol</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Timeframe</th>
                  <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase">Status</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Completeness</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Candles</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Gaps</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Missing</th>
                  <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase">Freshness</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Date Range</th>
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
                    <td className="px-4 py-3 text-sm text-right">
                      {(q.missing_candles || 0) > 0 ? (
                        <span className="text-red-600 font-medium">{formatNumber(q.missing_candles)}</span>
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
                      {q.oldest_candle && q.newest_candle ? (
                        <>
                          {new Date(q.oldest_candle).toLocaleDateString()}
                          {' - '}
                          {new Date(q.newest_candle).toLocaleDateString()}
                        </>
                      ) : (
                        'N/A'
                      )}
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
