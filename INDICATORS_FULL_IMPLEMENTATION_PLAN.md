# Complete Indicator System Implementation Plan

**Date**: January 20, 2026
**Status**: 📋 PLANNING

---

## Overview

Implement a complete, configurable indicator system with:
- **29 technical indicators** (all from catalog)
- **Two-level configuration**: Connector-level and Job-level
- **Inheritance model**: Job config overrides Connector config
- **UI/API management**: Enable/disable indicators per connector/job
- **Recalculation**: On-demand recalculation of all indicators
- **Query API**: Retrieve indicator data

---

## Phase 1: Database Schema & Models

### 1.1 Indicator Configuration Schema

```go
type IndicatorConfig struct {
    // Trend Indicators
    SMA      *SMAConfig      `bson:"sma,omitempty" json:"sma,omitempty"`
    EMA      *EMAConfig      `bson:"ema,omitempty" json:"ema,omitempty"`
    DEMA     *DEMAConfig     `bson:"dema,omitempty" json:"dema,omitempty"`
    TEMA     *TEMAConfig     `bson:"tema,omitempty" json:"tema,omitempty"`
    WMA      *WMAConfig      `bson:"wma,omitempty" json:"wma,omitempty"`
    HMA      *HMAConfig      `bson:"hma,omitempty" json:"hma,omitempty"`
    VWMA     *VWMAConfig     `bson:"vwma,omitempty" json:"vwma,omitempty"`
    Ichimoku *IchimokuConfig `bson:"ichimoku,omitempty" json:"ichimoku,omitempty"`
    ADX      *ADXConfig      `bson:"adx,omitempty" json:"adx,omitempty"`
    SuperTrend *SuperTrendConfig `bson:"supertrend,omitempty" json:"supertrend,omitempty"`

    // Momentum Indicators
    RSI         *RSIConfig         `bson:"rsi,omitempty" json:"rsi,omitempty"`
    Stochastic  *StochasticConfig  `bson:"stochastic,omitempty" json:"stochastic,omitempty"`
    MACD        *MACDConfig        `bson:"macd,omitempty" json:"macd,omitempty"`
    ROC         *ROCConfig         `bson:"roc,omitempty" json:"roc,omitempty"`
    CCI         *CCIConfig         `bson:"cci,omitempty" json:"cci,omitempty"`
    WilliamsR   *WilliamsRConfig   `bson:"williams_r,omitempty" json:"williams_r,omitempty"`
    Momentum    *MomentumConfig    `bson:"momentum,omitempty" json:"momentum,omitempty"`

    // Volatility Indicators
    Bollinger   *BollingerConfig   `bson:"bollinger,omitempty" json:"bollinger,omitempty"`
    ATR         *ATRConfig         `bson:"atr,omitempty" json:"atr,omitempty"`
    Keltner     *KeltnerConfig     `bson:"keltner,omitempty" json:"keltner,omitempty"`
    Donchian    *DonchianConfig    `bson:"donchian,omitempty" json:"donchian,omitempty"`
    StdDev      *StdDevConfig      `bson:"stddev,omitempty" json:"stddev,omitempty"`

    // Volume Indicators
    OBV         *OBVConfig         `bson:"obv,omitempty" json:"obv,omitempty"`
    VWAP        *VWAPConfig        `bson:"vwap,omitempty" json:"vwap,omitempty"`
    MFI         *MFIConfig         `bson:"mfi,omitempty" json:"mfi,omitempty"`
    CMF         *CMFConfig         `bson:"cmf,omitempty" json:"cmf,omitempty"`
    VolumeSMA   *VolumeSMAConfig   `bson:"volume_sma,omitempty" json:"volume_sma,omitempty"`
}

// Example config structs
type RSIConfig struct {
    Enabled bool  `bson:"enabled" json:"enabled"`
    Periods []int `bson:"periods" json:"periods"` // e.g., [6, 14, 24]
}

type MACDConfig struct {
    Enabled      bool `bson:"enabled" json:"enabled"`
    FastPeriod   int  `bson:"fast_period" json:"fast_period"`
    SlowPeriod   int  `bson:"slow_period" json:"slow_period"`
    SignalPeriod int  `bson:"signal_period" json:"signal_period"`
}
```

### 1.2 Update Connector Model

```go
type Connector struct {
    // ... existing fields ...
    IndicatorConfig *IndicatorConfig `bson:"indicator_config,omitempty" json:"indicator_config,omitempty"`
}
```

### 1.3 Update Job Model

