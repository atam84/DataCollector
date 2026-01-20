# ARCH-001: Data Collector Backend Architecture

## Document Information

| Field | Value |
|-------|-------|
| **Document ID** | ARCH-001 |
| **Version** | 1.0 |
| **Status** | Draft |
| **Author** | GoTrading Team |
| **Created** | 2025-01-17 |
| **Last Updated** | 2025-01-17 |

---

## 1. Overview

### 1.1 Purpose

This document describes the backend architecture for the GoTrading Data Collector service. The backend is responsible for collecting, storing, and enriching cryptocurrency market data from multiple exchanges using the CCXT library.

### 1.2 Technology Stack

| Component | Technology | Justification |
|-----------|------------|---------------|
| Language | Go 1.22+ | Performance, concurrency, type safety |
| Web Framework | Fiber v2 | Fast, Express-like, WebSocket support |
| Exchange Library | CCXT-Go | Unified exchange API access |
| Task Scheduler | go-cron | Cron-like job scheduling |
| WebSocket | gorilla/websocket | Robust WebSocket implementation |
| Database Driver | mongo-driver | Official MongoDB driver |
| Database Driver | pgx | High-performance PostgreSQL driver |
| Logging | zerolog | Structured, zero-allocation logging |
| Config | viper | Configuration management |
| Validation | validator | Struct validation |

---

## 2. High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           DATA COLLECTOR BACKEND                                 │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                            API LAYER                                     │   │
│  ├─────────────────────────────────────────────────────────────────────────┤   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │   │
│  │  │   REST API  │  │  WebSocket  │  │   GraphQL   │  │   Metrics   │    │   │
│  │  │   (Fiber)   │  │   Server    │  │  (Optional) │  │ (Prometheus)│    │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                        │                                        │
│                                        ▼                                        │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                          SERVICE LAYER                                   │   │
│  ├─────────────────────────────────────────────────────────────────────────┤   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │   │
│  │  │  Connector  │  │    Data     │  │  Indicator  │  │   Export    │    │   │
│  │  │   Service   │  │   Service   │  │   Service   │  │   Service   │    │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │   │
│  │  │  Exchange   │  │  Scheduler  │  │   Health    │  │    Alert    │    │   │
│  │  │   Service   │  │   Service   │  │   Service   │  │   Service   │    │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                        │                                        │
│                                        ▼                                        │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                           CORE LAYER                                     │   │
│  ├─────────────────────────────────────────────────────────────────────────┤   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │   │
│  │  │    CCXT     │  │ Rate Limit  │  │    Job      │  │   Event     │    │   │
│  │  │   Engine    │  │   Manager   │  │   Queue     │  │    Bus      │    │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │   │
│  │  │  Indicator  │  │    Gap      │  │   Cache     │  │   Circuit   │    │   │
│  │  │  Calculator │  │  Detector   │  │   Layer     │  │   Breaker   │    │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                        │                                        │
│                                        ▼                                        │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                        REPOSITORY LAYER                                  │   │
│  ├─────────────────────────────────────────────────────────────────────────┤   │
│  │  ┌─────────────────────────┐      ┌─────────────────────────┐          │   │
│  │  │     MongoDB Repos       │      │    PostgreSQL Repos     │          │   │
│  │  ├─────────────────────────┤      ├─────────────────────────┤          │   │
│  │  │ • OHLCVRepository       │      │ • ConnectorRepository   │          │   │
│  │  │ • IndicatorRepository   │      │ • ScheduleRepository    │          │   │
│  │  │ • MetadataRepository    │      │ • ExchangeConfigRepo    │          │   │
│  │  │ • GapRepository         │      │ • AlertRepository       │          │   │
│  │  └─────────────────────────┘      └─────────────────────────┘          │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                        │                                        │
│                                        ▼                                        │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                        INFRASTRUCTURE                                    │   │
│  ├─────────────────────────────────────────────────────────────────────────┤   │
│  │     ┌───────────────┐              ┌───────────────┐                    │   │
│  │     │    MongoDB    │              │  PostgreSQL   │                    │   │
│  │     │  (Time-Series)│              │   (Config)    │                    │   │
│  │     └───────────────┘              └───────────────┘                    │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Project Structure

