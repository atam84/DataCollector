# PRD — Data Collector (v3.1)

- **Last updated**: 2026-01-19
- **Supersedes**: PRD-004-DataCollector (original draft)

## 1. Summary

Data Collector ingests market data (candles) from multiple exchanges, enriches it with indicators, and stores it for downstream consumption.

This PRD adopts **one public connector per exchange** and models pair/timeframe ingestion as **jobs**. Rate limiting is **persisted** and enforced **per API request**, enabling safe horizontal scaling.

## 2. Goals

- Collect OHLCV for configured symbols/timeframes across exchanges.
- Support incremental sync (delta fetching) with resumability.
- Compute indicators (RSI/EMA/MACD initial set) on ingested candles.
- Enforce exchange rate limits reliably under concurrency.
- Provide admin UI + APIs to manage connectors and jobs.

## 3. Non-goals

- Trading / order execution.
- Portfolio management.
- Cross-exchange arbitrage logic.

## 4. Personas

- **Admin/Operator**: configures exchanges, jobs, monitors health.
- **Analyst/Developer**: consumes stored OHLCV + indicators.

## 5. Key Concepts

### 5.1 Connector (Exchange)
A connector represents an exchange integration and its credentials/config.

**Invariants**
- Exactly **one public connector per exchange** (per environment).
- Private connectors may exist in future (per user/tenant) but are out of scope.

### 5.2 Job (Symbol + Timeframe)
A job represents an ingestion stream: `(exchange, symbol, timeframe)` plus scheduling and cursor state.

### 5.3 Rate Limit State
Rate limiting is enforced via a token mechanism stored with the connector state. It must be concurrency-safe.

## 6. User Stories

### Connectors
- As an admin, I can create/enable/disable a connector for an exchange.
- As an admin, I can see rate limit configuration and current usage.

### Jobs
- As an admin, I can create jobs for symbol/timeframe under a connector.
- As an admin, I can start/stop a job.
- As an admin, I can backfill a job from a date.
- As an admin, I can see next run time, last run time, and last error.

### Data
- As a developer, I can query stored OHLCV and indicators via API.

## 7. Functional Requirements

### 7.1 Ingestion
- Support candle timeframes: `1m, 5m, 15m, 1h, 4h, 1d, 1w` (initial).
- Delta fetching: only fetch candles after the last stored timestamp.
- Deduplicate by `(exchange, symbol, timeframe, open_time)`.

### 7.2 Indicator computation
- Compute indicators on ingestion (or as a post-step per job run):
  - RSI(6/14/24)
  - EMA(12/26)
  - MACD(12,26,9)

### 7.3 Scheduling
- A job has a schedule policy (fixed interval by timeframe, or cron-like override).
- The scheduler must prevent overlapping runs of the same job.

### 7.4 Rate Limiting
- Rate limits are enforced **per API request**.
- Counters reset based on an exchange-defined **rate_limit_period**.
- On every request, a token must be acquired; if none available, the worker waits until the next reset.

### 7.5 Observability
- Structured logs for every job run.
- Metrics for job duration, errors, request counts, rate-limit waits.

## 8. Data Model (MongoDB)

### 8.1 `connectors` collection
```json
{
  "_id": "ObjectId",
  "exchange_id": "binance",
  "display_name": "Binance",
  "status": "active",
  "created_at": "2026-01-19T10:00:00Z",
  "updated_at": "2026-01-19T10:00:00Z",

  "rate_limit": {
    "limit": 1200,
    "period_ms": 60000,
    "usage": 0,
    "period_start": "2026-01-19T10:00:00Z",
    "last_job_run_time": null
  },

  "credentials_ref": {
    "mode": "env",
    "keys": ["BINANCE_API_KEY", "BINANCE_API_SECRET"]
  }
}
```

### 8.2 `jobs` collection
```json
{
  "_id": "ObjectId",
  "connector_exchange_id": "binance",
  "symbol": "BTC/USDT",
  "timeframe": "1h",
  "status": "active",

  "schedule": {
    "mode": "timeframe",
    "cron": null
  },

  "cursor": {
    "last_candle_time": "2026-01-18T00:00:00Z"
  },

  "run_state": {
    "locked_until": null,
    "last_run_time": null,
    "next_run_time": null,
    "last_error": null,
    "runs_total": 0
  },

  "created_at": "2026-01-19T10:00:00Z",
  "updated_at": "2026-01-19T10:00:00Z"
}
```

### 8.3 `ohlcv` collection
One document per candle:

```json
{
  "exchange_id": "binance",
  "symbol": "BTC/USDT",
  "timeframe": "1h",
  "open_time": "2026-01-18T10:00:00Z",
  "open": 42000.1,
  "high": 42500.0,
  "low": 41800.0,
  "close": 42300.5,
  "volume": 123.45,
  "indicators": {
    "rsi6": 55.2,
    "rsi14": 49.1,
    "ema12": 42123.4,
    "ema26": 41987.6,
    "macd": 135.8,
    "macd_signal": 120.2,
    "macd_hist": 15.6
  },
  "created_at": "2026-01-18T10:00:02Z"
}
```

## 9. APIs (summary)

See `04-API/API-Spec-v1.1.md` for full details.

## 10. Security & Secrets

- Store exchange API keys in environment variables or a secret manager.
- The connector references secrets by key name (no secrets stored in DB).

## 11. Acceptance Criteria

- Creating a connector for an exchange is idempotent (upsert).
- Job runs do not exceed rate limits under parallel workers.
- For a job, only new candles are fetched and stored (no duplicates).
- Indicators are computed and stored for each ingested candle.
- UI can manage connectors and jobs and shows key runtime status.