```go
type Job struct {
    // ... existing fields ...
    IndicatorConfig *IndicatorConfig `bson:"indicator_config,omitempty" json:"indicator_config,omitempty"`
}
```

### 1.4 Update OHLCV Model - Add ALL Indicator Fields

```go
type Indicators struct {
    // Trend Indicators
    SMA20   *float64 `bson:"sma20,omitempty" json:"sma20,omitempty"`
    SMA50   *float64 `bson:"sma50,omitempty" json:"sma50,omitempty"`
    SMA200  *float64 `bson:"sma200,omitempty" json:"sma200,omitempty"`
    EMA12   *float64 `bson:"ema12,omitempty" json:"ema12,omitempty"`
    EMA26   *float64 `bson:"ema26,omitempty" json:"ema26,omitempty"`
    DEMA    *float64 `bson:"dema,omitempty" json:"dema,omitempty"`
    TEMA    *float64 `bson:"tema,omitempty" json:"tema,omitempty"`
    WMA     *float64 `bson:"wma,omitempty" json:"wma,omitempty"`
    HMA     *float64 `bson:"hma,omitempty" json:"hma,omitempty"`
    VWMA    *float64 `bson:"vwma,omitempty" json:"vwma,omitempty"`

    // Ichimoku Cloud
    IchimokuTenkan   *float64 `bson:"ichimoku_tenkan,omitempty" json:"ichimoku_tenkan,omitempty"`
    IchimokuKijun    *float64 `bson:"ichimoku_kijun,omitempty" json:"ichimoku_kijun,omitempty"`
    IchimokuSenkouA  *float64 `bson:"ichimoku_senkou_a,omitempty" json:"ichimoku_senkou_a,omitempty"`
    IchimokuSenkouB  *float64 `bson:"ichimoku_senkou_b,omitempty" json:"ichimoku_senkou_b,omitempty"`
    IchimokuChikou   *float64 `bson:"ichimoku_chikou,omitempty" json:"ichimoku_chikou,omitempty"`

    // ADX/DMI
    ADX       *float64 `bson:"adx,omitempty" json:"adx,omitempty"`
    PlusDI    *float64 `bson:"plus_di,omitempty" json:"plus_di,omitempty"`
    MinusDI   *float64 `bson:"minus_di,omitempty" json:"minus_di,omitempty"`

    // SuperTrend
    SuperTrend      *float64 `bson:"supertrend,omitempty" json:"supertrend,omitempty"`
    SuperTrendSignal *int     `bson:"supertrend_signal,omitempty" json:"supertrend_signal,omitempty"` // 1=buy, -1=sell

    // Momentum Indicators
    RSI6    *float64 `bson:"rsi6,omitempty" json:"rsi6,omitempty"`
    RSI14   *float64 `bson:"rsi14,omitempty" json:"rsi14,omitempty"`
    RSI24   *float64 `bson:"rsi24,omitempty" json:"rsi24,omitempty"`

    // Stochastic
    StochK  *float64 `bson:"stoch_k,omitempty" json:"stoch_k,omitempty"`
    StochD  *float64 `bson:"stoch_d,omitempty" json:"stoch_d,omitempty"`

    // MACD
    MACD        *float64 `bson:"macd,omitempty" json:"macd,omitempty"`
    MACDSignal  *float64 `bson:"macd_signal,omitempty" json:"macd_signal,omitempty"`
    MACDHist    *float64 `bson:"macd_hist,omitempty" json:"macd_hist,omitempty"`

    // ROC
    ROC     *float64 `bson:"roc,omitempty" json:"roc,omitempty"`

    // CCI
    CCI     *float64 `bson:"cci,omitempty" json:"cci,omitempty"`

    // Williams %R
    WilliamsR *float64 `bson:"williams_r,omitempty" json:"williams_r,omitempty"`

    // Momentum
    Momentum *float64 `bson:"momentum,omitempty" json:"momentum,omitempty"`

    // Volatility Indicators
    BollingerUpper    *float64 `bson:"bb_upper,omitempty" json:"bb_upper,omitempty"`
    BollingerMiddle   *float64 `bson:"bb_middle,omitempty" json:"bb_middle,omitempty"`
    BollingerLower    *float64 `bson:"bb_lower,omitempty" json:"bb_lower,omitempty"`
    BollingerBandwidth *float64 `bson:"bb_bandwidth,omitempty" json:"bb_bandwidth,omitempty"`
    BollingerPercentB  *float64 `bson:"bb_percent_b,omitempty" json:"bb_percent_b,omitempty"`

    ATR     *float64 `bson:"atr,omitempty" json:"atr,omitempty"`

    // Keltner Channels
    KeltnerUpper  *float64 `bson:"keltner_upper,omitempty" json:"keltner_upper,omitempty"`
    KeltnerMiddle *float64 `bson:"keltner_middle,omitempty" json:"keltner_middle,omitempty"`
    KeltnerLower  *float64 `bson:"keltner_lower,omitempty" json:"keltner_lower,omitempty"`

    // Donchian Channels
    DonchianUpper  *float64 `bson:"donchian_upper,omitempty" json:"donchian_upper,omitempty"`
    DonchianMiddle *float64 `bson:"donchian_middle,omitempty" json:"donchian_middle,omitempty"`
    DonchianLower  *float64 `bson:"donchian_lower,omitempty" json:"donchian_lower,omitempty"`

    // Standard Deviation
    StdDev *float64 `bson:"stddev,omitempty" json:"stddev,omitempty"`

    // Volume Indicators
    OBV         *float64 `bson:"obv,omitempty" json:"obv,omitempty"`
    VWAP        *float64 `bson:"vwap,omitempty" json:"vwap,omitempty"`
    MFI         *float64 `bson:"mfi,omitempty" json:"mfi,omitempty"`
    CMF         *float64 `bson:"cmf,omitempty" json:"cmf,omitempty"`
    VolumeSMA   *float64 `bson:"volume_sma,omitempty" json:"volume_sma,omitempty"`
}
```