```
data-collector-backend/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── connector.go        # Connector endpoints
│   │   │   ├── exchange.go         # Exchange endpoints
│   │   │   ├── data.go             # Data query endpoints
│   │   │   ├── indicator.go        # Indicator endpoints
│   │   │   ├── export.go           # Export endpoints
│   │   │   └── system.go           # System/health endpoints
│   │   ├── middleware/
│   │   │   ├── auth.go             # Authentication (future)
│   │   │   ├── ratelimit.go        # API rate limiting
│   │   │   ├── logging.go          # Request logging
│   │   │   └── recovery.go         # Panic recovery
│   │   ├── websocket/
│   │   │   ├── hub.go              # WebSocket hub
│   │   │   ├── client.go           # WebSocket client
│   │   │   └── handlers.go         # WebSocket message handlers
│   │   ├── router.go               # Route definitions
│   │   └── server.go               # HTTP server setup
│   ├── services/
│   │   ├── connector/
│   │   │   ├── service.go          # Connector business logic
│   │   │   └── validator.go        # Connector validation
│   │   ├── exchange/
│   │   │   ├── service.go          # Exchange operations
│   │   │   └── cache.go            # Exchange info caching
│   │   ├── collector/
│   │   │   ├── service.go          # Data collection orchestration
│   │   │   ├── fetcher.go          # OHLCV fetching logic
│   │   │   ├── backfill.go         # Historical backfill
│   │   │   └── stream.go           # WebSocket streaming
│   │   ├── indicator/
│   │   │   ├── service.go          # Indicator orchestration
│   │   │   ├── calculator.go       # Indicator calculations
│   │   │   └── definitions.go      # Indicator definitions
│   │   ├── export/
│   │   │   ├── service.go          # Export orchestration
│   │   │   ├── csv.go              # CSV export
│   │   │   ├── json.go             # JSON export
│   │   │   └── parquet.go          # Parquet export
│   │   ├── scheduler/
│   │   │   ├── service.go          # Job scheduling
│   │   │   └── jobs.go             # Job definitions
│   │   ├── health/
│   │   │   └── service.go          # Health monitoring
│   │   └── alert/
│   │       └── service.go          # Alert management
│   ├── core/
│   │   ├── ccxt/
│   │   │   ├── engine.go           # CCXT wrapper
│   │   │   ├── exchange.go         # Exchange abstraction
│   │   │   └── types.go            # CCXT types
│   │   ├── ratelimit/
│   │   │   ├── manager.go          # Rate limit management
│   │   │   ├── bucket.go           # Token bucket implementation
│   │   │   └── config.go           # Exchange rate configs
│   │   ├── queue/
│   │   │   ├── job_queue.go        # Priority job queue
│   │   │   └── worker.go           # Worker pool
│   │   ├── events/
│   │   │   ├── bus.go              # Event bus
│   │   │   └── events.go           # Event definitions
│   │   ├── circuit/
│   │   │   └── breaker.go          # Circuit breaker
│   │   ├── gap/
│   │   │   └── detector.go         # Gap detection
│   │   └── cache/
│   │       └── cache.go            # In-memory cache
│   ├── repository/
│   │   ├── mongo/
│   │   │   ├── ohlcv.go            # OHLCV repository
│   │   │   ├── indicator.go        # Indicator repository
│   │   │   ├── metadata.go         # Metadata repository
│   │   │   └── gap.go              # Gap repository
│   │   ├── postgres/
│   │   │   ├── connector.go        # Connector repository
│   │   │   ├── schedule.go         # Schedule repository
│   │   │   ├── exchange_config.go  # Exchange config repository
│   │   │   └── alert.go            # Alert repository
│   │   └── interfaces.go           # Repository interfaces
│   ├── models/
│   │   ├── connector.go            # Connector model
│   │   ├── ohlcv.go                # OHLCV model
│   │   ├── indicator.go            # Indicator model
│   │   ├── exchange.go             # Exchange model
│   │   ├── schedule.go             # Schedule model
│   │   ├── job.go                  # Job model
│   │   ├── gap.go                  # Gap model
│   │   └── alert.go                # Alert model
│   └── config/
│       └── config.go               # Application configuration
├── pkg/
│   ├── logger/
│   │   └── logger.go               # Logging utilities
│   ├── validator/
│   │   └── validator.go            # Validation utilities
│   └── utils/
│       ├── time.go                 # Time utilities
│       └── convert.go              # Conversion utilities
├── migrations/
│   ├── postgres/
│   │   ├── 001_create_connectors.sql
│   │   ├── 002_create_schedules.sql
│   │   └── ...
│   └── mongo/
│       └── indexes.js              # MongoDB index definitions
├── scripts/
│   ├── setup.sh                    # Development setup
│   └── migrate.sh                  # Database migration
├── deployments/
│   ├── docker/
│   │   └── Dockerfile
│   └── kubernetes/
│       └── deployment.yaml
├── configs/
│   ├── config.yaml                 # Default configuration
│   ├── config.development.yaml     # Development overrides
│   └── config.production.yaml      # Production overrides
├── go.mod
├── go.sum
└── README.md
```

---

## 4. Component Details

### 4.1 API Layer

