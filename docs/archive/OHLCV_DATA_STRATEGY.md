# OHLCV Data Collection Strategy

## Current Implementation Analysis

### Data Structure (✓ Correct)
- **Storage**: One MongoDB document per candle
- **Collections**: Individual OHLCV candles in `ohlcv` collection
- **Example Document**:
```json
{
  "_id": "696f16f395abb8412d7e3952",
  "exchange_id": "bybit",
  "symbol": "BTC/USDT",
  "timeframe": "5m",
  "open_time": "2026-01-19T21:32:31.916Z",
  "open": 47295.98,
  "high": 47375.92,
  "low": 47236.60,
  "close": 47300.74,
  "volume": 485.69,
  "indicators": {},
  "created_at": "2026-01-20T05:47:31.916Z"
}
```

### Current Issues

1. **Mock Data Generation**
   - Currently generates fake OHLCV data
   - Not fetching real market data from exchanges
   - Random price movements (not realistic)

2. **Wrong Starting Point**
   - First execution: Starts from `-100 timeframes` ago
   - Should start from: Exchange listing date or configurable historical start date
   - Results in incomplete historical data

3. **Inefficient Subsequent Fetches**
   - Currently generates 10 candles starting from cursor
   - Can create overlapping data
   - Should fetch only NEW candles since last stored candle

4. **No Backfill Strategy**
   - No mechanism to backfill large historical gaps
   - First execution should fetch all historical data up to now
   - Subsequent executions should only fetch recent candles

## Proposed Solution

### 1. First Execution Strategy (No Cursor)

**Goal**: Fetch all historical data from a start date to current time

```
┌─────────────────────────────────────────────────────────┐
│ First Execution                                         │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Start Date          Fetch All Candles          Now    │
│  (configurable)    ──────────────────────────▶  │      │
│  │                                               │      │
│  │  Example: 1 year ago for daily data          │      │
│  │           1 month ago for 5m data            │      │
│  └──────────────────────────────────────────────┘      │
│                                                         │
│  Store cursor.last_candle_time = most recent candle    │
└─────────────────────────────────────────────────────────┘
```

**Implementation**:
```go
func (e *JobExecutor) FetchOHLCVData(connector *models.Connector, job *models.Job) ([]models.OHLCV, error) {
    var startTime time.Time

    if job.Cursor.LastCandleTime == nil {
        // FIRST EXECUTION: Fetch from configurable start date
        startTime = getHistoricalStartDate(job.Timeframe)
        // Examples:
        // - 1h timeframe: start 3 months ago
        // - 5m timeframe: start 1 month ago
        // - 1d timeframe: start 2 years ago
    } else {
        // SUBSEQUENT EXECUTIONS: Fetch only NEW candles
        startTime = *job.Cursor.LastCandleTime
    }

    endTime := time.Now()

    // Fetch real OHLCV data from CCXT
    ohlcvData := fetchFromCCXT(connector, job.Symbol, job.Timeframe, startTime, endTime)

    return ohlcvData, nil
}
```

### 2. Subsequent Execution Strategy (Has Cursor)

**Goal**: Fetch only new candles since last stored candle

```
┌─────────────────────────────────────────────────────────┐
│ Subsequent Executions                                   │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Already Stored      Fetch New Candles         Now     │
│  ─────────────────  ─────────────────────▶     │       │
│                  │                              │       │
│                  └─ last_candle_time            │       │
│                     (from cursor)               │       │
│                                                         │
│  Only fetch candles with open_time > last_candle_time  │
│  Update cursor.last_candle_time to newest candle       │
└─────────────────────────────────────────────────────────┘
```

### 3. Configuration for Historical Start Dates

Add to `internal/config/config.go`:

```go
type HistoricalDataConfig struct {
    // How far back to fetch on first execution
    StartDateByTimeframe map[string]string // e.g., "5m": "30d", "1h": "90d", "1d": "2y"

    // Maximum candles to fetch per execution (rate limit protection)
    MaxCandlesPerFetch int // e.g., 1000

    // Backfill batch size for large gaps
    BackfillBatchSize int // e.g., 500
}
```

Environment variables:
```env
# Historical data configuration
HISTORICAL_START_5m=30d    # 30 days for 5-minute candles
HISTORICAL_START_1h=90d    # 90 days for 1-hour candles
HISTORICAL_START_1d=2y     # 2 years for daily candles
MAX_CANDLES_PER_FETCH=1000
BACKFILL_BATCH_SIZE=500
```

### 4. Benefits for Indicator Calculation

**Why this matters**:

1. **Continuous Data**: No gaps in historical data
   - Indicators like RSI, MACD, EMA require continuous data
   - Missing candles break indicator calculations

