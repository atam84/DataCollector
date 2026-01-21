datacollector-api  | 2026/01/21 11:41:48 Connected to MongoDB successfully
datacollector-api  | 2026/01/21 11:41:48 Job scheduler started - checking for jobs every 30 seconds
datacollector-api  | 2026/01/21 11:41:48 Starting server on 0.0.0.0:8080 (Sandbox Mode: true)
datacollector-api  | 
datacollector-api  |  ┌───────────────────────────────────────────────────┐ 
datacollector-api  |  │                 DataCollector API                 │ 
datacollector-api  |  │                  Fiber v2.52.10                   │ 
datacollector-api  |  │               http://127.0.0.1:8080               │ 
datacollector-api  |  │       (bound on host 0.0.0.0 and port 8080)       │ 
datacollector-api  |  │                                                   │ 
datacollector-api  |  │ Handlers ............ 58  Processes ........... 1 │ 
datacollector-api  |  │ Prefork ....... Disabled  PID ................. 1 │ 
datacollector-api  |  └───────────────────────────────────────────────────┘ 
datacollector-api  | 
datacollector-api  | 2026/01/21 11:41:48 Found 1 job(s) ready to execute
datacollector-api  | 2026/01/21 11:41:48 Executing job 6970b161641cb4a2a65408f1
datacollector-api  | 2026/01/21 11:41:48 [EXEC] About to call FetchOHLCVData for 6970b161641cb4a2a65408f1
datacollector-api  | 2026/01/21 11:41:48 [FETCH_START] Starting FetchOHLCVData for ATOM/USDT/1d
datacollector-api  | 2026/01/21 11:41:48 [FETCH] First execution for ATOM/USDT/1d - fetching ALL available data (no since, no limit)
datacollector-api  | 2026/01/21 11:41:48 [CCXT] Fetching binance ATOM/USDT 1d - FIRST EXECUTION (ALL available data, no limit)
datacollector-api  | 2026/01/21 11:41:48 [CCXT] Calling Binance.FetchOHLCV(ATOM/USDT, timeframe=1d) - NO since, NO limit
datacollector-api  | 2026/01/21 11:41:50 [CCXT] Received 500 candles from exchange
datacollector-api  | 2026/01/21 11:41:50 [CCXT] Converted and reversed 500 candles (newest first)
datacollector-api  | 2026/01/21 11:41:50 [CCXT] Successfully converted 500 candles
datacollector-api  | 2026/01/21 11:41:50 [FETCH] First execution complete - fetched 500 historical candles
datacollector-api  | 2026/01/21 11:41:50 [EXEC] FetchOHLCVData returned 500 candles, err=<nil>
datacollector-api  | 2026/01/21 11:41:50 [EXEC] Calculating indicators for 500 candles
datacollector-api  | 2026/01/21 11:41:50 [INDICATORS] Calculating indicators for 500 candles
datacollector-api  | 2026/01/21 11:41:50 [INDICATORS] Calculated SMA(20)
datacollector-api  | 2026/01/21 11:41:50 [INDICATORS] Calculated EMA(20)
datacollector-api  | panic: runtime error: index out of range [-1]
datacollector-api  | 
datacollector-api  | goroutine 16 [running]:
datacollector-api  | github.com/yourusername/datacollector/internal/service/indicators.CalculateEMA({0xc00886e000, 0x1f4, 0x1f4}, 0x0, {0x42eadd6, 0x5})
datacollector-api  |    /app/internal/service/indicators/ema.go:25 +0x171
datacollector-api  | github.com/yourusername/datacollector/internal/service/indicators.CalculateDEMA({0xc00886e000, 0x1f4, 0x1f4}, 0xc0001c36e0?, {0x42eadd6, 0x5})
datacollector-api  |    /app/internal/service/indicators/dema.go:17 +0xd7
datacollector-api  | github.com/yourusername/datacollector/internal/service/indicators.(*Service).calculateTrendIndicators(0xc000120300?, {0xc00886e000, 0x1f4, 0x1f4}, 0xc0006001c0)
datacollector-api  |    /app/internal/service/indicators/service.go:106 +0xc6
datacollector-api  | github.com/yourusername/datacollector/internal/service/indicators.(*Service).CalculateAll(0x79c0720, {0xc008836000, 0x1f4, 0x1f4}, 0x7?)
datacollector-api  |    /app/internal/service/indicators/service.go:46 +0x305
datacollector-api  | github.com/yourusername/datacollector/internal/service.(*JobExecutor).ExecuteJob(0xc000278a80, {0x4839200, 0xc00009b260}, {0xc000038198, 0x18})
datacollector-api  |    /app/internal/service/job_executor.go:138 +0x8ce
datacollector-api  | github.com/yourusername/datacollector/internal/service.(*JobScheduler).checkAndExecuteJobs.func1({0xc000038198, 0x18})
datacollector-api  |    /app/internal/service/job_scheduler.go:85 +0x165
datacollector-api  | created by github.com/yourusername/datacollector/internal/service.(*JobScheduler).checkAndExecuteJobs in goroutine 41
datacollector-api  |    /app/internal/service/job_scheduler.go:79 +0x22d
datacollector-api exited with code 2                                                                                     
