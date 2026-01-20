# INDICATORS-CATALOG: Technical Indicator Specifications

## Document Information

| Field | Value |
|-------|-------|
| **Version** | 1.0 |
| **Last Updated** | 2026-01-17 |
| **Status** | Active |

---

## Overview

This catalog defines all technical indicators supported by the DataCollector. Indicators are computed on the **backend** and stored in MongoDB for efficient retrieval. The frontend selects and displays pre-computed indicator values.

### Computation Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    INDICATOR COMPUTATION FLOW                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. OHLCV Data Arrives (via Collector Service)                  │
│         │                                                        │
│         ▼                                                        │
│  2. Indicator Service Triggered                                  │
│         │                                                        │
│         ▼                                                        │
│  3. For Each Configured Indicator:                               │
│         │                                                        │
│         ├─► Load Required Historical Candles (lookback period)  │
│         │                                                        │
│         ├─► Compute Indicator Values                            │
│         │                                                        │
│         └─► Store in MongoDB (indicator_values collection)      │
│                                                                  │
│  4. Frontend Queries Pre-Computed Values via API                │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```


---

## 1. Trend Indicators (Smoothing / Direction)

Trend indicators smooth price data and identify the direction of the prevailing market trend.

---

### 1.1 SMA (Simple Moving Average)

| Property | Value |
|----------|-------|
| **ID** | `sma` |
| **Category** | Trend |
| **Description** | Arithmetic mean of prices over a specified period |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `period` | int | 20 | 2-500 | Number of periods to average |
| `source` | string | "close" | open, high, low, close, hl2, hlc3, ohlc4 | Price source |

#### Formula

```
SMA = (P1 + P2 + ... + Pn) / n

where:
  P = Price at period
  n = period parameter
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `value` | float64 | The SMA value |

#### Go Implementation

```go
func CalculateSMA(candles []OHLCV, period int, source string) []float64 {
    if len(candles) < period {
        return nil
    }

    result := make([]float64, len(candles))
    prices := extractPrices(candles, source)

    // Calculate initial SMA
    sum := 0.0
    for i := 0; i < period; i++ {
        sum += prices[i]
        result[i] = math.NaN() // Not enough data
    }
    result[period-1] = sum / float64(period)

    // Rolling calculation
    for i := period; i < len(candles); i++ {
        sum = sum - prices[i-period] + prices[i]
        result[i] = sum / float64(period)
    }

    return result
}
```

#### Use Cases
- Identifying trend direction (price above SMA = bullish)
- Dynamic support/resistance levels
- Crossover signals (SMA9 crossing SMA21)

---

### 1.2 EMA (Exponential Moving Average)

| Property | Value |
|----------|-------|
| **ID** | `ema` |
| **Category** | Trend |
| **Description** | Weighted moving average giving more weight to recent prices |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `period` | int | 20 | 2-500 | EMA period |
| `source` | string | "close" | open, high, low, close, hl2, hlc3, ohlc4 | Price source |

#### Formula

```
Multiplier = 2 / (period + 1)
EMA = (Price - EMA_prev) * Multiplier + EMA_prev

Initial EMA = SMA of first 'period' values
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `value` | float64 | The EMA value |

#### Go Implementation

```go
func CalculateEMA(candles []OHLCV, period int, source string) []float64 {
    if len(candles) < period {
        return nil
    }

    result := make([]float64, len(candles))
    prices := extractPrices(candles, source)
    multiplier := 2.0 / float64(period+1)

    // Initial SMA for first EMA value
    sum := 0.0
    for i := 0; i < period; i++ {
        sum += prices[i]
        result[i] = math.NaN()
    }
    result[period-1] = sum / float64(period)

    // EMA calculation
    for i := period; i < len(candles); i++ {
        result[i] = (prices[i]-result[i-1])*multiplier + result[i-1]
    }

    return result
}
```

#### Use Cases
- Faster trend detection than SMA
- MACD component
- Popular periods: 9, 12, 21, 26, 50, 200

---

### 1.3 DEMA (Double Exponential Moving Average)

| Property | Value |
|----------|-------|
| **ID** | `dema` |
| **Category** | Trend |
| **Description** | Reduces lag by applying EMA twice and adjusting |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `period` | int | 20 | 2-500 | DEMA period |
| `source` | string | "close" | open, high, low, close, hl2, hlc3, ohlc4 | Price source |

#### Formula

```
EMA1 = EMA(Price, period)
EMA2 = EMA(EMA1, period)
DEMA = 2 * EMA1 - EMA2
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `value` | float64 | The DEMA value |

#### Use Cases
- Reduced lag compared to EMA
- Trend following with faster response

---

### 1.4 TEMA (Triple Exponential Moving Average)

| Property | Value |
|----------|-------|
| **ID** | `tema` |
| **Category** | Trend |
| **Description** | Further reduces lag by applying EMA three times |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `period` | int | 20 | 2-500 | TEMA period |
| `source` | string | "close" | open, high, low, close, hl2, hlc3, ohlc4 | Price source |

