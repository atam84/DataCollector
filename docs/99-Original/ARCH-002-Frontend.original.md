# ARCH-002: Data Collector Frontend Architecture

## Document Information

| Field | Value |
|-------|-------|
| **Document ID** | ARCH-002 |
| **Version** | 1.0 |
| **Status** | Draft |
| **Author** | GoTrading Team |
| **Created** | 2025-01-17 |
| **Last Updated** | 2025-01-17 |

---

## 1. Overview

### 1.1 Purpose

This document describes the frontend architecture for the GoTrading Data Collector interface. The frontend provides a comprehensive UI for managing data connectors, exploring collected data, visualizing charts with indicators, and exporting data.

### 1.2 Technology Stack

| Component | Technology | Version | Justification |
|-----------|------------|---------|---------------|
| Framework | React | 18.x | Component-based, large ecosystem |
| Language | TypeScript | 5.x | Type safety, better DX |
| Build Tool | Vite | 5.x | Fast HMR, modern bundling |
| Styling | Tailwind CSS | 3.x | Utility-first, rapid development |
| State Management | Zustand | 4.x | Simple, performant, TypeScript-friendly |
| Data Fetching | TanStack Query | 5.x | Caching, background updates |
| Charting | Recharts / D3.js | Latest | Customizable financial charts |
| WebSocket | Native + reconnecting-websocket | - | Real-time updates |
| Routing | React Router | 6.x | Client-side routing |
| Forms | React Hook Form | 7.x | Performant form handling |
| Validation | Zod | 3.x | Schema validation |
| UI Components | Radix UI | Latest | Accessible primitives |
| Icons | Lucide React | Latest | Consistent icon set |

---

## 2. High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                        DATA COLLECTOR FRONTEND                                   │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                           PRESENTATION LAYER                             │   │
│  ├─────────────────────────────────────────────────────────────────────────┤   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │   │
│  │  │   Pages     │  │   Layouts   │  │  Components │  │   Charts    │    │   │
│  │  │             │  │             │  │             │  │             │    │   │
│  │  │ • Dashboard │  │ • MainLayout│  │ • Connector │  │ • Candle    │    │   │
│  │  │ • Connectors│  │ • AuthLayout│  │ • DataTable │  │ • Line      │    │   │
│  │  │ • Explorer  │  │             │  │ • StatusBadge│ │ • Volume    │    │   │
│  │  │ • Settings  │  │             │  │ • Modal     │  │ • Indicator │    │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                        │                                        │
│                                        ▼                                        │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                            FEATURE LAYER                                 │   │
│  ├─────────────────────────────────────────────────────────────────────────┤   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │   │
│  │  │  Connector  │  │   Exchange  │  │    Data     │  │   Export    │    │   │
│  │  │   Feature   │  │   Feature   │  │   Feature   │  │   Feature   │    │   │
│  │  │             │  │             │  │             │  │             │    │   │
│  │  │ • hooks     │  │ • hooks     │  │ • hooks     │  │ • hooks     │    │   │
│  │  │ • components│  │ • components│  │ • components│  │ • components│    │   │
│  │  │ • types     │  │ • types     │  │ • types     │  │ • types     │    │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                        │                                        │
│                                        ▼                                        │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                             CORE LAYER                                   │   │
│  ├─────────────────────────────────────────────────────────────────────────┤   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │   │
│  │  │    Store    │  │   API       │  │  WebSocket  │  │   Utils     │    │   │
│  │  │  (Zustand)  │  │  (TanStack) │  │   Client    │  │             │    │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                        │                                        │
│                                        ▼                                        │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                           EXTERNAL SERVICES                              │   │
│  ├─────────────────────────────────────────────────────────────────────────┤   │
│  │         ┌───────────────────────────────────────────────┐               │   │
│  │         │              Backend API (REST + WS)           │               │   │
│  │         └───────────────────────────────────────────────┘               │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Project Structure

