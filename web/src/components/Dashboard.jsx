import { useMemo } from 'react'

function Dashboard({ connectors, jobs, onRefresh }) {
  const stats = useMemo(() => {
    const activeConnectors = connectors.filter(c => c.status === 'active').length
    const sandboxConnectors = connectors.filter(c => c.sandbox_mode).length
    const activeJobs = jobs.filter(j => j.status === 'active').length
    const pausedJobs = jobs.filter(j => j.status === 'paused').length

    return {
      totalConnectors: connectors.length,
      activeConnectors,
      sandboxConnectors,
      totalJobs: jobs.length,
      activeJobs,
      pausedJobs
    }
  }, [connectors, jobs])

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold text-gray-900">Dashboard</h2>
        <button
          onClick={onRefresh}
          className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition"
        >
          Refresh
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
            <span className="text-yellow-600 font-medium">{stats.sandboxConnectors} sandbox</span>
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

        {/* Status Card */}
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">System Status</p>
              <p className="text-3xl font-bold text-green-600">Healthy</p>
            </div>
            <div className="p-3 bg-green-100 rounded-full">
              <svg className="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
          </div>
          <div className="mt-4 text-sm text-gray-600">
            All systems operational
          </div>
        </div>
      </div>

      {/* Recent Activity */}
      <div className="bg-white rounded-lg shadow">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-semibold text-gray-900">Quick Overview</h3>
        </div>
        <div className="p-6">
          <div className="space-y-4">
            {connectors.length === 0 && jobs.length === 0 ? (
              <p className="text-gray-500 text-center py-8">
                No connectors or jobs configured yet. Get started by creating a connector!
              </p>
            ) : (
              <>
                {connectors.slice(0, 3).map(connector => (
                  <div key={connector.id} className="flex items-center justify-between border-b border-gray-100 pb-3">
                    <div>
                      <p className="font-medium text-gray-900">{connector.display_name}</p>
                      <p className="text-sm text-gray-500">{connector.exchange_id}</p>
                    </div>
                    <div className="flex items-center space-x-2">
                      {connector.sandbox_mode && (
                        <span className="px-2 py-1 text-xs font-medium bg-yellow-100 text-yellow-800 rounded">
                          Sandbox
                        </span>
                      )}
                      <span className={`px-2 py-1 text-xs font-medium rounded ${
                        connector.status === 'active'
                          ? 'bg-green-100 text-green-800'
                          : 'bg-gray-100 text-gray-800'
                      }`}>
                        {connector.status}
                      </span>
                    </div>
                  </div>
                ))}
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default Dashboard
