# PRD-004: Data Collector Service (V3 — Consolidated + Enhanced)

## Document Information

| Field | Value |
|---|---|
| **Document ID** | PRD-004 |
| **Version** | 3.0 |
| **Status** | Consolidated (V1 + corrected V2) |
| **Created** | 2026-01-17 |
| **Last Updated** | 2026-01-17 |

## What’s new in V3 (vs V2)
- Re-added the strongest “product PRD” content from V1 (Problem Statement, User Stories, UI notes, extended FR coverage).
- Kept all V2 corrections (pagination backfill loop, incremental `since`, dedupe/upsert keys, gap detection/backfill, rate limiting + circuit breaker, purge confirmation).
- Expanded indicators into **Core (P0)** vs **Extended (P1)** while keeping the structured V2 catalog tables.
- Reintroduced missing NFR targets from V1 (chart rendering, export generation) and security NFRs.

---

## 1. Executive Summary

The Data Collector Service is a core component responsible for aggregating, storing, and enriching cryptocurrency **market data** from multiple exchanges. It collects **OHLCV** from public exchange APIs, supports historical backfill (earliest available) and continuous incremental refresh, computes and stores technical indicators, and provides REST + WebSocket APIs for exploration, charting, and exports.

### Key Capabilities
- Multi-exchange data collection via **CCXT-Go** (with defined fallback strategy if coverage is incomplete)
- Historical backfill with **proper pagination loop**
- Incremental updates with **idempotent upsert/dedup**
- **Gap detection** and prioritized **gap backfill**
- Pre-computed indicator library (Trend / Momentum / Volatility / Volume)
- Rate-limit compliant scheduling + retry/backoff + circuit breaker
- Data exploration + overlay indicators + export jobs (CSV/JSON; optional Parquet)

---

## 2. Problem Statement & User Needs

### 2.1 Current Challenges
1. Fragmented exchange APIs, data formats, and per-exchange rate limits
2. Historical data access requires pagination and gap handling to be reliable
3. Rate-limit violations cause temporary bans and incomplete datasets
4. Data quality issues (missing candles, timestamp drift, outliers) must be detected/handled
5. Computing indicators at scale is expensive if done on-demand only

### 2.2 User Needs
- Single interface to access data from multiple exchanges
- Reliable historical depth for research/backtesting
- Near-real-time data for monitoring (and later trading)
- Pre-computed indicators to reduce client load and keep results consistent
- Export-ready datasets for external analysis tools

---

## 3. Goals, Scope, Non-Goals

### 3.1 Goals
- **Comprehensive coverage**: support exchanges/pairs validated for OHLCV collection
- **Historical depth**: backfill from earliest available timestamp per pair/timeframe
- **Dual update modes**: scheduled polling everywhere; WebSocket streaming where available
- **Data enrichment**: compute indicators automatically; allow manual recomputation
- **Reliability**: respect exchange rate limits; robust retries; ensure integrity and continuity
- **Usability**: UI for connector management, exploration, charting, and export

### 3.2 Scope (Public Data Only)
The collector **only uses public endpoints** (no auth):
- Markets / instruments (pairs, metadata)
- OHLCV for charting (primary)
- Optional: tickers (light retention), trade history (aggregated or raw with retention)
- Optional: WebSocket streaming (where exchange supports it reliably)

### 3.3 Non-Goals
- No private endpoints (balances, orders, positions)
- No execution/trading logic (collector is data-only)
- No default persistence for high-volume data (order books) unless explicitly enabled

---

## 4. Success Metrics

| Metric | Target | Measurement |
|---|---:|---|
| System Uptime | 99.5% | Prometheus/Grafana |
| Data Freshness (polling) | < 5 minutes lag | last candle timestamp |
| Data Freshness (streaming) | < 5 seconds lag | WS latency |
| Gap Detection Accuracy | 99.9% | detected vs expected |
| Backfill Completion Rate | 100% to exchange limit | metadata tracking |
| API Response Time (p95) | < 200ms | APM metrics |
| Indicator Computation Lag | < 30 seconds | after new candle stored |

---

## 5. Solution Overview

### 5.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              DATA COLLECTOR                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐                   │
│  │   Frontend   │◄──►│   Backend    │◄──►│  Exchanges   │                   │
│  │   (React)    │    │   (Golang)   │    │   (CCXT)     │                   │
│  └──────────────┘    └──────────────┘    └──────────────┘                   │
│         │                   │                                                │
│         │                   ▼                                                │
│         │           ┌──────────────┐    ┌──────────────┐                    │
│         │           │   MongoDB    │    │  PostgreSQL  │                    │
│         │           │ (Time-Series)│    │   (Config)   │                    │
│         │           └──────────────┘    └──────────────┘                    │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 5.2 Core Components