```
data-collector-frontend/
├── public/
│   ├── favicon.ico
│   └── assets/
│       └── images/
├── src/
│   ├── app/
│   │   ├── App.tsx                    # Root component
│   │   ├── routes.tsx                 # Route definitions
│   │   └── providers.tsx              # Context providers
│   │
│   ├── pages/
│   │   ├── dashboard/
│   │   │   ├── index.tsx              # Dashboard page
│   │   │   └── components/
│   │   │       ├── StatsCards.tsx
│   │   │       ├── ActivityChart.tsx
│   │   │       ├── RecentJobs.tsx
│   │   │       └── ExchangeHealth.tsx
│   │   ├── connectors/
│   │   │   ├── index.tsx              # Connector list page
│   │   │   ├── [id].tsx               # Connector detail page
│   │   │   ├── create.tsx             # Create connector page
│   │   │   └── components/
│   │   │       ├── ConnectorCard.tsx
│   │   │       ├── ConnectorForm.tsx
│   │   │       ├── ConnectorStatus.tsx
│   │   │       └── PairSelector.tsx
│   │   ├── explorer/
│   │   │   ├── index.tsx              # Data explorer page
│   │   │   └── components/
│   │   │       ├── DataFilters.tsx
│   │   │       ├── OHLCVChart.tsx
│   │   │       ├── IndicatorOverlay.tsx
│   │   │       ├── DataTable.tsx
│   │   │       └── CoverageMap.tsx
│   │   ├── exports/
│   │   │   ├── index.tsx              # Exports page
│   │   │   └── components/
│   │   │       ├── ExportForm.tsx
│   │   │       └── ExportHistory.tsx
│   │   └── settings/
│   │       ├── index.tsx              # Settings page
│   │       └── components/
│   │           ├── IndicatorConfig.tsx
│   │           └── GeneralSettings.tsx
│   │
│   ├── features/
│   │   ├── connectors/
│   │   │   ├── api.ts                 # Connector API calls
│   │   │   ├── hooks.ts               # Connector hooks
│   │   │   ├── types.ts               # Connector types
│   │   │   └── store.ts               # Connector store slice
│   │   ├── exchanges/
│   │   │   ├── api.ts
│   │   │   ├── hooks.ts
│   │   │   └── types.ts
│   │   ├── data/
│   │   │   ├── api.ts
│   │   │   ├── hooks.ts
│   │   │   ├── types.ts
│   │   │   └── transforms.ts          # Data transformations
│   │   ├── indicators/
│   │   │   ├── api.ts
│   │   │   ├── hooks.ts
│   │   │   ├── types.ts
│   │   │   └── calculations.ts        # Client-side indicator calc
│   │   └── exports/
│   │       ├── api.ts
│   │       ├── hooks.ts
│   │       └── types.ts
│   │
│   ├── components/
│   │   ├── ui/                        # Base UI components
│   │   │   ├── Button.tsx
│   │   │   ├── Input.tsx
│   │   │   ├── Select.tsx
│   │   │   ├── Modal.tsx
│   │   │   ├── Dialog.tsx
│   │   │   ├── Toast.tsx
│   │   │   ├── Badge.tsx
│   │   │   ├── Card.tsx
│   │   │   ├── Table.tsx
│   │   │   ├── Tabs.tsx
│   │   │   ├── Dropdown.tsx
│   │   │   ├── Checkbox.tsx
│   │   │   ├── Switch.tsx
│   │   │   ├── Spinner.tsx
│   │   │   ├── Progress.tsx
│   │   │   ├── Skeleton.tsx
│   │   │   └── index.ts
│   │   ├── charts/                    # Chart components
│   │   │   ├── CandlestickChart.tsx
│   │   │   ├── LineChart.tsx
│   │   │   ├── VolumeChart.tsx
│   │   │   ├── AreaChart.tsx
│   │   │   ├── ChartContainer.tsx
│   │   │   ├── ChartTooltip.tsx
│   │   │   ├── ChartLegend.tsx
│   │   │   └── index.ts
│   │   ├── layout/                    # Layout components
│   │   │   ├── MainLayout.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   ├── Header.tsx
│   │   │   ├── Footer.tsx
│   │   │   └── index.ts
│   │   └── common/                    # Shared components
│   │       ├── DateRangePicker.tsx
│   │       ├── ExchangeLogo.tsx
│   │       ├── StatusBadge.tsx
│   │       ├── EmptyState.tsx
│   │       ├── ErrorBoundary.tsx
│   │       ├── LoadingOverlay.tsx
│   │       └── index.ts
│   │
│   ├── core/
│   │   ├── api/
│   │   │   ├── client.ts              # Axios instance
│   │   │   ├── endpoints.ts           # API endpoint constants
│   │   │   └── types.ts               # API response types
│   │   ├── websocket/
│   │   │   ├── client.ts              # WebSocket client
│   │   │   ├── hooks.ts               # WebSocket hooks
│   │   │   └── types.ts               # Message types
│   │   ├── store/
│   │   │   ├── index.ts               # Store setup
│   │   │   ├── slices/
│   │   │   │   ├── ui.ts              # UI state
│   │   │   │   └── realtime.ts        # Real-time data
│   │   │   └── middleware/
│   │   │       └── websocket.ts       # WS middleware
│   │   └── config/
│   │       └── index.ts               # App configuration
│   │
│   ├── hooks/
│   │   ├── useDebounce.ts
│   │   ├── useLocalStorage.ts
│   │   ├── useMediaQuery.ts
│   │   ├── useClickOutside.ts
│   │   └── index.ts
│   │
│   ├── utils/
│   │   ├── date.ts                    # Date utilities
│   │   ├── format.ts                  # Formatting utilities
│   │   ├── validation.ts              # Validation schemas
│   │   ├── cn.ts                      # Classname utility
│   │   └── index.ts
│   │
│   ├── types/
│   │   ├── global.d.ts
│   │   └── index.ts
│   │
│   ├── styles/
│   │   └── globals.css                # Global styles + Tailwind
│   │
│   ├── constants/
│   │   ├── timeframes.ts
│   │   ├── indicators.ts
│   │   └── index.ts
│   │
│   └── main.tsx                       # Entry point
│
├── tests/
│   ├── unit/
│   ├── integration/
│   └── e2e/
│
├── .env
├── .env.example
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
├── tailwind.config.js
├── postcss.config.js
└── README.md
```

---

## 4. Core Components

### 4.1 API Client

```typescript
// src/core/api/client.ts

import axios, { AxiosInstance, AxiosRequestConfig, AxiosError } from 'axios';
import { config } from '../config';

const createApiClient = (): AxiosInstance => {
  const instance = axios.create({
    baseURL: config.apiUrl,
    timeout: 30000,
    headers: {
      'Content-Type': 'application/json',
    },
  });

  // Request interceptor
  instance.interceptors.request.use(
    (config) => {
      // Add auth token if available (future use)
      const token = localStorage.getItem('token');
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      return config;
    },
    (error) => Promise.reject(error)
  );

  // Response interceptor
  instance.interceptors.response.use(
    (response) => response,
    (error: AxiosError) => {
      if (error.response) {
        switch (error.response.status) {
          case 401:
            // Handle unauthorized
            break;
          case 429:
            // Handle rate limit
            break;
          case 500:
            // Handle server error
            break;
        }
      }
      return Promise.reject(error);
    }
  );

  return instance;
};

export const apiClient = createApiClient();

// Typed request helpers
export const api = {
  get: <T>(url: string, config?: AxiosRequestConfig) =>
    apiClient.get<T>(url, config).then((res) => res.data),
  
  post: <T>(url: string, data?: unknown, config?: AxiosRequestConfig) =>
    apiClient.post<T>(url, data, config).then((res) => res.data),
  
  put: <T>(url: string, data?: unknown, config?: AxiosRequestConfig) =>
    apiClient.put<T>(url, data, config).then((res) => res.data),
  
  delete: <T>(url: string, config?: AxiosRequestConfig) =>
    apiClient.delete<T>(url, config).then((res) => res.data),
};
```

### 4.2 WebSocket Client

