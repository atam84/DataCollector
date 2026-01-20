# API Spec (v1.1)

- **Base URL**: `/api/v1`
- **Auth**: Admin endpoints require JWT (and optional 2FA).

## 1. Health

- `GET /health`

## 2. Connectors

A connector represents one exchange.

- `GET /connectors`
- `GET /connectors/{exchange_id}`
- `PUT /connectors/{exchange_id}` (create or update)
- `PATCH /connectors/{exchange_id}/status` (enable/disable)

### 2.1 Connector payload (PUT)

```json
{
  "exchange_id": "binance",
  "status": "active",
  "rate_limit": {
    "capacity": 1200,
    "period_seconds": 60
  },
  "notes": "optional"
}
```

## 3. Jobs

A job represents a **symbol + timeframe** ingestion loop for a given exchange.

- `GET /jobs?exchange_id=binance&status=active`
- `POST /jobs`
- `GET /jobs/{job_id}`
- `PATCH /jobs/{job_id}`
- `PATCH /jobs/{job_id}/status` (enable/disable/pause)
- `POST /jobs/{job_id}/run` (force run once)
- `POST /jobs/{job_id}/backfill` (bounded historical fill)

### 3.1 Job payload (POST)

```json
{
  "exchange_id": "binance",
  "symbol": "BTC/USDT",
  "timeframe": "1m",
  "priority": 5,
  "schedule": {
    "mode": "fixed",
    "every_seconds": 60
  },
  "start_from": "2024-01-01T00:00:00Z",
  "status": "active"
}
```
