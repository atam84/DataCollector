# TASK-TO-DO — Data Collector

This task plan is aligned with `PRD-DataCollector-v3.1.md` and the connector/job model in `ADR-0001`.

## Milestone 1 — Foundations
**Goal**: repo, CI, code style, base API skeleton.

### Sprint 1.1 — Project bootstrap
- [ ] Create repo structure (`cmd/`, `internal/`, `pkg/`, `web/`, `docs/`).
- [ ] Add linting (golangci-lint) + formatting + commit hooks.
- [ ] Add build pipeline (CI) with unit tests.
- [ ] Add config loader (.env + YAML) and typed config structs.
- [ ] Implement `/health` endpoint.

## Milestone 2 — Core domain (Connectors + Jobs)
**Goal**: store connectors/jobs, run scheduler loop, lock execution.

### Sprint 2.1 — Mongo models + repositories
- [ ] Define Mongo collections: `connectors`, `jobs`, `rate_limits`, `runs`.
- [ ] Implement repositories with indexes:
  - connectors: unique `(exchange_id)`
  - jobs: unique `(exchange_id, symbol, timeframe)`
  - jobs: index `(status, next_run_at)`
- [ ] CRUD endpoints for connectors.
- [ ] CRUD endpoints for jobs.

### Sprint 2.2 — Scheduler + job locking
- [ ] Scheduler selects due jobs (status=active AND next_run_at<=now).
- [ ] Acquire job lock (atomic update + lease).
- [ ] Record `run` document (start time, job_id).
- [ ] Release lock, compute next_run_at, persist run result.

## Milestone 3 — Rate limiting (Option B)
**Goal**: per-request token acquisition shared across workers.

### Sprint 3.1 — Token acquisition
- [ ] Implement `AcquireToken(exchange_id, weight)` with atomic DB update.
- [ ] Enforce rate limit across concurrent workers.
- [ ] Add metrics: denied requests, wait time, tokens used.

## Milestone 4 — Data ingestion engine
**Goal**: fetch OHLCV, normalize, store, compute indicators.

### Sprint 4.1 — Exchange adapter interface
- [ ] Define `ExchangeClient` interface: `FetchOHLCV`, `GetRateLimitWeight`.
- [ ] Implement first exchange (Binance) + mock.
- [ ] Add retry/backoff and error classification.

### Sprint 4.2 — Storage
- [ ] Store OHLCV with upsert by `(exchange_id, symbol, timeframe, ts)`.
- [ ] Add indicator pipeline (RSI/EMA/MACD) as post-step.

## Milestone 5 — Admin UI
**Goal**: manage connectors/jobs and view runs.

### Sprint 5.1 — UI skeleton
- [ ] Login + auth guard.
- [ ] Connectors list/detail.
- [ ] Jobs list/create/edit.

### Sprint 5.2 — Ops visibility
- [ ] Runs table and job history.
- [ ] Rate-limit dashboard panel.

## Milestone 6 — Quality + Ops
**Goal**: tests, observability, deployment.

### Sprint 6.1 — Testing
- [ ] Unit tests for repositories and AcquireToken.
- [ ] Integration tests with Mongo (testcontainer).

### Sprint 6.2 — Observability
- [ ] Structured logging (JSON).
- [ ] Prometheus metrics endpoints.
- [ ] Alerts (rate limit denial spikes, stalled jobs).

### Sprint 6.3 — Deployment
- [ ] Dockerfile + docker-compose.
- [ ] K8s manifests (optional).
