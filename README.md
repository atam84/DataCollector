# Data Collector

A cryptocurrency market data collection service that ingests OHLCV candles from multiple exchanges, computes technical indicators, and stores them for downstream analysis.

## Features

- ✅ Multi-exchange support via CCXT (100+ exchanges)
- ✅ OHLCV candle collection with configurable timeframes
- ✅ **Sandbox/Testnet mode with UI toggle**
- ✅ Built-in rate limiting
- ✅ MongoDB storage with optimized indexes
- ✅ RESTful API (15 endpoints)
- ✅ React + Tailwind Admin UI
- ✅ **Docker Compose deployment**
- ✅ Technical indicators (RSI, EMA, MACD)
- 🚧 Job scheduler and worker (coming soon)

## Architecture

See [docs/](./docs/) for complete documentation:
- [Project Overview](./docs/00-Overview/PROJECT-OVERVIEW.md)
- [PRD](./docs/01-PRD/PRD-DataCollector-v3.1.md)
- [Backend Architecture](./docs/02-Architecture/ARCH-Backend-v1.1.md)
- [CCXT Go Reference](./docs/CCXT-GO-PUBLIC-API-REFERENCE.md)

## Project Structure

```
.
├── cmd/
│   ├── api/           # API server
│   └── worker/        # Background workers
├── internal/
│   ├── api/           # HTTP handlers & middleware
│   ├── config/        # Configuration loader
│   ├── exchange/      # Exchange adapters (CCXT)
│   ├── models/        # Data models
│   ├── repository/    # Database layer
│   ├── scheduler/     # Job scheduler
│   └── service/       # Business logic
├── pkg/
│   └── logger/        # Logging utilities
├── web/               # React frontend
└── docs/              # Documentation
```

## Quick Start

### Option A: Docker Compose (Recommended) 🐳

**Prerequisites**: Docker 20.10+ and Docker Compose 2.0+

```bash
# Build and start all services (backend + frontend + mongodb)
docker-compose up -d

# View logs
docker-compose logs -f

# Access the application
# - Frontend: http://localhost:3000
# - Backend API: http://localhost:8080
# - MongoDB: localhost:27017

# Stop all services
docker-compose down
```

**That's it!** See [DOCKER.md](DOCKER.md) for complete Docker documentation.

### Option B: Manual Installation

**Prerequisites**: Go 1.21+, MongoDB 5.0+, Node.js 18+

#### Installation

**1. Clone the repository**
```bash
git clone <repo-url>
cd DataCollector
```

**2. Install Go dependencies**
```bash
go mod download
```

**3. Configure environment**
```bash
cp .env.example .env
# Edit .env with your configuration
```

**4. Start MongoDB**
```bash
make docker-up
# Or manually: docker run -d -p 27017:27017 --name mongodb mongo:latest
```

**5. Run the backend API**
```bash
make run
# Or manually: go run cmd/api/main.go
```

**6. Run the frontend UI** (optional)
```bash
cd web
npm install
npm run dev
# Access at http://localhost:3000
```

**7. Test the API**
```bash
curl http://localhost:8080/health
# Or run automated tests: ./scripts/test-api.sh
```

## Configuration

Configuration is loaded from environment variables or `.env` file.

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | HTTP server port | `8080` |
| `SERVER_HOST` | HTTP server host | `0.0.0.0` |
| `MONGODB_URI` | MongoDB connection URI | `mongodb://localhost:27017` |
| `MONGODB_DATABASE` | MongoDB database name | `datacollector` |
| `EXCHANGE_SANDBOX_MODE` | Use exchange sandbox/testnet | `true` |
| `EXCHANGE_ENABLE_RATE_LIMIT` | Enable built-in rate limiting | `true` |
| `EXCHANGE_REQUEST_TIMEOUT` | Request timeout (ms) | `30000` |

### Sandbox Mode

**Sandbox mode is enabled by default** for safety during development.

```bash
# .env file
EXCHANGE_SANDBOX_MODE=true
```

