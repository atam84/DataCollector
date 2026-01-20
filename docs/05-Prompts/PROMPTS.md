# PROMPTS — Data Collector

These prompts are designed so you can paste them into ChatGPT (or your IDE agent) and get concrete deliverables.

## Global constraints (always include)
- One public connector per exchange.
- Pair/timeframe ingestion is modeled as jobs.
- Rate limiting must be enforced per request with shared state (DB).
- Idempotent writes: same candle timestamp cannot create duplicates.
- Logs must include `request_id`, `exchange_id`, `job_id` when applicable.

---

## Prompt — Milestone 1 / Sprint 1.1 (Backend bootstrap)

**Role**: You are a senior Go backend engineer.

**Context**: Build the initial API skeleton for the Data Collector.

**Deliverables**:
1. Folder structure.
2. A minimal Fiber server with routes: `/health`, `/api/v1/connectors`, `/api/v1/jobs`.
3. A `config` package loading env vars.
4. Mongo connection helper with retry.
5. Unit test setup.

**Output format**: Provide a tree + code snippets for each file.

---

## Prompt — Milestone 2 / Sprint 2.1 (Connector + token acquisition)

**Role**: You are implementing a distributed rate limiter.

**Deliverables**:
1. Mongo schema for `connectors` including `rate_limit_state`.
2. Pseudocode + Go implementation for `AcquireToken(exchange_id, cost)` with atomic update.
3. Tests: concurrency (100 goroutines) + period reset.

---

## Prompt — Milestone 3 / Sprint 3.1 (Job scheduler)

**Role**: You are building a reliable scheduler.

**Deliverables**:
1. Mongo schema for `jobs` including `next_run_at`, `lease`, `cursor`.
2. Scheduler loop design (fairness across exchanges, priorities, backoff).
3. Worker design for OHLCV fetch + indicators + write.
4. Failure handling and retries.