#### 4.1.1 REST API (Fiber)

```go
// internal/api/router.go

package api

import (
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "github.com/gofiber/fiber/v2/middleware/recover"
)

func SetupRouter(app *fiber.App, h *Handlers) {
    // Middleware
    app.Use(recover.New())
    app.Use(logger.New())
    app.Use(cors.New())

    // API v1
    v1 := app.Group("/api/v1")

    // Connectors
    connectors := v1.Group("/connectors")
    connectors.Get("/", h.Connector.List)
    connectors.Post("/", h.Connector.Create)
    connectors.Get("/:id", h.Connector.Get)
    connectors.Put("/:id", h.Connector.Update)
    connectors.Delete("/:id", h.Connector.Delete)
    connectors.Post("/:id/start", h.Connector.Start)
    connectors.Post("/:id/stop", h.Connector.Stop)
    connectors.Post("/:id/refresh", h.Connector.Refresh)

    // Exchanges
    exchanges := v1.Group("/exchanges")
    exchanges.Get("/", h.Exchange.List)
    exchanges.Get("/:id", h.Exchange.Get)
    exchanges.Get("/:id/pairs", h.Exchange.ListPairs)
    exchanges.Get("/:id/timeframes", h.Exchange.ListTimeframes)
    exchanges.Get("/:id/status", h.Exchange.Status)

    // Data
    data := v1.Group("/data")
    data.Get("/ohlcv", h.Data.QueryOHLCV)
    data.Get("/indicators", h.Data.QueryIndicators)
    data.Get("/coverage", h.Data.Coverage)
    data.Get("/gaps", h.Data.ListGaps)
    data.Post("/export", h.Export.Create)
    data.Get("/export/:id", h.Export.Status)
    data.Get("/export/:id/download", h.Export.Download)

    // Indicators
    indicators := v1.Group("/indicators")
    indicators.Get("/", h.Indicator.List)
    indicators.Post("/compute", h.Indicator.Compute)
    indicators.Get("/config", h.Indicator.GetConfig)
    indicators.Put("/config", h.Indicator.UpdateConfig)

    // System
    system := v1.Group("/system")
    system.Get("/status", h.System.Status)
    system.Get("/stats", h.System.Stats)
    system.Get("/jobs", h.System.Jobs)
}
```

#### 4.1.2 WebSocket Server

```go
// internal/api/websocket/hub.go

package websocket

import (
    "sync"
    "github.com/gorilla/websocket"
)

type Hub struct {
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
    rooms      map[string]map[*Client]bool
    mu         sync.RWMutex
}

type Client struct {
    hub           *Hub
    conn          *websocket.Conn
    send          chan []byte
    subscriptions map[string]bool
}

type Message struct {
    Action    string `json:"action"`
    Channel   string `json:"channel"`
    Exchange  string `json:"exchange,omitempty"`
    Symbol    string `json:"symbol,omitempty"`
    Timeframe string `json:"timeframe,omitempty"`
    ConnectorID string `json:"connector_id,omitempty"`
}

func NewHub() *Hub {
    return &Hub{
        clients:    make(map[*Client]bool),
        broadcast:  make(chan []byte),
        register:   make(chan *Client),
        unregister: make(chan *Client),
        rooms:      make(map[string]map[*Client]bool),
    }
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.clients[client] = true
        case client := <-h.unregister:
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
                h.removeFromAllRooms(client)
            }
        case message := <-h.broadcast:
            for client := range h.clients {
                select {
                case client.send <- message:
                default:
                    close(client.send)
                    delete(h.clients, client)
                }
            }
        }
    }
}

func (h *Hub) BroadcastToRoom(room string, message []byte) {
    h.mu.RLock()
    defer h.mu.RUnlock()
    
    if clients, ok := h.rooms[room]; ok {
        for client := range clients {
            select {
            case client.send <- message:
            default:
                // Client buffer full, skip
            }
        }
    }
}
```

### 4.2 Service Layer

#### 4.2.1 Connector Service