When enabled:
- Binance → Binance Testnet
- Bybit → Bybit Testnet
- OKX → OKX Demo Trading
- Other exchanges with sandbox support

To use production APIs:
```bash
EXCHANGE_SANDBOX_MODE=false
```

## API Endpoints

### Health Check
```bash
GET /health
GET /api/v1/health
```

**Response:**
```json
{
  "status": "ok",
  "timestamp": 1674567890,
  "services": {
    "database": {
      "status": "healthy"
    }
  }
}
```

## Development

### Run in development mode
```bash
go run cmd/api/main.go
```

### Build for production
```bash
go build -o bin/api cmd/api/main.go
```

### Run with Docker
```bash
docker-compose up
```

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/exchange/
```

## Supported Exchanges

CCXT supports 100+ exchanges. Common ones:

- Binance (Spot, USDⓈ-M, COIN-M)
- Coinbase
- Bybit
- OKX
- Kraken
- KuCoin
- Bitfinex
- And many more...

See [CCXT-GO-PUBLIC-API-REFERENCE.md](./docs/CCXT-GO-PUBLIC-API-REFERENCE.md) for full list.

## Roadmap

**Phase 1 - Foundations** ✅
- [x] Project setup & configuration
- [x] MongoDB connection
- [x] Health check endpoint
- [x] Exchange adapter (CCXT)
- [x] Data models (Connector, Job, OHLCV)

**Phase 2 - Backend API** ✅
- [x] Connector CRUD API (6 endpoints)
- [x] Job CRUD API (8 endpoints)
- [x] Sandbox mode toggle endpoint
- [x] Rate limiting & job locking
- [x] MongoDB indexes & optimizations

**Phase 3 - Admin UI** ✅
- [x] React + Vite + Tailwind setup
- [x] Dashboard with statistics
- [x] Connector management UI
- [x] **Sandbox mode toggle switch**
- [x] Job management UI
- [x] Pause/resume job actions

**Phase 4 - Docker Deployment** ✅
- [x] Backend Dockerfile (multi-stage)
- [x] Frontend Dockerfile (Nginx)
- [x] Docker Compose setup
- [x] Development mode support
- [x] Health checks & networking

**Phase 5 - Future Enhancements** 🚧
- [ ] Job scheduler implementation
- [ ] OHLCV ingestion worker
- [ ] Indicator computation (RSI, EMA, MACD)
- [ ] Authentication & authorization
- [ ] Metrics & monitoring (Prometheus)
- [ ] Unit & integration tests

## Documentation

### Quick Start & Guides
- **[QUICKSTART.md](QUICKSTART.md)** - Get up and running in 3 minutes
- **[DOCKER.md](DOCKER.md)** - Complete Docker deployment guide
- **[STATUS.md](STATUS.md)** - Detailed implementation status
- **[PROJECT-SUMMARY.md](PROJECT-SUMMARY.md)** - Complete project overview

### Phase Summaries
- **[PHASE2-COMPLETE.md](PHASE2-COMPLETE.md)** - Backend API completion summary
- **[PHASE3-COMPLETE.md](PHASE3-COMPLETE.md)** - Frontend UI completion summary

### API Documentation
- **[API-ENDPOINTS.md](docs/04-API/API-ENDPOINTS.md)** - Complete API reference with examples
- **[CCXT-GO-PUBLIC-API-REFERENCE.md](docs/CCXT-GO-PUBLIC-API-REFERENCE.md)** - CCXT Go library guide

### Frontend
- **[web/README.md](web/README.md)** - Frontend documentation

### Architecture & Requirements
- **[PRD](docs/01-PRD/PRD-DataCollector-v3.1.md)** - Product requirements document
- **[Architecture](docs/02-Architecture/ARCH-Backend-v1.1.md)** - Backend architecture

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) (coming soon)

## License

MIT License - see [LICENSE](./LICENSE)

## Support

For questions or issues:
- Open an issue on GitHub
- Check the [docs/](./docs/) folder
- Review the [CCXT documentation](https://docs.ccxt.com/)
