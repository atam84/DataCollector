# Pending Tasks

**Last Updated:** 2026-01-22
**Current Version:** v1.0.5b
**Current Branch:** main

---

## Error Analysis Summary

| Error Type | Affected Exchanges | Root Cause | Impact |
|------------|-------------------|------------|--------|
| `RateLimitExceeded` | **ALL** | Too many API calls, insufficient delay between requests | **CRITICAL** - Jobs fail |
| `document is too large` | **ALL** | MongoDB 16MB limit exceeded (40k+ candles) | **HIGH** - Data loss |
| `date of query is too wide` | **ALL** | Exchange rejects wide historical date ranges | **MEDIUM** - First fetch fails |

**Example Errors Observed:**
- OKX: `{"msg":"Too Many Requests","code":"50011"}`
- BingX: `{"code":100204,"msg":"date of query is too wide."}`
- All: `failed to insert new OHLCV document: an inserted document is too large`

---

## ✅ P0 - Critical (COMPLETED 2026-01-22)

| # | Issue | Description | Status |
|---|-------|-------------|--------|
| **P0.1** | **Rate Limit - Per-Connector Throttling** | RateLimiter service with min delay between calls (MinDelayMs) | ✅ Done |
| **P0.2** | **Rate Limit - API Call Counter** | IncrementAPIUsage() tracks each call, LastAPICallAt timestamp | ✅ Done |
| **P0.3** | **Rate Limit - Exchange-Level Config** | Limit, PeriodMs, MinDelayMs in Connector model | ✅ Done |
| **P0.4** | **MongoDB Document Size Limit** | OHLCVChunk model with monthly chunking (ohlcv_chunks collection) | ✅ Done |

### Rate Limit Architecture (Target Design)