```go
// internal/services/connector/service.go

package connector

import (
    "context"
    "time"
    
    "github.com/google/uuid"
)

type Service struct {
    repo          ConnectorRepository
    exchangeSvc   *exchange.Service
    schedulerSvc  *scheduler.Service
    collectorSvc  *collector.Service
    eventBus      *events.Bus
}

type CreateConnectorInput struct {
    Name       string   `json:"name" validate:"required,min=1,max=100"`
    Exchange   string   `json:"exchange" validate:"required"`
    Pairs      []string `json:"pairs" validate:"required,min=1"`
    Timeframes []string `json:"timeframes" validate:"required,min=1"`
    Schedule   string   `json:"schedule" validate:"omitempty,cron"`
}

func (s *Service) Create(ctx context.Context, input CreateConnectorInput) (*Connector, error) {
    // Validate exchange exists
    if !s.exchangeSvc.Exists(input.Exchange) {
        return nil, ErrExchangeNotFound
    }
    
    // Validate pairs exist on exchange
    for _, pair := range input.Pairs {
        if !s.exchangeSvc.PairExists(input.Exchange, pair) {
            return nil, ErrPairNotFound
        }
    }
    
    // Validate timeframes
    for _, tf := range input.Timeframes {
        if !s.exchangeSvc.TimeframeExists(input.Exchange, tf) {
            return nil, ErrTimeframeNotFound
        }
    }
    
    connector := &Connector{
        ID:         uuid.New(),
        Name:       input.Name,
        Exchange:   input.Exchange,
        Pairs:      input.Pairs,
        Timeframes: input.Timeframes,
        Schedule:   input.Schedule,
        Status:     StatusCreated,
        CreatedAt:  time.Now(),
        UpdatedAt:  time.Now(),
    }
    
    if err := s.repo.Create(ctx, connector); err != nil {
        return nil, err
    }
    
    // Emit event
    s.eventBus.Publish(events.ConnectorCreated{
        ConnectorID: connector.ID,
        Exchange:    connector.Exchange,
    })
    
    return connector, nil
}

func (s *Service) Start(ctx context.Context, id uuid.UUID) error {
    connector, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return err
    }
    
    if connector.Status == StatusActive {
        return ErrAlreadyActive
    }
    
    // Start collection jobs
    if err := s.collectorSvc.StartConnector(ctx, connector); err != nil {
        return err
    }
    
    // Schedule future runs
    if connector.Schedule != "" {
        if err := s.schedulerSvc.Schedule(connector); err != nil {
            return err
        }
    }
    
    connector.Status = StatusActive
    connector.UpdatedAt = time.Now()
    
    if err := s.repo.Update(ctx, connector); err != nil {
        return err
    }
    
    s.eventBus.Publish(events.ConnectorStarted{ConnectorID: id})
    
    return nil
}

func (s *Service) Stop(ctx context.Context, id uuid.UUID) error {
    connector, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return err
    }
    
    // Stop collection jobs
    s.collectorSvc.StopConnector(ctx, connector)
    
    // Remove schedule
    s.schedulerSvc.Unschedule(connector.ID)
    
    connector.Status = StatusPaused
    connector.UpdatedAt = time.Now()
    
    if err := s.repo.Update(ctx, connector); err != nil {
        return err
    }
    
    s.eventBus.Publish(events.ConnectorStopped{ConnectorID: id})
    
    return nil
}
```

#### 4.2.2 Collector Service

