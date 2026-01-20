# Architecture — Backend (v1.1)

- **Last updated**: 2026-01-19
- **Goal**: A reliable, rate-limit-aware ingestion platform that can run continuously with recoverability.

## 1. Tech stack

- **Language**: Go
- **API Framework**: Fiber (or net/http; pick one and standardize)
- **Storage**: MongoDB (both configuration and time-series)
- **Queue/Scheduling**: internal scheduler + DB-backed job state (no external MQ required for MVP)
- **Observability**: structured logs + Prometheus metrics

## 2. High-level components

1. **API Service**
   - CRUD for `connectors` and `jobs`
   - Manual triggers: run job now, backfill, pause/resume

2. **Scheduler**
   - Periodically selects runnable jobs (`next_run_at <= now`, `status = active`)
   - Uses DB-backed locking to ensure only one worker executes a job at a time

3. **Worker Pool**
   - Executes ingestion steps:
     1) acquire rate-limit token
     2) fetch OHLCV via exchange adapter
     3) store raw candles
     4) compute indicators
     5) store enriched candles
     6) update job cursor (`last_ts`) and compute `next_run_at`

4. **RateLimit Service**
   - Per-request token acquisition based on connector state stored in MongoDB
   - Ensures global limits are respected across all workers/instances

5. **Exchange Adapters**
   - One adapter per exchange (ccxt-like behavior but native Go)
   - Provides unified methods: `FetchOHLCV`, `Ping`, `ListSymbols`, `RateLimitInfo`

## 3. Data stores

### 3.1 Collections (MongoDB)

- `connectors`
  - one document per exchange (public connector)
- `jobs`
  - one document per ingestion job (symbol + timeframe)
- `candles_{exchange}_{symbol}_{timeframe}` (optional sharding by exchange)
  - raw + enriched datasets
- `job_runs`
  - execution history and error tracking

## 4. Rate limiting model

- Connector stores:
  - `rate_limit` (max requests)
  - `rate_limit_period_ms`
  - `rate_limit_usage`
  - `period_start_at`
- Token acquisition is atomic: increment only if usage < limit, otherwise wait/retry.

See: `03-Implementation/IMPL-Connector-RateLimit-OptionB-v1.1.md`.

## 5. Reliability and recovery

- Job cursor (`cursor.last_ts`) is persisted; restart resumes from last successful timestamp.
- Worker uses retry policy (e.g., exponential backoff with jitter) for transient errors.
- Hard failures set `job.status = error` and emit an alert.

## 6. Security

- Admin API protected (JWT/OAuth2 is acceptable)
- Secrets stored outside the repo (env vars or secret manager)
- Audit logs for connector/job modifications