| Component | Technology | Purpose |
|---|---|---|
| Backend API | Go (Fiber/Gin) | REST + WebSocket, authn/z (admin), orchestration |
| Data Engine | Go + CCXT-Go | exchange connectivity, OHLCV fetch, markets/timeframes |
| Worker System | Go routines / queue | backfill jobs, incremental jobs, gap backfill, indicators |
| Time-Series DB | MongoDB | OHLCV + indicators + metadata (fast range queries) |
| Config DB | PostgreSQL | connectors, schedules, indicator configs, audit |
| Observability | Prometheus/Grafana | metrics, alerts, tracing/logs |

### 5.3 CCXT Support Strategy (Realistic Coverage)
- Primary: **CCXT-Go**
- Validation required (see §7.1)
- Optional fallback (only if needed): internal sidecar using official Node CCXT for missing exchanges/features

---

## 6. Data Scope & Retention Policy

### P0 — Must Persist (Core Data)
| Data Type | Description | Retention |
|---|---|---|
| OHLCV | candle series | indefinite |
| Markets / Instruments | symbols, precision/limits if available | refresh daily / store latest |
| Collection Metadata | first/last ts, counts, last fetch time, gaps | indefinite |
| Indicators (Core set) | per candle, params-hashed | indefinite |

### P1 — Optional Persist (with retention)
| Data Type | Description | Suggested Retention |
|---|---|---|
| Tickers | sampling snapshots | 30 days |
| Trades | aggregated or raw | 90 days |

### P2 — Cache Only (default)
| Data Type | Notes |
|---|---|
| Order book snapshots | store only if explicitly enabled (high volume) |
| WS tickers | stream/cache for UI; no default persistence |

---

## 7. Detailed Requirements

### 7.1 Exchange Support & Validation (P0)
An exchange is considered “supported” only if it passes:
1. `fetchMarkets()` returns a valid market list
2. timeframes are discoverable (via CCXT or known list) and validated
3. `fetchOHLCV(symbol, timeframe, since, limit)` returns valid candles for at least one symbol

Backend must expose capabilities:
- supported timeframes
- max OHLCV limit per call (if known)
- rate limit (`exchange.rateLimit`) and any overrides
- supported symbols/pairs

### 7.2 Connector Management (P0/P1/P2)
Connectors define:
- `name`, `exchange_id`, `pairs[]`, `timeframes[]`, `schedule` (cron), `status`
- optional: indicator overrides, retention overrides

**P0**
- Create, list, update connectors
- Start/stop (pause without losing data)
- Manual refresh (incremental) + manual backfill

**P1**
- Delete connector definition but keep collected data (default behavior)
- “Dangerous purge” that deletes data + metadata with explicit confirmation

**P2 (good V1 extras kept)**
- Clone connector
- Bulk operations: start/stop/delete for multiple connectors

### 7.3 Data Collection (P0)
#### 7.3.1 Polling (REST) Collection
- Default baseline update mode for all connectors.
- Incremental updates run on schedule and can be triggered manually.

#### 7.3.2 Backfill (Earliest → Now) with Pagination Loop
Backfill must iterate with `since` and stop conditions (no one-shot call).
- Fetch candles with `(since, limit)`
- Upsert/dedupe by unique key
- Advance `since = last_ts + timeframe_ms`
- Stop when:
  - returned candles < limit
  - reached “now” (latest closed candle)
  - exchange boundary (no older candles)
  - `max_pages` safety cap

#### 7.3.3 Incremental Updates (Since Last)
- Start from `metadata.last_ts + timeframe_ms` to avoid re-fetching last candle.
- Upsert/dedupe still required (exchanges sometimes overlap).

#### 7.3.4 Dedupe / Idempotency
Unique key per candle:
`(exchange_id, symbol, timeframe, candle_timestamp)`  
Write strategy: **upsert**.

#### 7.3.5 Gap Detection + Gap Backfill
- Detect gaps by comparing consecutive timestamps to expected interval (tolerance allowed).
- Store detected gaps and schedule backfill jobs.
- Priority: **gap backfill > scheduled incremental**.

#### 7.3.6 WebSocket Streaming (P0 where available)
- If the exchange supports streaming reliably, allow WS-based updates to reduce lag.
- Streaming is additive; polling remains baseline for consistency.

