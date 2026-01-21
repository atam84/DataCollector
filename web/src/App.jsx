import { useState, useEffect } from 'react'
import axios from 'axios'
import ConnectorList from './components/ConnectorList'
import JobList from './components/JobList'
import Dashboard from './components/Dashboard'
import JobQueue from './components/JobQueue'
import IndicatorsInfo from './components/IndicatorsInfo'

const API_BASE = '/api/v1'

function App() {
  const [activeTab, setActiveTab] = useState('dashboard')
  const [connectors, setConnectors] = useState([])
  const [jobs, setJobs] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  const fetchConnectors = async () => {
    try {
      setLoading(true)
      const response = await axios.get(`${API_BASE}/connectors`)
      setConnectors(response.data.data || [])
      setError(null)
    } catch (err) {
      setError('Failed to fetch connectors: ' + err.message)
    } finally {
      setLoading(false)
    }
  }

  const fetchJobs = async () => {
    try {
      setLoading(true)
      const response = await axios.get(`${API_BASE}/jobs`)
      setJobs(response.data.data || [])
      setError(null)
    } catch (err) {
      setError('Failed to fetch jobs: ' + err.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchConnectors()
    fetchJobs()
  }, [])

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <h1 className="text-3xl font-bold text-gray-900">
            Data Collector Admin
          </h1>
        </div>
      </header>

      {/* Navigation Tabs */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div className="border-b border-gray-200">
          <nav className="-mb-px flex space-x-8">
            <button
              onClick={() => setActiveTab('dashboard')}
              className={`${
                activeTab === 'dashboard'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm`}
            >
              Dashboard
            </button>
            <button
              onClick={() => setActiveTab('connectors')}
              className={`${
                activeTab === 'connectors'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm`}
            >
              Connectors
            </button>
            <button
              onClick={() => setActiveTab('jobs')}
              className={`${
                activeTab === 'jobs'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm`}
            >
              Jobs
            </button>
            <button
              onClick={() => setActiveTab('queue')}
              className={`${
                activeTab === 'queue'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm`}
            >
              Queue
            </button>
            <button
              onClick={() => setActiveTab('indicators')}
              className={`${
                activeTab === 'indicators'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm`}
            >
              Indicators
            </button>
          </nav>
        </div>

        {/* Error Display */}
        {error && (
          <div className="mt-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
            {error}
          </div>
        )}

        {/* Tab Content */}
        <div className="mt-6">
          {activeTab === 'dashboard' && (
            <Dashboard
              connectors={connectors}
              jobs={jobs}
              onRefresh={() => {
                fetchConnectors()
                fetchJobs()
              }}
              loading={loading}
            />
          )}
          {activeTab === 'connectors' && (
            <ConnectorList
              connectors={connectors}
              onRefresh={fetchConnectors}
              loading={loading}
            />
          )}
          {activeTab === 'jobs' && (
            <JobList
              jobs={jobs}
              connectors={connectors}
              onRefresh={fetchJobs}
              loading={loading}
            />
          )}
          {activeTab === 'queue' && (
            <JobQueue
              connectors={connectors}
            />
          )}
          {activeTab === 'indicators' && (
            <IndicatorsInfo />
          )}
        </div>
      </div>
    </div>
  )
}

export default App
