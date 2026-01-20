# Implementation Update Guide — One Public Connector per Exchange + Rate-Limit Accounting (per API request)

**Document:** IMPL-GUIDE-ConnectorModel  
**Version:** 1.0  
**Date:** 2026-01-18  
**Scope:** Update the Data Collector implementation to enforce **one public connector per exchange** and implement **exchange-wide rate limiting where usage is incremented per external API request** (Option B).

---

## 0) Goal (What we are adding)

### New Concept
- A **Connector** represents exactly **one exchange** (public access only).
- There must be **at most one connector per exchange** in the system.
- All jobs for that exchange share the same **rate limit window + usage counter**.
- **Rate-limit usage is incremented per API request**, not per job execution.
- Jobs are stored in a separate **`jobs`** collection and scheduled/executed via a queue/worker loop.
- If a job needs multiple requests (pagination/backfill), it must **acquire a token before every request** and be able to **pause/resume** if budget is exhausted.

---

## 1) Data Model (MongoDB)

### 1.1 `connectors` collection (one document per exchange)

A connector is the exchange-wide controller + limiter state.

**Fields (proposed)**

```json
{
  "_id": "ObjectId",
  "exchange_id": "binance",                   // unique (one per exchange)
  "creation_date": "2026-01-17T00:00:00Z",
  "status": "active|paused|deleted",          // connector-wide state
  "notes": "optional",

  "rate_limit_period_ms": 1000,               // window length (ms) (ex: 1000 for per-second)
  "rate_limit": 20,                           // max requests per window
  "rate_limit_usage": 0,                      // current requests used in window
  "rate_limit_window_start": 1700000000000,   // epoch ms
  "rate_limit_reset_at": 1700000001000,       // epoch ms (window_start + period)

  "last_job_run_time": "2026-01-17T10:30:00Z",
  "jobs_count": 0,
  "active_jobs_count": 0,

  "updated_at": "2026-01-17T10:30:00Z"
}
```

**Required indexes**
- **Unique** index on `exchange_id` (enforces 1 connector per exchange)
- Optional: index on `status`

Example (Mongo shell):
```js
db.connectors.createIndex({ exchange_id: 1 }, { unique: true })
db.connectors.createIndex({ status: 1 })
```

---

### 1.2 `jobs` collection (definitions + runtime state)

Jobs represent periodic work items (incremental refresh) and “long tasks” (backfill/gap fill) with resumable cursors.

**Fields (proposed)**

```json
{
  "_id": "ObjectId",
  "exchange_id": "binance",             // join key (or use connector_id)
  "type": "markets_refresh|ohlcv_incremental|ohlcv_backfill|gap_backfill",

  "symbol": "BTC/USDT",
  "timeframe": "1h",

  "status": "active|paused|deleted",
  "state": "idle|queued|running|waiting_rate_limit|success|failed",

  "schedule": {
    "mode": "interval",                 // interval-based scheduling (recommended for tf jobs)
    "interval_ms": 3600000              // typically timeframe_ms (with jitter)
  },
  "next_run_at": 1700003600000,         // epoch ms
  "priority": 50,                        // lower number = higher priority (manual/gap > incremental)

  "cursor": {
    "since_ms": 0,
    "until_ms": null,
    "page": 0,
    "last_ts": 0,
    "done": false
  },

  "stats": {
    "runs": 0,
    "success": 0,
    "fail": 0,
    "last_run_at": null,
    "last_success_at": null,
    "last_error": null
  },

  "lock": {
    "locked_by": null,
    "lock_until": null
  },

  "created_at": "2026-01-17T00:00:00Z",
  "updated_at": "2026-01-17T10:30:00Z"
}
```

**Required indexes**
- Scheduling:
  - compound index on `status`, `state`, `next_run_at`, `priority`
- Uniqueness (recommended):
  - unique partial index for active OHLCV incremental jobs per `{exchange_id, symbol, timeframe, type}`