---

## Phase 2: Implement All 29 Indicators

### 2.1 Complete Partial Implementations (3)
- ✅ SMA - Add to service integration
- ✅ Bollinger Bands - Add to service integration
- ✅ ATR - Add to service integration

### 2.2 Implement Trend Indicators (7)
1. **DEMA** - `dema.go`
2. **TEMA** - `tema.go`
3. **WMA** - `wma.go`
4. **HMA** - `hma.go`
5. **VWMA** - `vwma.go`
6. **Ichimoku Cloud** - `ichimoku.go`
7. **ADX/DMI** - `adx.go`
8. **SuperTrend** - `supertrend.go`

### 2.3 Implement Momentum Indicators (4)
1. **Stochastic Oscillator** - `stochastic.go`
2. **ROC** - `roc.go`
3. **CCI** - `cci.go`
4. **Williams %R** - `williams_r.go`
5. **Momentum** - `momentum.go`

### 2.4 Implement Volatility Indicators (3)
1. **Keltner Channels** - `keltner.go`
2. **Donchian Channels** - `donchian.go`
3. **Standard Deviation** - `stddev.go`

### 2.5 Implement Volume Indicators (5)
1. **OBV** - `obv.go`
2. **VWAP** - `vwap.go`
3. **MFI** - `mfi.go`
4. **CMF** - `cmf.go`
5. **Volume SMA** - `volume_sma.go`

---

## Phase 3: Configuration System

### 3.1 Default Configurations

```go
// internal/service/indicators/defaults.go
func DefaultConnectorConfig() *models.IndicatorConfig {
    return &models.IndicatorConfig{
        RSI: &models.RSIConfig{
            Enabled: true,
            Periods: []int{6, 14, 24},
        },
        EMA: &models.EMAConfig{
            Enabled: true,
            Periods: []int{12, 26},
        },
        MACD: &models.MACDConfig{
            Enabled: true,
            FastPeriod: 12,
            SlowPeriod: 26,
            SignalPeriod: 9,
        },
        // ... all others disabled by default
    }
}
```

### 3.2 Configuration Inheritance

```go
// internal/service/indicators/config_merge.go
func MergeConfigs(connectorConfig, jobConfig *models.IndicatorConfig) *models.IndicatorConfig {
    // Job config overrides connector config
    // If job config is nil, use connector config
    // If both nil, use defaults
}
```

### 3.3 Selective Calculation

```go
// Update service.go CalculateAll to only calculate enabled indicators
func (s *Service) CalculateAll(candles []models.Candle, config *models.IndicatorConfig) ([]models.Candle, error) {
    // Check each indicator's Enabled flag
    if config.RSI != nil && config.RSI.Enabled {
        // Calculate RSI
    }
    if config.MACD != nil && config.MACD.Enabled {
        // Calculate MACD
    }
    // etc...
}
```

---

## Phase 4: API Endpoints

