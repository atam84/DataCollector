# Project Overview — Data Collector

## Purpose

Data Collector is a service that continuously collects market data from multiple crypto exchanges (e.g., OHLCV candles), computes indicators, and stores both raw + enriched datasets for downstream use (dashboards, analysis, strategy engines).

## Non-goals

- Executing trades (no order placement).
- News/sentiment ingestion (optional future milestone).
- Providing a full analytics UI (only basic operational and data health UI).

## Users & key flows

- **Admin/Operator**
  - Adds/updates a **Connector** (one per exchange).
  - Creates and manages **Jobs** (symbol + timeframe + schedule/backfill rules).
  - Monitors job state, errors, and data freshness.

- **Downstream Consumer** (another service)
  - Reads stored candles/indicators via API or directly from DB (depending on deployment).

## Core concepts

### Connector (1 per exchange)
A connector represents access to an exchange and owns:
- Exchange identity and status (enabled/disabled).
- Credentials (if needed) and network configuration.
- Rate limiting parameters and state (token bucket window/counters).

### Job (symbol × timeframe)
A job is the unit of scheduled work:
- Target: `exchange_id + symbol + timeframe`.
- Schedule: periodic run plan + optional backfill.
- Cursor/state: last candle timestamp stored, next run time, retries.

### Rate limiting (per request)
All requests to an exchange must acquire a token **atomically** to prevent exceeding rate limits when multiple workers run in parallel.

## Glossary

- **OHLCV**: Open/High/Low/Close/Volume candle.
- **Timeframe**: Candle interval (e.g., 1m, 5m, 1h, 1d).
- **Backfill**: Fetching historical candles before live mode.
- **Cursor**: The last processed timestamp/marker for incremental collection.