Examples:
```js
db.jobs.createIndex({ status: 1, state: 1, next_run_at: 1, priority: 1 })
db.jobs.createIndex({ exchange_id: 1, type: 1, symbol: 1, timeframe: 1 })

db.jobs.createIndex(
  { exchange_id: 1, type: 1, symbol: 1, timeframe: 1 },
  { unique: true, partialFilterExpression: { status: "active" } }
)
```

---

## 2) Scheduling Model (How jobs get created)

### 2.1 One Connector per Exchange (public)
- “Create connector” becomes an **idempotent upsert** by `exchange_id`.
- Users enable streams by creating/updating **jobs** linked to that connector.

### 2.2 Jobs per pair/timeframe
For each `{exchange_id, symbol, timeframe}` you typically create:
- `ohlcv_incremental` (active, interval ≈ timeframe_ms)
- `ohlcv_backfill` (stateful until done; then paused/deleted)
Gap jobs are created dynamically.

### 2.3 Period calculation
- Default: `interval_ms = timeframe_ms`
- Add jitter (±5%) to avoid bursts.
- `next_run_at = now + interval_ms + jitter`

---

## 3) Rate Limiting (Option B — per API request)

### 3.1 Rule
Before **any** external exchange API request, worker must acquire a token from the connector.

- Each token increments `rate_limit_usage` by 1
- Budget resets at the end of the `rate_limit_period_ms` window

### 3.2 Consequence
Backfill/pagination may consume many tokens; when budget is exhausted:
- job transitions to `waiting_rate_limit`
- job sets `next_run_at = rate_limit_reset_at`
- job resumes from cursor later

---

## 4) Atomic Token Acquisition (MongoDB-safe)

### 4.1 Connector window fields
Use:
- `rate_limit_window_start`
- `rate_limit_reset_at`
- `rate_limit_usage`

### 4.2 AcquireToken (two atomic attempts)
Attempt A (reset+consume):
- Filter: `exchange_id=X AND rate_limit_reset_at <= now`
- Update: set new window, usage=1

Attempt B (consume in active window):
- Filter: `exchange_id=X AND rate_limit_reset_at > now AND rate_limit_usage < rate_limit`
- Update: `inc(rate_limit_usage, 1)`

If both fail → exhausted. Return `retry_after_ms = rate_limit_reset_at - now` (read connector to compute).

---

## 5) Worker Execution Flow (with rate-limit pauses)

### 5.1 Job selection
Worker pulls the next job with:
- `status=active`
- `state in (idle, queued, waiting_rate_limit)`
- `next_run_at <= now`
Sort by: `priority ASC, next_run_at ASC`

### 5.2 Locking
Lock job with `lock_until` to prevent duplicate execution.

### 5.3 Execution loop (per request)
For each external API call:
1. `AcquireToken(exchange_id)`
2. If OK → execute API call
3. If not OK → set job to `waiting_rate_limit`, set `next_run_at = now + retry_after_ms`, unlock, stop

> Backfill loops must check token for every page request.

---

## 6) Job Types (Recommended semantics)

### 6.1 `markets_refresh`
- interval: 6–24h
- updates markets/timeframes metadata
- token cost: 1+

### 6.2 `ohlcv_incremental`
- interval: timeframe_ms (plus jitter)
- fetch new candles since `last_ts + timeframe_ms`
- usually 1 request per run, but always token-gated

### 6.3 `ohlcv_backfill` (resumable)
- pagination loop from earliest → now
- cursor persisted after each page: `since_ms`, `last_ts`, `page`
- can pause/resume when rate-limited

### 6.4 `gap_backfill`
- created for missing ranges
- high priority
- deletes itself or marks success when done

---

## 7) Required Code Changes (Update existing implementation)

### 7.1 ConnectorManager (new module)
Responsibilities:
- `EnsureConnector(exchange_id, defaults...)`
- `AcquireToken(exchange_id) -> (ok, retry_after_ms)`
- `PauseConnector(exchange_id)` / `ResumeConnector(exchange_id)`
- `UpdateConnectorJobCounts(exchange_id)` (optional periodic maintenance)