```typescript
// src/core/websocket/client.ts

import ReconnectingWebSocket from 'reconnecting-websocket';
import { config } from '../config';

export type MessageHandler = (data: unknown) => void;

interface Subscription {
  channel: string;
  params: Record<string, string>;
  handler: MessageHandler;
}

class WebSocketClient {
  private ws: ReconnectingWebSocket | null = null;
  private subscriptions: Map<string, Subscription> = new Map();
  private messageQueue: unknown[] = [];
  private isConnected = false;

  connect() {
    if (this.ws) return;

    this.ws = new ReconnectingWebSocket(config.wsUrl, [], {
      maxReconnectionDelay: 10000,
      minReconnectionDelay: 1000,
      reconnectionDelayGrowFactor: 1.3,
      maxRetries: Infinity,
    });

    this.ws.onopen = () => {
      this.isConnected = true;
      console.log('WebSocket connected');
      
      // Resubscribe on reconnect
      this.subscriptions.forEach((sub, key) => {
        this.sendSubscribe(sub);
      });
      
      // Send queued messages
      while (this.messageQueue.length > 0) {
        const msg = this.messageQueue.shift();
        this.send(msg);
      }
    };

    this.ws.onclose = () => {
      this.isConnected = false;
      console.log('WebSocket disconnected');
    };

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        this.handleMessage(data);
      } catch (err) {
        console.error('Failed to parse WebSocket message', err);
      }
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error', error);
    };
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
      this.isConnected = false;
    }
  }

  subscribe(
    channel: string,
    params: Record<string, string>,
    handler: MessageHandler
  ): () => void {
    const key = this.getSubscriptionKey(channel, params);
    
    const subscription: Subscription = { channel, params, handler };
    this.subscriptions.set(key, subscription);
    
    if (this.isConnected) {
      this.sendSubscribe(subscription);
    }
    
    // Return unsubscribe function
    return () => {
      this.subscriptions.delete(key);
      if (this.isConnected) {
        this.sendUnsubscribe(subscription);
      }
    };
  }

  private send(data: unknown) {
    if (this.isConnected && this.ws) {
      this.ws.send(JSON.stringify(data));
    } else {
      this.messageQueue.push(data);
    }
  }

  private sendSubscribe(sub: Subscription) {
    this.send({
      action: 'subscribe',
      channel: sub.channel,
      ...sub.params,
    });
  }

  private sendUnsubscribe(sub: Subscription) {
    this.send({
      action: 'unsubscribe',
      channel: sub.channel,
      ...sub.params,
    });
  }

  private handleMessage(data: unknown) {
    const msg = data as { channel: string; [key: string]: unknown };
    
    // Find matching subscriptions
    this.subscriptions.forEach((sub) => {
      if (this.matchesSubscription(msg, sub)) {
        sub.handler(msg);
      }
    });
  }

  private matchesSubscription(
    msg: Record<string, unknown>,
    sub: Subscription
  ): boolean {
    if (msg.channel !== sub.channel) return false;
    
    for (const [key, value] of Object.entries(sub.params)) {
      if (msg[key] !== value) return false;
    }
    
    return true;
  }

  private getSubscriptionKey(
    channel: string,
    params: Record<string, string>
  ): string {
    return `${channel}:${JSON.stringify(params)}`;
  }
}

export const wsClient = new WebSocketClient();
```

### 4.3 Global Store (Zustand)

```typescript
// src/core/store/index.ts

import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';

// UI State
interface UIState {
  sidebarOpen: boolean;
  theme: 'light' | 'dark' | 'system';
  toggleSidebar: () => void;
  setTheme: (theme: UIState['theme']) => void;
}

export const useUIStore = create<UIState>()(
  devtools(
    persist(
      (set) => ({
        sidebarOpen: true,
        theme: 'system',
        toggleSidebar: () =>
          set((state) => ({ sidebarOpen: !state.sidebarOpen })),
        setTheme: (theme) => set({ theme }),
      }),
      { name: 'ui-storage' }
    )
  )
);

// Real-time Data State
interface RealtimeState {
  ohlcvUpdates: Record<string, OHLCV>;
  statusUpdates: Record<string, ConnectorStatus>;
  updateOHLCV: (key: string, data: OHLCV) => void;
  updateStatus: (connectorId: string, status: ConnectorStatus) => void;
  clearUpdates: () => void;
}

export const useRealtimeStore = create<RealtimeState>()(
  immer((set) => ({
    ohlcvUpdates: {},
    statusUpdates: {},
    updateOHLCV: (key, data) =>
      set((state) => {
        state.ohlcvUpdates[key] = data;
      }),
    updateStatus: (connectorId, status) =>
      set((state) => {
        state.statusUpdates[connectorId] = status;
      }),
    clearUpdates: () =>
      set((state) => {
        state.ohlcvUpdates = {};
        state.statusUpdates = {};
      }),
  }))
);

// Filter State
interface FilterState {
  explorer: {
    exchange: string | null;
    pair: string | null;
    timeframe: string | null;
    dateRange: [Date | null, Date | null];
    indicators: string[];
  };
  setExplorerFilter: (filters: Partial<FilterState['explorer']>) => void;
  resetExplorerFilter: () => void;
}

const defaultExplorerFilter: FilterState['explorer'] = {
  exchange: null,
  pair: null,
  timeframe: null,
  dateRange: [null, null],
  indicators: [],
};

export const useFilterStore = create<FilterState>()(
  devtools((set) => ({
    explorer: defaultExplorerFilter,
    setExplorerFilter: (filters) =>
      set((state) => ({
        explorer: { ...state.explorer, ...filters },
      })),
    resetExplorerFilter: () =>
      set({ explorer: defaultExplorerFilter }),
  }))
);
```

---

## 5. Feature Modules

### 5.1 Connectors Feature

```typescript
// src/features/connectors/types.ts

export interface Connector {
  id: string;
  name: string;
  exchange: string;
  pairs: string[];
  timeframes: string[];
  schedule: string | null;
  status: ConnectorStatus;
  createdAt: string;
  updatedAt: string;
  stats?: ConnectorStats;
}

export type ConnectorStatus = 'created' | 'active' | 'paused' | 'error';

export interface ConnectorStats {
  totalCandles: number;
  lastSync: string | null;
  gapsCount: number;
  errorCount: number;
}

export interface CreateConnectorInput {
  name: string;
  exchange: string;
  pairs: string[];
  timeframes: string[];
  schedule?: string;
}

export interface UpdateConnectorInput extends Partial<CreateConnectorInput> {}
```

