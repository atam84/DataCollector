# CCXT Go - Quick Reference

## Full Documentation

For complete CCXT Go public API reference, see:
📖 **[CCXT-GO-PUBLIC-API-REFERENCE.md](./CCXT-GO-PUBLIC-API-REFERENCE.md)**

This comprehensive guide covers:
- Installation & Setup
- All Public API Methods (OHLCV, Tickers, Order Books, Trades)
- Complete Data Structures
- Error Handling & Retry Logic
- Rate Limiting
- Best Practices
- Full Examples

---

## Official Resources

**CCXT Manual** (concepts apply to Go)
https://docs.ccxt.com/
Covers the unified API concepts (markets, tickers, orders, params, errors, rate limits, etc.).

**Go API reference on pkg.go.dev** — Go-specific usage
https://pkg.go.dev/github.com/ccxt/ccxt/go/v4
Shows exported types/functions and examples. Actively published.

**GitHub repo + examples**
https://github.com/ccxt/ccxt
Main CCXT repo with examples in multiple languages.

---

## Important Notes

**Go Support:**
- Official Go package with **REST capabilities**
- WebSockets available via `github.com/ccxt/ccxt/go/v4/pro` (CCXT Pro)
- For your data collection needs, REST API is sufficient

**Installation:**
```bash
go get github.com/ccxt/ccxt/go/v4
```

---

## Quick Start Pattern

```go
package main

import (
  "fmt"
  "log"
  "github.com/ccxt/ccxt/go/v4"
)

func main() {
  // Create exchange
  exchange := ccxt.NewBinance(map[string]interface{}{
    "enableRateLimit": true,  // Enable built-in rate limiting
  })

  // Load markets (important!)
  _, err := exchange.LoadMarkets()
  if err != nil {
    log.Fatal(err)
  }

  // Fetch OHLCV candles
  ohlcv, err := exchange.FetchOHLCV(
    "BTC/USDT",
    ccxt.WithFetchOHLCVTimeframe("1h"),
    ccxt.WithFetchOHLCVLimit(100),
  )

  if err != nil {
    log.Fatal(err)
  }

  // Process candles
  for _, candle := range ohlcv {
    fmt.Printf("Time: %d, Open: %.2f, Close: %.2f\n",
      candle.Timestamp, candle.Open, candle.Close)
  }
}
```

**Core Flow:**
1. `LoadMarkets()` - Always call first
2. `FetchOHLCV()` / `FetchTicker()` / etc. - Fetch data
3. Process and store

---

## Key Methods for Data Collection

| Method | Purpose | Example |
|--------|---------|---------|
| `LoadMarkets()` | Load all trading pairs | `exchange.LoadMarkets()` |
| `FetchOHLCV()` | Get candle data | `exchange.FetchOHLCV("BTC/USDT", WithFetchOHLCVTimeframe("1h"))` |
| `FetchTicker()` | Get 24h stats | `exchange.FetchTicker("BTC/USDT")` |
| `FetchTrades()` | Get recent trades | `exchange.FetchTrades("BTC/USDT")` |
| `FetchOrderBook()` | Get order book | `exchange.FetchOrderBook("BTC/USDT")` |

For detailed information on all methods, parameters, and options, see **[CCXT-GO-PUBLIC-API-REFERENCE.md](./CCXT-GO-PUBLIC-API-REFERENCE.md)**
