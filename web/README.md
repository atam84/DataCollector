# Data Collector - Admin UI

Modern React admin interface for managing cryptocurrency data collection.

## Tech Stack

- **React 18** - UI framework
- **Vite** - Build tool and dev server
- **Tailwind CSS** - Utility-first CSS framework
- **Axios** - HTTP client

## Features

### Dashboard
- Overview of connectors and jobs
- Real-time statistics
- System health status
- Connector health monitoring (uptime, error rate, response time)
- Data quality summary with distribution chart

### Connector Management
- Create, view, and delete connectors
- Rate limit monitoring with cooldown visualization
- Health status synced with Dashboard API
- Support for 111 exchanges via CCXT

### Job Management
- Create, pause, resume, and delete jobs
- Monitor job execution status
- View last run times and errors
- **Filters**: Timeframe, Status, Exchange
- **Candles column**: Shows total candle count
- Support for multiple timeframes (1m, 5m, 15m, 30m, 1h, 4h, 1d)

### Data Quality
- Gap detection and completeness scoring
- Quality status: Excellent/Good/Fair/Poor
- Background quality checks with progress tracking
- Gap filling (fill missing candles)
- Historical backfill (fetch past data)

### Job Queue
- View pending/running jobs in queue
- Clickable symbols → job details modal
- Rate limit status per connector

### Charts (CandlestickChart)
- Professional candlestick visualization
- Volume histogram with color coding
- **Period selection**: 1D, 1W, 1M, 3M, 6M, 1Y, All
- **Zoom controls**: Zoom in/out/reset buttons
- Mouse wheel zoom and drag to pan
- Indicator overlays (SMA, EMA, Bollinger Bands, etc.)
- Separate panes for momentum indicators (RSI, MACD, Stochastic)

## Getting Started

### Prerequisites
- Node.js 18+ and npm

### Installation

```bash
# Install dependencies
npm install
```

### Development

```bash
# Start development server (with API proxy)
npm run dev
```

The dev server will start on `http://localhost:3000` and proxy API requests to `http://localhost:8080`.

**Make sure the backend API is running on port 8080!**

### Build for Production

```bash
# Build optimized production bundle
npm run build

# Preview production build
npm run preview
```

## Project Structure

```
web/
├── src/
│   ├── components/
│   │   ├── Dashboard.jsx         # Dashboard overview with health & quality
│   │   ├── ConnectorList.jsx     # Connector management with health sync
│   │   ├── ConnectorWizard.jsx   # Multi-step connector creation
│   │   ├── JobList.jsx           # Job management with filters
│   │   ├── JobDetails.jsx        # Job details modal (Overview, Data, Charts)
│   │   ├── JobWizard.jsx         # Multi-step job creation
│   │   ├── JobQueue.jsx          # Queue view with rate limits
│   │   ├── DataQuality.jsx       # Quality analysis & gap filling
│   │   ├── CandlestickChart.jsx  # Professional charts with zoom
│   │   └── IndicatorsInfo.jsx    # Indicators documentation
│   ├── App.jsx                   # Main app with tab navigation
│   ├── main.jsx                  # Entry point
│   └── index.css                 # Tailwind styles
├── index.html                    # HTML template
├── vite.config.js                # Vite configuration
├── tailwind.config.js            # Tailwind configuration
└── package.json                  # Dependencies
```

## Key Features

### Data Quality Monitoring

The Data Quality tab provides comprehensive monitoring:

- **Quality Status**: Excellent (>99%), Good (>95%), Fair (>90%), Poor (<90%)
- **Gap Detection**: Identifies missing candles in data
- **Completeness Score**: Percentage of expected candles present
- **Freshness Tracking**: How recent is the latest data
- **Background Checks**: Non-blocking quality analysis

### Gap Filling & Backfill

- **Gap Fill**: Fill missing candles (first 5 or all gaps)
- **Backfill**: Fetch historical data (months/years back)
- Both run in background with progress tracking

### API Integration

All components use Axios to communicate with the backend API:

- **GET** `/api/v1/connectors` - Fetch all connectors
- **POST** `/api/v1/connectors` - Create connector
- **PATCH** `/api/v1/connectors/:id/sandbox` - Toggle sandbox mode
- **DELETE** `/api/v1/connectors/:id` - Delete connector
- **GET** `/api/v1/jobs` - Fetch all jobs
- **POST** `/api/v1/jobs` - Create job
- **POST** `/api/v1/jobs/:id/pause` - Pause job
- **POST** `/api/v1/jobs/:id/resume` - Resume job
- **DELETE** `/api/v1/jobs/:id` - Delete job

### Development Workflow

1. Start MongoDB: `make docker-up` (from project root)
2. Start backend API: `make run` (from project root)
3. Start frontend dev server: `npm run dev` (from web/ directory)
4. Open browser to `http://localhost:3000`

## Environment Variables

The Vite dev server is configured to proxy `/api` requests to the backend.

To change the backend URL, edit `vite.config.js`:

```javascript
server: {
  proxy: {
    '/api': {
      target: 'http://your-backend-url:8080',
      changeOrigin: true
    }
  }
}
```

## Styling

This project uses Tailwind CSS for styling. Key utility classes:

- `bg-blue-500` - Background colors
- `text-gray-900` - Text colors
- `rounded-lg` - Border radius
- `shadow-md` - Box shadows
- `hover:bg-blue-600` - Hover states
- `transition` - Smooth transitions

To customize the design, edit `tailwind.config.js`.

## Components

### Dashboard.jsx
- Summary statistics (connectors, jobs, candles)
- Connector health monitoring (uptime, errors, response time)
- Data quality distribution chart
- Rate limit overview for all connectors

### ConnectorList.jsx
- Grid view of all connectors
- Health status synced with Dashboard API
- Health stats: Uptime, Error Rate, Response Time
- Rate limit visualization with cooldown timer
- Create/Edit/Delete connector actions

### JobList.jsx
- Table view of all jobs with sorting
- **Timeframe filter**: Filter by 1m, 5m, 15m, etc.
- **Status filter**: Active, Paused, Stopped
- **Candles column**: Shows total candle count
- Job details modal with charts
- Pause/Resume/Delete actions

### DataQuality.jsx
- Quality summary with status distribution
- Gap detection and completeness scores
- Background quality checks with progress
- Gap filling (fill first 5 or all gaps)
- Historical backfill feature
- Clickable symbols → job details

### CandlestickChart.jsx
- Professional candlestick visualization (lightweight-charts v5)
- Volume histogram with color coding
- Period selection: 1D, 1W, 1M, 3M, 6M, 1Y, All
- Zoom controls: In/Out/Reset buttons
- Mouse wheel zoom, drag to pan
- Indicator overlays and separate panes

### JobQueue.jsx
- Queue status per connector
- Pending/Running job counts
- Clickable symbols → job details
- Rate limit indicators

## Browser Support

- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)

## Troubleshooting

### API connection errors
- Ensure backend is running on port 8080
- Check browser console for CORS errors
- Verify MongoDB is running

### Build errors
- Delete `node_modules` and run `npm install` again
- Clear Vite cache: `rm -rf node_modules/.vite`

### Styling issues
- Ensure Tailwind CSS is properly configured
- Check that `@tailwind` directives are in `index.css`
- Rebuild: `npm run dev` or `npm run build`

## License

MIT