```typescript
// src/features/connectors/api.ts

import { api } from '@/core/api/client';
import {
  Connector,
  CreateConnectorInput,
  UpdateConnectorInput,
} from './types';

const BASE_URL = '/api/v1/connectors';

export const connectorApi = {
  list: () => api.get<Connector[]>(BASE_URL),
  
  get: (id: string) => api.get<Connector>(`${BASE_URL}/${id}`),
  
  create: (data: CreateConnectorInput) =>
    api.post<Connector>(BASE_URL, data),
  
  update: (id: string, data: UpdateConnectorInput) =>
    api.put<Connector>(`${BASE_URL}/${id}`, data),
  
  delete: (id: string) => api.delete(`${BASE_URL}/${id}`),
  
  start: (id: string) => api.post(`${BASE_URL}/${id}/start`),
  
  stop: (id: string) => api.post(`${BASE_URL}/${id}/stop`),
  
  refresh: (id: string) => api.post(`${BASE_URL}/${id}/refresh`),
};
```

```typescript
// src/features/connectors/hooks.ts

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { connectorApi } from './api';
import { CreateConnectorInput, UpdateConnectorInput } from './types';
import { toast } from '@/components/ui/Toast';

// Query keys
export const connectorKeys = {
  all: ['connectors'] as const,
  lists: () => [...connectorKeys.all, 'list'] as const,
  list: (filters: string) => [...connectorKeys.lists(), { filters }] as const,
  details: () => [...connectorKeys.all, 'detail'] as const,
  detail: (id: string) => [...connectorKeys.details(), id] as const,
};

// Hooks
export const useConnectors = () => {
  return useQuery({
    queryKey: connectorKeys.lists(),
    queryFn: connectorApi.list,
    staleTime: 30 * 1000, // 30 seconds
  });
};

export const useConnector = (id: string) => {
  return useQuery({
    queryKey: connectorKeys.detail(id),
    queryFn: () => connectorApi.get(id),
    enabled: !!id,
  });
};

export const useCreateConnector = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (data: CreateConnectorInput) => connectorApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: connectorKeys.lists() });
      toast.success('Connector created successfully');
    },
    onError: (error) => {
      toast.error('Failed to create connector');
    },
  });
};

export const useUpdateConnector = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateConnectorInput }) =>
      connectorApi.update(id, data),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: connectorKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: connectorKeys.lists() });
      toast.success('Connector updated successfully');
    },
  });
};

export const useDeleteConnector = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (id: string) => connectorApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: connectorKeys.lists() });
      toast.success('Connector deleted');
    },
  });
};

export const useStartConnector = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (id: string) => connectorApi.start(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: connectorKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: connectorKeys.lists() });
      toast.success('Connector started');
    },
  });
};

export const useStopConnector = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (id: string) => connectorApi.stop(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: connectorKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: connectorKeys.lists() });
      toast.success('Connector stopped');
    },
  });
};
```

### 5.2 Data Feature

```typescript
// src/features/data/types.ts

export interface OHLCV {
  timestamp: number;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

export interface OHLCVWithIndicators extends OHLCV {
  sma20?: number;
  sma50?: number;
  ema12?: number;
  ema26?: number;
  rsi?: number;
  macd?: number;
  macdSignal?: number;
  macdHistogram?: number;
  bbUpper?: number;
  bbMiddle?: number;
  bbLower?: number;
}

export interface DataQuery {
  exchange: string;
  pair: string;
  timeframe: string;
  start?: string;
  end?: string;
  limit?: number;
  indicators?: string[];
}

export interface DataCoverage {
  exchange: string;
  pair: string;
  timeframe: string;
  firstTimestamp: number;
  lastTimestamp: number;
  candleCount: number;
  gaps: Gap[];
}

export interface Gap {
  start: number;
  end: number;
  missingCandles: number;
}
```

```typescript
// src/features/data/hooks.ts

import { useQuery, useInfiniteQuery } from '@tanstack/react-query';
import { dataApi } from './api';
import { DataQuery, OHLCV, OHLCVWithIndicators } from './types';
import { useEffect } from 'react';
import { wsClient } from '@/core/websocket/client';
import { useRealtimeStore } from '@/core/store';

export const dataKeys = {
  all: ['data'] as const,
  ohlcv: (query: DataQuery) => [...dataKeys.all, 'ohlcv', query] as const,
  coverage: (exchange: string, pair: string, timeframe: string) =>
    [...dataKeys.all, 'coverage', exchange, pair, timeframe] as const,
  gaps: (exchange?: string) => [...dataKeys.all, 'gaps', exchange] as const,
};

export const useOHLCVData = (query: DataQuery, options?: { enabled?: boolean }) => {
  return useQuery({
    queryKey: dataKeys.ohlcv(query),
    queryFn: () => dataApi.queryOHLCV(query),
    enabled: options?.enabled !== false && !!query.exchange && !!query.pair && !!query.timeframe,
    staleTime: 60 * 1000, // 1 minute
  });
};

export const useInfiniteOHLCVData = (query: Omit<DataQuery, 'start' | 'limit'>) => {
  return useInfiniteQuery({
    queryKey: [...dataKeys.ohlcv(query as DataQuery), 'infinite'],
    queryFn: ({ pageParam }) =>
      dataApi.queryOHLCV({
        ...query,
        end: pageParam,
        limit: 1000,
      }),
    getNextPageParam: (lastPage) =>
      lastPage.length === 1000
        ? new Date(lastPage[0].timestamp).toISOString()
        : undefined,
    initialPageParam: undefined as string | undefined,
    enabled: !!query.exchange && !!query.pair && !!query.timeframe,
  });
};

export const useRealtimeOHLCV = (
  exchange: string,
  pair: string,
  timeframe: string,
  enabled = true
) => {
  const updateOHLCV = useRealtimeStore((state) => state.updateOHLCV);
  
  useEffect(() => {
    if (!enabled) return;
    
    const unsubscribe = wsClient.subscribe(
      'ohlcv',
      { exchange, symbol: pair, timeframe },
      (data) => {
        const key = `${exchange}:${pair}:${timeframe}`;
        updateOHLCV(key, data as OHLCV);
      }
    );
    
    return unsubscribe;
  }, [exchange, pair, timeframe, enabled, updateOHLCV]);
  
  const key = `${exchange}:${pair}:${timeframe}`;
  return useRealtimeStore((state) => state.ohlcvUpdates[key]);
};

export const useDataCoverage = (
  exchange: string,
  pair: string,
  timeframe: string
) => {
  return useQuery({
    queryKey: dataKeys.coverage(exchange, pair, timeframe),
    queryFn: () => dataApi.getCoverage(exchange, pair, timeframe),
    enabled: !!exchange && !!pair && !!timeframe,
  });
};
```

