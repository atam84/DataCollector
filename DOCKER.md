# Data Collector - Docker Deployment Guide

Complete guide for running the Data Collector application using Docker and Docker Compose.

---

## Prerequisites

- Docker 20.10+
- Docker Compose 2.0+

Check your installation:
```bash
docker --version
docker-compose --version
```

---

## Quick Start

### Production Deployment (Recommended)

Build and start all services:
```bash
# Build all images
make docker-compose-build

# Start all services (detached mode)
make docker-compose-up

# Or manually
docker-compose up -d
```

**Access the application**:
- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **MongoDB**: localhost:27017

### Development Mode (with hot reload)

For active development with code changes reflected immediately:
```bash
# Start in development mode (attached mode with logs)
make docker-compose-dev

# Or manually
docker-compose -f docker-compose.dev.yml up
```

---

## Services Overview

### 1. MongoDB (`mongodb`)
- **Image**: `mongo:latest`
- **Port**: `27017`
- **Container**: `datacollector-mongodb`
- **Volumes**:
  - `mongodb_data` - Database files
  - `mongodb_config` - Configuration files
- **Health Check**: Ping database every 10s

### 2. Backend API (`api`)
- **Build**: `./Dockerfile`
- **Port**: `8080`
- **Container**: `datacollector-api`
- **Language**: Go 1.21
- **Framework**: Fiber v2
- **Dependencies**: MongoDB
- **Health Check**: HTTP GET /health every 10s

### 3. Frontend Web (`web`)
- **Build**: `./web/Dockerfile`
- **Port**: `3000` (mapped to 80 in container)
- **Container**: `datacollector-web`
- **Framework**: React 18 + Vite
- **Server**: Nginx
- **Dependencies**: Backend API
- **Health Check**: HTTP GET / every 10s

---

## Docker Compose Files

### `docker-compose.yml` (Production)
Multi-stage builds with optimized images:
- Go binary (Alpine-based, ~20MB)
- React static files served by Nginx
- Persistent MongoDB volumes
- Health checks for all services
- Automatic service dependencies

### `docker-compose.dev.yml` (Development)
Development-optimized setup:
- Volume mounts for hot reload
- Go application runs with `go run`
- Vite dev server with HMR
- Source code changes reflected immediately
- Faster iteration cycles

---

## Common Commands

### Starting Services

```bash
# Production mode (detached)
docker-compose up -d

# Development mode (with logs)
docker-compose -f docker-compose.dev.yml up

# Build and start
docker-compose up -d --build

# Start specific service
docker-compose up -d api
```

### Stopping Services

```bash
# Stop all services
docker-compose down

# Stop and remove volumes
docker-compose down -v

# Stop specific service
docker-compose stop api
```

### Viewing Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f api
docker-compose logs -f web
docker-compose logs -f mongodb

# Last 100 lines
docker-compose logs --tail=100 api
```

### Rebuilding Images

```bash
# Rebuild all images
docker-compose build

# Rebuild specific service
docker-compose build api

# Rebuild without cache
docker-compose build --no-cache

# Rebuild and restart
docker-compose up -d --build
```

### Checking Status

```bash
# List running containers
docker-compose ps

# View resource usage
docker-compose stats

# Inspect service configuration
docker-compose config
```

### Executing Commands

```bash
# Access backend container shell
docker-compose exec api sh

# Access MongoDB shell
docker-compose exec mongodb mongosh datacollector

# Access frontend container
docker-compose exec web sh

# Run Go tests
docker-compose exec api go test ./...
```

---

## Environment Variables

Edit `docker-compose.yml` to configure:

### Backend API
```yaml
environment:
  SERVER_PORT: 8080
  SERVER_HOST: 0.0.0.0
  MONGODB_URI: mongodb://mongodb:27017
  MONGODB_DATABASE: datacollector
  EXCHANGE_SANDBOX_MODE: "true"    # Default to sandbox for safety
  EXCHANGE_ENABLE_RATE_LIMIT: "true"
  EXCHANGE_REQUEST_TIMEOUT: "30000"
