# Changelog

All notable changes to the DataCollector project will be documented in this file.

## [v1.1.0] - 2026-01-23

### Added
- **ML Export System**: Comprehensive machine learning data export with feature engineering
  - Multiple export formats: CSV, Parquet, NumPy (.npz), JSON-L
  - Feature engineering: Price features, lagged features, rolling statistics, temporal features
  - Target variable generation: Future returns, direction, multi-class classification, volatility
  - Preprocessing: Normalization (min-max, z-score, robust), NaN handling, outlier clipping
  - Train/validation/test splitting with time-based option (no look-ahead bias)
  - Sequence generation for LSTM/Transformer models
  - Multi-job dataset creation for combining multiple symbols/timeframes
  - Background export processing with progress tracking
  - Built-in presets: Basic, Technical, Full Features, LSTM Ready, Classification, Scalping, Swing Trading

- **ML Export API Endpoints**:
  - `POST /api/v1/ml/export/start` - Start background export job
  - `GET /api/v1/ml/export/jobs` - List export jobs
  - `GET /api/v1/ml/export/jobs/:id` - Get export job status
  - `GET /api/v1/ml/export/jobs/:id/download` - Download exported file
  - `GET /api/v1/ml/export/jobs/:id/metadata` - Get export metadata
  - `POST /api/v1/ml/export/jobs/:id/cancel` - Cancel export job
  - `DELETE /api/v1/ml/export/jobs/:id` - Delete export job
  - `GET /api/v1/ml/profiles` - List saved profiles
  - `GET /api/v1/ml/profiles/presets` - Get built-in presets
  - `POST /api/v1/ml/profiles` - Create custom profile
  - `GET /api/v1/ml/formats` - Get supported formats
  - `GET /api/v1/ml/features` - Get available features
  - `POST /api/v1/ml/datasets` - Create combined dataset

- **ML Export Frontend**:
  - New "ML Export" tab in navigation
  - Job selection with multi-select
  - Quick preset buttons for common configurations
  - Format selection (CSV, Parquet, NumPy, JSON-L)
  - Feature configuration with category toggles and specific indicators
  - Price features, temporal features, lagged features, rolling statistics
  - Target variable configuration with type and lookahead periods
  - Advanced settings: Preprocessing, train/val/test split, sequence generation
  - Recent exports table with status, progress, download actions

### Changed
- Updated navigation to include ML Export tab
- Added MLExportRepository, MLFeatureEngine, MLExportService, MLExportWriter
- Added MLExportHandler with comprehensive API endpoints

## [v1.0.6] - 2026-01-23

### Added
- **Data Quality System**: Comprehensive data quality monitoring and analysis
  - Background quality checks with progress tracking
  - Gap detection and completeness scoring
  - Freshness tracking and stale data alerts
  - Quality status caching for fast dashboard display
- **Gap Filling**: Background gap-filling to fetch missing candles
  - Fill first 5 gaps or all gaps options
  - Non-blocking API with job status tracking
  - Prevents 504 timeouts on large datasets
- **Historical Backfill**: Fetch historical data going back months/years
  - Configurable months back or target date
  - Background processing with progress updates
  - Batch fetching respecting rate limits
- **CandlestickChart Enhancements**:
  - Period selection buttons (1D, 1W, 1M, 3M, 6M, 1Y, All)
  - Zoom controls (zoom in, zoom out, reset)
  - Mouse wheel zoom and drag to pan support
- **JobList Improvements**:
  - Timeframe filter
  - Status filter (active, paused, stopped)
  - Candles count column
- **Clickable Symbols**: Symbols in Queue and Data Quality pages link to job details modal
- **Connector Health Sync**: Connectors page now uses same health API as Dashboard
  - Shows Uptime, Error Rate, Response Time
  - Health status matches Dashboard (Healthy/Degraded/Unhealthy)

### Fixed
- **504 Gateway Timeout**: Increased timeouts for quality analysis (10s/30s -> 120s)
- **Data Quality Dropdown**: Fixed exchange selector showing proper options
- **Running Checks Display**: Now shows real progress with intermediate results
- **Health Indicator Mismatch**: Connectors and Dashboard now show consistent health status

### Changed
- Quality checks run in background to prevent UI blocking
- Gap fill operations are non-blocking with status polling
- Backfill operations use background processing

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
