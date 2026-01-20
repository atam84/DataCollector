# UI Improvements Plan - Based on User Feedback

## User Requirements Summary

### 1. Dashboard - Quick Overview Enhancements
**Current**: Shows basic connector info
**Required**:
- Number of active jobs per connector
- Last execution time for each connector
- More detailed connector status

### 2. Connectors Page Enhancements
**Current**: Basic connector cards with sandbox toggle
**Required**:
- ✅ Rate limit display with tiny progress bar
- ✅ Rate limit usage visualization
- ✅ "Suspend" button to suspend connector activity
- ✅ Number of jobs attached to connector

### 3. Job Queue Display
**Current**: Not implemented
**Required**:
- ✅ Queue showing upcoming job executions
- ✅ Display next run time for each job
- ✅ Show job priority/order

### 4. Job Execution Features
**Current**: Jobs created but not executing (no worker implemented)
**Required**:
- ✅ Manual "Run" button for each job
- ✅ Execute job on-demand via API
- ⚠️  Automatic execution (requires worker implementation)

---

## Implementation Plan

### Phase 1: Backend API Enhancements (30 minutes)

#### 1.1 Add Manual Job Execution Endpoint
```go
POST /api/v1/jobs/:id/execute
```
- Triggers immediate job execution
- Fetches OHLCV data from exchange
- Stores data in MongoDB
- Returns execution result

#### 1.2 Add Connector Suspend/Resume Endpoints
```go
POST /api/v1/connectors/:id/suspend
POST /api/v1/connectors/:id/resume
```
- Changes connector status to "suspended" or "active"
- Prevents jobs from running when connector is suspended

#### 1.3 Enhance Connector Response with Job Stats
```go
GET /api/v1/connectors
GET /api/v1/connectors/:id
```
Response includes:
- `job_count`: Number of jobs attached
- `active_job_count`: Number of active jobs
- `last_execution`: Last execution timestamp
- `rate_limit_usage`: Current usage percentage

#### 1.4 Add Job Queue Endpoint
```go
GET /api/v1/jobs/queue
```
Returns upcoming jobs sorted by next_run_time

---

### Phase 2: Frontend UI Improvements (45 minutes)

#### 2.1 Dashboard - Quick Overview
**File**: `web/src/components/Dashboard.jsx`

**Changes**:
- Show active job count per connector
- Display last execution time
- Add visual indicators for connector health
- Show rate limit usage

**New UI Elements**:
```jsx
<div className="connector-detail">
  <p>Active Jobs: {connector.active_job_count}</p>
  <p>Last Run: {formatDate(connector.last_execution)}</p>
  <RateLimitBar usage={connector.rate_limit_usage} />
</div>
```

#### 2.2 Connectors Page - Enhanced Cards
**File**: `web/src/components/ConnectorList.jsx`

**Changes**:
- Add rate limit progress bar
- Add Suspend/Resume button
- Show job count badge
- Display rate limit usage percentage

**New UI Elements**:
```jsx
{/* Rate Limit Progress Bar */}
<div className="rate-limit">
  <p className="text-xs">Rate Limit Usage</p>
  <div className="progress-bar">
    <div className="progress" style={{width: `${usage}%`}} />
  </div>
  <p className="text-xs">{available} / {limit} requests</p>
</div>

{/* Suspend Button */}
<button onClick={suspendConnector}>
  {connector.status === 'suspended' ? 'Resume' : 'Suspend'}
</button>

{/* Job Count Badge */}
<span className="badge">{connector.job_count} jobs</span>
```

#### 2.3 Job Queue Component (NEW)
**File**: `web/src/components/JobQueue.jsx`

**Features**:
- Shows upcoming job executions
- Displays next run time for each job
- Shows countdown to next execution
- Sortable by next run time

**UI Design**:
```jsx
<div className="job-queue">
  <h3>Upcoming Executions</h3>
  <table>
    <thead>
      <tr>
        <th>Job</th>
        <th>Symbol</th>
        <th>Next Run</th>
        <th>Countdown</th>
      </tr>
    </thead>
    <tbody>
      {queuedJobs.map(job => (
        <tr>
          <td>{job.id}</td>
          <td>{job.symbol}</td>
          <td>{formatTime(job.next_run_time)}</td>
          <td>{countdown(job.next_run_time)}</td>
        </tr>
      ))}
    </tbody>
  </table>
</div>
```

#### 2.4 Job List - Manual Run Button
**File**: `web/src/components/JobList.jsx`

**Changes**:
- Add "Run Now" button for each job
- Show execution spinner while running
- Display last execution result
- Show execution success/failure status

**New UI Elements**:
```jsx
<button
  onClick={() => executeJob(job.id)}
  disabled={executing}
  className="btn-run"
>
  {executing ? 'Running...' : 'Run Now'}
</button>
```

---

### Phase 3: Job Execution Logic (Backend)

#### 3.1 Job Execution Service
**File**: `internal/service/job_executor.go` (NEW)

**Responsibilities**:
- Fetch OHLCV data from exchange via CCXT
- Validate data
- Store in MongoDB
- Update job run state
- Handle errors and retries

**Key Methods**:
```go
func ExecuteJob(jobID string) (*ExecutionResult, error)
func FetchOHLCVData(job *Job, connector *Connector) ([]OHLCV, error)
func StoreOHLCVData(data []OHLCV) error
```

#### 3.2 OHLCV Repository
**File**: `internal/repository/ohlcv_repository.go` (NEW)