#### Formula

```
EMA1 = EMA(Price, period)
EMA2 = EMA(EMA1, period)
EMA3 = EMA(EMA2, period)
TEMA = 3 * EMA1 - 3 * EMA2 + EMA3
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `value` | float64 | The TEMA value |

---

### 1.5 WMA (Weighted Moving Average)

| Property | Value |
|----------|-------|
| **ID** | `wma` |
| **Category** | Trend |
| **Description** | Linear weighted average, recent prices have higher weight |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `period` | int | 20 | 2-500 | WMA period |
| `source` | string | "close" | open, high, low, close, hl2, hlc3, ohlc4 | Price source |

#### Formula

```
WMA = (P1*1 + P2*2 + ... + Pn*n) / (1 + 2 + ... + n)

Denominator = n * (n + 1) / 2
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `value` | float64 | The WMA value |

---

### 1.6 HMA (Hull Moving Average)

| Property | Value |
|----------|-------|
| **ID** | `hma` |
| **Category** | Trend |
| **Description** | Low-lag moving average using weighted moving averages |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `period` | int | 20 | 2-500 | HMA period |
| `source` | string | "close" | open, high, low, close, hl2, hlc3, ohlc4 | Price source |

#### Formula

```
HMA = WMA(2 * WMA(Price, period/2) - WMA(Price, period), sqrt(period))
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `value` | float64 | The HMA value |

#### Use Cases
- Very responsive to price changes
- Good for identifying trend reversals quickly

---

### 1.7 VWMA (Volume Weighted Moving Average)

| Property | Value |
|----------|-------|
| **ID** | `vwma` |
| **Category** | Trend |
| **Description** | Moving average weighted by volume |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `period` | int | 20 | 2-500 | VWMA period |
| `source` | string | "close" | open, high, low, close, hl2, hlc3, ohlc4 | Price source |

#### Formula

```
VWMA = Σ(Price * Volume) / Σ(Volume)

over the specified period
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `value` | float64 | The VWMA value |

#### Use Cases
- More accurate representation of average price when volume is considered
- Divergence from price can indicate trend strength

---

### 1.8 Ichimoku Cloud

| Property | Value |
|----------|-------|
| **ID** | `ichimoku` |
| **Category** | Trend |
| **Description** | Comprehensive trend system with multiple components |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `tenkan_period` | int | 9 | 2-100 | Tenkan-sen (Conversion Line) period |
| `kijun_period` | int | 26 | 2-100 | Kijun-sen (Base Line) period |
| `senkou_b_period` | int | 52 | 2-200 | Senkou Span B period |
| `displacement` | int | 26 | 1-100 | Cloud displacement (Chikou, Senkou forward) |

#### Formula

```
Tenkan-sen = (Highest High + Lowest Low) / 2 over tenkan_period
Kijun-sen = (Highest High + Lowest Low) / 2 over kijun_period
Senkou Span A = (Tenkan-sen + Kijun-sen) / 2, plotted displacement periods ahead
Senkou Span B = (Highest High + Lowest Low) / 2 over senkou_b_period, plotted displacement periods ahead
Chikou Span = Close, plotted displacement periods behind
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `tenkan` | float64 | Tenkan-sen (Conversion Line) |
| `kijun` | float64 | Kijun-sen (Base Line) |
| `senkou_a` | float64 | Senkou Span A (Leading Span A) |
| `senkou_b` | float64 | Senkou Span B (Leading Span B) |
| `chikou` | float64 | Chikou Span (Lagging Span) |

#### Go Implementation

```go
func CalculateIchimoku(candles []OHLCV, tenkan, kijun, senkouB, displacement int) *IchimokuResult {
    n := len(candles)
    result := &IchimokuResult{
        Tenkan:  make([]float64, n),
        Kijun:   make([]float64, n),
        SenkouA: make([]float64, n+displacement),
        SenkouB: make([]float64, n+displacement),
        Chikou:  make([]float64, n),
    }

    // Helper: (highest high + lowest low) / 2
    midpoint := func(start, period int) float64 {
        high := candles[start].High
        low := candles[start].Low
        for i := start + 1; i < start+period && i < n; i++ {
            if candles[i].High > high {
                high = candles[i].High
            }
            if candles[i].Low < low {
                low = candles[i].Low
            }
        }
        return (high + low) / 2
    }

    for i := 0; i < n; i++ {
        // Tenkan-sen
        if i >= tenkan-1 {
            result.Tenkan[i] = midpoint(i-tenkan+1, tenkan)
        }

        // Kijun-sen
        if i >= kijun-1 {
            result.Kijun[i] = midpoint(i-kijun+1, kijun)
        }

        // Senkou Span A (displaced forward)
        if i >= kijun-1 {
            result.SenkouA[i+displacement] = (result.Tenkan[i] + result.Kijun[i]) / 2
        }

        // Senkou Span B (displaced forward)
        if i >= senkouB-1 {
            result.SenkouB[i+displacement] = midpoint(i-senkouB+1, senkouB)
        }

        // Chikou Span (current close, will be plotted displacement bars back)
        result.Chikou[i] = candles[i].Close
    }

    return result
}
```

#### Use Cases
- Complete trend analysis system
- Cloud provides support/resistance
- TK cross for entry signals

---

### 1.9 ADX/DMI (Average Directional Index / Directional Movement Index)

| Property | Value |
|----------|-------|
| **ID** | `adx` |
| **Category** | Trend |
| **Description** | Measures trend strength regardless of direction |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `period` | int | 14 | 2-100 | ADX smoothing period |

#### Formula

```
+DM = Current High - Previous High (if positive and > -DM, else 0)
-DM = Previous Low - Current Low (if positive and > +DM, else 0)
TR = max(High-Low, |High-PrevClose|, |Low-PrevClose|)

