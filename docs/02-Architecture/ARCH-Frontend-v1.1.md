# Architecture — Frontend (v1.1)

- **Last updated**: 2026-01-19
- **Goal**: Admin UI to manage exchanges (connectors), ingestion jobs, and operational visibility.

## 1. Tech stack

- React + TypeScript
- Routing: React Router
- Data fetching: TanStack Query
- UI: MUI / Chakra / shadcn (choose one)
- Auth: JWT (admin-only), optional 2FA

## 2. Navigation

- Dashboard
- Connectors
  - List / Detail
- Jobs
  - List / Create / Edit / Run now
- Data Explorer (read-only)
- Settings

## 3. Key data models

### Connector

A connector represents **one exchange**.

```ts
type Connector = {
  id: string;
  exchange_id: string; // "binance", "okx", "kucoin", ...
  status: "active" | "paused" | "disabled";
  created_at: string;
  updated_at: string;
  rate_limit: {
    limit: number;
    period_ms: number;
  };
  last_job_run_at?: string;
  jobs_count: number;
  active_jobs_count: number;
};
```

### Job

```ts
type Job = {
  id: string;
  connector_id: string;
  symbol: string;       // "BTC/USDT"
  timeframe: string;    // "1m", "5m", "1h", ...
  status: "active" | "paused" | "disabled";
  schedule: { kind: "interval"; every_ms: number } | { kind: "cron"; cron: string };
  cursor: { since_ms: number; last_candle_open_ms?: number };
  next_run_at: string;
  last_run_at?: string;
  last_success_at?: string;
  last_error?: { at: string; message: string; code?: string };
};
```

## 4. Screens (MVP)

### 4.1 Dashboard

- Cards: active connectors, active jobs, backlog (jobs behind cursor), last 24h errors
- Tables: last job runs, recent errors

### 4.2 Connectors — list

- Table: exchange, status, RL (limit/period), active jobs, last job run
- Actions: Create / Pause / Resume / Disable

### 4.3 Connector detail

- Connector header + status
- Embedded jobs table filtered by connector
- Buttons: Create job, Pause all jobs, Resume all jobs

### 4.4 Jobs — list

- Filters: exchange, symbol, timeframe, status, error-only
- Actions: Run now, Pause/Resume, Edit

### 4.5 Job create/edit

- Fields: Exchange (connector), symbol, timeframe, schedule, start date, indicators toggle
- Validation: timeframe allowed by exchange; schedule not faster than timeframe

## 5. Error UX

- If an operation fails, show:
  - user-friendly message
  - technical details (expandable)
  - correlation id (if backend provides one)

## 6. Frontend observability

- Capture API errors with a client-side error boundary
- Optional: send UI error events to backend `/api/v1/events/ui`