### 7.4 Scheduling & Rate Limiting (P0)
- Cron-based scheduling per connector, persisted across restarts.
- Per-exchange token bucket limiter (load default from `exchange.rateLimit`).
- Priority queue: manual user-triggered > gap backfill > scheduled incremental.
- Exponential backoff on rate limit and transient errors.

**Circuit breaker (P0)**
- Open after N consecutive failures (default 5)
- Half-open after cooldown (default 60s)
- Close after M successes (default 3)

### 7.5 Data Storage & Metadata (P0)
- Store OHLCV organized by exchange/pair/timeframe.
- Track metadata: first_ts, last_ts, candle_count, expected_count, pages_fetched, last_fetch_at, last_success_at.
- Retention policies supported for P1 optional datasets.
- Optional cold storage archival for old optional datasets (future/P3).

### 7.6 Indicators / Enrichment (P0/P1/P2)
**Indicator architecture (P0)**
- Backend computes and stores indicators.
- Frontend selects which indicators to display.

Storage key:
`(exchange_id, symbol, timeframe, timestamp, indicator_name, params_hash)`

**Triggers**
- Auto: after new candles stored
- Manual: API endpoint triggers recomputation
- Batch: after backfill completion

**Configuration**
- Global defaults + per-connector overrides (params stored & hashed).

**Future (P2)**
- Custom indicator definitions (user-defined expressions)
- Plug-in system for custom computations

### 7.7 Data Exploration & Export (P0/P2/P3)
**P0**
- Browse OHLCV by exchange/pair/timeframe
- Interactive charting with zoom/pan
- Overlay indicators
- Export CSV and JSON with date range filtering

**P2**
- Export Parquet
- Additional derived datasets (returns, log returns, z-scores)
- Export templates (OHLCV only, OHLCV+Indicators)

**P3**
- Scheduled exports (cron) and delivery to storage/endpoint

### 7.8 Monitoring & Alerts (P0/P1/P2)  *(restored from V1)*
**P0**
- Dashboard showing connector status and last successful sync
- System stats endpoint and basic metrics

**P1**
- Real-time job progress indicators
- Alert on collection failure
- Alert on detected gaps

**P2**
- Exchange health/latency monitoring
- SLO dashboards for freshness and WS latency

### 7.9 Security (NFR)  *(restored from V1)*
- No exchange credentials stored (public API only)
- Input validation and sanitization on all endpoints
- Internal API abuse protection / rate limiting for admin endpoints

---

## 8. Indicator Catalog (Core vs Extended)

### 8.1 General Notes
- Default source: `close` unless indicator requires H/L/C/V
- Warmup period: return `null` until enough candles exist
- All indicators support configurable parameters with sensible defaults

### 8.2 Trend Indicators

| Indicator | Parameters | Inputs | Outputs | Notes |
|---|---|---|---|---|
| **SMA** | `length=20` | close | `sma` | Core |
| **EMA** | `length=20` | close | `ema` | Core |
| **DEMA** | `length=20` | close | `dema` | Core |
| **TEMA** | `length=20` | close | `tema` | Core |
| **WMA** | `length=20` | close | `wma` | Core |
| **HMA** | `length=20` | close | `hma` | Core |
| **VWMA** | `length=20` | close, volume | `vwma` | Core |
| **ADX/DMI** | `length=14` | high, low, close | `adx`, `plus_di`, `minus_di` | Core |
| **SuperTrend** | `atr_length=10`, `multiplier=3` | high, low, close | `supertrend`, `direction` | Core |
| **Ichimoku** | `tenkan=9`, `kijun=26`, `senkou_b=52`, `displacement=26` | high, low, close | `tenkan_sen`, `kijun_sen`, `senkou_a`, `senkou_b`, `chikou` | Extended |
| **KAMA** | `length=10`, `fast=2`, `slow=30` | close | `kama` | Extended |
| **PSAR** | `step=0.02`, `max=0.2` | high, low | `psar` | Extended |
| **Aroon** | `length=25` | high, low | `aroon_up`, `aroon_down` | Extended |

### 8.3 Momentum Indicators