2. **Sufficient History**: Proper warm-up period
   - RSI(14) needs 14+ candles to calculate
   - MACD(12,26) needs 26+ candles
   - First execution provides this historical context

3. **Efficient Updates**: Only fetch new data
   - Subsequent executions only fetch recent candles
   - Recalculate indicators only for new candles
   - Or recalculate last N candles if indicator needs historical context

4. **Chronological Order**: Data stored in time order
   - Easy to query for indicator calculation windows
   - Example: Get last 26 candles for MACD calculation

## Implementation Plan

### Phase 1: Configuration (Priority: High)
- [ ] Add HistoricalDataConfig to config structure
- [ ] Add environment variables for start dates
- [ ] Create helper function `getHistoricalStartDate(timeframe)`

### Phase 2: CCXT Integration (Priority: High)
- [ ] Replace mock data generator with real CCXT calls
- [ ] Implement `fetchFromCCXT(connector, symbol, timeframe, since, until)`
- [ ] Handle CCXT rate limits
- [ ] Handle exchange-specific limits (max candles per request)

### Phase 3: Smart Fetching Logic (Priority: High)
- [ ] Modify `FetchOHLCVData` to check cursor state
- [ ] First execution: fetch from historical start date
- [ ] Subsequent executions: fetch from last_candle_time
- [ ] Handle pagination if result exceeds exchange limits

### Phase 4: Backfill Strategy (Priority: Medium)
- [ ] Detect large time gaps between executions
- [ ] Implement batched backfill for gaps
- [ ] Add manual backfill API endpoint for specific date ranges

### Phase 5: Indicator Calculation (Priority: Medium)
- [ ] Create indicator calculation service
- [ ] Calculate indicators for new candles
- [ ] Store indicators in `indicators` field
- [ ] API endpoint to recalculate indicators for date range

## API Endpoints Needed

### Backfill Historical Data
```http
POST /api/v1/jobs/{id}/backfill
Content-Type: application/json

{
  "start_date": "2025-01-01T00:00:00Z",
  "end_date": "2026-01-01T00:00:00Z",
  "batch_size": 500
}
```

### Get OHLCV Data
```http
GET /api/v1/ohlcv?exchange_id=bybit&symbol=BTC/USDT&timeframe=5m&limit=100&since=2026-01-01
```

### Calculate Indicators
```http
POST /api/v1/jobs/{id}/calculate-indicators
Content-Type: application/json

{
  "start_date": "2026-01-01T00:00:00Z",
  "end_date": "2026-01-20T00:00:00Z",
  "indicators": ["rsi14", "macd", "ema12", "ema26"]
}
```

## Data Volume Estimates

### Storage Requirements (BTC/USDT example)

| Timeframe | Candles/Day | Candles/Month | Candles/Year | Size/Year (approx) |
|-----------|-------------|---------------|--------------|-------------------|
| 1m        | 1,440       | 43,200        | 525,600      | ~50 MB            |
| 5m        | 288         | 8,640         | 105,120      | ~10 MB            |
| 15m       | 96          | 2,880         | 35,040       | ~3.5 MB           |
| 1h        | 24          | 720           | 8,760        | ~900 KB           |
| 4h        | 6           | 180           | 2,190        | ~220 KB           |
| 1d        | 1           | 30            | 365          | ~40 KB            |

*Based on ~100 bytes per OHLCV document without indicators*

### Fetching Strategy by Timeframe

| Timeframe | First Fetch  | Candles | Subsequent Fetch | Update Frequency |
|-----------|-------------|---------|------------------|------------------|
| 1m        | 7 days      | ~10,080 | Last 15 min     | Every 1 minute   |
| 5m        | 30 days     | ~8,640  | Last 1 hour     | Every 5 minutes  |
| 15m       | 90 days     | ~8,640  | Last 3 hours    | Every 15 minutes |
| 1h        | 180 days    | ~4,320  | Last 12 hours   | Every 1 hour     |
| 4h        | 1 year      | ~2,190  | Last 2 days     | Every 4 hours    |
| 1d        | 3 years     | ~1,095  | Last 7 days     | Every 1 day      |

## Summary

**Current State**:
- ✓ Correct data structure (one document per candle)
- ✗ Mock data generation
- ✗ Wrong starting point (-100 timeframes)
- ✗ No historical backfill strategy

**Required Changes**:
1. Replace mock data with real CCXT integration
2. First execution: Fetch from configurable historical start date
3. Subsequent executions: Fetch only new candles from cursor
4. Add configuration for start dates by timeframe
5. Implement backfill strategy for gaps

**Benefits**:
- Complete historical data for indicator calculations
- Efficient incremental updates
- No duplicate or overlapping data
- Proper warm-up period for indicators
- Scalable for multiple symbols and timeframes
