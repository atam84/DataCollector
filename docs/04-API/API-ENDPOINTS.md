# API Endpoints Reference

**Base URL**: `http://localhost:8080/api/v1`

---

## Table of Contents

1. [Health Check](#health-check)
2. [Connectors](#connectors)
3. [Jobs](#jobs)

---

## Health Check

### Get Health Status

Check the health of the application and its services.

**Endpoint**: `GET /api/v1/health`

**Response**:
```json
{
  "status": "ok",
  "timestamp": 1674567890,
  "services": {
    "database": {
      "status": "healthy",
      "error": ""
    }
  }
}
```

---

## Connectors

Connectors represent exchange integrations with sandbox mode support.

### Create Connector

Create a new connector for an exchange.

**Endpoint**: `POST /api/v1/connectors`

**Request Body**:
```json
{
  "exchange_id": "binance",
  "display_name": "Binance",
  "sandbox_mode": true,
  "rate_limit": {
    "limit": 1200,
    "period_ms": 60000
  }
}
```

**Response** (201 Created):
```json
{
  "id": "507f1f77bcf86cd799439011",
  "exchange_id": "binance",
  "display_name": "Binance",
  "status": "active",
  "sandbox_mode": true,
  "created_at": "2026-01-20T10:00:00Z",
  "updated_at": "2026-01-20T10:00:00Z",
  "rate_limit": {
    "limit": 1200,
    "period_ms": 60000,
    "usage": 0,
    "period_start": "2026-01-20T10:00:00Z",
    "last_job_run_at": null
  },
  "credentials_ref": {
    "mode": "",
    "keys": null
  }
}
```

**cURL Example**:
```bash
curl -X POST http://localhost:8080/api/v1/connectors \
  -H "Content-Type: application/json" \
  -d '{
    "exchange_id": "binance",
    "display_name": "Binance",
    "sandbox_mode": true,
    "rate_limit": {
      "limit": 1200,
      "period_ms": 60000
    }
  }'
```

---

### Get All Connectors

Retrieve all connectors with optional filters.

**Endpoint**: `GET /api/v1/connectors`

**Query Parameters**:
- `status` (optional): Filter by status (`active`, `disabled`)
- `sandbox_mode` (optional): Filter by sandbox mode (`true`, `false`)

**Response**:
```json
{
  "data": [
    {
      "id": "507f1f77bcf86cd799439011",
      "exchange_id": "binance",
      "display_name": "Binance",
      "status": "active",
      "sandbox_mode": true,
      "created_at": "2026-01-20T10:00:00Z",
      "updated_at": "2026-01-20T10:00:00Z",
      "rate_limit": {
        "limit": 1200,
        "period_ms": 60000,
        "usage": 0,
        "period_start": "2026-01-20T10:00:00Z"
      }
    }
  ],
  "count": 1
}
```

**cURL Examples**:
```bash
# Get all connectors
curl http://localhost:8080/api/v1/connectors

# Get active connectors only
curl http://localhost:8080/api/v1/connectors?status=active

# Get sandbox connectors only
curl http://localhost:8080/api/v1/connectors?sandbox_mode=true
```

---

### Get Connector by ID

Retrieve a specific connector.

**Endpoint**: `GET /api/v1/connectors/:id`

**Response**:
```json
{
  "id": "507f1f77bcf86cd799439011",
  "exchange_id": "binance",
  "display_name": "Binance",
  "status": "active",
  "sandbox_mode": true,
  "created_at": "2026-01-20T10:00:00Z",
  "updated_at": "2026-01-20T10:00:00Z",
  "rate_limit": {
    "limit": 1200,
    "period_ms": 60000,
    "usage": 0,
    "period_start": "2026-01-20T10:00:00Z"
  }
}
```

**cURL Example**:
```bash
curl http://localhost:8080/api/v1/connectors/507f1f77bcf86cd799439011
```

---

### Update Connector

Update connector properties including sandbox mode.

**Endpoint**: `PUT /api/v1/connectors/:id`

**Request Body** (all fields optional):
```json
{
  "display_name": "Binance Main",
  "status": "active",
  "sandbox_mode": false,
  "rate_limit": {
    "limit": 2400,
    "period_ms": 60000
  }
}
```

**Response**:
```json
{
  "id": "507f1f77bcf86cd799439011",
  "exchange_id": "binance",
  "display_name": "Binance Main",
  "status": "active",
  "sandbox_mode": false,
  "created_at": "2026-01-20T10:00:00Z",
  "updated_at": "2026-01-20T10:05:00Z",
  "rate_limit": {
    "limit": 2400,
    "period_ms": 60000,
    "usage": 0,
    "period_start": "2026-01-20T10:05:00Z"
  }
}
```

**cURL Example**:
```bash
curl -X PUT http://localhost:8080/api/v1/connectors/507f1f77bcf86cd799439011 \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "Binance Main",
    "sandbox_mode": false
  }'
```

---

### Toggle Sandbox Mode

Dedicated endpoint to toggle sandbox mode (useful for UI switches).

**Endpoint**: `PATCH /api/v1/connectors/:id/sandbox`

**Request Body**:
```json
{
  "sandbox_mode": false
}
```

**Response**:
```json
{
  "message": "Sandbox mode updated successfully",
  "sandbox_mode": false,
  "connector": {
    "id": "507f1f77bcf86cd799439011",
    "exchange_id": "binance",
    "display_name": "Binance",
    "status": "active",
    "sandbox_mode": false,
    "created_at": "2026-01-20T10:00:00Z",
    "updated_at": "2026-01-20T10:10:00Z",
    "rate_limit": {
      "limit": 1200,
      "period_ms": 60000,
      "usage": 0,
      "period_start": "2026-01-20T10:10:00Z"
    }
  }
}
```

**cURL Example**:
```bash
# Enable sandbox mode
curl -X PATCH http://localhost:8080/api/v1/connectors/507f1f77bcf86cd799439011/sandbox \
  -H "Content-Type: application/json" \
  -d '{"sandbox_mode": true}'

# Disable sandbox mode (production)
curl -X PATCH http://localhost:8080/api/v1/connectors/507f1f77bcf86cd799439011/sandbox \
  -H "Content-Type: application/json" \
  -d '{"sandbox_mode": false}'
```

---

### Delete Connector

Delete a connector.

**Endpoint**: `DELETE /api/v1/connectors/:id`

**Response**: `204 No Content`

**cURL Example**:
```bash
curl -X DELETE http://localhost:8080/api/v1/connectors/507f1f77bcf86cd799439011
```

---

## Jobs

Jobs represent data collection tasks for a specific symbol and timeframe.

### Create Job

Create a new job for collecting OHLCV data.

**Endpoint**: `POST /api/v1/jobs`

**Request Body**:
```json
{
  "connector_exchange_id": "binance",
  "symbol": "BTC/USDT",
  "timeframe": "1h",
  "status": "active"
}
```

**Supported Timeframes**: `1m`, `5m`, `15m`, `30m`, `1h`, `4h`, `1d`, `1w`

**Response** (201 Created):
```json
{
  "id": "507f1f77bcf86cd799439012",
  "connector_exchange_id": "binance",
  "symbol": "BTC/USDT",
  "timeframe": "1h",
  "status": "active",
  "created_at": "2026-01-20T10:15:00Z",
  "updated_at": "2026-01-20T10:15:00Z",
  "schedule": {
    "mode": "timeframe",
    "cron": null
  },
  "cursor": {
    "last_candle_time": null
  },
  "run_state": {
    "locked_until": null,
    "last_run_time": null,
    "next_run_time": null,
    "last_error": null,
    "runs_total": 0
  }
}
```

**cURL Example**:
```bash
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "connector_exchange_id": "binance",
    "symbol": "BTC/USDT",
    "timeframe": "1h",
    "status": "active"
  }'
```

---

### Get All Jobs

Retrieve all jobs with optional filters.

**Endpoint**: `GET /api/v1/jobs`

**Query Parameters**:
- `status` (optional): Filter by status (`active`, `paused`, `error`)
- `exchange_id` (optional): Filter by exchange
- `symbol` (optional): Filter by symbol
- `timeframe` (optional): Filter by timeframe

**Response**:
```json
{
  "data": [
    {
      "id": "507f1f77bcf86cd799439012",
      "connector_exchange_id": "binance",
      "symbol": "BTC/USDT",
      "timeframe": "1h",
      "status": "active",
      "created_at": "2026-01-20T10:15:00Z",
      "updated_at": "2026-01-20T10:15:00Z",
      "schedule": {
        "mode": "timeframe"
      },
      "cursor": {
        "last_candle_time": null
      },
      "run_state": {
        "runs_total": 0
      }
    }
  ],
  "count": 1
}
```

**cURL Examples**:
```bash
# Get all jobs
curl http://localhost:8080/api/v1/jobs

# Get active jobs only
curl http://localhost:8080/api/v1/jobs?status=active

# Get jobs for specific exchange
curl http://localhost:8080/api/v1/jobs?exchange_id=binance

# Get jobs for specific symbol
curl http://localhost:8080/api/v1/jobs?symbol=BTC/USDT

# Combined filters
curl http://localhost:8080/api/v1/jobs?exchange_id=binance&timeframe=1h
```

---

### Get Job by ID

Retrieve a specific job.

**Endpoint**: `GET /api/v1/jobs/:id`

**Response**:
```json
{
  "id": "507f1f77bcf86cd799439012",
  "connector_exchange_id": "binance",
  "symbol": "BTC/USDT",
  "timeframe": "1h",
  "status": "active",
  "created_at": "2026-01-20T10:15:00Z",
  "updated_at": "2026-01-20T10:15:00Z",
  "schedule": {
    "mode": "timeframe",
    "cron": null
  },
  "cursor": {
    "last_candle_time": "2026-01-20T09:00:00Z"
  },
  "run_state": {
    "locked_until": null,
    "last_run_time": "2026-01-20T10:00:00Z",
    "next_run_time": "2026-01-20T11:00:00Z",
    "last_error": null,
    "runs_total": 5
  }
}
```

**cURL Example**:
```bash
curl http://localhost:8080/api/v1/jobs/507f1f77bcf86cd799439012
```

---

### Get Jobs by Connector

Retrieve all jobs for a specific connector/exchange.

**Endpoint**: `GET /api/v1/connectors/:exchangeId/jobs`

**Response**:
```json
{
  "data": [
    {
      "id": "507f1f77bcf86cd799439012",
      "connector_exchange_id": "binance",
      "symbol": "BTC/USDT",
      "timeframe": "1h",
      "status": "active"
    },
    {
      "id": "507f1f77bcf86cd799439013",
      "connector_exchange_id": "binance",
      "symbol": "ETH/USDT",
      "timeframe": "1h",
      "status": "active"
    }
  ],
  "count": 2,
  "exchange_id": "binance"
}
```

**cURL Example**:
```bash
curl http://localhost:8080/api/v1/connectors/binance/jobs
```

---

### Update Job

Update job properties.

**Endpoint**: `PUT /api/v1/jobs/:id`

**Request Body** (all fields optional):
```json
{
  "status": "paused",
  "timeframe": "4h"
}
```

**Response**:
```json
{
  "id": "507f1f77bcf86cd799439012",
  "connector_exchange_id": "binance",
  "symbol": "BTC/USDT",
  "timeframe": "4h",
  "status": "paused",
  "created_at": "2026-01-20T10:15:00Z",
  "updated_at": "2026-01-20T10:20:00Z",
  "schedule": {
    "mode": "timeframe"
  }
}
```

**cURL Example**:
```bash
curl -X PUT http://localhost:8080/api/v1/jobs/507f1f77bcf86cd799439012 \
  -H "Content-Type: application/json" \
  -d '{
    "status": "paused"
  }'
```

---

### Pause Job

Pause a running job.

**Endpoint**: `POST /api/v1/jobs/:id/pause`

**Response**:
```json
{
  "message": "Job paused successfully",
  "job": {
    "id": "507f1f77bcf86cd799439012",
    "status": "paused"
  }
}
```

**cURL Example**:
```bash
curl -X POST http://localhost:8080/api/v1/jobs/507f1f77bcf86cd799439012/pause
```

---

### Resume Job

Resume a paused job.

**Endpoint**: `POST /api/v1/jobs/:id/resume`

**Response**:
```json
{
  "message": "Job resumed successfully",
  "job": {
    "id": "507f1f77bcf86cd799439012",
    "status": "active"
  }
}
```

**cURL Example**:
```bash
curl -X POST http://localhost:8080/api/v1/jobs/507f1f77bcf86cd799439012/resume
```

---

### Delete Job

Delete a job.

**Endpoint**: `DELETE /api/v1/jobs/:id`

**Response**: `204 No Content`

**cURL Example**:
```bash
curl -X DELETE http://localhost:8080/api/v1/jobs/507f1f77bcf86cd799439012
```

---

## Error Responses

All endpoints return consistent error responses:

**4xx Client Errors**:
```json
{
  "error": "Invalid request body"
}
```

**5xx Server Errors**:
```json
{
  "error": "Internal server error message"
}
```

**Common Status Codes**:
- `200 OK` - Success
- `201 Created` - Resource created
- `204 No Content` - Success with no response body
- `400 Bad Request` - Invalid input
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error
- `503 Service Unavailable` - Service down (e.g., database)

---

## Complete Workflow Example

### 1. Create a Connector (Sandbox Mode)

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

### 2. Create Jobs for Data Collection

```bash
# BTC/USDT hourly
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "connector_exchange_id": "binance",
    "symbol": "BTC/USDT",
    "timeframe": "1h"
  }'

# ETH/USDT 15-minute
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "connector_exchange_id": "binance",
    "symbol": "ETH/USDT",
    "timeframe": "15m"
  }'
```

### 3. List All Jobs

```bash
curl http://localhost:8080/api/v1/jobs
```

### 4. Toggle to Production Mode

```bash
curl -X PATCH http://localhost:8080/api/v1/connectors/507f1f77bcf86cd799439011/sandbox \
  -H "Content-Type: application/json" \
  -d '{"sandbox_mode": false}'
```

### 5. Pause a Job

```bash
curl -X POST http://localhost:8080/api/v1/jobs/507f1f77bcf86cd799439012/pause
```

---

## Notes

- All timestamps are in ISO 8601 format (UTC)
- MongoDB ObjectIDs are returned as hex strings
- Sandbox mode can be toggled at any time
- Jobs are automatically scheduled based on timeframe
- Rate limiting is enforced per connector

---

**API Version**: v1
**Last Updated**: 2026-01-20