| Indicator | Parameters | Inputs | Outputs | Notes |
|---|---|---|---|---|
| **RSI** | `length=14` | close | `rsi` | Core |
| **Stochastic** | `k_period=14`, `d_period=3`, `smooth=3` | high, low, close | `stoch_k`, `stoch_d` | Core |
| **MACD** | `fast=12`, `slow=26`, `signal=9` | close | `macd`, `macd_signal`, `macd_hist` | Core |
| **ROC** | `length=12` | close | `roc` | Core |
| **CCI** | `length=20` | high, low, close | `cci` | Core |
| **Williams %R** | `length=14` | high, low, close | `willr` | Core |
| **Momentum** | `length=10` | close | `momentum` | Core |
| **TRIX** | `length=15` | close | `trix` | Extended |
| **Ultimate Osc** | `(7,14,28)` | high, low, close | `ultosc` | Extended |
| **PPO** | `fast=12`, `slow=26`, `signal=9` | close | `ppo`, `ppo_signal`, `ppo_hist` | Extended |

### 8.4 Volatility Indicators

| Indicator | Parameters | Inputs | Outputs | Notes |
|---|---|---|---|---|
| **Bollinger Bands** | `length=20`, `std_dev=2` | close | `bb_mid`, `bb_upper`, `bb_lower`, `bb_width`, `bb_percent_b` | Core |
| **ATR** | `length=14` | high, low, close | `atr` | Core |
| **Keltner Channels** | `ema_length=20`, `atr_length=10`, `multiplier=2` | high, low, close | `kc_mid`, `kc_upper`, `kc_lower` | Core |
| **Donchian Channels** | `length=20` | high, low | `dc_upper`, `dc_lower`, `dc_mid` | Core |
| **StdDev** | `length=20` | close | `stdev` | Core |
| **NATR** | `length=14` | high, low, close | `natr` | Extended |
| **HV** | `length=20/50` | close | `hv` | Extended |

### 8.5 Volume Indicators

| Indicator | Parameters | Inputs | Outputs | Notes |
|---|---|---|---|---|
| **OBV** | none | close, volume | `obv` | Core |
| **VWAP** | rolling `length=20` or session | high, low, close, volume | `vwap` | Core |
| **MFI** | `length=14` | high, low, close, volume | `mfi` | Core |
| **CMF** | `length=20` | high, low, close, volume | `cmf` | Core |
| **Volume SMA** | `length=20` | volume | `vol_sma` | Core |
| **Volume EMA** | `length=20` | volume | `vol_ema` | Core |
| **ADL** | none | high, low, close, volume | `adl` | Extended |
| **VPT** | none | close, volume | `vpt` | Extended |

---

## 9. API Specification

### 9.1 Connector Endpoints

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/connectors` | Create connector |
| GET | `/api/v1/connectors` | List connectors |
| GET | `/api/v1/connectors/{id}` | Get connector details |
| PUT | `/api/v1/connectors/{id}` | Update connector |
| DELETE | `/api/v1/connectors/{id}` | Delete connector definition (keeps data) |
| POST | `/api/v1/connectors/{id}/purge` | Purge connector data (dangerous) |
| POST | `/api/v1/connectors/{id}/start` | Start collection |
| POST | `/api/v1/connectors/{id}/stop` | Stop collection |
| POST | `/api/v1/connectors/{id}/refresh` | Trigger incremental fetch |
| POST | `/api/v1/connectors/{id}/backfill` | Trigger full backfill |
| POST | `/api/v1/connectors/{id}/clone` | Clone connector (P2) |
| POST | `/api/v1/connectors/bulk` | Bulk actions (P2) |

**Purge confirmation payload example**
```json
{"confirm":"PURGE-<connector_name>"}
```

### 9.2 Exchange Endpoints

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/exchanges` | List exchanges |
| GET | `/api/v1/exchanges/{id}` | Exchange info |
| GET | `/api/v1/exchanges/{id}/pairs` | Trading pairs |
| GET | `/api/v1/exchanges/{id}/timeframes` | Supported timeframes |
| GET | `/api/v1/exchanges/{id}/capabilities` | Limits, rate limits, etc |
| GET | `/api/v1/exchanges/{id}/status` | Health status |

### 9.3 Data Endpoints

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/data/ohlcv` | Query OHLCV |
| GET | `/api/v1/data/indicators` | Query indicators |
| GET | `/api/v1/data/coverage` | Coverage info |
| GET | `/api/v1/data/gaps` | Detected gaps |
| POST | `/api/v1/data/export` | Create export job |
| GET | `/api/v1/data/export/{id}` | Export job status |
| GET | `/api/v1/data/export/{id}/download` | Download export |

### 9.4 Indicator Endpoints

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/indicators` | List indicators |
| GET | `/api/v1/indicators/config` | Get configs |
| PUT | `/api/v1/indicators/config` | Update configs |
| POST | `/api/v1/indicators/compute` | Recompute |