```go
// internal/services/collector/service.go

package collector

import (
    "context"
    "sync"
    "time"
)

type Service struct {
    ccxt          *ccxt.Engine
    rateLimiter   *ratelimit.Manager
    ohlcvRepo     OHLCVRepository
    metadataRepo  MetadataRepository
    gapDetector   *gap.Detector
    indicatorSvc  *indicator.Service
    eventBus      *events.Bus
    jobQueue      *queue.JobQueue
    activeJobs    map[uuid.UUID]context.CancelFunc
    mu            sync.RWMutex
}

func (s *Service) StartConnector(ctx context.Context, connector *Connector) error {
    for _, pair := range connector.Pairs {
        for _, timeframe := range connector.Timeframes {
            job := &CollectionJob{
                ID:          uuid.New(),
                ConnectorID: connector.ID,
                Exchange:    connector.Exchange,
                Pair:        pair,
                Timeframe:   timeframe,
                Type:        JobTypeBackfill,
                Priority:    PriorityNormal,
            }
            s.jobQueue.Enqueue(job)
        }
    }
    return nil
}

func (s *Service) ProcessJob(ctx context.Context, job *CollectionJob) error {
    // Acquire rate limit token
    if err := s.rateLimiter.Acquire(ctx, job.Exchange); err != nil {
        return err
    }
    
    // Get last collected timestamp
    metadata, err := s.metadataRepo.Get(ctx, job.Exchange, job.Pair, job.Timeframe)
    if err != nil && err != ErrNotFound {
        return err
    }
    
    var since int64
    if metadata != nil && job.Type == JobTypeIncremental {
        since = metadata.LastTimestamp
    }
    
    // Fetch OHLCV data
    candles, err := s.ccxt.FetchOHLCV(ctx, job.Exchange, job.Pair, job.Timeframe, since, 1000)
    if err != nil {
        s.eventBus.Publish(events.CollectionFailed{
            ConnectorID: job.ConnectorID,
            Exchange:    job.Exchange,
            Pair:        job.Pair,
            Error:       err.Error(),
        })
        return err
    }
    
    if len(candles) == 0 {
        return nil
    }
    
    // Store candles
    if err := s.ohlcvRepo.BulkUpsert(ctx, job.Exchange, job.Pair, job.Timeframe, candles); err != nil {
        return err
    }
    
    // Update metadata
    newMetadata := &Metadata{
        Exchange:       job.Exchange,
        Pair:           job.Pair,
        Timeframe:      job.Timeframe,
        FirstTimestamp: candles[0].Timestamp,
        LastTimestamp:  candles[len(candles)-1].Timestamp,
        CandleCount:    len(candles),
        UpdatedAt:      time.Now(),
    }
    
    if metadata != nil {
        newMetadata.CandleCount = metadata.CandleCount + len(candles)
        if candles[0].Timestamp < metadata.FirstTimestamp {
            newMetadata.FirstTimestamp = metadata.FirstTimestamp
        }
    }
    
    if err := s.metadataRepo.Upsert(ctx, newMetadata); err != nil {
        return err
    }
    
    // Detect gaps
    gaps := s.gapDetector.Detect(candles, job.Timeframe)
    for _, g := range gaps {
        s.eventBus.Publish(events.GapDetected{
            Exchange:  job.Exchange,
            Pair:      job.Pair,
            Timeframe: job.Timeframe,
            Start:     g.Start,
            End:       g.End,
        })
    }
    
    // Trigger indicator computation
    s.indicatorSvc.ComputeForRange(ctx, job.Exchange, job.Pair, job.Timeframe,
        candles[0].Timestamp, candles[len(candles)-1].Timestamp)
    
    // Emit success event
    s.eventBus.Publish(events.CollectionCompleted{
        ConnectorID: job.ConnectorID,
        Exchange:    job.Exchange,
        Pair:        job.Pair,
        Timeframe:   job.Timeframe,
        Count:       len(candles),
    })
    
    // If backfill and more data available, queue next job
    if job.Type == JobTypeBackfill && len(candles) == 1000 {
        nextJob := *job
        nextJob.ID = uuid.New()
        s.jobQueue.Enqueue(&nextJob)
    }
    
    return nil
}
```

### 4.3 Core Layer

#### 4.3.1 Rate Limit Manager

```go
// internal/core/ratelimit/manager.go

package ratelimit

import (
    "context"
    "sync"
    "time"
)

type Manager struct {
    buckets map[string]*TokenBucket
    configs map[string]*ExchangeConfig
    mu      sync.RWMutex
}

type TokenBucket struct {
    tokens     float64
    maxTokens  float64
    refillRate float64 // tokens per second
    lastRefill time.Time
    mu         sync.Mutex
}

type ExchangeConfig struct {
    Exchange       string
    RequestsPerMin int
    BurstSize      int
}

func NewManager(configs []ExchangeConfig) *Manager {
    m := &Manager{
        buckets: make(map[string]*TokenBucket),
        configs: make(map[string]*ExchangeConfig),
    }
    
    for _, cfg := range configs {
        m.configs[cfg.Exchange] = &cfg
        m.buckets[cfg.Exchange] = &TokenBucket{
            tokens:     float64(cfg.BurstSize),
            maxTokens:  float64(cfg.BurstSize),
            refillRate: float64(cfg.RequestsPerMin) / 60.0,
            lastRefill: time.Now(),
        }
    }
    
    return m
}

func (m *Manager) Acquire(ctx context.Context, exchange string) error {
    m.mu.RLock()
    bucket, ok := m.buckets[exchange]
    m.mu.RUnlock()
    
    if !ok {
        return ErrExchangeNotConfigured
    }
    
    return bucket.Acquire(ctx)
}

func (b *TokenBucket) Acquire(ctx context.Context) error {
    for {
        b.mu.Lock()
        b.refill()
        
        if b.tokens >= 1 {
            b.tokens--
            b.mu.Unlock()
            return nil
        }
        
        // Calculate wait time
        waitTime := time.Duration((1 - b.tokens) / b.refillRate * float64(time.Second))
        b.mu.Unlock()
        
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(waitTime):
            // Retry
        }
    }
}

func (b *TokenBucket) refill() {
    now := time.Now()
    elapsed := now.Sub(b.lastRefill).Seconds()
    b.tokens = min(b.maxTokens, b.tokens+elapsed*b.refillRate)
    b.lastRefill = now
}
```

#### 4.3.2 Job Queue

