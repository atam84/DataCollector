# Changelog

All notable changes to the DataCollector project will be documented in this file.

## [v1.0.5b] - 2026-01-21

### Added
- **Dynamic Exchange Support**: Now supports 111 exchanges dynamically discovered from CCXT Go library
- **Historical Data Collection**: Full pagination support to fetch all available historical data (up to 5 years)
- **CandlestickChart Component**: Professional charting with lightweight-charts v5
  - Candlestick visualization with volume histogram
  - Indicator overlays (SMA, EMA, Bollinger Bands)
  - Separate panes for momentum indicators (RSI, MACD, Stochastic)
  - Collapsible indicator groups with enable/disable checkboxes
- **Exchange Debug Endpoint**: `GET /api/v1/exchanges/:id/debug` for troubleshooting
- **Cache Refresh Endpoint**: `POST /api/v1/exchanges/refresh` to clear metadata cache
- **Refresh Buttons**: Added to all pages (Dashboard, Connectors, Jobs, Queue) with spin animation

### Fixed
- **OKX/BingX Support**: Fixed "exchange not yet supported" error by replacing hardcoded exchange list with dynamic adapter
- **CCXTService**: Refactored to use generic `exchange.CCXTAdapter` instead of hardcoded bybit/binance functions
- **lightweight-charts v5**: Fixed API compatibility (`addSeries()` instead of deprecated `addCandlestickSeries()`)
- **4h Timeframe**: Fixed duration calculation bug (was 4 minutes instead of 4 hours)

### Changed
- Exchange metadata now fetched dynamically using `GetTimeframes()`, `GetHas()`, `GetFeatures()`
- Thread-safe caching implemented for exchange metadata and supported exchange list
- Historical data fetching now uses batched pagination with exchange-specific limits

## [v1.0.5a] - 2026-01-20

### Added
- Wizard-based workflows for connector and job creation
- Batch job creation endpoint (`/api/v1/jobs/batch`)
- Data export endpoints (CSV, JSON, ML-optimized formats)
- Job search and multi-connector filtering
- JobDetails component with Overview, Raw Data, and Charts tabs
- Heroicons integration for UI buttons
- Indicators documentation page

### Fixed
- KuCoin exchange support
- MongoDB disk space issues
- Type conversion errors in CCXT API calls

## [v1.0.4] - 2026-01-19

### Added
- Exchange validation system
- Rate limit management for connectors
- Automatic job scheduling with next run time calculation

## [v1.0.3] - 2026-01-18

### Added
- 29 technical indicators with automatic calculation
- Indicator recalculation endpoints
- OHLCV data storage with embedded indicators

## [v1.0.2] - 2026-01-17

### Added
- Job execution with CCXT integration
- Connector management (create, update, delete, suspend, resume)
- Basic job management

## [v1.0.1] - 2026-01-16

### Added
- Initial MongoDB integration
- Basic API structure with Go Fiber
- Health check endpoints

## [v1.0.0] - 2026-01-15

### Added
- Initial project setup
- Docker configuration
- React frontend scaffold
