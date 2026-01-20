# Backend Dockerfile - Multi-stage build for Go API

# Stage 1: Build
FROM golang:1.23-alpine AS builder

# Set GOTOOLCHAIN to auto to allow downloading newer Go versions if needed
ENV GOTOOLCHAIN=auto

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies and install the required Go toolchain
RUN go mod download && go install golang.org/dl/go1.24.12@latest && go1.24.12 download

# Copy source code
COPY . .

# Update go.mod and go.sum for the new Go version
RUN go1.24.12 mod tidy

# Build the application using the correct Go version
RUN CGO_ENABLED=0 GOOS=linux go1.24.12 build -a -installsuffix cgo -o /app/bin/api ./cmd/api

# Stage 2: Runtime
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create app directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/api /app/api

# Copy .env.example as template (will be overridden by docker-compose)
COPY .env.example /app/.env.example

# Expose port
EXPOSE 8080

# Run the application
CMD ["/app/api"]