```

### MongoDB
```yaml
environment:
  MONGO_INITDB_DATABASE: datacollector
```

---

## Data Persistence

### Named Volumes

MongoDB data is persisted in Docker volumes:
```yaml
volumes:
  mongodb_data:       # Database files
  mongodb_config:     # MongoDB configuration
```

**View volumes**:
```bash
docker volume ls | grep datacollector
```

**Inspect volume**:
```bash
docker volume inspect datacollector_mongodb_data
```

**Backup MongoDB data**:
```bash
# Export database
docker-compose exec mongodb mongodump --out /data/backup

# Copy to host
docker cp datacollector-mongodb:/data/backup ./backup
```

**Restore MongoDB data**:
```bash
# Copy backup to container
docker cp ./backup datacollector-mongodb:/data/restore

# Restore database
docker-compose exec mongodb mongorestore /data/restore
```

---

## Networking

### Internal Network

Services communicate via `datacollector-network`:
- `api` connects to `mongodb:27017`
- `web` (nginx) proxies to `api:8080`

**Inspect network**:
```bash
docker network inspect datacollector_datacollector-network
```

### Port Mapping

| Service  | Container Port | Host Port | URL                     |
|----------|----------------|-----------|-------------------------|
| Frontend | 80             | 3000      | http://localhost:3000   |
| Backend  | 8080           | 8080      | http://localhost:8080   |
| MongoDB  | 27017          | 27017     | mongodb://localhost:27017 |

---

## Health Checks

All services have health checks configured:

### Backend API
```yaml
healthcheck:
  test: wget --quiet --tries=1 --spider http://localhost:8080/health
  interval: 10s
  timeout: 5s
  retries: 5
```

### Frontend
```yaml
healthcheck:
  test: wget --quiet --tries=1 --spider http://localhost:80
  interval: 10s
  timeout: 5s
  retries: 3
```

### MongoDB
```yaml
healthcheck:
  test: echo 'db.runCommand("ping").ok' | mongosh localhost:27017/datacollector --quiet
  interval: 10s
  timeout: 5s
  retries: 5
```

**Check health status**:
```bash
docker-compose ps
```

---

## Troubleshooting

### Service won't start

**Check logs**:
```bash
docker-compose logs api
```

**Check if port is already in use**:
```bash
# Linux/Mac
lsof -i :8080
lsof -i :3000

# Or stop conflicting container
docker ps
docker stop <container_id>
```

### Cannot connect to MongoDB

**Verify MongoDB is running**:
```bash
docker-compose ps mongodb
```

**Test connection**:
```bash
docker-compose exec mongodb mongosh datacollector --eval "db.runCommand({ ping: 1 })"
```

**Check network**:
```bash
docker-compose exec api ping mongodb
```

### Frontend can't reach backend

**Check nginx configuration**:
```bash
docker-compose exec web cat /etc/nginx/conf.d/default.conf
```

**Test API from frontend container**:
```bash
docker-compose exec web wget -O- http://api:8080/health
```

### Build failures

**Clean build cache**:
```bash
docker-compose build --no-cache
```

**Remove old images**:
```bash
docker image prune -f
```

**Remove all project containers and volumes**:
```bash
make docker-compose-clean
# Or
docker-compose down -v --rmi all
```

### Permission issues

**Fix volume permissions** (Linux):
```bash
sudo chown -R $USER:$USER .
```

---

## Production Deployment

### Building for Production

```bash
# Build optimized images
docker-compose build

# Tag images for registry
docker tag datacollector-api:latest your-registry/datacollector-api:latest
docker tag datacollector-web:latest your-registry/datacollector-web:latest