```go
// internal/core/queue/job_queue.go

package queue

import (
    "container/heap"
    "context"
    "sync"
)

type Priority int

const (
    PriorityLow    Priority = 0
    PriorityNormal Priority = 1
    PriorityHigh   Priority = 2
)

type Job interface {
    GetID() uuid.UUID
    GetPriority() Priority
    GetCreatedAt() time.Time
}

type JobQueue struct {
    heap    *jobHeap
    cond    *sync.Cond
    workers int
    handler func(context.Context, Job) error
}

type jobHeap []Job

func (h jobHeap) Len() int { return len(h) }
func (h jobHeap) Less(i, j int) bool {
    // Higher priority first, then older first
    if h[i].GetPriority() != h[j].GetPriority() {
        return h[i].GetPriority() > h[j].GetPriority()
    }
    return h[i].GetCreatedAt().Before(h[j].GetCreatedAt())
}
func (h jobHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *jobHeap) Push(x interface{}) { *h = append(*h, x.(Job)) }
func (h *jobHeap) Pop() interface{} {
    old := *h
    n := len(old)
    x := old[n-1]
    *h = old[0 : n-1]
    return x
}

func NewJobQueue(workers int, handler func(context.Context, Job) error) *JobQueue {
    jq := &JobQueue{
        heap:    &jobHeap{},
        workers: workers,
        handler: handler,
    }
    jq.cond = sync.NewCond(&sync.Mutex{})
    heap.Init(jq.heap)
    return jq
}

func (jq *JobQueue) Start(ctx context.Context) {
    for i := 0; i < jq.workers; i++ {
        go jq.worker(ctx)
    }
}

func (jq *JobQueue) Enqueue(job Job) {
    jq.cond.L.Lock()
    heap.Push(jq.heap, job)
    jq.cond.L.Unlock()
    jq.cond.Signal()
}

func (jq *JobQueue) worker(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
        }
        
        jq.cond.L.Lock()
        for jq.heap.Len() == 0 {
            jq.cond.Wait()
        }
        job := heap.Pop(jq.heap).(Job)
        jq.cond.L.Unlock()
        
        if err := jq.handler(ctx, job); err != nil {
            // Log error, maybe requeue with backoff
        }
    }
}
```

#### 4.3.3 Circuit Breaker

```go
// internal/core/circuit/breaker.go

package circuit

import (
    "sync"
    "time"
)

type State int

const (
    StateClosed State = iota
    StateOpen
    StateHalfOpen
)

type Breaker struct {
    name          string
    maxFailures   int
    timeout       time.Duration
    halfOpenMax   int
    
    state         State
    failures      int
    successes     int
    lastFailure   time.Time
    mu            sync.RWMutex
}

func NewBreaker(name string, maxFailures int, timeout time.Duration) *Breaker {
    return &Breaker{
        name:        name,
        maxFailures: maxFailures,
        timeout:     timeout,
        halfOpenMax: 3,
        state:       StateClosed,
    }
}

func (b *Breaker) Execute(fn func() error) error {
    if !b.AllowRequest() {
        return ErrCircuitOpen
    }
    
    err := fn()
    
    if err != nil {
        b.RecordFailure()
    } else {
        b.RecordSuccess()
    }
    
    return err
}

func (b *Breaker) AllowRequest() bool {
    b.mu.RLock()
    defer b.mu.RUnlock()
    
    switch b.state {
    case StateClosed:
        return true
    case StateOpen:
        if time.Since(b.lastFailure) > b.timeout {
            b.mu.RUnlock()
            b.mu.Lock()
            b.state = StateHalfOpen
            b.successes = 0
            b.mu.Unlock()
            b.mu.RLock()
            return true
        }
        return false
    case StateHalfOpen:
        return true
    }
    return false
}

func (b *Breaker) RecordSuccess() {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    if b.state == StateHalfOpen {
        b.successes++
        if b.successes >= b.halfOpenMax {
            b.state = StateClosed
            b.failures = 0
        }
    } else {
        b.failures = 0
    }
}

func (b *Breaker) RecordFailure() {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    b.failures++
    b.lastFailure = time.Now()
    
    if b.failures >= b.maxFailures {
        b.state = StateOpen
    }
    
    if b.state == StateHalfOpen {
        b.state = StateOpen
    }
}
```

### 4.4 Indicator Calculator