Smoothed +DM, -DM, TR using Wilder's smoothing

+DI = 100 * Smoothed(+DM) / Smoothed(TR)
-DI = 100 * Smoothed(-DM) / Smoothed(TR)

DX = 100 * |+DI - -DI| / (+DI + -DI)
ADX = Wilder's smoothed average of DX
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `adx` | float64 | ADX value (0-100) |
| `plus_di` | float64 | +DI value |
| `minus_di` | float64 | -DI value |

#### Use Cases
- ADX > 25 indicates trending market
- ADX < 20 indicates ranging/sideways market
- +DI/-DI crossovers for direction signals

---

### 1.10 SuperTrend

| Property | Value |
|----------|-------|
| **ID** | `supertrend` |
| **Category** | Trend |
| **Description** | Trend-following indicator based on ATR bands |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `period` | int | 10 | 2-100 | ATR period |
| `multiplier` | float64 | 3.0 | 0.5-10.0 | ATR multiplier |

#### Formula

```
HL2 = (High + Low) / 2
ATR = Average True Range over period

Upper Band = HL2 + (Multiplier * ATR)
Lower Band = HL2 - (Multiplier * ATR)

SuperTrend = Upper Band if Close <= Previous SuperTrend AND Close > Upper Band
           = Lower Band if Close >= Previous SuperTrend AND Close < Lower Band
           = Continue previous trend otherwise

Trend = 1 (bullish) if price is above SuperTrend
      = -1 (bearish) if price is below SuperTrend
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `value` | float64 | SuperTrend value (support/resistance level) |
| `trend` | int | 1 for bullish, -1 for bearish |

#### Go Implementation

```go
func CalculateSuperTrend(candles []OHLCV, period int, multiplier float64) *SuperTrendResult {
    n := len(candles)
    atr := CalculateATR(candles, period)

    result := &SuperTrendResult{
        Value: make([]float64, n),
        Trend: make([]int, n),
    }

    upperBand := make([]float64, n)
    lowerBand := make([]float64, n)

    for i := period - 1; i < n; i++ {
        hl2 := (candles[i].High + candles[i].Low) / 2
        upperBand[i] = hl2 + multiplier*atr[i]
        lowerBand[i] = hl2 - multiplier*atr[i]

        if i == period-1 {
            result.Trend[i] = 1
            result.Value[i] = lowerBand[i]
            continue
        }

        // Adjust bands based on previous values
        if lowerBand[i] > lowerBand[i-1] || candles[i-1].Close < lowerBand[i-1] {
            lowerBand[i] = lowerBand[i]
        } else {
            lowerBand[i] = lowerBand[i-1]
        }

        if upperBand[i] < upperBand[i-1] || candles[i-1].Close > upperBand[i-1] {
            upperBand[i] = upperBand[i]
        } else {
            upperBand[i] = upperBand[i-1]
        }

        // Determine trend
        if result.Trend[i-1] == 1 {
            if candles[i].Close < lowerBand[i] {
                result.Trend[i] = -1
                result.Value[i] = upperBand[i]
            } else {
                result.Trend[i] = 1
                result.Value[i] = lowerBand[i]
            }
        } else {
            if candles[i].Close > upperBand[i] {
                result.Trend[i] = 1
                result.Value[i] = lowerBand[i]
            } else {
                result.Trend[i] = -1
                result.Value[i] = upperBand[i]
            }
        }
    }

    return result
}
```

#### Use Cases
- Clear buy/sell signals on trend changes
- Trailing stop-loss levels
- Works well in trending markets

---

## 2. Momentum Indicators (Strength / Oscillators)

Momentum indicators measure the rate of price change and identify overbought/oversold conditions.

---

### 2.1 RSI (Relative Strength Index)

| Property | Value |
|----------|-------|
| **ID** | `rsi` |
| **Category** | Momentum |
| **Description** | Measures speed and magnitude of price changes |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `period` | int | 14 | 2-100 | RSI period |
| `source` | string | "close" | open, high, low, close | Price source |

#### Formula

```
Change = Current Price - Previous Price
Gain = Change if Change > 0, else 0
Loss = -Change if Change < 0, else 0