# Push to registry
docker push your-registry/datacollector-api:latest
docker push your-registry/datacollector-web:latest
```

### Security Best Practices

1. **Use environment files** instead of hardcoding secrets:
   ```bash
   # Create .env file
   echo "MONGODB_URI=mongodb://user:pass@mongodb:27017" > .env

   # Reference in docker-compose.yml
   env_file:
     - .env
   ```

2. **Enable MongoDB authentication**:
   ```yaml
   environment:
     MONGO_INITDB_ROOT_USERNAME: admin
     MONGO_INITDB_ROOT_PASSWORD: ${MONGODB_PASSWORD}
   ```

3. **Use specific image tags** (not `latest`):
   ```yaml
   image: mongo:7.0.5
   ```

4. **Limit container resources**:
   ```yaml
   deploy:
     resources:
       limits:
         cpus: '0.5'
         memory: 512M
   ```

### Running Behind Reverse Proxy

If using Traefik, Nginx, or Caddy as reverse proxy:

```yaml
services:
  web:
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.datacollector.rule=Host(`datacollector.example.com`)"
      - "traefik.http.services.datacollector.loadbalancer.server.port=80"
```

---

## Performance Optimization

### Build Cache

Speed up builds by optimizing Dockerfile layer order:
- Dependencies first (cached layer)
- Source code last (changes frequently)

### Multi-stage Builds

Both Dockerfiles use multi-stage builds:
- **Builder stage**: Compiles code with all dev dependencies
- **Runtime stage**: Minimal Alpine image with only the binary/static files

**Result**:
- Backend: ~20MB (vs ~500MB with full Go image)
- Frontend: ~25MB (vs ~200MB with Node image)

### Resource Limits

Set appropriate limits in production:
```yaml
services:
  api:
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M
```

---

## Monitoring

### View Resource Usage

```bash
# Real-time stats
docker-compose stats

# Memory usage
docker-compose exec api free -h

# Disk usage
docker system df
```

### Logs Management

**Limit log size** in docker-compose.yml:
```yaml
services:
  api:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

---

## Development Workflow

### Hot Reload Setup

1. **Start development environment**:
   ```bash
   docker-compose -f docker-compose.dev.yml up
   ```

2. **Edit code** - changes are automatically reflected:
   - Backend: Go restarts on file change
   - Frontend: Vite HMR updates browser

3. **Access logs** in real-time:
   ```bash
   docker-compose -f docker-compose.dev.yml logs -f
   ```

### Testing in Container

```bash
# Run backend tests
docker-compose exec api go test -v ./...

# Run with coverage
docker-compose exec api go test -cover ./...

# Format code
docker-compose exec api go fmt ./...
```

---

## Makefile Commands Summary

```bash
# Production
make docker-compose-build      # Build all images
make docker-compose-up          # Start all services (detached)
make docker-compose-down        # Stop all services
make docker-compose-logs        # View logs
make docker-compose-restart     # Restart services
make docker-compose-clean       # Remove everything

# Development
make docker-compose-dev         # Start in dev mode
```

---

## Architecture Diagram

```
┌─────────────────────────────────────────┐
│  Browser (localhost:3000)               │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│  Frontend Container (Nginx)             │
│  - Serves React app                     │
│  - Proxies /api/* to backend            │
│  Port: 80 → 3000                        │
└────────────────┬────────────────────────┘
                 │ /api/* requests
                 ▼
┌─────────────────────────────────────────┐
│  Backend Container (Go + Fiber)         │
│  - REST API endpoints                   │
│  - Business logic                       │
│  Port: 8080                             │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│  MongoDB Container                      │
│  - Database                             │
│  - Persistent volumes                   │
│  Port: 27017                            │
└─────────────────────────────────────────┘
```

---

## Summary

Docker Compose provides a complete containerized environment for the Data Collector application:

✅ **One-command deployment**: `docker-compose up -d`
✅ **Isolated services**: Backend, Frontend, Database
✅ **Data persistence**: MongoDB volumes
✅ **Health checks**: Automatic service monitoring
✅ **Development mode**: Hot reload support
✅ **Production ready**: Optimized multi-stage builds
✅ **Easy scaling**: Docker Swarm or Kubernetes ready

**Next Steps**:
- Deploy to production with CI/CD
- Add SSL/TLS with Let's Encrypt
- Configure monitoring (Prometheus + Grafana)
- Set up automated backups

---

**Questions or issues?** Check the logs: `docker-compose logs -f`