### 4.1 Connector Indicator Configuration

```
POST   /api/v1/connectors/:id/indicators/config
GET    /api/v1/connectors/:id/indicators/config
PUT    /api/v1/connectors/:id/indicators/config
PATCH  /api/v1/connectors/:id/indicators/config
POST   /api/v1/connectors/:id/indicators/recalculate
```

### 4.2 Job Indicator Configuration

```
POST   /api/v1/jobs/:id/indicators/config
GET    /api/v1/jobs/:id/indicators/config
PUT    /api/v1/jobs/:id/indicators/config
PATCH  /api/v1/jobs/:id/indicators/config
POST   /api/v1/jobs/:id/indicators/recalculate
```

### 4.3 Indicator Data Retrieval

```
GET /api/v1/indicators/:exchange/:symbol/:timeframe/latest
GET /api/v1/indicators/:exchange/:symbol/:timeframe/range?from=X&to=Y
GET /api/v1/indicators/:exchange/:symbol/:timeframe/:indicator?limit=100
```

### 4.4 Bulk Operations

```
POST /api/v1/indicators/recalculate-all  # Recalculate for all jobs
```

---

## Phase 5: Service Updates

### 5.1 Update JobExecutor

```go
func (e *JobExecutor) ExecuteJob(ctx context.Context, jobID string) (*models.JobExecutionResult, error) {
    // ... fetch candles ...

    // Get merged indicator config (job overrides connector)
    config := e.getMergedIndicatorConfig(job, connector)

    // Calculate only enabled indicators
    candles, err = e.indicatorService.CalculateAll(candles, config)

    // ... store candles ...
}

func (e *JobExecutor) getMergedIndicatorConfig(job *models.Job, connector *models.Connector) *models.IndicatorConfig {
    return indicators.MergeConfigs(connector.IndicatorConfig, job.IndicatorConfig)
}
```

### 5.2 Recalculation Service

```go
// internal/service/recalculator.go
type RecalculatorService struct {
    jobRepo   *repository.JobRepository
    ohlcvRepo *repository.OHLCVRepository
    indService *indicators.Service
}

func (r *RecalculatorService) RecalculateJob(ctx context.Context, jobID string) error {
    // 1. Fetch all candles for job
    // 2. Get indicator config
    // 3. Recalculate all indicators
    // 4. Update candles in database
}

func (r *RecalculatorService) RecalculateConnector(ctx context.Context, connectorID string) error {
    // 1. Find all jobs for connector
    // 2. Recalculate each job
}
```

---

## Phase 6: Frontend UI

### 6.1 Connector Configuration Page

```
/connectors/:id/indicators

Components:
- Indicator toggle switches (enable/disable)
- Period/parameter inputs
- Save button
- Recalculate button (icon: 🔄)
```

### 6.2 Job Configuration Page

```
/jobs/:id/indicators

Components:
- "Use connector defaults" checkbox
- Override indicator toggles
- Parameter customization
- Recalculate button
```

### 6.3 Job/Connector List Enhancement

Add recalculate icon button next to each job/connector:
```jsx
<IconButton onClick={handleRecalculate}>
  <RefreshIcon />
</IconButton>
```

---

## Phase 7: Testing Strategy

### 7.1 Unit Tests
- Test each indicator calculation
- Test configuration merging
- Test selective calculation

### 7.2 Integration Tests
- Test connector config → job inheritance
- Test recalculation endpoints
- Test indicator data retrieval

### 7.3 Performance Tests
- Benchmark all 29 indicators on 200 candles
- Memory usage profiling
- Database query optimization

---

## Implementation Order

### Day 1: Foundation (6-8 hours)
1. ✅ Update database models (Connector, Job, OHLCV)
2. ✅ Create configuration structs
3. ✅ Implement configuration merging logic
4. ✅ Complete 3 partial indicators (SMA, Bollinger, ATR)

### Day 2: Indicator Implementation Part 1 (8 hours)
5. ✅ Implement all Trend indicators (8)
6. ✅ Test trend indicators

### Day 3: Indicator Implementation Part 2 (8 hours)
7. ✅ Implement all Momentum indicators (4)
8. ✅ Implement all Volatility indicators (3)
9. ✅ Test momentum and volatility indicators

### Day 4: Indicator Implementation Part 3 (6 hours)
10. ✅ Implement all Volume indicators (5)
11. ✅ Test volume indicators
12. ✅ Update service.go for all 29 indicators