### 9.5 System Endpoints

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/system/status` | Health |
| GET | `/api/v1/system/stats` | Stats |
| GET | `/api/v1/system/jobs` | Running/queued jobs |

---

## 10. Supported Timeframes (UI Rule)

Timeframes are **exchange-specific**. The UI must only display timeframes supported by the selected exchange.

**Validation Rule**
1. `GET /api/v1/exchanges/{id}/timeframes`
2. Populate timeframe selector with only supported values
3. Hide/disable unsupported values

---

## 11. WebSocket API

### 11.1 Connection
```
ws://host:port/ws
```

### 11.2 Messages

**Subscribe OHLCV**
```json
{"action":"subscribe","channel":"ohlcv","exchange":"binance","symbol":"BTC/USDT","timeframe":"1h"}
```

**OHLCV update**
```json
{
  "type":"ohlcv",
  "exchange":"binance",
  "symbol":"BTC/USDT",
  "timeframe":"1h",
  "data":{"timestamp":1704067200000,"open":42150.5,"high":42300.0,"low":42100.0,"close":42250.0,"volume":1234.56}
}
```

**Subscribe connector status**
```json
{"action":"subscribe","channel":"status","connector_id":"uuid"}
```

---

## 12. Non-Functional Requirements (NFR)

### 12.1 Performance
| Metric | Requirement |
|---|---|
| API Response (p95) | < 200ms |
| OHLCV Query (1000 candles) | < 100ms |
| Indicator Compute (1000 candles) | < 500ms |
| WebSocket Latency | < 50ms |
| Chart Rendering | < 1s for 10,000 candles |
| Export Generation | < 30s for 1M rows |

### 12.2 Scalability
| Metric | Requirement |
|---|---|
| Exchanges Supported | 100+ (via CCXT) |
| Concurrent Connectors | 500+ |
| OHLCV Storage | 1B+ candles |
| WebSocket Connections | 1000+ |

### 12.3 Reliability
| Metric | Requirement |
|---|---|
| System Uptime | 99.5% |
| Data Durability | 99.999% |
| RTO | < 1 hour |
| RPO | < 5 minutes |

### 12.4 Security
| Requirement | Description |
|---|---|
| No exchange secrets | public APIs only |
| Input validation | sanitize all user inputs |
| Internal abuse protection | rate limit admin APIs |

---

## 13. User Stories (Best of V1)

### Connector Management
- US-001: Create connector by selecting exchange/pairs/timeframes
- US-002: View connector list with status and last sync
- US-003: Pause connector without data loss
- US-004: Delete connector but keep data

### Data Collection
- US-010: Auto historical backfill on connector creation
- US-011: Manual refresh trigger
- US-012: View backfill progress
- US-013: Auto detect/fill gaps

### Exploration & Export
- US-020: Browse by exchange/pair/timeframe
- US-021: Interactive OHLCV chart
- US-022: Overlay indicators
- US-023: Export filtered dataset (CSV/JSON)

### Monitoring
- US-030: Dashboard with stats
- US-031: Notification on failure
- US-032: Identify exchanges with issues

---

## 14. Implementation Phases

### Phase 1: Foundation (Weeks 1-4)
- Project skeleton + DBs
- CCXT-Go integration + validation
- Connector CRUD
- Basic OHLCV collection with pagination
- MongoDB schema + dedupe

### Phase 2: Core Features (Weeks 5-8)
- Rate limiting + circuit breaker
- Job queue + priorities
- Gap detection + backfill
- Scheduler persistence
- Stats endpoints

### Phase 3: Enrichment (Weeks 9-12)
- Indicator engine (core set)
- Explorer UI + chart overlays
- Export jobs (CSV/JSON)
- WebSocket streaming where supported

### Phase 4: Polish (Weeks 13-16)
- Monitoring + alerting
- Performance tuning
- Docs + tests
- Retention & optional datasets hardening

---

## 15. Appendix: Backfill Metadata Schema

```json
{
  "exchange_id": "binance",
  "symbol": "BTC/USDT",
  "timeframe": "1h",
  "status": "in_progress|completed|failed|partial",
  "first_ts": 1609459200000,
  "last_ts": 1704067200000,
  "candle_count": 8760,
  "expected_count": 8800,
  "pages_fetched": 9,
  "gaps": [
    {"start": 1650000000000, "end": 1650003600000, "count": 1}
  ],
  "last_fetch_at": "2026-01-17T12:00:00Z",
  "last_success_at": "2026-01-17T12:00:00Z",
  "completed_at": null,
  "error_message": null
}
```

---

*Document Version: 3.0 — Consolidated V1 strengths + corrected V2 logic*
