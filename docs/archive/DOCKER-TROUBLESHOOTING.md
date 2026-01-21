# Docker Build Troubleshooting - Network Issues

## Problem
Cannot connect to Docker Hub (`registry-1.docker.io`) when building images.

**Error**: `failed to solve: rpc error: code = Unknown desc = failed to solve with frontend dockerfile.v0: failed to create LLB definition: failed to do request: Head "https://registry-1.docker.io/v2/library/alpine/manifests/latest": dial tcp: lookup registry-1.docker.io: i/o timeout`

---

## Solutions

### Solution 1: Configure Docker DNS (Recommended)

Docker might be using a DNS server that can't resolve Docker Hub properly. Configure it to use Google DNS:

**1. Create or edit Docker daemon configuration:**
```bash
sudo mkdir -p /etc/docker
sudo nano /etc/docker/daemon.json
```

**2. Add this configuration:**
```json
{
  "dns": ["8.8.8.8", "8.8.4.4"],
  "registry-mirrors": ["https://mirror.gcr.io"]
}
```

**3. Restart Docker:**
```bash
sudo systemctl restart docker
```

**4. Try building again:**
```bash
docker compose build
```

---

### Solution 2: Use Docker Registry Mirror

If Docker Hub is slow or blocked in your region, use a mirror:

**For China/Asia regions:**
```json
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com"
  ]
}
```

**For Middle East/Africa regions:**
```json
{
  "registry-mirrors": [
    "https://mirror.gcr.io"
  ]
}
```

Apply the same steps as Solution 1.

---

### Solution 3: Use VPN or Proxy

If Docker Hub is blocked, use a VPN or configure Docker to use a proxy:

**1. Edit Docker daemon configuration:**
```json
{
  "proxies": {
    "http-proxy": "http://your-proxy:port",
    "https-proxy": "http://your-proxy:port",
    "no-proxy": "localhost,127.0.0.1"
  }
}
```

**2. Restart Docker and try again.**

---

### Solution 4: Pull Images Manually First

Try pulling the base images manually to test connectivity:

```bash
# Try pulling Alpine (used by backend)
docker pull alpine:latest

# Try pulling Golang (used by backend builder)
docker pull golang:1.21-alpine

# Try pulling Node (used by frontend builder)
docker pull node:18-alpine

# Try pulling Nginx (used by frontend server)
docker pull nginx:alpine

# Try pulling MongoDB (used by database)
docker pull mongo:latest
```

If these work, then try the build again:
```bash
docker compose build
```

---

### Solution 5: Build Without Docker (Alternative)

If Docker Hub remains inaccessible, run the application without Docker:

**Backend:**
```bash
# Install Go dependencies
go mod download

# Run the API
go run cmd/api/main.go
```

**Frontend:**
```bash
cd web
npm install
npm run dev
```

**MongoDB:**
```bash
# Use local MongoDB installation or download from mongodb.com
sudo systemctl start mongod
```

---

### Solution 6: Check Firewall

Your firewall might be blocking Docker Hub:

```bash
# Temporarily disable firewall to test (not recommended for production)
sudo ufw disable

# Try build
docker compose build

# Re-enable firewall
sudo ufw enable
```

If this works, you need to allow Docker Hub in your firewall:
```bash
# Allow Docker Hub IPs
sudo ufw allow out to 44.195.163.50
sudo ufw allow out to 34.193.215.148
```

---

### Solution 7: Use Alternative Base Images

Edit Dockerfiles to use alternative registries:

**For `Dockerfile` (Backend):**
```dockerfile
# Instead of: FROM golang:1.21-alpine
FROM mirror.gcr.io/golang:1.21-alpine

# Instead of: FROM alpine:latest
FROM mirror.gcr.io/alpine:latest
```

**For `web/Dockerfile` (Frontend):**
```dockerfile
# Instead of: FROM node:18-alpine
FROM mirror.gcr.io/node:18-alpine

# Instead of: FROM nginx:alpine
FROM mirror.gcr.io/nginx:alpine
```

---

## Quick Test Script

Run this to diagnose the issue:

```bash
#!/bin/bash
echo "Testing Docker connectivity..."
echo ""

echo "1. Testing internet connectivity..."
ping -c 2 8.8.8.8

echo ""
echo "2. Testing DNS resolution..."
nslookup registry-1.docker.io

echo ""
echo "3. Testing Docker Hub connectivity..."
ping -c 2 registry-1.docker.io

echo ""
echo "4. Testing Docker daemon status..."
sudo systemctl status docker

echo ""
echo "5. Checking Docker configuration..."
cat /etc/docker/daemon.json 2>/dev/null || echo "No daemon.json found"

echo ""
echo "6. Testing image pull..."
timeout 30 docker pull alpine:latest
```

---

## Recommended Immediate Solution

Based on your network location, try this:

```bash
# 1. Configure Docker DNS
sudo mkdir -p /etc/docker
echo '{
  "dns": ["8.8.8.8", "8.8.4.4", "1.1.1.1"],
  "registry-mirrors": ["https://mirror.gcr.io"]
}' | sudo tee /etc/docker/daemon.json

# 2. Restart Docker
sudo systemctl restart docker

# 3. Wait a few seconds
sleep 5

# 4. Test pull
docker pull alpine:latest

# 5. If successful, build
docker compose build
```

---

## If Nothing Works: Manual Deployment

See [QUICKSTART.md](QUICKSTART.md) Option B for manual installation without Docker.

---

## Support

If you continue having issues:
1. Check your network/firewall configuration
2. Contact your network administrator
3. Use a VPN to access Docker Hub
4. Use the manual installation method (no Docker required)
