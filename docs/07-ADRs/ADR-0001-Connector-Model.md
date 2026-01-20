# ADR-0001: One Public Connector per Exchange + Per-request Rate Limiting

- **Status**: Accepted
- **Date**: 2026-01-19

## Context

Earlier drafts mixed two approaches:
- **Connector contains trading pairs/timeframes and schedules**, and rate limiting was described as **in-memory**.
- A newer guide proposes **one public connector per exchange**, with **jobs** containing pair/timeframe/schedule, and a **persistent token acquisition** mechanism in the database.

We need one consistent model.

## Decision

1. **Connector = exchange-level configuration object** (exactly one public connector per exchange).
2. **Jobs = periodic tasks** scoped to (exchange, symbol, timeframe) with scheduling + cursor/state.
3. **Rate limiting is enforced per API request** via an atomic `AcquireToken` operation backed by persistent storage (MongoDB), shared by all workers.

## Rationale

- Avoids duplicated connector config per pair/timeframe.
- Makes scheduling and progress tracking explicit and debuggable.
- Prevents exceeding exchange limits when multiple workers run concurrently.

## Consequences

- UI and API must expose **Jobs** as first-class objects.
- Migration: existing connector definitions that include pairs/timeframes must be converted into jobs.
- Rate limit state must be persisted and updated atomically.

## Alternatives considered

- Keep connector = pair/timeframe bundle (rejected: duplication + harder scaling).
- Keep rate limiting in-memory (rejected: not safe with multi-process/multi-node workers).
