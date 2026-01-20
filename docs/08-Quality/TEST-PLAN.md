# Test Plan

## 1. Unit tests
- Rate limiter token acquisition (happy path, contention, reset behavior)
- Job scheduler (next_run_at calculation)
- OHLCV upsert + dedupe

## 2. Integration tests
- Mock exchange API (deterministic candles)
- Backfill correctness (no gaps, no duplicates)
- Crash recovery (resume from cursor)

## 3. Non-functional
- Load test: N jobs across M exchanges, with rate limit constraints
- Observability: metrics available for tokens, job lag, errors