```go
// internal/services/indicator/calculator.go

package indicator

import (
    "math"
)

type Calculator struct{}

// SMA - Simple Moving Average
func (c *Calculator) SMA(prices []float64, period int) []float64 {
    if len(prices) < period {
        return nil
    }
    
    result := make([]float64, len(prices)-period+1)
    sum := 0.0
    
    for i := 0; i < period; i++ {
        sum += prices[i]
    }
    result[0] = sum / float64(period)
    
    for i := period; i < len(prices); i++ {
        sum = sum - prices[i-period] + prices[i]
        result[i-period+1] = sum / float64(period)
    }
    
    return result
}

// EMA - Exponential Moving Average
func (c *Calculator) EMA(prices []float64, period int) []float64 {
    if len(prices) < period {
        return nil
    }
    
    result := make([]float64, len(prices))
    multiplier := 2.0 / float64(period+1)
    
    // First EMA is SMA
    sum := 0.0
    for i := 0; i < period; i++ {
        sum += prices[i]
    }
    result[period-1] = sum / float64(period)
    
    // Calculate EMA
    for i := period; i < len(prices); i++ {
        result[i] = (prices[i]-result[i-1])*multiplier + result[i-1]
    }
    
    return result
}

// RSI - Relative Strength Index
func (c *Calculator) RSI(prices []float64, period int) []float64 {
    if len(prices) < period+1 {
        return nil
    }
    
    changes := make([]float64, len(prices)-1)
    for i := 1; i < len(prices); i++ {
        changes[i-1] = prices[i] - prices[i-1]
    }
    
    result := make([]float64, len(prices)-period)
    
    var avgGain, avgLoss float64
    for i := 0; i < period; i++ {
        if changes[i] > 0 {
            avgGain += changes[i]
        } else {
            avgLoss -= changes[i]
        }
    }
    avgGain /= float64(period)
    avgLoss /= float64(period)
    
    if avgLoss == 0 {
        result[0] = 100
    } else {
        rs := avgGain / avgLoss
        result[0] = 100 - (100 / (1 + rs))
    }
    
    for i := period; i < len(changes); i++ {
        gain := 0.0
        loss := 0.0
        if changes[i] > 0 {
            gain = changes[i]
        } else {
            loss = -changes[i]
        }
        
        avgGain = (avgGain*float64(period-1) + gain) / float64(period)
        avgLoss = (avgLoss*float64(period-1) + loss) / float64(period)
        
        if avgLoss == 0 {
            result[i-period+1] = 100
        } else {
            rs := avgGain / avgLoss
            result[i-period+1] = 100 - (100 / (1 + rs))
        }
    }
    
    return result
}

// MACD - Moving Average Convergence Divergence
func (c *Calculator) MACD(prices []float64, fastPeriod, slowPeriod, signalPeriod int) (macd, signal, histogram []float64) {
    fastEMA := c.EMA(prices, fastPeriod)
    slowEMA := c.EMA(prices, slowPeriod)
    
    if len(fastEMA) == 0 || len(slowEMA) == 0 {
        return nil, nil, nil
    }
    
    // Align arrays
    offset := slowPeriod - fastPeriod
    macdLine := make([]float64, len(slowEMA))
    for i := 0; i < len(slowEMA); i++ {
        macdLine[i] = fastEMA[i+offset] - slowEMA[i]
    }
    
    signalLine := c.EMA(macdLine, signalPeriod)
    
    hist := make([]float64, len(signalLine))
    for i := 0; i < len(signalLine); i++ {
        hist[i] = macdLine[i+signalPeriod-1] - signalLine[i]
    }
    
    return macdLine, signalLine, hist
}

// BollingerBands
func (c *Calculator) BollingerBands(prices []float64, period int, stdDev float64) (upper, middle, lower []float64) {
    middle = c.SMA(prices, period)
    if middle == nil {
        return nil, nil, nil
    }
    
    upper = make([]float64, len(middle))
    lower = make([]float64, len(middle))
    
    for i := 0; i < len(middle); i++ {
        // Calculate standard deviation
        sum := 0.0
        for j := 0; j < period; j++ {
            diff := prices[i+j] - middle[i]
            sum += diff * diff
        }
        std := math.Sqrt(sum / float64(period))
        
        upper[i] = middle[i] + stdDev*std
        lower[i] = middle[i] - stdDev*std
    }
    
    return upper, middle, lower
}
```

---

## 5. Data Flow Diagrams

### 5.1 Connector Creation Flow

```
┌──────────┐      ┌──────────┐      ┌──────────┐      ┌──────────┐
│  Client  │      │   API    │      │ Service  │      │   Repo   │
└────┬─────┘      └────┬─────┘      └────┬─────┘      └────┬─────┘
     │                 │                 │                 │
     │  POST /connectors                 │                 │
     │────────────────►│                 │                 │
     │                 │                 │                 │
     │                 │  Create()       │                 │
     │                 │────────────────►│                 │
     │                 │                 │                 │
     │                 │                 │  Validate       │
     │                 │                 │─────────────────│
     │                 │                 │                 │
     │                 │                 │  Save()         │
     │                 │                 │────────────────►│
     │                 │                 │                 │
     │                 │                 │  Emit Event     │
     │                 │                 │─────────────────│
     │                 │                 │                 │
     │                 │  Connector      │                 │
     │                 │◄────────────────│                 │
     │                 │                 │                 │
     │  201 Created    │                 │                 │
     │◄────────────────│                 │                 │
     │                 │                 │                 │
```