### 7.2 Wrap every CCXT call
Every public request must be token-gated:
- `fetchMarkets`
- `fetchOHLCV`
- `fetchTickers` / `fetchTrades` (if enabled)
- any other public endpoint

### 7.3 Convert “big loops” into resumable jobs
For backfill/gap backfill:
- persist cursor after every successful request
- when rate-limited, save cursor and reschedule
- on restart, resume from cursor

### 7.4 API layer changes
- `POST /connectors` becomes idempotent upsert per `exchange_id`
- Pair/timeframe selection becomes job creation:
  - `POST /jobs` / `PUT /jobs/{id}`
- Ensure DB unique index prevents duplicate connector per exchange

### 7.5 Metrics changes
Expose:
- per exchange: `rate_limit_usage`, `reset_at`, saturation count
- jobs: queued/running/waiting_rate_limit
- request counters per exchange + error rates

---

## 8) Migration Plan (Existing system → new model)

### Step 1 — Collections + indexes
- Create `connectors` collection with unique index on `exchange_id`
- Create `jobs` scheduling indexes and optional uniqueness index

### Step 2 — Seed connectors (one per exchange)
- For each existing exchange in config:
  - upsert connector with defaults:
    - `rate_limit_period_ms` (from exchange or default)
    - `rate_limit` (from exchange or default)
    - initialize window fields

### Step 3 — Convert existing periodic tasks into jobs
For each configured `{exchange_id, symbol, timeframe}`:
- Create `ohlcv_incremental` job
- Create `ohlcv_backfill` job if historical missing (or allow user-triggerged backfill)

### Step 4 — Enforce token gating
- Deploy code requiring `AcquireToken()` before any CCXT call
- Validate limiter saturation behavior under load

### Step 5 — Make backfill resumable (cursor)
- Replace non-resumable pagination loops with cursor-based execution
- Validate pause/resume works and no duplicates occur (upsert key)

### Step 6 — Remove legacy limiter logic
- Remove per-job-only limiter (if any)
- Ensure all limiter accounting is exchange-wide via connector

---

## 9) Acceptance Criteria Checklist

### One connector per exchange
- [ ] DB unique index prevents duplicates
- [ ] “create connector” is idempotent and returns existing connector for same exchange_id

### Rate limiting per request
- [ ] Every CCXT request calls `AcquireToken()` first
- [ ] `rate_limit_usage` never exceeds `rate_limit` within window under concurrency
- [ ] `rate_limit_usage` resets when window expires

### Resumable backfill
- [ ] Backfill pauses on rate limit and resumes later from cursor
- [ ] No duplicate candles (upsert key is enforced)
- [ ] Backfill reaches earliest available history (subject to exchange limits)

### Job scheduling + priorities
- [ ] Incremental runs at timeframe cadence
- [ ] Gap backfill preempts incremental (priority lower number)
- [ ] waiting_rate_limit jobs schedule resume at reset time

---

## 10) Recommended Enhancements (Optional but valuable)

- Add **priority queue** behavior explicitly in job selection query
- Add **jitter** to job schedules to reduce bursts
- Add a **global system limiter** (optional) to protect your own infra
- Add “per-exchange max concurrent jobs” (e.g., 1–2) to reduce thrash

---

## Appendix A — Minimal API shapes (if UI/API exists)

### Connectors
- Create/Upsert connector
  - `POST /api/v1/connectors` -> `{ "exchange_id": "binance" }`

### Jobs
- Create job
  - `POST /api/v1/jobs` -> `{ "exchange_id": "binance", "type": "ohlcv_incremental", "symbol": "BTC/USDT", "timeframe": "1h", "schedule": { "mode": "interval", "interval_ms": 3600000 } }`
- Pause job
  - `POST /api/v1/jobs/{job_id}/pause`
- Resume job
  - `POST /api/v1/jobs/{job_id}/resume`
- Run now (manual trigger)
  - `POST /api/v1/jobs/{job_id}/run`

---
