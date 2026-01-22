import { useMemo, useState, useEffect } from 'react'
import axios from 'axios'
import { ArrowPathIcon, BoltIcon, ExclamationTriangleIcon, ChartBarIcon, CircleStackIcon, HeartIcon, CheckCircleIcon, XCircleIcon, ClockIcon } from '@heroicons/react/24/outline'

const API_BASE = '/api/v1'

// Format large numbers with K, M suffixes
const formatNumber = (num) => {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M'
  }
  if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'K'
  }
  return num.toString()
}

function Dashboard({ connectors, jobs, onRefresh, loading }) {
  const [globalStats, setGlobalStats] = useState(null)
  const [loadingStats, setLoadingStats] = useState(false)
  const [healthData, setHealthData] = useState(null)
  const [loadingHealth, setLoadingHealth] = useState(false)

  // Fetch global stats
  useEffect(() => {
    const fetchStats = async () => {
      setLoadingStats(true)
      try {
        const response = await axios.get(`${API_BASE}/stats`)
        setGlobalStats(response.data)
      } catch (err) {
        console.error('Failed to fetch stats:', err)
      } finally {
        setLoadingStats(false)
      }
    }
    fetchStats()
  }, [connectors, jobs]) // Refresh when connectors or jobs change

  // Fetch health data
  useEffect(() => {
    const fetchHealth = async () => {
      setLoadingHealth(true)
      try {
        const response = await axios.get(`${API_BASE}/connectors/health`)
        setHealthData(response.data)
      } catch (err) {
        console.error('Failed to fetch health data:', err)
      } finally {
        setLoadingHealth(false)
      }
    }
    fetchHealth()
  }, [connectors]) // Refresh when connectors change
  const stats = useMemo(() => {
    const activeConnectors = connectors.filter(c => c.status === 'active').length
    const disabledConnectors = connectors.filter(c => c.status === 'disabled').length
    const activeJobs = jobs.filter(j => j.status === 'active').length
    const pausedJobs = jobs.filter(j => j.status === 'paused').length
    const failedJobs = jobs.filter(j => j.run_state?.last_error && j.run_state.last_error.length > 0).length

    // Rate limit stats
    const rateLimitStats = connectors.map(c => {
      const usage = c.rate_limit?.usage || 0
      const limit = c.rate_limit?.limit || 0
      const percentage = limit > 0 ? (usage / limit) * 100 : 0
      return { connector: c, usage, limit, percentage }
    })
    const highUsageConnectors = rateLimitStats.filter(r => r.percentage > 80).length

    return {
      totalConnectors: connectors.length,
      activeConnectors,
      disabledConnectors,
      totalJobs: jobs.length,
      activeJobs,
      pausedJobs,
      failedJobs,
      highUsageConnectors,
      rateLimitStats
    }
  }, [connectors, jobs])

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold text-gray-900">Dashboard</h2>
        <button
          onClick={onRefresh}
          className="p-2 bg-gray-100 text-gray-700 rounded hover:bg-gray-200 transition"
          title="Refresh"
          disabled={loading}
        >
          <ArrowPathIcon className={`w-5 h-5 ${loading ? 'animate-spin' : ''}`} />
        </button>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
        {/* Connectors Card */}
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Total Connectors</p>
              <p className="text-3xl font-bold text-gray-900">{stats.totalConnectors}</p>
            </div>
            <div className="p-3 bg-blue-100 rounded-full">
              <svg className="w-8 h-8 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
          </div>
          <div className="mt-4 flex items-center text-sm">
            <span className="text-green-600 font-medium">{stats.activeConnectors} active</span>
            <span className="mx-2 text-gray-400">•</span>
            <span className="text-gray-600 font-medium">{stats.disabledConnectors} disabled</span>
          </div>
        </div>

        {/* Jobs Card */}
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Total Jobs</p>
              <p className="text-3xl font-bold text-gray-900">{stats.totalJobs}</p>
            </div>
            <div className="p-3 bg-green-100 rounded-full">
              <svg className="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
              </svg>
            </div>
          </div>
          <div className="mt-4 flex items-center text-sm">
            <span className="text-green-600 font-medium">{stats.activeJobs} active</span>
            <span className="mx-2 text-gray-400">•</span>
            <span className="text-gray-600 font-medium">{stats.pausedJobs} paused</span>
          </div>
        </div>

        {/* Rate Limit Status Card */}
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Rate Limits</p>
              <p className={`text-3xl font-bold ${stats.highUsageConnectors > 0 ? 'text-orange-600' : 'text-green-600'}`}>
                {stats.highUsageConnectors > 0 ? `${stats.highUsageConnectors} Warning` : 'All Clear'}
              </p>
            </div>
            <div className={`p-3 rounded-full ${stats.highUsageConnectors > 0 ? 'bg-orange-100' : 'bg-green-100'}`}>
              <BoltIcon className={`w-8 h-8 ${stats.highUsageConnectors > 0 ? 'text-orange-600' : 'text-green-600'}`} />
            </div>
          </div>
          <div className="mt-4 text-sm">
            {stats.highUsageConnectors > 0 ? (
              <span className="text-orange-600 flex items-center">
                <ExclamationTriangleIcon className="w-4 h-4 mr-1" />
                {stats.highUsageConnectors} connector(s) at high usage
              </span>
            ) : (
              <span className="text-gray-600">All rate limits healthy</span>
            )}
          </div>
        </div>

        {/* Data Volume Card */}
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Total Candles</p>
              <p className="text-3xl font-bold text-purple-600">
                {loadingStats ? '...' : formatNumber(globalStats?.data?.total_candles || 0)}
              </p>
            </div>
            <div className="p-3 bg-purple-100 rounded-full">
              <CircleStackIcon className="w-8 h-8 text-purple-600" />
            </div>
          </div>
          <div className="mt-4 text-sm text-gray-600">
            {globalStats?.data ? (
              <span>{globalStats.data.unique_symbols || 0} symbols across {globalStats.data.unique_timeframes || 0} timeframes</span>
            ) : (
              <span>Loading statistics...</span>
            )}
          </div>
        </div>

        {/* Failed Jobs Card */}
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Failed Jobs</p>
              <p className={`text-3xl font-bold ${stats.failedJobs > 0 ? 'text-red-600' : 'text-green-600'}`}>
                {stats.failedJobs}
              </p>
            </div>
            <div className={`p-3 rounded-full ${stats.failedJobs > 0 ? 'bg-red-100' : 'bg-green-100'}`}>
              {stats.failedJobs > 0 ? (
                <ExclamationTriangleIcon className="w-8 h-8 text-red-600" />
              ) : (
                <svg className="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              )}
            </div>
          </div>
          <div className="mt-4 text-sm">
            {stats.failedJobs > 0 ? (
              <span className="text-red-600">Check job errors in Jobs tab</span>
            ) : (
              <span className="text-gray-600">All jobs running normally</span>
            )}
          </div>
        </div>
      </div>

      {/* Data Statistics Overview */}
      {globalStats?.data && (
        <div className="bg-white rounded-lg shadow mb-8">
          <div className="px-6 py-4 border-b border-gray-200">
            <h3 className="text-lg font-semibold text-gray-900 flex items-center">
              <CircleStackIcon className="w-5 h-5 mr-2 text-purple-600" />
              Data Statistics
            </h3>
          </div>
          <div className="p-6">
            <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
              <div className="text-center">
                <p className="text-3xl font-bold text-purple-600">{formatNumber(globalStats.data.total_candles || 0)}</p>
                <p className="text-sm text-gray-600">Total Candles</p>
              </div>
              <div className="text-center">
                <p className="text-3xl font-bold text-blue-600">{globalStats.data.unique_symbols || 0}</p>
                <p className="text-sm text-gray-600">Unique Symbols</p>
              </div>
              <div className="text-center">
                <p className="text-3xl font-bold text-green-600">{globalStats.data.unique_timeframes || 0}</p>
                <p className="text-sm text-gray-600">Timeframes</p>
              </div>
              <div className="text-center">
                <p className="text-3xl font-bold text-orange-600">{globalStats.data.total_chunks || 0}</p>
                <p className="text-sm text-gray-600">Data Chunks</p>
              </div>
            </div>
            {(globalStats.data.oldest_data || globalStats.data.newest_data) && (
              <div className="mt-4 pt-4 border-t border-gray-200 flex justify-between text-sm text-gray-600">
                <div>
                  <span className="font-medium">Oldest Data:</span>{' '}
                  {globalStats.data.oldest_data ? new Date(globalStats.data.oldest_data).toLocaleDateString() : 'N/A'}
                </div>
                <div>
                  <span className="font-medium">Newest Data:</span>{' '}
                  {globalStats.data.newest_data ? new Date(globalStats.data.newest_data).toLocaleDateString() : 'N/A'}
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Rate Limit Overview - Only shown if there are connectors */}
      {connectors.length > 0 && (
        <div className="bg-white rounded-lg shadow mb-8">
          <div className="px-6 py-4 border-b border-gray-200">
            <h3 className="text-lg font-semibold text-gray-900 flex items-center">
              <BoltIcon className="w-5 h-5 mr-2 text-blue-600" />
              Rate Limit Overview
            </h3>
          </div>
          <div className="p-6">
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {stats.rateLimitStats.map(({ connector, usage, limit, percentage }) => (
                <div key={connector.id} className="border border-gray-200 rounded-lg p-4">
                  <div className="flex justify-between items-center mb-2">
                    <span className="font-medium text-gray-900 text-sm">{connector.display_name}</span>
                    <span className={`text-xs px-2 py-0.5 rounded ${
                      percentage > 80 ? 'bg-red-100 text-red-700' :
                      percentage > 50 ? 'bg-yellow-100 text-yellow-700' :
                      'bg-green-100 text-green-700'
                    }`}>
                      {percentage.toFixed(0)}%
                    </span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2 mb-2">
                    <div
                      className={`h-2 rounded-full transition-all ${
                        percentage > 80 ? 'bg-red-500' :
                        percentage > 50 ? 'bg-yellow-500' :
                        'bg-green-500'
                      }`}
                      style={{ width: `${Math.min(percentage, 100)}%` }}
                    />
                  </div>
                  <div className="flex justify-between text-xs text-gray-500">
                    <span>{usage} / {limit} calls</span>
                    <span>Min: {(connector.rate_limit?.min_delay_ms || 3000) / 1000}s</span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Health Monitoring Section */}
      {connectors.length > 0 && (
        <div className="bg-white rounded-lg shadow mb-8">
          <div className="px-6 py-4 border-b border-gray-200">
            <div className="flex justify-between items-center">
              <h3 className="text-lg font-semibold text-gray-900 flex items-center">
                <HeartIcon className="w-5 h-5 mr-2 text-red-500" />
                Connector Health Monitoring
              </h3>
              {healthData?.summary && (
                <div className="flex items-center space-x-4 text-sm">
                  <span className="flex items-center text-green-600">
                    <CheckCircleIcon className="w-4 h-4 mr-1" />
                    {healthData.summary.healthy} healthy
                  </span>
                  {healthData.summary.degraded > 0 && (
                    <span className="flex items-center text-yellow-600">
                      <ExclamationTriangleIcon className="w-4 h-4 mr-1" />
                      {healthData.summary.degraded} degraded
                    </span>
                  )}
                  {healthData.summary.unhealthy > 0 && (
                    <span className="flex items-center text-red-600">
                      <XCircleIcon className="w-4 h-4 mr-1" />
                      {healthData.summary.unhealthy} unhealthy
                    </span>
                  )}
                </div>
              )}
            </div>
          </div>
          <div className="p-6">
            {loadingHealth ? (
              <div className="text-center py-4">
                <ArrowPathIcon className="w-6 h-6 animate-spin mx-auto text-gray-400" />
                <p className="text-sm text-gray-500 mt-2">Loading health data...</p>
              </div>
            ) : healthData?.connectors ? (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {healthData.connectors.map(conn => (
                  <div
                    key={conn.exchange_id}
                    className={`border rounded-lg p-4 ${
                      conn.health_status === 'unhealthy' ? 'border-red-300 bg-red-50' :
                      conn.health_status === 'degraded' ? 'border-yellow-300 bg-yellow-50' :
                      'border-green-300 bg-green-50'
                    }`}
                  >
                    <div className="flex justify-between items-start mb-3">
                      <div>
                        <p className="font-medium text-gray-900">{conn.display_name}</p>
                        <p className="text-xs text-gray-500">{conn.exchange_id}</p>
                      </div>
                      <span className={`px-2 py-0.5 text-xs font-medium rounded ${
                        conn.health_status === 'unhealthy' ? 'bg-red-200 text-red-800' :
                        conn.health_status === 'degraded' ? 'bg-yellow-200 text-yellow-800' :
                        'bg-green-200 text-green-800'
                      }`}>
                        {conn.health_status || 'healthy'}
                      </span>
                    </div>

                    <div className="grid grid-cols-2 gap-2 text-xs">
                      <div className="bg-white bg-opacity-60 rounded p-2">
                        <p className="text-gray-500">Uptime</p>
                        <p className="font-semibold text-gray-900">
                          {conn.uptime_percentage ? conn.uptime_percentage.toFixed(1) : 100}%
                        </p>
                      </div>
                      <div className="bg-white bg-opacity-60 rounded p-2">
                        <p className="text-gray-500">Avg Response</p>
                        <p className="font-semibold text-gray-900">
                          {conn.average_response_ms ? Math.round(conn.average_response_ms) : 0}ms
                        </p>
                      </div>
                      <div className="bg-white bg-opacity-60 rounded p-2">
                        <p className="text-gray-500">Total Calls</p>
                        <p className="font-semibold text-gray-900">{formatNumber(conn.total_calls || 0)}</p>
                      </div>
                      <div className="bg-white bg-opacity-60 rounded p-2">
                        <p className="text-gray-500">Error Rate</p>
                        <p className={`font-semibold ${
                          conn.error_rate_percentage > 10 ? 'text-red-600' :
                          conn.error_rate_percentage > 5 ? 'text-yellow-600' :
                          'text-green-600'
                        }`}>
                          {conn.error_rate_percentage ? conn.error_rate_percentage.toFixed(1) : 0}%
                        </p>
                      </div>
                    </div>

                    {conn.consecutive_failures > 0 && (
                      <div className="mt-2 p-2 bg-red-100 rounded text-xs text-red-700">
                        <ExclamationTriangleIcon className="w-3 h-3 inline mr-1" />
                        {conn.consecutive_failures} consecutive failures
                      </div>
                    )}

                    {conn.last_error && conn.health_status !== 'healthy' && (
                      <div className="mt-2 p-2 bg-gray-100 rounded text-xs text-gray-600 truncate" title={conn.last_error}>
                        Last error: {conn.last_error.substring(0, 50)}...
                      </div>
                    )}

                    <div className="mt-2 pt-2 border-t border-gray-200 flex justify-between text-xs text-gray-500">
                      <span className="flex items-center">
                        <ClockIcon className="w-3 h-3 mr-1" />
                        {conn.last_successful_call
                          ? new Date(conn.last_successful_call).toLocaleTimeString()
                          : 'Never'}
                      </span>
                      <span>{conn.active_job_count} / {conn.job_count} jobs active</span>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-gray-500 text-center py-4">No health data available</p>
            )}
          </div>
        </div>
      )}

      {/* Recent Activity */}
      <div className="bg-white rounded-lg shadow">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-semibold text-gray-900">Connector Details - Quick Overview</h3>
        </div>
        <div className="p-6">
          <div className="space-y-4">
            {connectors.length === 0 && jobs.length === 0 ? (
              <p className="text-gray-500 text-center py-8">
                No connectors or jobs configured yet. Get started by creating a connector!
              </p>
            ) : (
              <>
                {connectors.slice(0, 5).map(connector => {
                  const connectorJobs = jobs.filter(j => j.connector_exchange_id === connector.exchange_id)
                  const activeJobs = connectorJobs.filter(j => j.status === 'active').length
                  const lastRun = connectorJobs
                    .map(j => j.run_state?.last_run_time)
                    .filter(Boolean)
                    .sort((a, b) => new Date(b) - new Date(a))[0]
                  const rateLimitUsage = connector.rate_limit?.limit
                    ? ((connector.rate_limit.usage || 0) / connector.rate_limit.limit * 100).toFixed(1)
                    : 0

                  return (
                    <div key={connector.id} className="border border-gray-200 rounded-lg p-4 hover:shadow-md transition">
                      <div className="flex items-start justify-between mb-3">
                        <div>
                          <p className="font-medium text-gray-900">{connector.display_name}</p>
                          <p className="text-sm text-gray-500">{connector.exchange_id}</p>
                        </div>
                        <span className={`px-2 py-1 text-xs font-medium rounded ${
                          connector.status === 'active'
                            ? 'bg-green-100 text-green-800'
                            : 'bg-gray-100 text-gray-800'
                        }`}>
                          {connector.status}
                        </span>
                      </div>

                      <div className="grid grid-cols-3 gap-4 text-sm">
                        <div>
                          <p className="text-gray-500">Active Jobs</p>
                          <p className="font-semibold text-gray-900">{activeJobs} / {connectorJobs.length}</p>
                        </div>
                        <div>
                          <p className="text-gray-500">Last Execution</p>
                          <p className="font-semibold text-gray-900">
                            {lastRun ? new Date(lastRun).toLocaleDateString() : 'Never'}
                          </p>
                        </div>
                        <div>
                          <p className="text-gray-500">Rate Limit Usage</p>
                          <div className="flex items-center space-x-2">
                            <div className="flex-1 bg-gray-200 rounded-full h-2">
                              <div
                                className={`h-2 rounded-full ${
                                  rateLimitUsage > 80 ? 'bg-red-500' : rateLimitUsage > 50 ? 'bg-yellow-500' : 'bg-green-500'
                                }`}
                                style={{ width: `${rateLimitUsage}%` }}
                              />
                            </div>
                            <span className="text-xs font-medium">{rateLimitUsage}%</span>
                          </div>
                        </div>
                      </div>
                    </div>
                  )
                })}
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default Dashboard