### 5.2 Data Collection Flow

```
┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
│Scheduler │   │ Collector│   │Rate Limit│   │   CCXT   │   │ MongoDB  │
└────┬─────┘   └────┬─────┘   └────┬─────┘   └────┬─────┘   └────┬─────┘
     │              │              │              │              │
     │  Trigger     │              │              │              │
     │─────────────►│              │              │              │
     │              │              │              │              │
     │              │  Acquire()   │              │              │
     │              │─────────────►│              │              │
     │              │              │              │              │
     │              │  Token OK    │              │              │
     │              │◄─────────────│              │              │
     │              │              │              │              │
     │              │  FetchOHLCV()│              │              │
     │              │─────────────────────────────►              │
     │              │              │              │              │
     │              │  Candles[]   │              │              │
     │              │◄─────────────────────────────              │
     │              │              │              │              │
     │              │  BulkUpsert()│              │              │
     │              │─────────────────────────────────────────────►
     │              │              │              │              │
     │              │  Compute Indicators         │              │
     │              │─────────────────────────────────────────────►
     │              │              │              │              │
```

---

## 6. Configuration

```yaml
# configs/config.yaml

server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

mongodb:
  uri: "mongodb://localhost:27017"
  database: "gotrading_data"
  max_pool_size: 100
  min_pool_size: 10

postgres:
  host: "localhost"
  port: 5432
  database: "gotrading_config"
  user: "gotrading"
  password: "${POSTGRES_PASSWORD}"
  max_connections: 50
  min_connections: 10

collector:
  workers: 10
  batch_size: 1000
  backfill_enabled: true
  gap_detection_enabled: true

indicators:
  auto_compute: true
  workers: 5
  default_config:
    sma_periods: [10, 20, 50, 200]
    ema_periods: [12, 26, 50]
    rsi_period: 14
    macd:
      fast: 12
      slow: 26
      signal: 9
    bollinger:
      period: 20
      std_dev: 2

rate_limits:
  default:
    requests_per_min: 60
    burst_size: 10
  binance:
    requests_per_min: 1200
    burst_size: 20
  coinbase:
    requests_per_min: 300
    burst_size: 10

websocket:
  read_buffer_size: 1024
  write_buffer_size: 1024
  ping_interval: 30s

logging:
  level: "info"
  format: "json"
  output: "stdout"

metrics:
  enabled: true
  port: 9090
```

---

## 7. Deployment

### 7.1 Docker

```dockerfile
# deployments/docker/Dockerfile

FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /data-collector ./cmd/server

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /data-collector .
COPY configs/config.yaml ./configs/

EXPOSE 8080 9090

ENTRYPOINT ["./data-collector"]
```

### 7.2 Docker Compose

```yaml
# docker-compose.yml

version: '3.8'

services:
  data-collector:
    build:
      context: .
      dockerfile: deployments/docker/Dockerfile
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - MONGODB_URI=mongodb://mongodb:27017
      - POSTGRES_HOST=postgres
    depends_on:
      - mongodb
      - postgres

  mongodb:
    image: mongo:7
    ports:
      - "27017:27017"
    volumes:
      - mongodb_data:/data/db

  postgres:
    image: postgres:16
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_USER=gotrading
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - POSTGRES_DB=gotrading_config
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  mongodb_data:
  postgres_data:
```

---

## 8. Monitoring & Observability

### 8.1 Prometheus Metrics

```go
// internal/metrics/metrics.go

package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    CandlesCollected = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "data_collector_candles_total",
            Help: "Total number of candles collected",
        },
        []string{"exchange", "pair", "timeframe"},
    )
    
    CollectionDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "data_collector_fetch_duration_seconds",
            Help:    "Duration of OHLCV fetch operations",
            Buckets: prometheus.DefBuckets,
        },
        []string{"exchange"},
    )
    
    RateLimitWaits = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "data_collector_rate_limit_waits_total",
            Help: "Total number of rate limit waits",
        },
        []string{"exchange"},
    )
    
    ActiveConnectors = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "data_collector_active_connectors",
            Help: "Number of active connectors",
        },
    )
    
    JobQueueSize = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "data_collector_job_queue_size",
            Help: "Current size of the job queue",
        },
    )
    
    GapsDetected = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "data_collector_gaps_detected_total",
            Help: "Total number of gaps detected",
        },
        []string{"exchange", "pair", "timeframe"},
    )
)
```

---

*End of Document*