---

## 6. Chart Components

### 6.1 Candlestick Chart

```tsx
// src/components/charts/CandlestickChart.tsx

import React, { useMemo, useCallback } from 'react';
import {
  ComposedChart,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
  Bar,
  Line,
} from 'recharts';
import { OHLCVWithIndicators } from '@/features/data/types';
import { formatDate, formatPrice, formatVolume } from '@/utils/format';

interface CandlestickChartProps {
  data: OHLCVWithIndicators[];
  indicators?: {
    sma?: boolean;
    ema?: boolean;
    bollinger?: boolean;
  };
  height?: number;
  onCandleClick?: (candle: OHLCVWithIndicators) => void;
}

// Custom candlestick shape
const CandlestickShape = (props: any) => {
  const { x, y, width, payload } = props;
  const { open, high, low, close } = payload;
  
  const isGreen = close >= open;
  const color = isGreen ? '#22c55e' : '#ef4444';
  
  const candleWidth = Math.max(width * 0.8, 1);
  const wickWidth = 1;
  
  const bodyTop = Math.min(open, close);
  const bodyBottom = Math.max(open, close);
  const bodyHeight = Math.max(bodyBottom - bodyTop, 0.5);
  
  // Scale calculations would depend on your domain
  // This is simplified for illustration
  
  return (
    <g>
      {/* Wick */}
      <line
        x1={x + width / 2}
        y1={y - (high - bodyTop) * 10}
        x2={x + width / 2}
        y2={y - (low - bodyTop) * 10}
        stroke={color}
        strokeWidth={wickWidth}
      />
      {/* Body */}
      <rect
        x={x + (width - candleWidth) / 2}
        y={y}
        width={candleWidth}
        height={bodyHeight * 10}
        fill={isGreen ? color : color}
        stroke={color}
      />
    </g>
  );
};

export const CandlestickChart: React.FC<CandlestickChartProps> = ({
  data,
  indicators = {},
  height = 400,
  onCandleClick,
}) => {
  const chartData = useMemo(() => {
    return data.map((candle) => ({
      ...candle,
      date: new Date(candle.timestamp),
      // Calculate range for candlestick rendering
      range: [candle.low, candle.high],
      body: [Math.min(candle.open, candle.close), Math.max(candle.open, candle.close)],
      isGreen: candle.close >= candle.open,
    }));
  }, [data]);

  const [yDomain, volumeMax] = useMemo(() => {
    if (chartData.length === 0) return [[0, 100], 0];
    
    const prices = chartData.flatMap((d) => [d.high, d.low]);
    const min = Math.min(...prices);
    const max = Math.max(...prices);
    const padding = (max - min) * 0.1;
    
    const maxVol = Math.max(...chartData.map((d) => d.volume));
    
    return [[min - padding, max + padding], maxVol];
  }, [chartData]);

  const CustomTooltip = useCallback(({ active, payload }: any) => {
    if (!active || !payload?.length) return null;
    
    const candle = payload[0].payload;
    
    return (
      <div className="bg-white dark:bg-gray-800 p-3 rounded-lg shadow-lg border">
        <p className="text-sm text-gray-500 mb-2">
          {formatDate(candle.timestamp, 'full')}
        </p>
        <div className="grid grid-cols-2 gap-x-4 gap-y-1 text-sm">
          <span className="text-gray-500">Open:</span>
          <span className="font-mono">{formatPrice(candle.open)}</span>
          <span className="text-gray-500">High:</span>
          <span className="font-mono">{formatPrice(candle.high)}</span>
          <span className="text-gray-500">Low:</span>
          <span className="font-mono">{formatPrice(candle.low)}</span>
          <span className="text-gray-500">Close:</span>
          <span className={`font-mono ${candle.isGreen ? 'text-green-500' : 'text-red-500'}`}>
            {formatPrice(candle.close)}
          </span>
          <span className="text-gray-500">Volume:</span>
          <span className="font-mono">{formatVolume(candle.volume)}</span>
        </div>
        {candle.rsi && (
          <div className="mt-2 pt-2 border-t">
            <span className="text-gray-500 text-sm">RSI: </span>
            <span className="font-mono">{candle.rsi.toFixed(2)}</span>
          </div>
        )}
      </div>
    );
  }, []);

  return (
    <div className="w-full" style={{ height }}>
      <ResponsiveContainer width="100%" height="100%">
        <ComposedChart data={chartData} margin={{ top: 10, right: 30, left: 0, bottom: 0 }}>
          <XAxis
            dataKey="timestamp"
            tickFormatter={(ts) => formatDate(ts, 'short')}
            tick={{ fontSize: 12 }}
            tickLine={false}
            axisLine={false}
          />
          <YAxis
            domain={yDomain}
            tickFormatter={(val) => formatPrice(val)}
            tick={{ fontSize: 12 }}
            tickLine={false}
            axisLine={false}
            orientation="right"
          />
          <Tooltip content={<CustomTooltip />} />
          
          {/* Candlesticks rendered as bars */}
          <Bar
            dataKey="body"
            shape={<CandlestickShape />}
            onClick={(data) => onCandleClick?.(data)}
          />
          
          {/* SMA Lines */}
          {indicators.sma && (
            <>
              <Line
                type="monotone"
                dataKey="sma20"
                stroke="#3b82f6"
                dot={false}
                strokeWidth={1}
              />
              <Line
                type="monotone"
                dataKey="sma50"
                stroke="#8b5cf6"
                dot={false}
                strokeWidth={1}
              />
            </>
          )}
          
          {/* EMA Lines */}
          {indicators.ema && (
            <>
              <Line
                type="monotone"
                dataKey="ema12"
                stroke="#f59e0b"
                dot={false}
                strokeWidth={1}
              />
              <Line
                type="monotone"
                dataKey="ema26"
                stroke="#10b981"
                dot={false}
                strokeWidth={1}
              />
            </>
          )}
          
          {/* Bollinger Bands */}
          {indicators.bollinger && (
            <>
              <Line
                type="monotone"
                dataKey="bbUpper"
                stroke="#94a3b8"
                strokeDasharray="3 3"
                dot={false}
                strokeWidth={1}
              />
              <Line
                type="monotone"
                dataKey="bbMiddle"
                stroke="#64748b"
                dot={false}
                strokeWidth={1}
              />
              <Line
                type="monotone"
                dataKey="bbLower"
                stroke="#94a3b8"
                strokeDasharray="3 3"
                dot={false}
                strokeWidth={1}
              />
            </>
          )}
        </ComposedChart>
      </ResponsiveContainer>
    </div>
  );
};
```