Avg Gain = Wilder's smoothed average of Gains
Avg Loss = Wilder's smoothed average of Losses

RS = Avg Gain / Avg Loss
RSI = 100 - (100 / (1 + RS))
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `value` | float64 | RSI value (0-100) |

#### Go Implementation

```go
func CalculateRSI(candles []OHLCV, period int, source string) []float64 {
    n := len(candles)
    if n < period+1 {
        return nil
    }

    result := make([]float64, n)
    prices := extractPrices(candles, source)

    gains := make([]float64, n)
    losses := make([]float64, n)

    // Calculate gains and losses
    for i := 1; i < n; i++ {
        change := prices[i] - prices[i-1]
        if change > 0 {
            gains[i] = change
        } else {
            losses[i] = -change
        }
    }

    // Initial averages (SMA)
    avgGain := 0.0
    avgLoss := 0.0
    for i := 1; i <= period; i++ {
        avgGain += gains[i]
        avgLoss += losses[i]
        result[i-1] = math.NaN()
    }
    avgGain /= float64(period)
    avgLoss /= float64(period)

    // First RSI
    if avgLoss == 0 {
        result[period] = 100
    } else {
        rs := avgGain / avgLoss
        result[period] = 100 - (100 / (1 + rs))
    }

    // Subsequent RSI values using Wilder's smoothing
    for i := period + 1; i < n; i++ {
        avgGain = (avgGain*float64(period-1) + gains[i]) / float64(period)
        avgLoss = (avgLoss*float64(period-1) + losses[i]) / float64(period)

        if avgLoss == 0 {
            result[i] = 100
        } else {
            rs := avgGain / avgLoss
            result[i] = 100 - (100 / (1 + rs))
        }
    }

    return result
}
```

#### Use Cases
- Overbought (>70) / Oversold (<30) signals
- Divergence detection
- Trend confirmation

---

### 2.2 Stochastic Oscillator

| Property | Value |
|----------|-------|
| **ID** | `stochastic` |
| **Category** | Momentum |
| **Description** | Compares closing price to price range over a period |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `k_period` | int | 14 | 2-100 | %K lookback period |
| `k_smooth` | int | 1 | 1-10 | %K smoothing (1 = Fast Stochastic) |
| `d_period` | int | 3 | 2-50 | %D smoothing period (signal line) |

#### Formula

```
%K = 100 * (Close - Lowest Low) / (Highest High - Lowest Low)

Fast Stochastic: %K as calculated, %D = SMA(%K, d_period)
Slow Stochastic: %K = SMA(Fast %K, k_smooth), %D = SMA(%K, d_period)
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `k` | float64 | %K line value (0-100) |
| `d` | float64 | %D signal line value (0-100) |

#### Use Cases
- Overbought (>80) / Oversold (<20) signals
- %K/%D crossovers for entry/exit
- Divergence analysis

---

### 2.3 MACD (Moving Average Convergence Divergence)

| Property | Value |
|----------|-------|
| **ID** | `macd` |
| **Category** | Momentum |
| **Description** | Trend-following momentum indicator using EMA differences |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `fast_period` | int | 12 | 2-100 | Fast EMA period |
| `slow_period` | int | 26 | 2-200 | Slow EMA period |
| `signal_period` | int | 9 | 2-50 | Signal line EMA period |
| `source` | string | "close" | open, high, low, close | Price source |

#### Formula

```
MACD Line = EMA(fast_period) - EMA(slow_period)
Signal Line = EMA(MACD Line, signal_period)
Histogram = MACD Line - Signal Line
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `macd` | float64 | MACD line value |
| `signal` | float64 | Signal line value |
| `histogram` | float64 | MACD histogram |

#### Go Implementation

```go
func CalculateMACD(candles []OHLCV, fast, slow, signal int, source string) *MACDResult {
    n := len(candles)
    prices := extractPrices(candles, source)

    fastEMA := CalculateEMA(candles, fast, source)
    slowEMA := CalculateEMA(candles, slow, source)

    result := &MACDResult{
        MACD:      make([]float64, n),
        Signal:    make([]float64, n),
        Histogram: make([]float64, n),
    }

    // Calculate MACD line
    for i := slow - 1; i < n; i++ {
        result.MACD[i] = fastEMA[i] - slowEMA[i]
    }

    // Calculate Signal line (EMA of MACD)
    signalMultiplier := 2.0 / float64(signal+1)
    result.Signal[slow+signal-2] = result.MACD[slow-1]
    for i := slow; i < slow+signal-1; i++ {
        result.Signal[slow+signal-2] += result.MACD[i]
    }
    result.Signal[slow+signal-2] /= float64(signal)

    for i := slow + signal - 1; i < n; i++ {
        result.Signal[i] = (result.MACD[i]-result.Signal[i-1])*signalMultiplier + result.Signal[i-1]
        result.Histogram[i] = result.MACD[i] - result.Signal[i]
    }

    return result
}
```

