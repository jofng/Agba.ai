# Multi-stage build for FreeSWITCH Service
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o freeswitch-service ./cmd/main.go

# Final stage with FreeSWITCH
FROM signalwire/freeswitch:latest

# Install additional packages
USER root
RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Create app user
RUN useradd -r -s /bin/false agba

# Create necessary directories
RUN mkdir -p /app/configs /app/logs /usr/local/freeswitch/recordings \
    && chown -R agba:agba /app /usr/local/freeswitch/recordings

# Copy the binary from builder stage
COPY --from=builder /app/freeswitch-service /app/freeswitch-service

# Copy configuration files
COPY configs/ /app/configs/
COPY configs/freeswitch/ /usr/local/freeswitch/conf/

# Set permissions
RUN chmod +x /app/freeswitch-service \
    && chown -R agba:agba /app

# Create startup script
RUN cat > /app/start.sh << 'EOF'
#!/bin/bash
set -e

# Start FreeSWITCH in background
echo "Starting FreeSWITCH..."
/usr/local/freeswitch/bin/freeswitch -nc -nf &
FREESWITCH_PID=$!

# Wait for FreeSWITCH to start
sleep 10

# Start the Go service
echo "Starting FreeSWITCH Service..."
exec /app/freeswitch-service
EOF

RUN chmod +x /app/start.sh && chown agba:agba /app/start.sh

# Switch to app user
USER agba

# Set working directory
WORKDIR /app

# Expose ports
EXPOSE 8080 5060/udp 5080/udp 8021 16384-32768/udp

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Set environment variables
ENV GIN_MODE=release
ENV PORT=8080

# Start the application
CMD ["/app/start.sh"]