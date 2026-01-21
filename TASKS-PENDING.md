# Pending Tasks

## Recent Completed Work (2026-01-21)

### ✅ v1.0.5b - Dynamic Exchange Support & Historical Data Collection

**Dynamic CCXT Exchange Support (Completed)**
- ✅ Dynamic exchange discovery from `ccxt.Exchanges` (111 exchanges supported)
- ✅ Auto-detection of OHLCV support via `exchange.GetHas()`
- ✅ Dynamic metadata fetching: `GetTimeframes()`, `GetFeatures()`, `GetHas()`
- ✅ Thread-safe caching for exchange metadata and supported list
- ✅ Cache refresh endpoint: `POST /api/v1/exchanges/refresh`
- ✅ Debug endpoint: `GET /api/v1/exchanges/:id/debug`
- ✅ Fixed CCXTService to use generic adapter (was hardcoded to only bybit/binance)

**Historical Data Collection (Completed)**
- ✅ Full historical data fetching with pagination
- ✅ Starts from 5 years ago (or exchange's earliest data)
- ✅ Batched fetching using exchange's OHLCV limit
- ✅ Forward pagination until reaching present time
- ✅ Rate limit respect with delays between batches
- ✅ Graceful error handling (returns partial data on errors)

**CandlestickChart Component (Completed)**
- ✅ Professional candlestick visualization with lightweight-charts v5
- ✅ Volume histogram with color coding (green/red)
- ✅ Indicator overlay support (SMA, EMA, Bollinger Bands, etc.)
- ✅ Separate panes for momentum indicators (RSI, MACD, Stochastic)
- ✅ Collapsible indicator groups with enable/disable checkboxes
- ✅ Interactive legend showing active indicators
- ✅ Fixed v5 API compatibility (`addSeries()` instead of `addCandlestickSeries()`)

**UI Enhancements (Completed)**
- ✅ Refresh buttons on all pages (Dashboard, Connectors, Jobs, Queue)
- ✅ Consistent icon-style buttons with spin animation during loading
- ✅ Exchange selection in wizards now uses dynamic list

---

## Recent Completed Work (2026-01-20)

### ✅ Completed Features

**Exchange Validation & Support (Completed)**
- ✅ Exchange validation system with `TestExchangeAvailability()`
- ✅ Support for 111 exchanges (dynamic from CCXT)
- ✅ `/api/v1/exchanges` endpoint returning supported exchanges
- ✅ `/api/v1/exchanges/test` endpoint for testing availability
- ✅ `/api/v1/exchanges/metadata` endpoint for all exchange metadata
- ✅ Exchange ID mapping (gate → gateio, etc.)

**Wizard-Based Workflows (Completed)**
- ✅ ConnectorWizard with 2-step flow (exchange + indicators)
- ✅ JobWizard with 4-step flow (connector + pairs + timeframes + indicators)
- ✅ Visual exchange selection grid with rate limits
- ✅ Progress bars and step validation
- ✅ Edit rate limit functionality for connectors

**Batch Operations & Data Export (Completed)**
- ✅ `/api/v1/jobs/batch` endpoint (up to 100 jobs)
- ✅ Automatic job start time staggering
- ✅ Multi-select for cryptocurrency pairs and timeframes
- ✅ `/api/v1/jobs/:id/ohlcv` endpoint with pagination
- ✅ `/api/v1/jobs/:id/export` endpoint (CSV/JSON formats)
- ✅ `/api/v1/jobs/:id/export/ml` endpoint (ML-optimized format)

**Job Management Enhancements (Completed)**
- ✅ Search bar for filtering jobs by symbol
- ✅ Multi-connector filter with checkboxes
- ✅ JobDetails component with 3 tabs (Overview, Raw Data, Charts)
- ✅ lightweight-charts integration for data visualization
- ✅ Clickable symbols opening detailed view
- ✅ Export buttons with file download handling

**UI Improvements (Completed)**
- ✅ Heroicons integration replacing text in buttons
- ✅ Indicators documentation page with descriptions
- ✅ Info tooltips for indicators in configuration modals
- ✅ Run now button in queue items

---

## Indicator Configuration - Affectation Issues

**Status:** ⚠️ NOT WORKING AS EXPECTED

**Problem:**
The indicator configuration system has been implemented with GET/PUT/PATCH endpoints and UI, but the actual affectation/application of configurations is not working correctly. When indicators are enabled/disabled via the configuration UI, they are not being properly enforced during calculation.

**What's Implemented:**
- ✅ Configuration UI for connectors and jobs
- ✅ GET/PUT/PATCH API endpoints
- ✅ Config saved to database
- ✅ Config merge logic (job > connector > defaults)

**What's NOT Working:**
- ❌ Configurations not properly applied during indicator calculation
- ❌ Disabling indicators doesn't prevent their calculation
- ❌ Only enabled indicators should be calculated
- ❌ Need verification that config is actually used by indicator service

**Files Involved:**
- `/internal/service/indicators/service.go` - CalculateAll method
- `/internal/service/indicators/config.go` - MergeConfigs logic
- `/internal/service/job_executor.go` - GetEffectiveConfig usage
- `/internal/service/recalculator.go` - Recalculation with configs

**Next Steps:**
1. Add debug logging to verify config is passed to indicator service
2. Verify CalculateAll actually checks config.Enabled flags
3. Test with only 1-2 indicators enabled
4. Verify database stores config correctly
5. Test recalculation applies new configs

**Priority:** Medium (system works with defaults, but user configs not respected)

---

## Pending Items

### High Priority

- [ ] **Fix Indicator Configuration Affectation** - Main blocker for full system functionality
- [ ] Test full end-to-end indicator config workflow
- [ ] Verify config inheritance (job overrides connector)
- [ ] Performance testing with all 29 indicators enabled
- [ ] Add connector statistics page showing data volume, job count, last run times

### Medium Priority

- [ ] Add more comprehensive error handling in API
- [ ] Add validation for indicator parameters (min/max values)
- [ ] Add unit tests for config merge logic
- [ ] Implement job dependency management (run job A before job B)
- [ ] Add alerting system for failed jobs or rate limit violations
- [ ] Implement data retention policies (auto-cleanup old candles)

### Low Priority

- [ ] API documentation (Swagger/OpenAPI)
- [ ] Add more popular trading pairs to JobWizard
- [ ] Implement connector health monitoring dashboard
- [ ] Add data quality metrics (missing candles, data gaps)
- [ ] Support for custom timeframes beyond predefined ones
- [ ] WebSocket support for real-time data streaming
- [ ] Multi-user support with authentication/authorization

### Future Enhancements

- [ ] Machine learning model training integration
- [ ] Strategy backtesting framework
- [ ] Alert system for indicator threshold breaches
- [ ] Correlation analysis between different pairs
- [ ] Market regime detection
- [ ] Custom indicator formula builder

---

## Known Issues

### Fixed Issues (v1.0.5b)
- ✅ OKX exchange "not yet supported" error - Fixed by using dynamic adapter
- ✅ BingX exchange "not yet supported" error - Fixed by using dynamic adapter
- ✅ CCXTService hardcoded to only bybit/binance - Now uses generic adapter
- ✅ Historical data not being collected - Implemented pagination
- ✅ lightweight-charts v5 API incompatibility - Fixed `addSeries()` usage

### Fixed Issues (Previous Sessions)
- ✅ KuCoin exchange "not yet supported" error
- ✅ Exchange validation rejecting all exchanges except Binance
- ✅ MongoDB disk space issues causing crashes
- ✅ Type conversion errors in CCXT API calls (int → int64)

### Active Issues
- ⚠️ Indicator configurations not being enforced during calculation
- ⚠️ No validation for conflicting job schedules on same connector

---

## Testing Checklist

### Exchange Integration
- [x] Dynamic exchange discovery (111 exchanges)
- [x] Binance connector creation and data fetching
- [x] Multiple exchange support verification
- [x] Exchange availability testing endpoint
- [x] OKX and BingX support verified
- [ ] Sandbox mode testing for each exchange
- [ ] Rate limit enforcement testing

### Historical Data Collection
- [x] First execution fetches all historical data
- [x] Pagination working correctly
- [x] Subsequent executions fetch only new data
- [ ] Very large dataset collection (100k+ candles)
- [ ] Error recovery during pagination

### Wizard Workflows
- [x] Connector wizard 2-step flow
- [x] Job wizard 4-step flow
- [x] Batch job creation (multiple pairs × timeframes)
- [x] Dynamic exchange list in wizards
- [ ] Wizard validation edge cases
- [ ] Indicator configuration in wizards

### Data Visualization
- [x] CandlestickChart rendering
- [x] Volume histogram display
- [x] Indicator overlays (SMA, EMA, BB)
- [x] Separate indicator panes (RSI, MACD)
- [ ] Large dataset performance (10k+ candles)

### Data Export
- [x] CSV export format
- [x] JSON export format
- [x] ML-optimized export format
- [ ] Large dataset export (10k+ candles)
- [ ] Export with all indicators enabled

### Job Management
- [x] Job search and filtering
- [x] Job details view with charts
- [x] Manual job execution
- [ ] Job pause/resume functionality
- [ ] Job error recovery

---

## Architecture Improvements Needed

### Code Quality
- [ ] Add comprehensive unit tests for services
- [ ] Add integration tests for API endpoints
- [ ] Implement proper error types instead of string errors
- [ ] Add request/response DTOs for all endpoints

### Performance
- [x] Implemented caching for exchange metadata
- [ ] Optimize indicator calculations for large datasets
- [ ] Add database indexes for common queries
- [ ] Implement connection pooling optimization

### Observability
- [x] Added comprehensive logging for CCXT operations
- [ ] Implement metrics collection (Prometheus)
- [ ] Add distributed tracing
- [ ] Create health check dashboard

### Security
- [ ] Add API authentication
- [ ] Implement rate limiting per user/IP
- [ ] Add input sanitization
- [ ] Secure sensitive configuration (API keys)

---

## Documentation Needs

- [ ] API documentation with examples
- [ ] Architecture decision records (ADRs)
- [ ] Deployment guide
- [ ] Troubleshooting guide
- [ ] Indicator calculation formulas documentation
- [ ] Contributing guidelines

---

**Last Updated:** 2026-01-21

**Current Version:** v1.0.5b

**Current Branch:** main

**Next Session Priorities:**
1. Fix indicator configuration affectation issue
2. Test historical data collection with various exchanges
3. Add connector statistics/monitoring dashboard
4. Performance testing with large datasets
