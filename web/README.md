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

### Connector Management
- Create, view, and delete connectors
- **Sandbox mode toggle switch** - Switch between testnet and production
- Rate limit monitoring
- Support for multiple exchanges (Binance, Bybit, Coinbase, Kraken, KuCoin)

### Job Management
- Create, pause, resume, and delete jobs
- Monitor job execution status
- View last run times and errors
- Support for multiple timeframes (1m, 5m, 15m, 30m, 1h, 4h, 1d)

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
│   │   ├── Dashboard.jsx       # Dashboard overview
│   │   ├── ConnectorList.jsx   # Connector management + sandbox toggle
│   │   └── JobList.jsx         # Job management
│   ├── App.jsx                 # Main app component
│   ├── main.jsx                # Entry point
│   └── index.css               # Tailwind styles
├── index.html                  # HTML template
├── vite.config.js              # Vite configuration
├── tailwind.config.js          # Tailwind configuration
└── package.json                # Dependencies
```

## Key Features

### Sandbox Mode Toggle

The connector management page includes a **visual toggle switch** for sandbox mode:

- **Yellow toggle** = Sandbox mode ON (using testnet)
- **Green toggle** = Sandbox mode OFF (using production)

Simply click the toggle to switch between modes. The change is saved immediately to the database.

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
- Displays summary statistics
- Shows active/sandbox connectors
- Shows active/paused jobs
- Quick overview of recent connectors

### ConnectorList.jsx
- Grid view of all connectors
- **Sandbox mode toggle switch** (key feature)
- Create new connector modal
- Delete connector action
- Rate limit display

### JobList.jsx
- Table view of all jobs
- Create new job modal
- Pause/resume job actions
- Delete job action
- Last run time and error display

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
