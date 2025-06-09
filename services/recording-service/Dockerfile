# Multi-stage build for Recording Service
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
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o recording-service ./cmd/main.go

# Final stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl ffmpeg

# Create non-root user
RUN addgroup -g 1000 agba && \
    adduser -D -s /bin/sh -u 1000 -G agba agba

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/recording-service .

# Copy configuration files
COPY --from=builder /app/configs ./configs

# Create necessary directories
RUN mkdir -p recordings temp models && \
    chown -R agba:agba recordings temp models

# Change ownership
RUN chown -R agba:agba /app

# Switch to non-root user
USER agba

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run the application
CMD ["./recording-service"]