**Methods**:
```go
func Create(ohlcv *OHLCV) error
func BulkInsert(ohlcvList []OHLCV) error
func FindBySymbolAndTimeframe(symbol, timeframe string) ([]OHLCV, error)
```

---

## Implementation Priority

### High Priority (Implement First)
1. ✅ Manual job execution endpoint (POST /api/v1/jobs/:id/execute)
2. ✅ Job execution service (fetch OHLCV + store)
3. ✅ "Run Now" button in Job List
4. ✅ Connector suspend/resume endpoints
5. ✅ Rate limit progress bar in Connectors

### Medium Priority
6. ✅ Enhanced connector stats (job count, last execution)
7. ✅ Dashboard improvements with detailed stats
8. ✅ Job Queue component
9. ✅ Suspend button in Connectors

### Low Priority (Future)
10. ⚠️  Automatic job scheduler (requires worker implementation)
11. ⚠️  Real-time job execution updates (WebSocket)
12. ⚠️  Advanced queue management (priority, reordering)

---

## Technical Details

### Rate Limit Progress Bar
```jsx
const RateLimitBar = ({ limit, available }) => {
  const usage = ((limit - available) / limit) * 100;
  const color = usage > 80 ? 'red' : usage > 50 ? 'yellow' : 'green';

  return (
    <div className="w-full bg-gray-200 rounded-full h-2">
      <div
        className={`h-2 rounded-full bg-${color}-500`}
        style={{ width: `${usage}%` }}
      />
    </div>
  );
};
```

### Job Execution Flow
```
1. User clicks "Run Now" button
2. Frontend calls POST /api/v1/jobs/:id/execute
3. Backend:
   - Validates job and connector
   - Acquires rate limit token
   - Acquires job lock
   - Fetches data from exchange via CCXT
   - Stores OHLCV data in MongoDB
   - Updates job run state
   - Releases lock
4. Frontend shows success/error message
```

### Countdown Timer
```jsx
const useCountdown = (targetTime) => {
  const [countdown, setCountdown] = useState('');

  useEffect(() => {
    const interval = setInterval(() => {
      const diff = new Date(targetTime) - new Date();
      if (diff <= 0) {
        setCountdown('Now');
      } else {
        const minutes = Math.floor(diff / 60000);
        const seconds = Math.floor((diff % 60000) / 1000);
        setCountdown(`${minutes}m ${seconds}s`);
      }
    }, 1000);

    return () => clearInterval(interval);
  }, [targetTime]);

  return countdown;
};
```

---

## API Changes Summary

### New Endpoints
```
POST   /api/v1/jobs/:id/execute        - Execute job manually
POST   /api/v1/connectors/:id/suspend  - Suspend connector
POST   /api/v1/connectors/:id/resume   - Resume connector
GET    /api/v1/jobs/queue               - Get upcoming jobs
```

### Enhanced Responses
```json
// GET /api/v1/connectors/:id
{
  "id": "...",
  "exchange_id": "binance",
  "status": "active",
  "job_count": 5,              // NEW
  "active_job_count": 3,       // NEW
  "last_execution": "2026-...", // NEW
  "rate_limit": {
    "limit": 1200,
    "available_tokens": 800,
    "usage_percent": 33.3      // NEW
  }
}

// POST /api/v1/jobs/:id/execute
{
  "success": true,
  "message": "Job executed successfully",
  "records_fetched": 100,
  "execution_time_ms": 1234,
  "next_run_time": "2026-01-20T05:00:00Z"
}
```

---

## Files to Create/Modify

### Backend (Go)
- ✅ `internal/service/job_executor.go` (NEW)
- ✅ `internal/repository/ohlcv_repository.go` (NEW)
- ✅ `internal/api/handlers/job_handler.go` (MODIFY - add execute, queue)
- ✅ `internal/api/handlers/connector_handler.go` (MODIFY - add suspend/resume)
- ✅ `internal/repository/connector_repository.go` (MODIFY - add stats)
- ✅ `internal/repository/job_repository.go` (MODIFY - add queue method)
- ✅ `cmd/api/main.go` (MODIFY - wire up new endpoints)

### Frontend (React)
- ✅ `web/src/components/Dashboard.jsx` (MODIFY - enhanced stats)
- ✅ `web/src/components/ConnectorList.jsx` (MODIFY - rate bar, suspend)
- ✅ `web/src/components/JobList.jsx` (MODIFY - run button)
- ✅ `web/src/components/JobQueue.jsx` (NEW)
- ✅ `web/src/components/RateLimitBar.jsx` (NEW)
- ✅ `web/src/App.jsx` (MODIFY - add JobQueue)

---

## Testing Checklist

### Backend
- [ ] Test manual job execution endpoint
- [ ] Verify OHLCV data is stored correctly
- [ ] Test suspend/resume connector
- [ ] Verify rate limit enforcement during execution
- [ ] Test job queue endpoint

### Frontend
- [ ] Test "Run Now" button functionality
- [ ] Verify rate limit progress bar displays correctly
- [ ] Test suspend/resume button
- [ ] Verify job queue updates
- [ ] Test countdown timers

---

## Estimated Time

- **Backend API**: 2-3 hours
- **Frontend UI**: 2-3 hours
- **Testing**: 1 hour
- **Total**: 5-7 hours

---

## Next Steps

1. Implement backend job execution service
2. Add new API endpoints
3. Update frontend components
4. Test end-to-end functionality
5. Update documentation

Would you like me to start implementing these improvements?