#### Use Cases
- Signal line crossovers for buy/sell
- Zero line crossovers for trend direction
- Histogram divergence for momentum shifts

---

### 2.4 ROC (Rate of Change)

| Property | Value |
|----------|-------|
| **ID** | `roc` |
| **Category** | Momentum |
| **Description** | Percentage change in price over a period |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `period` | int | 14 | 1-200 | Lookback period |
| `source` | string | "close" | open, high, low, close | Price source |

#### Formula

```
ROC = ((Current Price - Price n periods ago) / Price n periods ago) * 100
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `value` | float64 | ROC percentage |

#### Use Cases
- Momentum divergence detection
- Overbought/oversold when extreme
- Zero line crossovers

---

### 2.5 CCI (Commodity Channel Index)

| Property | Value |
|----------|-------|
| **ID** | `cci` |
| **Category** | Momentum |
| **Description** | Measures price deviation from statistical mean |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `period` | int | 20 | 2-200 | CCI period |

#### Formula

```
Typical Price (TP) = (High + Low + Close) / 3
SMA_TP = SMA of Typical Price over period
Mean Deviation = Average of |TP - SMA_TP| over period

CCI = (TP - SMA_TP) / (0.015 * Mean Deviation)
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `value` | float64 | CCI value (unbounded, typically -200 to +200) |

#### Use Cases
- Overbought (>100) / Oversold (<-100)
- Trend identification
- Divergence analysis

---

### 2.6 Williams %R

| Property | Value |
|----------|-------|
| **ID** | `williams_r` |
| **Category** | Momentum |
| **Description** | Shows relationship of close to high-low range |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `period` | int | 14 | 2-100 | Lookback period |

#### Formula

```
%R = (Highest High - Close) / (Highest High - Lowest Low) * -100

Range: -100 to 0
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `value` | float64 | Williams %R value (-100 to 0) |

#### Use Cases
- Overbought (> -20) / Oversold (< -80)
- Failure swings
- Divergence detection

---

### 2.7 Momentum

| Property | Value |
|----------|-------|
| **ID** | `momentum` |
| **Category** | Momentum |
| **Description** | Simple difference between current and past price |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `period` | int | 10 | 1-200 | Lookback period |
| `source` | string | "close" | open, high, low, close | Price source |

#### Formula

```
Momentum = Current Price - Price n periods ago
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `value` | float64 | Momentum value |

#### Use Cases
- Simple trend direction (positive = up, negative = down)
- Zero line crossovers
- Momentum shifts

---

## 3. Volatility Indicators (Range / Bands / Regimes)

Volatility indicators measure the degree of price variation over time.

---

### 3.1 Bollinger Bands

| Property | Value |
|----------|-------|
| **ID** | `bollinger` |
| **Category** | Volatility |
| **Description** | Dynamic bands based on standard deviation around a moving average |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `period` | int | 20 | 2-200 | Moving average period |
| `std_dev` | float64 | 2.0 | 0.5-5.0 | Standard deviation multiplier |
| `source` | string | "close" | open, high, low, close | Price source |

#### Formula

```
Middle Band = SMA(Price, period)
Upper Band = Middle Band + (std_dev * StdDev(Price, period))
Lower Band = Middle Band - (std_dev * StdDev(Price, period))
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `upper` | float64 | Upper band |
| `middle` | float64 | Middle band (SMA) |
| `lower` | float64 | Lower band |
| `bandwidth` | float64 | (Upper - Lower) / Middle * 100 |
| `percent_b` | float64 | (Price - Lower) / (Upper - Lower) |

#### Go Implementation

```go
func CalculateBollingerBands(candles []OHLCV, period int, stdDev float64, source string) *BollingerResult {
    n := len(candles)
    prices := extractPrices(candles, source)
    sma := CalculateSMA(candles, period, source)

    result := &BollingerResult{
        Upper:     make([]float64, n),
        Middle:    sma,
        Lower:     make([]float64, n),
        Bandwidth: make([]float64, n),
        PercentB:  make([]float64, n),
    }

    for i := period - 1; i < n; i++ {
        // Calculate standard deviation
        sum := 0.0
        for j := i - period + 1; j <= i; j++ {
            diff := prices[j] - sma[i]
            sum += diff * diff
        }
        sd := math.Sqrt(sum / float64(period))

        result.Upper[i] = sma[i] + stdDev*sd
        result.Lower[i] = sma[i] - stdDev*sd
        result.Bandwidth[i] = (result.Upper[i] - result.Lower[i]) / sma[i] * 100

        if result.Upper[i] != result.Lower[i] {
            result.PercentB[i] = (prices[i] - result.Lower[i]) / (result.Upper[i] - result.Lower[i])
        }
    }

    return result
}
```

#### Use Cases
- Band squeezes indicate impending volatility
- Percent B for overbought/oversold
- Mean reversion strategies

---

### 3.2 ATR (Average True Range)

| Property | Value |
|----------|-------|
| **ID** | `atr` |
| **Category** | Volatility |
| **Description** | Measures market volatility using true range |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `period` | int | 14 | 2-100 | ATR smoothing period |

#### Formula

```
True Range = max(
    High - Low,
    |High - Previous Close|,
    |Low - Previous Close|
)