### Day 5: Backend API (8 hours)
13. ✅ Create API endpoints for config management
14. ✅ Create API endpoints for data retrieval
15. ✅ Implement recalculation service
16. ✅ Test all API endpoints

### Day 6: Frontend (8 hours)
17. ✅ Create indicator config UI for connectors
18. ✅ Create indicator config UI for jobs
19. ✅ Add recalculate buttons
20. ✅ Test UI integration

### Day 7: Testing & Documentation (6 hours)
21. ✅ End-to-end testing
22. ✅ Performance optimization
23. ✅ Update documentation
24. ✅ Create user guide

---

## File Structure

```
internal/
├── models/
│   ├── connector.go          # Add IndicatorConfig field
│   ├── job.go                # Add IndicatorConfig field
│   ├── ohlcv.go              # Add all 60+ indicator fields
│   └── indicator_config.go   # NEW: Config structs
│
├── service/
│   ├── indicators/
│   │   ├── helpers.go
│   │   ├── service.go        # Update with all 29 indicators
│   │   ├── config.go         # NEW: Config management
│   │   ├── defaults.go       # NEW: Default configs
│   │   ├── merge.go          # NEW: Config merging
│   │   │
│   │   # Existing (update)
│   │   ├── rsi.go
│   │   ├── ema.go
│   │   ├── macd.go
│   │   ├── sma.go
│   │   ├── bollinger.go
│   │   ├── atr.go
│   │   │
│   │   # NEW Trend Indicators
│   │   ├── dema.go
│   │   ├── tema.go
│   │   ├── wma.go
│   │   ├── hma.go
│   │   ├── vwma.go
│   │   ├── ichimoku.go
│   │   ├── adx.go
│   │   ├── supertrend.go
│   │   │
│   │   # NEW Momentum Indicators
│   │   ├── stochastic.go
│   │   ├── roc.go
│   │   ├── cci.go
│   │   ├── williams_r.go
│   │   ├── momentum.go
│   │   │
│   │   # NEW Volatility Indicators
│   │   ├── keltner.go
│   │   ├── donchian.go
│   │   ├── stddev.go
│   │   │
│   │   # NEW Volume Indicators
│   │   ├── obv.go
│   │   ├── vwap.go
│   │   ├── mfi.go
│   │   ├── cmf.go
│   │   └── volume_sma.go
│   │
│   ├── job_executor.go       # Update to use config
│   └── recalculator.go       # NEW: Recalculation service
│
├── handler/
│   ├── connector_handler.go  # Add indicator config endpoints
│   ├── job_handler.go        # Add indicator config endpoints
│   └── indicator_handler.go  # NEW: Indicator data endpoints
│
└── repository/
    └── ohlcv_repository.go   # Add indicator query methods

web/
└── src/
    ├── components/
    │   ├── IndicatorConfig.jsx      # NEW
    │   ├── IndicatorToggle.jsx      # NEW
    │   └── RecalculateButton.jsx    # NEW
    │
    └── pages/
        ├── ConnectorDetail.jsx       # Add indicator config
        └── JobDetail.jsx             # Add indicator config
```

---

## Success Criteria

✅ All 29 indicators implemented and tested
✅ Configuration system working (Connector + Job levels)
✅ Job config properly overrides Connector config
✅ Recalculate functionality working for both levels
✅ API endpoints functional
✅ Frontend UI complete with toggles and recalculate buttons
✅ Performance acceptable (<50ms for all indicators on 200 candles)
✅ Documentation updated
✅ End-to-end tested with real data

---

## Estimated Timeline

- **Minimum (focused work)**: 5-6 days
- **Realistic (with breaks/testing)**: 7-10 days
- **Conservative (with issues)**: 10-14 days

---

## Risks & Mitigation

### Risk 1: Performance degradation with 29 indicators
**Mitigation**:
- Selective calculation (only enabled indicators)
- Parallel processing for independent indicators
- Caching intermediate results

### Risk 2: Complex configuration inheritance
**Mitigation**:
- Clear merge rules (job always overrides)
- Validation at API level
- Default fallbacks

### Risk 3: Database storage size
**Mitigation**:
- Use pointers (omitempty) for null values
- MongoDB compression
- TTL for old data

### Risk 4: Frontend complexity
**Mitigation**:
- Reusable components
- Clear grouping by indicator category
- Collapsible sections

---

## Next Immediate Steps

1. Get approval for this plan
2. Start with Phase 1: Database models
3. Implement indicators in batches
4. Continuous testing as we build

---

Ready to start implementation?
