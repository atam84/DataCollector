# Implementation Guide — Connector Model + Rate Limit (Option B) (v1.1)

- **Last updated**: 2026-01-19
- **Related**: `07-ADRs/ADR-0001-Connector-Model.md`

## 1. Target behavior

1. Only **one** public connector per exchange exists.
2. Pair/timeframe ingestion is managed by **jobs**.
3. Rate limiting is enforced **per request**, using a shared state in the database (atomic acquisition of a token).
4. Jobs must never exceed the exchange rate limit; the scheduler must honor token availability.

## 2. Data model

### 2.1 connectors collection

```json
{
  "_id": "binance",                // exchange_id
  "exchange_id": "binance",
  "creation_date": "2026-01-19T00:00:00Z",
  "status": "active",             // active | paused | disabled
  "rate_limit_period_ms": 60000,
  "rate_limit": 1200,
  "rate_limit_usage": 0,
  "rate_limit_period_start": "2026-01-19T00:00:00Z",
  "last_job_run_time": "2026-01-19T00:00:00Z",
  "jobs_count": 0,
  "active_jobs_count": 0
}
```

### 2.2 jobs collection

```json
{
  "_id": "binance:BTC/USDT:1m:ohlcv",
  "exchange_id": "binance",
  "job_type": "ohlcv",
  "symbol": "BTC/USDT",
  "timeframe": "1m",
  "enabled": true,
  "priority": 50,
  "next_run_at": "2026-01-19T00:00:00Z",
  "locked_until": null,
  "cursor": {
    "since_ms": 1705622400000
  },
  "stats": {
    "runs": 0,
    "failures": 0,
    "last_run_at": null,
    "last_success_at": null,
    "last_error": null
  }
}
```

## 3. Rate-limit token acquisition (atomic)

### 3.1 Rule

A job **must** call `AcquireToken(exchange_id)` **before each upstream API call**. If no token is available, the job must either:
- sleep until the next reset moment, or
- yield and reschedule itself (preferred for distributed workers).

### 3.2 Pseudocode

```text
function AcquireToken(exchange_id):
  now = utc_now()
  connector = connectors.findOne({_id: exchange_id})

  // If period expired, reset usage atomically
  if now - connector.rate_limit_period_start >= connector.rate_limit_period_ms:
      connectors.updateOne(
        {_id: exchange_id, rate_limit_period_start: connector.rate_limit_period_start},
        {$set: {rate_limit_period_start: now, rate_limit_usage: 0}}
      )

  // Atomically increment usage if under limit
  res = connectors.findOneAndUpdate(
    {_id: exchange_id, status: 'active', rate_limit_usage: {$lt: rate_limit}},
    {$inc: {rate_limit_usage: 1}},
    {returnDocument: 'after'}
  )

  if res is null:
      return {ok:false, retry_at: connector.rate_limit_period_start + period_ms}

  return {ok:true}
```

## 4. Scheduler behavior

- `next_run_at` is computed from `timeframe` (e.g., 1m jobs run every 60s, 1h every 3600s) with jitter.
- Before running a job, acquire a short lock:

```text
findOneAndUpdate(
  {_id: job_id, enabled:true, next_run_at:{$lte:now}, locked_until:{$lte:now_or_null}},
  {$set:{locked_until: now+lock_ttl}}
)
```

- Job does work, updates cursor, sets `next_run_at`, clears lock.

## 5. Migration / correction steps (from earlier drafts)

1. Remove pair/timeframe arrays from the connector definition.
2. Create jobs from existing connector configs (one job per pair x timeframe x type).
3. Replace in-memory limiter with DB-based `AcquireToken`.
4. Update UI: connector = exchange; jobs manage pairs/timeframes.

## 6. Acceptance checks

- Creating connector `binance` twice updates the same document (no duplicates).
- With 2 workers, rate_limit_usage never exceeds rate_limit within a period.
- A paused connector prevents job execution.
- Cursor persists across restarts.