### 6.2 Volume Chart

```tsx
// src/components/charts/VolumeChart.tsx

import React, { useMemo } from 'react';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  ResponsiveContainer,
  Cell,
  Tooltip,
} from 'recharts';
import { OHLCVWithIndicators } from '@/features/data/types';
import { formatDate, formatVolume } from '@/utils/format';

interface VolumeChartProps {
  data: OHLCVWithIndicators[];
  height?: number;
  syncId?: string;
}

export const VolumeChart: React.FC<VolumeChartProps> = ({
  data,
  height = 100,
  syncId,
}) => {
  const chartData = useMemo(() => {
    return data.map((candle) => ({
      timestamp: candle.timestamp,
      volume: candle.volume,
      isGreen: candle.close >= candle.open,
    }));
  }, [data]);

  return (
    <div className="w-full" style={{ height }}>
      <ResponsiveContainer width="100%" height="100%">
        <BarChart
          data={chartData}
          syncId={syncId}
          margin={{ top: 0, right: 30, left: 0, bottom: 0 }}
        >
          <XAxis
            dataKey="timestamp"
            tickFormatter={(ts) => formatDate(ts, 'short')}
            tick={{ fontSize: 10 }}
            tickLine={false}
            axisLine={false}
            hide
          />
          <YAxis
            tickFormatter={(val) => formatVolume(val)}
            tick={{ fontSize: 10 }}
            tickLine={false}
            axisLine={false}
            orientation="right"
            width={60}
          />
          <Tooltip
            formatter={(value: number) => [formatVolume(value), 'Volume']}
            labelFormatter={(ts) => formatDate(ts, 'full')}
          />
          <Bar dataKey="volume" maxBarSize={20}>
            {chartData.map((entry, index) => (
              <Cell
                key={`cell-${index}`}
                fill={entry.isGreen ? '#22c55e' : '#ef4444'}
                fillOpacity={0.7}
              />
            ))}
          </Bar>
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
};
```

---

## 7. Page Components

### 7.1 Dashboard Page

```tsx
// src/pages/dashboard/index.tsx

import React from 'react';
import { useConnectors } from '@/features/connectors/hooks';
import { useSystemStats } from '@/features/system/hooks';
import { StatsCards } from './components/StatsCards';
import { ActivityChart } from './components/ActivityChart';
import { RecentJobs } from './components/RecentJobs';
import { ExchangeHealth } from './components/ExchangeHealth';
import { Skeleton } from '@/components/ui/Skeleton';

export const DashboardPage: React.FC = () => {
  const { data: connectors, isLoading: connectorsLoading } = useConnectors();
  const { data: stats, isLoading: statsLoading } = useSystemStats();

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
          Dashboard
        </h1>
        <p className="text-gray-500 dark:text-gray-400">
          Overview of your data collection status
        </p>
      </div>

      {/* Stats Cards */}
      {statsLoading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {[...Array(4)].map((_, i) => (
            <Skeleton key={i} className="h-24" />
          ))}
        </div>
      ) : (
        <StatsCards
          activeConnectors={connectors?.filter((c) => c.status === 'active').length ?? 0}
          totalPairs={stats?.totalPairs ?? 0}
          dataPoints={stats?.totalCandles ?? 0}
          gapsDetected={stats?.gapsCount ?? 0}
        />
      )}

      {/* Activity Chart */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold mb-4">Collection Activity (24h)</h2>
        <ActivityChart />
      </div>

      {/* Two Column Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent Jobs */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-lg font-semibold">Recent Jobs</h2>
            <a
              href="/jobs"
              className="text-sm text-blue-600 hover:text-blue-700"
            >
              View All →
            </a>
          </div>
          <RecentJobs />
        </div>

        {/* Exchange Health */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold mb-4">Exchange Health</h2>
          <ExchangeHealth />
        </div>
      </div>
    </div>
  );
};
```

### 7.2 Connector Create Page