ATR = Wilder's smoothed average of True Range
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `value` | float64 | ATR value |
| `tr` | float64 | True Range (unsmoothed) |

#### Go Implementation

```go
func CalculateATR(candles []OHLCV, period int) []float64 {
    n := len(candles)
    if n < period {
        return nil
    }

    result := make([]float64, n)
    tr := make([]float64, n)

    // First TR is just High - Low
    tr[0] = candles[0].High - candles[0].Low

    // Calculate True Range
    for i := 1; i < n; i++ {
        hl := candles[i].High - candles[i].Low
        hpc := math.Abs(candles[i].High - candles[i-1].Close)
        lpc := math.Abs(candles[i].Low - candles[i-1].Close)
        tr[i] = math.Max(hl, math.Max(hpc, lpc))
    }

    // Initial ATR (SMA of first period TRs)
    sum := 0.0
    for i := 0; i < period; i++ {
        sum += tr[i]
    }
    result[period-1] = sum / float64(period)

    // Wilder's smoothing
    for i := period; i < n; i++ {
        result[i] = (result[i-1]*float64(period-1) + tr[i]) / float64(period)
    }

    return result
}
```

#### Use Cases
- Position sizing
- Stop-loss placement
- Volatility regime detection

---

### 3.3 Keltner Channels

| Property | Value |
|----------|-------|
| **ID** | `keltner` |
| **Category** | Volatility |
| **Description** | Volatility bands based on EMA and ATR |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `ema_period` | int | 20 | 2-200 | EMA period for middle line |
| `atr_period` | int | 10 | 2-100 | ATR calculation period |
| `multiplier` | float64 | 2.0 | 0.5-5.0 | ATR multiplier |

#### Formula

```
Middle = EMA(Close, ema_period)
Upper = Middle + (multiplier * ATR(atr_period))
Lower = Middle - (multiplier * ATR(atr_period))
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `upper` | float64 | Upper channel |
| `middle` | float64 | Middle line (EMA) |
| `lower` | float64 | Lower channel |

#### Use Cases
- Trend following (price breaking above/below channels)
- Squeeze detection with Bollinger Bands
- Support/resistance levels

---

### 3.4 Donchian Channels

| Property | Value |
|----------|-------|
| **ID** | `donchian` |
| **Category** | Volatility |
| **Description** | Highest high and lowest low over a period |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `period` | int | 20 | 2-200 | Lookback period |

#### Formula

```
Upper = Highest High over period
Lower = Lowest Low over period
Middle = (Upper + Lower) / 2
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `upper` | float64 | Upper channel (highest high) |
| `middle` | float64 | Middle line |
| `lower` | float64 | Lower channel (lowest low) |
| `width` | float64 | Upper - Lower |

#### Use Cases
- Turtle trading strategy
- Breakout detection
- Volatility measurement

---

### 3.5 Standard Deviation

| Property | Value |
|----------|-------|
| **ID** | `stddev` |
| **Category** | Volatility |
| **Description** | Statistical measure of price dispersion |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `period` | int | 20 | 2-200 | Calculation period |
| `source` | string | "close" | open, high, low, close | Price source |

#### Formula

```
StdDev = sqrt(Σ(Price - Mean)² / n)

where Mean = SMA of Price over period
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `value` | float64 | Standard deviation |

#### Use Cases
- Volatility measurement
- Building other indicators (Bollinger Bands)
- Risk assessment

---

## 4. Volume Indicators (Flow / Liquidity / Volume-Weighted)

Volume indicators incorporate trading volume to confirm price movements.

---

### 4.1 OBV (On-Balance Volume)

| Property | Value |
|----------|-------|
| **ID** | `obv` |
| **Category** | Volume |
| **Description** | Cumulative volume based on price direction |

#### Parameters

*No configurable parameters*

#### Formula

```
If Close > Previous Close: OBV = Previous OBV + Volume
If Close < Previous Close: OBV = Previous OBV - Volume
If Close = Previous Close: OBV = Previous OBV
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `value` | float64 | OBV value |

#### Go Implementation

```go
func CalculateOBV(candles []OHLCV) []float64 {
    n := len(candles)
    result := make([]float64, n)

    result[0] = candles[0].Volume

    for i := 1; i < n; i++ {
        if candles[i].Close > candles[i-1].Close {
            result[i] = result[i-1] + candles[i].Volume
        } else if candles[i].Close < candles[i-1].Close {
            result[i] = result[i-1] - candles[i].Volume
        } else {
            result[i] = result[i-1]
        }
    }

    return result
}
```

#### Use Cases
- Confirm trend with volume
- Divergence between OBV and price
- Breakout confirmation

