# Data Collector - Quick Start Guide

Get up and running in 3 minutes! ⚡

---

## 🐳 Option A: Docker Compose (Easiest!)

**Prerequisites**: Docker 20.10+ and Docker Compose 2.0+

```bash
# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Access the application
# - Frontend: http://localhost:3000
# - Backend API: http://localhost:8080
```

**That's it!** Everything is running. Skip to section 4 to try the Admin UI.

For complete Docker documentation, see [DOCKER.md](DOCKER.md).

---

## 🔧 Option B: Manual Setup

### Prerequisites

- ✅ Go 1.21+
- ✅ MongoDB 5.0+
- ✅ Node.js 18+ and npm (for frontend)
- ✅ jq (for test script)

### 1. Start MongoDB

```bash
make docker-up
```

Or manually:
```bash
docker run -d -p 27017:27017 --name mongodb-datacollector mongo:latest
```

---

### 2. Run the API Server

```bash
make run
```

Or:
```bash
go run cmd/api/main.go
```

**Expected output:**
```
Connected to MongoDB successfully
Starting server on 0.0.0.0:8080 (Sandbox Mode: true)
```

---

### 3. Test the API

### Option A: Quick Health Check
```bash
curl http://localhost:8080/health
```

### Option B: Run Full Test Suite
```bash
./scripts/test-api.sh
```

This will test all 20 operations automatically! ✨

---

## 4. Try the API Manually

### Create a Connector (Sandbox Mode)
```bash
curl -X POST http://localhost:8080/api/v1/connectors \
  -H "Content-Type: application/json" \
  -d '{
    "exchange_id": "binance",
    "display_name": "Binance Testnet",
    "sandbox_mode": true,
    "rate_limit": {
      "limit": 1200,
      "period_ms": 60000
    }
  }'
```

**Response:**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "exchange_id": "binance",
  "display_name": "Binance Testnet",
  "status": "active",
  "sandbox_mode": true,
  ...
}
```

### Create a Job
```bash
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "connector_exchange_id": "binance",
    "symbol": "BTC/USDT",
    "timeframe": "1h"
  }'
```

### List Everything
```bash
# List connectors
curl http://localhost:8080/api/v1/connectors | jq

# List jobs
curl http://localhost:8080/api/v1/jobs | jq
```

### Toggle Sandbox Mode
```bash
# Get connector ID from previous response
CONNECTOR_ID="507f1f77bcf86cd799439011"

# Toggle to production
curl -X PATCH http://localhost:8080/api/v1/connectors/$CONNECTOR_ID/sandbox \
  -H "Content-Type: application/json" \
  -d '{"sandbox_mode": false}' | jq
```

---

## 4. Run the Admin UI (Optional)

### Start the React Frontend

```bash
cd web
npm install
npm run dev
```

**Access the UI**: `http://localhost:3000`

### What You Can Do:

- ✅ View dashboard with statistics
- ✅ Create and manage connectors
- ✅ **Toggle sandbox mode with visual switch** 🎯
- ✅ Create and manage jobs
- ✅ Pause/resume jobs
- ✅ Monitor job execution

The UI automatically proxies API requests to `http://localhost:8080`.

---

## 📚 Full Documentation

- **API Reference**: [`docs/04-API/API-ENDPOINTS.md`](docs/04-API/API-ENDPOINTS.md)
- **CCXT Guide**: [`docs/CCXT-GO-PUBLIC-API-REFERENCE.md`](docs/CCXT-GO-PUBLIC-API-REFERENCE.md)
- **Architecture**: [`docs/02-Architecture/ARCH-Backend-v1.1.md`](docs/02-Architecture/ARCH-Backend-v1.1.md)

---

## 🛠️ Useful Commands

```bash
# Build the binary
make build

# Run tests
make test

# Format code
make fmt

# Stop MongoDB
make docker-down

# Clean build artifacts
make clean
```

---

## 🎯 All Available Endpoints

### Connectors
- `POST /api/v1/connectors` - Create
- `GET /api/v1/connectors` - List all
- `GET /api/v1/connectors/:id` - Get one
- `PUT /api/v1/connectors/:id` - Update
- `DELETE /api/v1/connectors/:id` - Delete
- `PATCH /api/v1/connectors/:id/sandbox` - Toggle sandbox

### Jobs
- `POST /api/v1/jobs` - Create
- `GET /api/v1/jobs` - List all
- `GET /api/v1/jobs/:id` - Get one
- `PUT /api/v1/jobs/:id` - Update
- `DELETE /api/v1/jobs/:id` - Delete
- `POST /api/v1/jobs/:id/pause` - Pause
- `POST /api/v1/jobs/:id/resume` - Resume
- `GET /api/v1/connectors/:exchangeId/jobs` - Get connector jobs

### Health
- `GET /health` - Health check
- `GET /api/v1/health` - Health check (versioned)

---

## 🔧 Configuration

Edit `.env` file:

```bash
# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# MongoDB
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=datacollector

# Exchange (default sandbox for safety!)
EXCHANGE_SANDBOX_MODE=true
EXCHANGE_ENABLE_RATE_LIMIT=true
EXCHANGE_REQUEST_TIMEOUT=30000
```

---

## 🎉 You're Ready!

The application is fully functional! ✅

**What's Working**:
1. ✅ **Backend API** - 15 REST endpoints
2. ✅ **Admin UI** - React + Tailwind with sandbox toggle
3. ✅ **Database** - MongoDB with indexes and optimizations

**Next Steps** (Optional):
1. **Worker**: Implement scheduler and data ingestion
2. **Deploy**: Docker compose for production
3. **Authentication**: Add user auth and authorization

---

**Need help?** Check [`docs/`](docs/) or run `./scripts/test-api.sh` to see examples!