```tsx
// src/pages/connectors/create.tsx

import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useCreateConnector } from '@/features/connectors/hooks';
import { useExchanges, useExchangePairs, useExchangeTimeframes } from '@/features/exchanges/hooks';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Select } from '@/components/ui/Select';
import { MultiSelect } from '@/components/ui/MultiSelect';
import { Card } from '@/components/ui/Card';

const schema = z.object({
  name: z.string().min(1, 'Name is required').max(100),
  exchange: z.string().min(1, 'Exchange is required'),
  pairs: z.array(z.string()).min(1, 'At least one pair is required'),
  timeframes: z.array(z.string()).min(1, 'At least one timeframe is required'),
  schedule: z.string().optional(),
});

type FormData = z.infer<typeof schema>;

export const ConnectorCreatePage: React.FC = () => {
  const navigate = useNavigate();
  const createConnector = useCreateConnector();
  
  const {
    register,
    control,
    handleSubmit,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: {
      pairs: [],
      timeframes: [],
    },
  });

  const selectedExchange = watch('exchange');
  
  const { data: exchanges } = useExchanges();
  const { data: pairs } = useExchangePairs(selectedExchange);
  const { data: timeframes } = useExchangeTimeframes(selectedExchange);

  const onSubmit = async (data: FormData) => {
    try {
      await createConnector.mutateAsync(data);
      navigate('/connectors');
    } catch (error) {
      // Error handled by mutation
    }
  };

  return (
    <div className="max-w-2xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
          Create Connector
        </h1>
        <p className="text-gray-500 dark:text-gray-400">
          Set up a new data collection connector
        </p>
      </div>

      <Card>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6 p-6">
          {/* Name */}
          <div>
            <label className="block text-sm font-medium mb-2">
              Connector Name
            </label>
            <Input
              {...register('name')}
              placeholder="e.g., Binance Main Pairs"
              error={errors.name?.message}
            />
          </div>

          {/* Exchange */}
          <div>
            <label className="block text-sm font-medium mb-2">
              Exchange
            </label>
            <Controller
              name="exchange"
              control={control}
              render={({ field }) => (
                <Select
                  {...field}
                  options={exchanges?.map((e) => ({
                    value: e.id,
                    label: e.name,
                  })) ?? []}
                  placeholder="Select exchange"
                  error={errors.exchange?.message}
                />
              )}
            />
          </div>

          {/* Pairs */}
          <div>
            <label className="block text-sm font-medium mb-2">
              Trading Pairs
            </label>
            <Controller
              name="pairs"
              control={control}
              render={({ field }) => (
                <MultiSelect
                  {...field}
                  options={pairs?.map((p) => ({
                    value: p.symbol,
                    label: p.symbol,
                  })) ?? []}
                  placeholder="Select pairs"
                  disabled={!selectedExchange}
                  error={errors.pairs?.message}
                  searchable
                />
              )}
            />
            <p className="text-sm text-gray-500 mt-1">
              {pairs?.length ?? 0} pairs available
            </p>
          </div>

          {/* Timeframes */}
          <div>
            <label className="block text-sm font-medium mb-2">
              Timeframes
            </label>
            <Controller
              name="timeframes"
              control={control}
              render={({ field }) => (
                <MultiSelect
                  {...field}
                  options={timeframes?.map((tf) => ({
                    value: tf,
                    label: tf,
                  })) ?? []}
                  placeholder="Select timeframes"
                  disabled={!selectedExchange}
                  error={errors.timeframes?.message}
                />
              )}
            />
          </div>

          {/* Schedule (Optional) */}
          <div>
            <label className="block text-sm font-medium mb-2">
              Schedule (Optional)
            </label>
            <Input
              {...register('schedule')}
              placeholder="e.g., */5 * * * * (every 5 minutes)"
            />
            <p className="text-sm text-gray-500 mt-1">
              Leave empty for manual refresh only
            </p>
          </div>

          {/* Actions */}
          <div className="flex justify-end gap-3 pt-4 border-t">
            <Button
              type="button"
              variant="outline"
              onClick={() => navigate('/connectors')}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={isSubmitting}
            >
              Create Connector
            </Button>
          </div>
        </form>
      </Card>
    </div>
  );
};
```

### 7.3 Data Explorer Page

```tsx
// src/pages/explorer/index.tsx

import React, { useState } from 'react';
import { useFilterStore } from '@/core/store';
import { useOHLCVData, useDataCoverage, useRealtimeOHLCV } from '@/features/data/hooks';
import { useExchanges, useExchangePairs, useExchangeTimeframes } from '@/features/exchanges/hooks';
import { DataFilters } from './components/DataFilters';
import { CandlestickChart } from '@/components/charts/CandlestickChart';
import { VolumeChart } from '@/components/charts/VolumeChart';
import { CoverageInfo } from './components/CoverageInfo';
import { ExportModal } from './components/ExportModal';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Checkbox } from '@/components/ui/Checkbox';
import { Download, RefreshCw } from 'lucide-react';

export const ExplorerPage: React.FC = () => {
  const [showExportModal, setShowExportModal] = useState(false);
  const [indicators, setIndicators] = useState({
    sma: true,
    ema: false,
    bollinger: false,
  });
  
  const { explorer: filters, setExplorerFilter } = useFilterStore();
  
  const { data: exchanges } = useExchanges();
  const { data: pairs } = useExchangePairs(filters.exchange ?? '');
  const { data: timeframes } = useExchangeTimeframes(filters.exchange ?? '');
  
  const canQuery = filters.exchange && filters.pair && filters.timeframe;
  
  const { data: ohlcvData, isLoading, refetch } = useOHLCVData(
    {
      exchange: filters.exchange!,
      pair: filters.pair!,
      timeframe: filters.timeframe!,
      start: filters.dateRange[0]?.toISOString(),
      end: filters.dateRange[1]?.toISOString(),
      indicators: Object.entries(indicators)
        .filter(([_, enabled]) => enabled)
        .map(([name]) => name),
    },
    { enabled: canQuery }
  );
  
  const { data: coverage } = useDataCoverage(
    filters.exchange!,
    filters.pair!,
    filters.timeframe!
  );
  
  // Real-time updates
  const realtimeCandle = useRealtimeOHLCV(
    filters.exchange ?? '',
    filters.pair ?? '',
    filters.timeframe ?? '',
    canQuery
  );

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
            Data Explorer
          </h1>
          <p className="text-gray-500 dark:text-gray-400">
            Browse and visualize collected market data
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={() => refetch()}
            disabled={!canQuery}
          >
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </Button>
          <Button
            onClick={() => setShowExportModal(true)}
            disabled={!canQuery || !ohlcvData?.length}
          >
            <Download className="w-4 h-4 mr-2" />
            Export
          </Button>
        </div>
      </div>

      {/* Filters */}
      <Card className="p-4">
        <DataFilters
          exchanges={exchanges ?? []}
          pairs={pairs ?? []}
          timeframes={timeframes ?? []}
          filters={filters}
          onChange={setExplorerFilter}
        />
      </Card>

      {/* Charts */}
      {canQuery && (
        <Card className="p-4">
          {isLoading ? (
            <div className="h-[500px] flex items-center justify-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
            </div>
          ) : ohlcvData?.length ? (
            <div className="space-y-4">
              {/* Indicator Toggles */}
              <div className="flex gap-4 pb-4 border-b">
                <label className="flex items-center gap-2">
                  <Checkbox
                    checked={indicators.sma}
                    onCheckedChange={(checked) =>
                      setIndicators((prev) => ({ ...prev, sma: !!checked }))
                    }
                  />
                  <span className="text-sm">SMA (20, 50)</span>
                </label>
                <label className="flex items-center gap-2">
                  <Checkbox
                    checked={indicators.ema}
                    onCheckedChange={(checked) =>
                      setIndicators((prev) => ({ ...prev, ema: !!checked }))
                    }
                  />
                  <span className="text-sm">EMA (12, 26)</span>
                </label>
                <label className="flex items-center gap-2">
                  <Checkbox
                    checked={indicators.bollinger}
                    onCheckedChange={(checked) =>
                      setIndicators((prev) => ({ ...prev, bollinger: !!checked }))
                    }
                  />
                  <span className="text-sm">Bollinger Bands</span>
                </label>
              </div>

              {/* Price Chart */}
              <CandlestickChart
                data={ohlcvData}
                indicators={indicators}
                height={400}
              />

              {/* Volume Chart */}
              <VolumeChart
                data={ohlcvData}
                height={100}
                syncId="explorer-charts"
              />

              {/* Coverage Info */}
              {coverage && <CoverageInfo coverage={coverage} />}
            </div>
          ) : (
            <div className="h-[500px] flex items-center justify-center text-gray-500">
              No data available for the selected filters
            </div>
          )}
        </Card>
      )}

      {/* Export Modal */}
      <ExportModal
        open={showExportModal}
        onClose={() => setShowExportModal(false)}
        filters={filters}
      />
    </div>
  );
};
```