---

### 4.2 VWAP (Volume Weighted Average Price)

| Property | Value |
|----------|-------|
| **ID** | `vwap` |
| **Category** | Volume |
| **Description** | Average price weighted by volume, typically reset daily |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `anchor` | string | "session" | session, week, month | Reset period |

#### Formula

```
VWAP = Σ(Typical Price * Volume) / Σ(Volume)

Typical Price = (High + Low + Close) / 3

Reset at start of each anchor period
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `value` | float64 | VWAP value |
| `upper_1` | float64 | VWAP + 1 StdDev |
| `lower_1` | float64 | VWAP - 1 StdDev |
| `upper_2` | float64 | VWAP + 2 StdDev |
| `lower_2` | float64 | VWAP - 2 StdDev |

#### Go Implementation

```go
func CalculateVWAP(candles []OHLCV, anchor string) *VWAPResult {
    n := len(candles)
    result := &VWAPResult{
        Value:   make([]float64, n),
        Upper1:  make([]float64, n),
        Lower1:  make([]float64, n),
        Upper2:  make([]float64, n),
        Lower2:  make([]float64, n),
    }

    cumVolume := 0.0
    cumTPV := 0.0  // TP * Volume
    cumTPV2 := 0.0 // For StdDev calculation

    for i := 0; i < n; i++ {
        // Check for anchor reset
        if shouldReset(candles, i, anchor) {
            cumVolume = 0
            cumTPV = 0
            cumTPV2 = 0
        }

        tp := (candles[i].High + candles[i].Low + candles[i].Close) / 3
        cumVolume += candles[i].Volume
        cumTPV += tp * candles[i].Volume
        cumTPV2 += tp * tp * candles[i].Volume

        if cumVolume > 0 {
            result.Value[i] = cumTPV / cumVolume

            // Standard deviation bands
            variance := (cumTPV2 / cumVolume) - (result.Value[i] * result.Value[i])
            stdDev := math.Sqrt(math.Max(0, variance))

            result.Upper1[i] = result.Value[i] + stdDev
            result.Lower1[i] = result.Value[i] - stdDev
            result.Upper2[i] = result.Value[i] + 2*stdDev
            result.Lower2[i] = result.Value[i] - 2*stdDev
        }
    }

    return result
}
```

#### Use Cases
- Institutional benchmark price
- Intraday support/resistance
- Trade execution quality measurement

---

### 4.3 MFI (Money Flow Index)

| Property | Value |
|----------|-------|
| **ID** | `mfi` |
| **Category** | Volume |
| **Description** | Volume-weighted RSI, measures buying/selling pressure |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `period` | int | 14 | 2-100 | MFI period |

#### Formula

```
Typical Price = (High + Low + Close) / 3
Raw Money Flow = Typical Price * Volume

If TP > Previous TP: Positive Money Flow += Raw Money Flow
If TP < Previous TP: Negative Money Flow += Raw Money Flow

Money Flow Ratio = Positive MF / Negative MF (over period)
MFI = 100 - (100 / (1 + Money Flow Ratio))
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `value` | float64 | MFI value (0-100) |

#### Use Cases
- Overbought (>80) / Oversold (<20)
- Divergence with price
- Volume confirmation of RSI signals

---

### 4.4 CMF (Chaikin Money Flow)

| Property | Value |
|----------|-------|
| **ID** | `cmf` |
| **Category** | Volume |
| **Description** | Measures accumulation/distribution over a period |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `period` | int | 20 | 2-100 | CMF period |

#### Formula

