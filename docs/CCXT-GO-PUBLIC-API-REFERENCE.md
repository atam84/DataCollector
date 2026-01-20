# CCXT Go - Public API Reference

**Version**: v4
**Package**: `github.com/ccxt/ccxt/go/v4`
**Last Updated**: 2026-01-20

This document focuses on the public API methods relevant for cryptocurrency market data collection.

---

## Table of Contents

1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [Core Concepts](#core-concepts)
4. [Exchange Initialization](#exchange-initialization)
5. [Public API Methods](#public-api-methods)
   - [Market Data](#market-data)
   - [OHLCV Candles](#ohlcv-candles)
   - [Order Books](#order-books)
   - [Trades](#trades)
   - [Tickers](#tickers)
6. [Data Structures](#data-structures)
7. [Error Handling](#error-handling)
8. [Rate Limiting](#rate-limiting)
9. [Supported Exchanges](#supported-exchanges)
10. [Best Practices](#best-practices)

---

## Installation

```bash
go get github.com/ccxt/ccxt/go/v4
```

---

## Quick Start

```go
package main

import (
    "fmt"
    ccxt "github.com/ccxt/ccxt/go/v4"
)

func main() {
    // Create exchange instance
    exchange := ccxt.NewBinance(nil)

    // Load markets (required before most operations)
    markets, err := exchange.LoadMarkets()
    if err != nil {
        panic(err)
    }

    // Fetch OHLCV candles
    ohlcv, err := exchange.FetchOHLCV(
        "BTC/USDT",
        ccxt.WithFetchOHLCVTimeframe("1h"),
        ccxt.WithFetchOHLCVLimit(100),
    )
    if err != nil {
        panic(err)
    }

    fmt.Println("Latest candle:", ohlcv[len(ohlcv)-1])
}
```

---

## Core Concepts

### Unified API
CCXT provides a unified interface across 100+ exchanges. All exchanges implement the `IExchange` interface.

### Markets
- **Symbol Format**: `BASE/QUOTE` (e.g., `BTC/USDT`, `ETH/BTC`)
- **Market Types**: `spot`, `swap`, `future`, `option`
- **Always call `LoadMarkets()` before trading operations**

### Timeframes (OHLCV)
Common timeframes: `1m`, `5m`, `15m`, `30m`, `1h`, `4h`, `1d`, `1w`, `1M`

**Note**: Available timeframes vary by exchange. Check `exchange.GetTimeframes()` for supported intervals.

---

## Exchange Initialization

### Method 1: Direct Constructor
```go
// Binance
exchange := ccxt.NewBinance(nil)

// With configuration
exchange := ccxt.NewBinance(map[string]interface{}{
    "apiKey":    "YOUR_API_KEY",     // Optional for public endpoints
    "secret":    "YOUR_SECRET",      // Optional for public endpoints
    "enableRateLimit": true,         // Enable built-in rate limiting
    "timeout":   30000,              // 30 seconds
})
```

### Method 2: Dynamic Creation
```go
exchangeId := "binance"  // lowercase
exchange := ccxt.CreateExchange(exchangeId, nil)
```

### Configuration Options
```go
options := map[string]interface{}{
    "enableRateLimit": true,          // Enable rate limiting (recommended)
    "timeout":         30000,         // Request timeout (ms)
    "verbose":         false,         // Debug logging
    "sandbox":         false,         // Use testnet/sandbox
}

exchange := ccxt.NewBinance(options)
```

### Sandbox Mode (Testnet)
```go
exchange := ccxt.NewBinance(map[string]interface{}{
    "sandbox": true,
})
// Or
exchange.SetSandboxMode(true)
```

---

## Public API Methods

Public endpoints **do not require API keys** and are accessible without authentication.

### Market Data

#### LoadMarkets
**Load all available trading pairs from the exchange.**

```go
func (e *Exchange) LoadMarkets(params ...interface{}) (map[string]MarketInterface, error)
```

**Example:**
```go
markets, err := exchange.LoadMarkets()
if err != nil {
    log.Fatal(err)
}

// Access market info
for symbol, market := range markets {
    fmt.Printf("Symbol: %s, Type: %s\n", *market.Symbol, *market.Type)
}
```

**Returns**: `map[string]MarketInterface`

---

#### FetchMarkets
**Fetch market information (alternative to LoadMarkets).**

```go
func (e *Exchange) FetchMarkets(params ...interface{}) ([]MarketInterface, error)
```

**Example:**
```go
markets, err := exchange.FetchMarkets()
```

**Returns**: `[]MarketInterface`

---

#### GetMarket
**Get market information for a specific symbol.**

```go
func (e *Exchange) GetMarket(symbol string) MarketInterface
```

**Example:**
```go
market := exchange.GetMarket("BTC/USDT")
fmt.Printf("Spot: %v, Swap: %v\n", *market.Spot, *market.Swap)
```

---

### OHLCV Candles

#### FetchOHLCV
**Fetch historical candlestick data (OHLCV).**

```go
func (e *Exchange) FetchOHLCV(symbol string, options ...FetchOHLCVOptions) ([]OHLCV, error)
```

**Options:**
- `WithFetchOHLCVTimeframe(timeframe string)` - Candle interval (default: "1m")
- `WithFetchOHLCVSince(since int64)` - Start timestamp in milliseconds
- `WithFetchOHLCVLimit(limit int)` - Number of candles to fetch
- `WithFetchOHLCVParams(params map[string]interface{})` - Exchange-specific params

**Example:**
```go
// Fetch 100 hourly candles for BTC/USDT
ohlcv, err := exchange.FetchOHLCV(
    "BTC/USDT",
    ccxt.WithFetchOHLCVTimeframe("1h"),
    ccxt.WithFetchOHLCVLimit(100),
)

if err != nil {
    log.Fatal(err)
}

// Print latest candle
for _, candle := range ohlcv {
    fmt.Printf("Timestamp: %d, Open: %.2f, High: %.2f, Low: %.2f, Close: %.2f, Volume: %.2f\n",
        candle.Timestamp,
        candle.Open,
        candle.High,
        candle.Low,
        candle.Close,
        candle.Volume,
    )
}
```

**Fetch from specific timestamp:**
```go
since := time.Now().Add(-24 * time.Hour).UnixMilli() // 24 hours ago

ohlcv, err := exchange.FetchOHLCV(
    "BTC/USDT",
    ccxt.WithFetchOHLCVTimeframe("5m"),
    ccxt.WithFetchOHLCVSince(since),
    ccxt.WithFetchOHLCVLimit(288), // 5m * 288 = 24 hours
)
```

**Returns**: `[]OHLCV`

---

### Order Books

#### FetchOrderBook
**Fetch the current order book (bids and asks).**

```go
func (e *Exchange) FetchOrderBook(symbol string, options ...FetchOrderBookOptions) (OrderBook, error)
```

**Options:**
- `WithFetchOrderBookLimit(limit int)` - Depth of order book (e.g., 20, 50, 100)

**Example:**
```go
orderBook, err := exchange.FetchOrderBook(
    "BTC/USDT",
    ccxt.WithFetchOrderBookLimit(20),
)

if err != nil {
    log.Fatal(err)
}

fmt.Println("Best Bid:", orderBook.Bids[0])   // [price, amount]
fmt.Println("Best Ask:", orderBook.Asks[0])   // [price, amount]
fmt.Println("Timestamp:", *orderBook.Timestamp)
```

**Returns**: `OrderBook`

---

#### FetchOrderBooks
**Fetch order books for multiple symbols.**

```go
func (e *Exchange) FetchOrderBooks(options ...FetchOrderBooksOptions) (OrderBooks, error)
```

**Example:**
```go
orderBooks, err := exchange.FetchOrderBooks(
    ccxt.WithFetchOrderBooksSymbols([]string{"BTC/USDT", "ETH/USDT"}),
)
```

**Returns**: `OrderBooks` (map of symbol -> OrderBook)

---

### Trades

#### FetchTrades
**Fetch recent public trades.**

```go
func (e *Exchange) FetchTrades(symbol string, options ...FetchTradesOptions) ([]Trade, error)
```

**Options:**
- `WithFetchTradesLimit(limit int)` - Number of trades
- `WithFetchTradesSince(since int64)` - Start timestamp

**Example:**
```go
trades, err := exchange.FetchTrades(
    "BTC/USDT",
    ccxt.WithFetchTradesLimit(50),
)

if err != nil {
    log.Fatal(err)
}

for _, trade := range trades {
    fmt.Printf("ID: %s, Price: %.2f, Amount: %.6f, Side: %s, Time: %d\n",
        *trade.Id,
        *trade.Price,
        *trade.Amount,
        *trade.Side,
        *trade.Timestamp,
    )
}
```

**Returns**: `[]Trade`

---

### Tickers

#### FetchTicker
**Fetch current ticker (24h statistics) for a symbol.**

```go
func (e *Exchange) FetchTicker(symbol string, options ...FetchTickerOptions) (Ticker, error)
```

**Example:**
```go
ticker, err := exchange.FetchTicker("BTC/USDT")

if err != nil {
    log.Fatal(err)
}

fmt.Printf("Symbol: %s\n", *ticker.Symbol)
fmt.Printf("Last: %.2f\n", *ticker.Last)
fmt.Printf("Bid: %.2f, Ask: %.2f\n", *ticker.Bid, *ticker.Ask)
fmt.Printf("24h High: %.2f, Low: %.2f\n", *ticker.High, *ticker.Low)
fmt.Printf("24h Volume: %.2f\n", *ticker.BaseVolume)
fmt.Printf("24h Change: %.2f%%\n", *ticker.Percentage)
```

**Returns**: `Ticker`

---

#### FetchTickers
**Fetch tickers for multiple (or all) symbols.**

```go
func (e *Exchange) FetchTickers(options ...FetchTickersOptions) (Tickers, error)
```

**Options:**
- `WithFetchTickersSymbols(symbols []string)` - Specific symbols (if omitted, fetches all)

**Example:**
```go
// Fetch all tickers
tickers, err := exchange.FetchTickers()

// Fetch specific tickers
tickers, err := exchange.FetchTickers(
    ccxt.WithFetchTickersSymbols([]string{"BTC/USDT", "ETH/USDT"}),
)

if err != nil {
    log.Fatal(err)
}

for symbol, ticker := range tickers.Tickers {
    fmt.Printf("%s: Last=%.2f, Volume=%.2f\n",
        symbol,
        *ticker.Last,
        *ticker.BaseVolume,
    )
}
```

**Returns**: `Tickers` (map of symbol -> Ticker)

---

### Time & Status

#### FetchTime
**Fetch server time from the exchange.**

```go
func (e *Exchange) FetchTime(params ...interface{}) (int64, error)
```

**Example:**
```go
serverTime, err := exchange.FetchTime()
fmt.Println("Server time (ms):", serverTime)
```

**Returns**: `int64` (timestamp in milliseconds)

---

#### FetchStatus
**Check exchange status (operational, maintenance, etc.).**

```go
func (e *Exchange) FetchStatus(params ...interface{}) (map[string]interface{}, error)
```

**Example:**
```go
status, err := exchange.FetchStatus()
fmt.Println("Status:", status["status"])  // "ok", "maintenance", etc.
```

---

## Data Structures

### OHLCV
```go
type OHLCV struct {
    Timestamp int64   // Unix timestamp in milliseconds
    Open      float64 // Opening price
    High      float64 // Highest price
    Low       float64 // Lowest price
    Close     float64 // Closing price
    Volume    float64 // Base currency volume
}
```

**Usage:**
```go
candle := ohlcv[0]
fmt.Printf("Open: %.2f, Close: %.2f\n", candle.Open, candle.Close)
```

---

### Ticker
```go
type Ticker struct {
    Symbol        *string  // Trading pair (e.g., "BTC/USDT")
    Timestamp     *int64   // Unix timestamp (ms)
    Datetime      *string  // ISO 8601 datetime
    High          *float64 // 24h highest price
    Low           *float64 // 24h lowest price
    Bid           *float64 // Best bid price
    BidVolume     *float64 // Bid volume
    Ask           *float64 // Best ask price
    AskVolume     *float64 // Ask volume
    Vwap          *float64 // Volume weighted average price
    Open          *float64 // 24h opening price
    Close         *float64 // 24h closing price
    Last          *float64 // Last traded price
    PreviousClose *float64 // Previous day closing price
    Change        *float64 // Absolute price change
    Percentage    *float64 // Percentage price change
    Average       *float64 // Average price
    BaseVolume    *float64 // Base currency volume (e.g., BTC)
    QuoteVolume   *float64 // Quote currency volume (e.g., USDT)
    Info          map[string]interface{} // Raw exchange data
}
```

**Usage:**
```go
if ticker.Last != nil {
    fmt.Printf("Last price: %.2f\n", *ticker.Last)
}
```

---

### OrderBook
```go
type OrderBook struct {
    Bids      [][]float64 // [[price, amount], ...]
    Asks      [][]float64 // [[price, amount], ...]
    Symbol    *string
    Timestamp *int64
    Datetime  *string
    Nonce     *int64
}
```

**Usage:**
```go
// Best bid (highest buy price)
bestBid := orderBook.Bids[0]
bidPrice := bestBid[0]
bidAmount := bestBid[1]

// Best ask (lowest sell price)
bestAsk := orderBook.Asks[0]
askPrice := bestAsk[0]
askAmount := bestAsk[1]

// Spread
spread := askPrice - bidPrice
```

---

### Trade
```go
type Trade struct {
    Amount       *float64
    Price        *float64
    Cost         *float64  // Amount * Price
    Id           *string   // Trade ID
    Order        *string   // Order ID (if available)
    Timestamp    *int64    // Unix timestamp (ms)
    Datetime     *string   // ISO 8601
    Symbol       *string   // Trading pair
    Type         *string   // "limit", "market"
    Side         *string   // "buy", "sell"
    TakerOrMaker *string   // "taker", "maker"
    Fee          Fee
    Info         map[string]interface{} // Raw data
}
```

---

### MarketInterface
```go
type MarketInterface struct {
    Info           map[string]interface{}
    Id             *string   // Exchange-specific market ID
    Symbol         *string   // Unified symbol (e.g., "BTC/USDT")
    BaseCurrency   *string   // Base currency (e.g., "BTC")
    QuoteCurrency  *string   // Quote currency (e.g., "USDT")
    BaseId         *string   // Exchange-specific base ID
    QuoteId        *string   // Exchange-specific quote ID
    Active         *bool     // Trading enabled
    Type           *string   // "spot", "swap", "future", "option"
    Spot           *bool
    Margin         *bool
    Swap           *bool
    Future         *bool
    Option         *bool
    Contract       *bool
    Settle         *string   // Settlement currency
    ContractSize   *float64  // Contract value
    Linear         *bool
    Inverse        *bool
    Expiry         *int64    // Expiration timestamp (futures)
    Strike         *float64  // Strike price (options)
    Taker          *float64  // Taker fee
    Maker          *float64  // Maker fee
    Limits         Limits    // Min/max limits
    Created        *int64
}
```

**Usage:**
```go
market := exchange.GetMarket("BTC/USDT")

if *market.Spot {
    fmt.Println("Spot market")
}

if *market.Active {
    fmt.Println("Market is active")
}
```

---

## Error Handling

### Common Errors

```go
ohlcv, err := exchange.FetchOHLCV("BTC/USDT")
if err != nil {
    // Check error type
    switch {
    case strings.Contains(err.Error(), "rate limit"):
        fmt.Println("Rate limit exceeded, wait and retry")
    case strings.Contains(err.Error(), "network"):
        fmt.Println("Network error, retry")
    case strings.Contains(err.Error(), "timeout"):
        fmt.Println("Request timeout")
    case strings.Contains(err.Error(), "Invalid symbol"):
        fmt.Println("Symbol not found")
    default:
        fmt.Printf("Unknown error: %v\n", err)
    }
}
```

### Best Practice Pattern

```go
func fetchWithRetry(exchange ccxt.IExchange, symbol string, maxRetries int) ([]ccxt.OHLCV, error) {
    var lastErr error

    for i := 0; i < maxRetries; i++ {
        ohlcv, err := exchange.FetchOHLCV(
            symbol,
            ccxt.WithFetchOHLCVTimeframe("1h"),
        )

        if err == nil {
            return ohlcv, nil
        }

        lastErr = err

        // Exponential backoff
        time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Second)
    }

    return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
```

---

## Rate Limiting

### Built-in Rate Limiter

```go
exchange := ccxt.NewBinance(map[string]interface{}{
    "enableRateLimit": true,  // Enables automatic rate limiting
})
```

**How it works:**
- CCXT tracks API calls and automatically throttles requests
- Prevents hitting exchange rate limits
- **Highly recommended for production**

### Manual Throttling

If you manage rate limiting externally:

```go
exchange := ccxt.NewBinance(map[string]interface{}{
    "enableRateLimit": false,
})

// Implement your own rate limiting
time.Sleep(100 * time.Millisecond)
```

### Rate Limit Info

```go
// Check rate limit weight for an endpoint
rateLimitWeight := exchange.GetApi()["fetchOHLCV"].(map[string]interface{})["weight"]
```

---

## Supported Exchanges

CCXT Go v4 supports 100+ exchanges. Common ones:

| Exchange | ID | Spot | Futures | OHLCV | Notes |
|----------|-------|------|---------|-------|-------|
| Binance | `binance` | ✅ | ✅ | ✅ | Most liquid |
| Binance US | `binanceus` | ✅ | ❌ | ✅ | US version |
| Binance USDⓈ-M | `binanceusdm` | ❌ | ✅ | ✅ | USDT futures |
| Binance COIN-M | `binancecoinm` | ❌ | ✅ | ✅ | Coin futures |
| Coinbase | `coinbase` | ✅ | ❌ | ✅ | US-based |
| Bybit | `bybit` | ✅ | ✅ | ✅ | Derivatives |
| OKX | `okx` | ✅ | ✅ | ✅ | Global |
| Kraken | `kraken` | ✅ | ✅ | ✅ | Europe |
| KuCoin | `kucoin` | ✅ | ✅ | ✅ | Altcoins |
| Bitfinex | `bitfinex` | ✅ | ❌ | ✅ | Margin trading |
| Bitget | `bitget` | ✅ | ✅ | ✅ | Copy trading |
| Gate.io | `gate` | ✅ | ✅ | ✅ | Altcoins |
| HTX (Huobi) | `htx` | ✅ | ✅ | ✅ | Asian market |
| MEXC | `mexc` | ✅ | ✅ | ✅ | Altcoins |

**Full list**: See `exchange_typed_interface.go:217-559` or use:

```go
exchanges := []string{
    "binance", "coinbase", "kraken", "bybit", "okx",
    "kucoin", "bitfinex", "gate", "htx", "mexc",
    // ... 100+ more
}
```

**Creating dynamically:**

```go
for _, exchangeId := range exchanges {
    exchange := ccxt.CreateExchange(exchangeId, nil)
    fmt.Println("Created:", exchange.GetId())
}
```

---

## Best Practices

### 1. Always Load Markets First
```go
exchange := ccxt.NewBinance(nil)
_, err := exchange.LoadMarkets()  // IMPORTANT
if err != nil {
    log.Fatal(err)
}
```

### 2. Enable Rate Limiting
```go
exchange := ccxt.NewBinance(map[string]interface{}{
    "enableRateLimit": true,
})
```

### 3. Handle Nil Pointers
Many fields in CCXT structs are pointers. Always check for nil:

```go
if ticker.Last != nil {
    fmt.Printf("Last: %.2f\n", *ticker.Last)
}
```

### 4. Use Pagination for Historical Data
When fetching large amounts of historical OHLCV:

```go
func fetchAllOHLCV(exchange ccxt.IExchange, symbol, timeframe string, since int64) ([]ccxt.OHLCV, error) {
    var allCandles []ccxt.OHLCV
    currentSince := since

    for {
        candles, err := exchange.FetchOHLCV(
            symbol,
            ccxt.WithFetchOHLCVTimeframe(timeframe),
            ccxt.WithFetchOHLCVSince(currentSince),
            ccxt.WithFetchOHLCVLimit(1000),
        )

        if err != nil {
            return nil, err
        }

        if len(candles) == 0 {
            break
        }

        allCandles = append(allCandles, candles...)

        // Update since to last candle timestamp + 1ms
        currentSince = candles[len(candles)-1].Timestamp + 1

        // Prevent infinite loops
        if len(candles) < 1000 {
            break
        }
    }

    return allCandles, nil
}
```

### 5. Implement Retry Logic
```go
func fetchWithExponentialBackoff(fn func() error, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        err := fn()
        if err == nil {
            return nil
        }

        if i == maxRetries-1 {
            return err
        }

        backoff := time.Duration(math.Pow(2, float64(i))) * time.Second
        time.Sleep(backoff)
    }
    return nil
}
```

### 6. Validate Symbols
```go
markets, _ := exchange.LoadMarkets()

symbol := "BTC/USDT"
if _, exists := markets[symbol]; !exists {
    log.Fatalf("Symbol %s not found", symbol)
}
```

### 7. Use Timeframe Constants
```go
const (
    Timeframe1m  = "1m"
    Timeframe5m  = "5m"
    Timeframe15m = "15m"
    Timeframe1h  = "1h"
    Timeframe4h  = "4h"
    Timeframe1d  = "1d"
)
```

### 8. Graceful Shutdown
```go
defer exchange.Close()  // Close connections
```

### 9. Log API Calls (Debug)
```go
exchange := ccxt.NewBinance(map[string]interface{}{
    "verbose": true,  // Enable debug logging
})
```

### 10. Store Raw Response
```go
ticker, err := exchange.FetchTicker("BTC/USDT")
rawData := ticker.Info  // Original exchange response
```

---

## Complete Example: Data Collector

```go
package main

import (
    "fmt"
    "log"
    "time"

    ccxt "github.com/ccxt/ccxt/go/v4"
)

type Candle struct {
    ExchangeID string
    Symbol     string
    Timeframe  string
    Timestamp  int64
    Open       float64
    High       float64
    Low        float64
    Close      float64
    Volume     float64
}

func main() {
    // Initialize exchange
    exchange := ccxt.NewBinance(map[string]interface{}{
        "enableRateLimit": true,
        "timeout":         30000,
    })

    // Load markets
    _, err := exchange.LoadMarkets()
    if err != nil {
        log.Fatal("Failed to load markets:", err)
    }

    // Collect data
    symbol := "BTC/USDT"
    timeframe := "1h"

    ohlcvData, err := fetchOHLCV(exchange, symbol, timeframe, 100)
    if err != nil {
        log.Fatal("Failed to fetch OHLCV:", err)
    }

    // Process and store
    for _, candle := range ohlcvData {
        fmt.Printf("Candle: Time=%d, O=%.2f, H=%.2f, L=%.2f, C=%.2f, V=%.2f\n",
            candle.Timestamp, candle.Open, candle.High, candle.Low, candle.Close, candle.Volume)

        // Store in database here
    }
}

func fetchOHLCV(exchange ccxt.IExchange, symbol, timeframe string, limit int) ([]Candle, error) {
    ohlcv, err := exchange.FetchOHLCV(
        symbol,
        ccxt.WithFetchOHLCVTimeframe(timeframe),
        ccxt.WithFetchOHLCVLimit(limit),
    )

    if err != nil {
        return nil, err
    }

    candles := make([]Candle, len(ohlcv))
    for i, bar := range ohlcv {
        candles[i] = Candle{
            ExchangeID: exchange.GetId(),
            Symbol:     symbol,
            Timeframe:  timeframe,
            Timestamp:  bar.Timestamp,
            Open:       bar.Open,
            High:       bar.High,
            Low:        bar.Low,
            Close:      bar.Close,
            Volume:     bar.Volume,
        }
    }

    return candles, nil
}
```

---

## Additional Resources

- **Official CCXT Manual**: https://docs.ccxt.com/
- **CCXT GitHub**: https://github.com/ccxt/ccxt
- **Go Package Docs**: https://pkg.go.dev/github.com/ccxt/ccxt/go/v4
- **Exchange-Specific Docs**: Check each exchange's API documentation for quirks

---

**Note**: This reference focuses on public REST API methods. CCXT also supports:
- **WebSocket (Pro)**: Real-time streaming via `github.com/ccxt/ccxt/go/v4/pro`
- **Private API**: Trading, orders, balances (requires API keys)
- **Advanced Features**: Margin trading, futures, options

For your data collection use case, the **public REST API methods** documented above are sufficient.