---

## 8. Configuration

### 8.1 Vite Configuration

```typescript
// vite.config.ts

import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom', 'react-router-dom'],
          charts: ['recharts', 'd3'],
          ui: ['@radix-ui/react-dialog', '@radix-ui/react-dropdown-menu'],
        },
      },
    },
  },
});
```

### 8.2 Tailwind Configuration

```javascript
// tailwind.config.js

/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#eff6ff',
          100: '#dbeafe',
          200: '#bfdbfe',
          300: '#93c5fd',
          400: '#60a5fa',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
          800: '#1e40af',
          900: '#1e3a8a',
        },
        success: '#22c55e',
        warning: '#f59e0b',
        danger: '#ef4444',
        candle: {
          green: '#22c55e',
          red: '#ef4444',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace'],
      },
    },
  },
  plugins: [],
};
```

### 8.3 Environment Variables

```bash
# .env.example

# API Configuration
VITE_API_URL=http://localhost:8080/api/v1
VITE_WS_URL=ws://localhost:8080/ws

# Feature Flags
VITE_ENABLE_REALTIME=true
VITE_ENABLE_EXPORTS=true

# Chart Configuration
VITE_DEFAULT_CANDLE_LIMIT=500
VITE_MAX_CANDLE_LIMIT=5000
```

---

## 9. Performance Optimizations

### 9.1 Chart Performance

```typescript
// src/utils/chartOptimizations.ts

import { OHLCVWithIndicators } from '@/features/data/types';

// Downsample data for large datasets
export function downsampleOHLCV(
  data: OHLCVWithIndicators[],
  maxPoints: number
): OHLCVWithIndicators[] {
  if (data.length <= maxPoints) return data;

  const step = Math.ceil(data.length / maxPoints);
  const result: OHLCVWithIndicators[] = [];

  for (let i = 0; i < data.length; i += step) {
    const chunk = data.slice(i, Math.min(i + step, data.length));
    
    // Aggregate chunk into single candle
    result.push({
      timestamp: chunk[0].timestamp,
      open: chunk[0].open,
      high: Math.max(...chunk.map((c) => c.high)),
      low: Math.min(...chunk.map((c) => c.low)),
      close: chunk[chunk.length - 1].close,
      volume: chunk.reduce((sum, c) => sum + c.volume, 0),
      // Average indicators
      ...(chunk[0].sma20 && {
        sma20: chunk.reduce((sum, c) => sum + (c.sma20 ?? 0), 0) / chunk.length,
      }),
    });
  }

  return result;
}

// Virtual rendering for large datasets
export function getVisibleRange(
  dataLength: number,
  viewportStart: number,
  viewportEnd: number,
  buffer = 50
): [number, number] {
  const start = Math.max(0, viewportStart - buffer);
  const end = Math.min(dataLength, viewportEnd + buffer);
  return [start, end];
}
```

### 9.2 Memoization

```typescript
// src/hooks/useMemoizedChartData.ts

import { useMemo } from 'react';
import { OHLCVWithIndicators } from '@/features/data/types';
import { downsampleOHLCV } from '@/utils/chartOptimizations';

export function useMemoizedChartData(
  data: OHLCVWithIndicators[] | undefined,
  maxPoints = 1000
) {
  return useMemo(() => {
    if (!data) return [];
    return downsampleOHLCV(data, maxPoints);
  }, [data, maxPoints]);
}
```

---

## 10. Testing Strategy

### 10.1 Unit Tests

```typescript
// tests/unit/features/connectors/hooks.test.ts

import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useConnectors, useCreateConnector } from '@/features/connectors/hooks';
import { connectorApi } from '@/features/connectors/api';

jest.mock('@/features/connectors/api');

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
};

describe('useConnectors', () => {
  it('should fetch connectors', async () => {
    const mockConnectors = [
      { id: '1', name: 'Test Connector', exchange: 'binance' },
    ];
    (connectorApi.list as jest.Mock).mockResolvedValue(mockConnectors);

    const { result } = renderHook(() => useConnectors(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockConnectors);
  });
});
```

### 10.2 E2E Tests

```typescript
// tests/e2e/connectors.spec.ts

import { test, expect } from '@playwright/test';

test.describe('Connectors', () => {
  test('should create a new connector', async ({ page }) => {
    await page.goto('/connectors/create');

    await page.fill('input[name="name"]', 'Test Connector');
    await page.selectOption('select[name="exchange"]', 'binance');
    
    await page.click('[data-testid="pairs-select"]');
    await page.click('[data-value="BTC/USDT"]');
    
    await page.click('[data-testid="timeframes-select"]');
    await page.click('[data-value="1h"]');
    
    await page.click('button[type="submit"]');

    await expect(page).toHaveURL('/connectors');
    await expect(page.locator('text=Test Connector')).toBeVisible();
  });
});
```

---

*End of Document*