```
Money Flow Multiplier = ((Close - Low) - (High - Close)) / (High - Low)
Money Flow Volume = MF Multiplier * Volume

CMF = Σ(Money Flow Volume over period) / Σ(Volume over period)
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `value` | float64 | CMF value (-1 to +1) |

#### Use Cases
- Positive CMF = buying pressure
- Negative CMF = selling pressure
- Trend confirmation

---

### 4.5 Volume SMA/EMA

| Property | Value |
|----------|-------|
| **ID** | `volume_ma` |
| **Category** | Volume |
| **Description** | Moving averages of volume for comparison |

#### Parameters

| Parameter | Type | Default | Range | Description |
|-----------|------|---------|-------|-------------|
| `period` | int | 20 | 2-200 | MA period |
| `type` | string | "sma" | sma, ema | Moving average type |

#### Formula

```
Volume SMA = SMA(Volume, period)
Volume EMA = EMA(Volume, period)
```

#### Output Fields

| Field | Type | Description |
|-------|------|-------------|
| `value` | float64 | Volume MA value |
| `ratio` | float64 | Current Volume / MA (volume spike detection) |

#### Use Cases
- Identify volume spikes (ratio > 2)
- Confirm breakouts with above-average volume
- Filter out low-volume noise

---

## Appendix A: Price Source Definitions

| Source | Formula |
|--------|---------|
| `open` | Open price |
| `high` | High price |
| `low` | Low price |
| `close` | Close price |
| `hl2` | (High + Low) / 2 |
| `hlc3` | (High + Low + Close) / 3 |
| `ohlc4` | (Open + High + Low + Close) / 4 |

```go
func extractPrices(candles []OHLCV, source string) []float64 {
    result := make([]float64, len(candles))
    for i, c := range candles {
        switch source {
        case "open":
            result[i] = c.Open
        case "high":
            result[i] = c.High
        case "low":
            result[i] = c.Low
        case "close":
            result[i] = c.Close
        case "hl2":
            result[i] = (c.High + c.Low) / 2
        case "hlc3":
            result[i] = (c.High + c.Low + c.Close) / 3
        case "ohlc4":
            result[i] = (c.Open + c.High + c.Low + c.Close) / 4
        default:
            result[i] = c.Close
        }
    }
    return result
}
```

---

## Appendix B: Lookback Requirements

Each indicator requires a minimum number of historical candles before producing valid output.

| Indicator | Minimum Candles Required |
|-----------|--------------------------|
| SMA(n) | n |
| EMA(n) | n |
| DEMA(n) | 2n - 1 |
| TEMA(n) | 3n - 2 |
| WMA(n) | n |
| HMA(n) | n + sqrt(n) |
| VWMA(n) | n |
| Ichimoku | senkou_b_period + displacement |
| ADX(n) | 2n |
| SuperTrend(n) | n |
| RSI(n) | n + 1 |
| Stochastic(k, d) | k + d |
| MACD(fast, slow, signal) | slow + signal - 1 |
| ROC(n) | n + 1 |
| CCI(n) | n |
| Williams %R(n) | n |
| Momentum(n) | n + 1 |
| Bollinger(n) | n |
| ATR(n) | n |
| Keltner(ema, atr) | max(ema, atr) |
| Donchian(n) | n |
| StdDev(n) | n |
| OBV | 1 |
| VWAP | 1 (per anchor period) |
| MFI(n) | n + 1 |
| CMF(n) | n |
| Volume MA(n) | n |

---

## Appendix C: Indicator Service Interface

```go
// IndicatorService computes and stores indicator values
type IndicatorService interface {
    // ComputeIndicator computes a single indicator for given candles
    ComputeIndicator(ctx context.Context, req *ComputeRequest) (*ComputeResult, error)

    // ComputeAll computes all configured indicators for a connector
    ComputeAll(ctx context.Context, connectorID uuid.UUID, candles []OHLCV) error

    // GetValues retrieves stored indicator values
    GetValues(ctx context.Context, req *GetValuesRequest) ([]IndicatorValue, error)

    // GetLatest retrieves the most recent indicator value
    GetLatest(ctx context.Context, exchange, symbol, timeframe, indicator string, params map[string]any) (*IndicatorValue, error)
}

type ComputeRequest struct {
    Candles   []OHLCV
    Indicator string         // e.g., "rsi", "macd"
    Params    map[string]any // e.g., {"period": 14}
}

type ComputeResult struct {
    Indicator string
    Params    map[string]any
    Values    []map[string]float64 // One map per candle
}

type GetValuesRequest struct {
    Exchange   string
    Symbol     string
    Timeframe  string
    Indicator  string
    Params     map[string]any
    StartTime  int64
    EndTime    int64
    Limit      int
}
```

---

## Appendix D: Frontend Display Configuration

Each indicator type has recommended chart display settings:

```typescript
interface IndicatorDisplayConfig {
  id: string;
  name: string;
  overlay: boolean;        // true = on price chart, false = separate panel
  defaultColor: string;
  fields: FieldConfig[];
  yAxisRange?: [number, number];  // Fixed range for oscillators
}

const INDICATOR_CONFIGS: Record<string, IndicatorDisplayConfig> = {
  sma: {
    id: 'sma',
    name: 'Simple Moving Average',
    overlay: true,
    defaultColor: '#2196F3',
    fields: [{ key: 'value', type: 'line' }],
  },
  rsi: {
    id: 'rsi',
    name: 'Relative Strength Index',
    overlay: false,
    defaultColor: '#9C27B0',
    fields: [{ key: 'value', type: 'line' }],
    yAxisRange: [0, 100],
  },
  bollinger: {
    id: 'bollinger',
    name: 'Bollinger Bands',
    overlay: true,
    defaultColor: '#FF9800',
    fields: [
      { key: 'upper', type: 'line', style: 'dashed' },
      { key: 'middle', type: 'line' },
      { key: 'lower', type: 'line', style: 'dashed' },
    ],
  },
  macd: {
    id: 'macd',
    name: 'MACD',
    overlay: false,
    defaultColor: '#4CAF50',
    fields: [
      { key: 'macd', type: 'line', color: '#2196F3' },
      { key: 'signal', type: 'line', color: '#FF5722' },
      { key: 'histogram', type: 'histogram' },
    ],
  },
  // ... additional configurations
};
```

---

*End of INDICATORS-CATALOG.md*