```
┌─────────────────────────────────────────────────────────────────┐
│                    RATE LIMIT MANAGEMENT                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Exchange Config (stored in DB):                                │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ exchange_id: "okx"                                       │   │
│  │ rate_limit_per_minute: 20                                │   │
│  │ min_delay_ms: 3000  (calculated: 60000/20 = 3000ms)     │   │
│  │ api_calls_used: 15                                       │   │
│  │ api_calls_reset_at: timestamp (next minute)             │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                  │
│  Before Each API Call:                                          │
│  1. Check time since last call on this connector                │
│  2. If < min_delay_ms → sleep(remaining)                        │
│  3. Check api_calls_used < rate_limit_per_minute               │
│  4. If limit reached → sleep until reset_at                     │
│  5. Make API call                                               │
│  6. Update api_calls_used++                                     │
│  7. Update last_call_timestamp                                  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🟠 P1 - High Priority (Mostly COMPLETED 2026-01-22)

| # | Issue | Description | Status |
|---|-------|-------------|--------|
| **P1.1** | **Exchange Date Range Fallback** | dateRangeFallbacks array (5y→1y→6m→3m→1m), isDateRangeError() detection | ✅ Done |
| **P1.2** | **Error Display - Modal for Long Errors** | ErrorModal component, extractMainError(), truncateText() helpers | ✅ Done |
| **P1.3** | **Cryptocurrency Pairs Management** | Dynamic popular pairs, symbol validation endpoints, improved UX | ✅ Done |
| **P1.4** | **Remove Sandbox Mode** | Removed from frontend, docker-compose (Go code had none) | ✅ Done |

---

## 🟡 P2 - Medium Priority

| # | Issue | Description | Status |
|---|-------|-------------|--------|
| P2.1 | Indicator Config Affectation | Configs saved but not enforced during calculation | ✅ Done |
| P2.2 | Connector Statistics Dashboard | Data volume, job count, last run times, API calls used | ✅ Done |
| P2.3 | Job Error Recovery | Automatic retry with exponential backoff on transient errors | ✅ Done |
| P2.4 | Rate Limit Visualization | Show remaining calls, cooldown timer in UI | ✅ Done |
| P2.5 | Comprehensive error handling in API | Better error types and messages | ✅ Done |
| P2.6 | Validation for indicator parameters | Min/max values enforcement | ✅ Done |
| P2.7 | Unit tests for config merge logic | Test coverage | ✅ Done |
| P2.8 | Job dependency management | Run job A before job B | ✅ Done |
| P2.9 | Alerting system for failed jobs | Notifications on failures | ✅ Done |
| P2.10 | Data retention policies | Auto-cleanup old candles | ✅ Done |

---

## 🟢 P3 - Low Priority

| # | Issue | Description | Status |
|---|-------|-------------|--------|
| P3.1 | API documentation (Swagger/OpenAPI) | Swagger UI at /swagger/index.html | ✅ Done |
| P3.2 | Add more trading pairs to JobWizard | | |
| P3.3 | Connector health monitoring dashboard | | |
| P3.4 | Data quality metrics (missing candles, gaps) | | |
| P3.5 | Custom timeframes support | | |
| P3.6 | WebSocket real-time updates | | |
| P3.7 | Multi-user authentication | | |

---

## 🔵 Future Enhancements

| # | Issue | Description |
|---|-------|-------------|
| F1 | Machine learning model training integration | |
| F2 | Strategy backtesting framework | |
| F3 | Alert system for indicator threshold breaches | |
| F4 | Correlation analysis between pairs | |
| F5 | Market regime detection | |
| F6 | Custom indicator formula builder | |

---

## 🗑️ Removed Items

| Item | Reason |
|------|--------|
| Sandbox mode testing | Never use sandbox - removed from scope |

---

## Recent Completed Work (2026-01-21)

### v1.0.5b - Dynamic Exchange Support & Historical Data Collection

**Dynamic CCXT Exchange Support:**
- ✅ Dynamic exchange discovery from `ccxt.Exchanges` (111 exchanges supported)
- ✅ Auto-detection of OHLCV support via `exchange.GetHas()`
- ✅ Dynamic metadata fetching: `GetTimeframes()`, `GetFeatures()`, `GetHas()`
- ✅ Thread-safe caching for exchange metadata and supported list
- ✅ Cache refresh endpoint: `POST /api/v1/exchanges/refresh`
- ✅ Debug endpoint: `GET /api/v1/exchanges/:id/debug`

**Historical Data Collection:**
- ✅ Full historical data fetching with pagination
- ✅ Batched fetching using exchange's OHLCV limit
- ✅ Forward pagination until reaching present time

**CandlestickChart Component:**
- ✅ Professional candlestick visualization with lightweight-charts v5
- ✅ Volume histogram with color coding
- ✅ Indicator overlay support (SMA, EMA, Bollinger Bands)
- ✅ Separate panes for momentum indicators (RSI, MACD, Stochastic)

**UI Enhancements:**
- ✅ Refresh buttons on all pages
- ✅ Exchange selection in wizards uses dynamic list

---

## Files Modified (2026-01-22)

### Rate Limit System (P0.1-P0.3): ✅ DONE
- `/internal/models/connector.go` - Added MinDelayMs, LastAPICallAt fields
- `/internal/service/rate_limiter.go` - **NEW** RateLimiter service with WaitForSlot()
- `/internal/service/ccxt_service.go` - Integrated rate limiter, context support
- `/internal/service/job_executor.go` - Uses rate-limited CCXT service
- `/internal/repository/connector_repository.go` - Added IncrementAPIUsage(), ResetRateLimitPeriod()
- `/internal/api/handlers/connector_handler.go` - Added rate limit status/reset endpoints
- `/cmd/api/main.go` - Added routes for rate limit endpoints

### MongoDB Chunking (P0.4): ✅ DONE
- `/internal/models/ohlcv.go` - Added OHLCVChunk model, GetYearMonthFromTimestamp()
- `/internal/repository/ohlcv_repository.go` - Chunked storage (ohlcv_chunks collection)

### Date Range Fallback (P1.1): ✅ DONE
- `/internal/service/ccxt_service.go` - Added dateRangeFallbacks, isDateRangeError()

### Error Modal (P1.2): ✅ DONE
- `/web/src/components/JobList.jsx` - Added error modal, extractMainError(), truncateText()

### Remove Sandbox (P1.4): ✅ DONE
- `/web/src/components/JobDetails.jsx` - Removed sandbox_mode reference
- `/docker-compose.dev.yml` - Removed EXCHANGE_SANDBOX_MODE

### Cryptocurrency Pairs Management (P1.3): ✅ DONE
- `/internal/api/handlers/health.go` - Added ValidateSymbol, ValidateSymbols, GetPopularSymbols endpoints
- `/cmd/api/main.go` - Added routes for new symbol validation endpoints
- `/web/src/components/JobWizard.jsx` - Dynamic popular pairs, validation before job creation, quick select actions

### Rate Limit Visualization (P2.4): ✅ DONE
- `/web/src/components/ConnectorList.jsx` - Real-time rate limit status, cooldown indicator, detailed settings modal
- `/web/src/components/Dashboard.jsx` - Rate limit overview section with all connectors

### Connector Statistics Dashboard (P2.2): ✅ DONE
- `/internal/models/ohlcv.go` - Added OHLCVStats model
- `/internal/repository/ohlcv_repository.go` - Added GetStatsByExchange(), GetAllStats() methods
- `/internal/api/handlers/connector_handler.go` - Added GetConnectorStats(), GetAllStats() endpoints
- `/cmd/api/main.go` - Added /stats and /connectors/:id/stats routes
- `/web/src/components/Dashboard.jsx` - Data statistics section with candle counts, symbols, timeframes

### Job Error Recovery (P2.3): ✅ DONE
- `/internal/models/job.go` - Added ConsecutiveFailures and LastFailureTime to RunState
- `/internal/service/job_executor.go` - Added handleExecutionError() with exponential backoff, isTransientError(), calculateBackoff()
- `/internal/repository/job_repository.go` - Added IncrementConsecutiveFailures(), ResetConsecutiveFailures(), GetJobsWithFailures()

### Comprehensive Error Handling (P2.5): ✅ DONE
- `/internal/api/errors/errors.go` - **NEW** Standardized error types and response format
- `/internal/api/handlers/job_handler.go` - Updated all endpoints to use standardized errors
- `/internal/api/handlers/connector_handler.go` - Updated all endpoints to use standardized errors

### Alerting System (P2.9): ✅ DONE
- `/internal/models/alert.go` - **NEW** Alert model with types, severities, and statuses
- `/internal/repository/alert_repository.go` - **NEW** Alert CRUD operations and summary
- `/internal/service/alert_service.go` - **NEW** Alert generation for job failures, connector issues
- `/internal/api/handlers/alert_handler.go` - **NEW** Alert API endpoints
- `/cmd/api/main.go` - Added alert routes and service initialization

### Data Retention Policies (P2.10): ✅ DONE
- `/internal/models/retention.go` - **NEW** Retention policy and config models
- `/internal/repository/retention_repository.go` - **NEW** Policy CRUD, chunk deletion, usage stats
- `/internal/service/retention_service.go` - **NEW** Cleanup operations, data usage tracking
- `/internal/api/handlers/retention_handler.go` - **NEW** Retention API endpoints
- `/cmd/api/main.go` - Added retention routes and service initialization

### Indicator Config Affectation (P2.1): ✅ DONE
- `/internal/models/indicator_config.go` - **NEW** Full config model with TrendConfig, MomentumConfig, VolatilityConfig, VolumeConfig
- `/internal/repository/indicator_config_repository.go` - **NEW** Config CRUD, SetDefault(), FindDefault()
- `/internal/service/indicators/service.go` - Updated to use config, added CalculateWithConfig(), configurable periods
- `/internal/api/handlers/indicator_config_handler.go` - **NEW** Config API endpoints (CRUD, set default, builtin defaults)
- `/cmd/api/main.go` - Added indicator config routes

### Indicator Parameter Validation (P2.6): ✅ DONE
- `/internal/models/indicator_config.go` - Added Validate() methods with min/max constraints, ValidationError type
- `/internal/api/handlers/indicator_config_handler.go` - Added validation on create/update, GetValidationRules(), ValidateConfig() endpoints
- `/cmd/api/main.go` - Added validation-rules and validate routes

### Unit Tests for Config Validation (P2.7): ✅ DONE
- `/internal/models/indicator_config_test.go` - **NEW** 12 test cases covering all validation functions

### Job Dependency Management (P2.8): ✅ DONE
- `/internal/models/job.go` - Added DependsOn field, DependencyStatus type
- `/internal/repository/job_repository.go` - Added SetDependencies(), FindByIDs(), GetDependencyStatus(), CheckCircularDependency(), FindJobsDependingOn()
- `/internal/api/handlers/job_handler.go` - Added dependency support to Create/Update, GetJobDependencies(), SetJobDependencies(), GetJobDependents()
- `/internal/service/job_executor.go` - Added dependency checking before job execution
- `/cmd/api/main.go` - Added dependency routes

---

## Progress Tracking

| Priority | Total | Done | In Progress | Remaining |
|----------|-------|------|-------------|-----------|
| P0 | 4 | 4 | 0 | 0 |
| P1 | 4 | 4 | 0 | 0 |
| P2 | 10 | 10 | 0 | 0 |
| P3 | 7 | 0 | 0 | 7 |

**Current Focus:** All P2 Tasks Complete - Ready for P3
