# API Gateway Service

The API Gateway Service is the central entry point for all external requests to the Agba Voice AI Platform. It handles request routing, authentication, rate limiting, and provides a unified API interface for all platform services.

## Overview

The API Gateway acts as a reverse proxy and API management layer, providing:
- Centralized authentication and authorization
- Request routing to appropriate microservices
- Rate limiting and throttling
- API versioning and documentation
- Request/response transformation
- Monitoring and analytics

## Architecture

### Technology Stack
- **Language**: Go (Golang)
- **Framework**: Gin HTTP framework
- **Authentication**: JWT with OAuth2/OpenID Connect
- **Rate Limiting**: Redis-based token bucket
- **Documentation**: OpenAPI 3.0 (Swagger)
- **Monitoring**: Prometheus metrics

### Key Components
1. **Router**: Path-based request routing
2. **Auth Middleware**: JWT validation and user context
3. **Rate Limiter**: Redis-backed rate limiting
4. **Load Balancer**: Service discovery and load balancing
5. **Transformer**: Request/response transformation
6. **Logger**: Structured logging with correlation IDs

## Features

### Authentication & Authorization
- JWT token validation
- OAuth2/OpenID Connect integration
- Role-based access control (RBAC)
- API key authentication for service-to-service calls
- Multi-factor authentication support

### Request Routing
- Path-based routing to microservices
- Header-based routing for API versioning
- Weighted routing for canary deployments
- Circuit breaker pattern for fault tolerance
- Retry logic with exponential backoff

### Rate Limiting
- Per-user rate limiting
- Per-API endpoint rate limiting
- Burst capacity handling
- Distributed rate limiting across instances
- Custom rate limiting rules

### API Management
- OpenAPI 3.0 specification
- Interactive API documentation
- Request/response validation
- API versioning (v1, v2, etc.)
- Deprecation notices and migration guides

## Configuration

### Environment Variables
```bash
# Server Configuration
PORT=8080
HOST=0.0.0.0
ENV=production

# Database
POSTGRES_URL=postgresql://user:pass@localhost:5432/agba_core
REDIS_URL=redis://localhost:6379

# Authentication
JWT_SECRET=your-jwt-secret
OAUTH2_CLIENT_ID=your-oauth2-client-id
OAUTH2_CLIENT_SECRET=your-oauth2-client-secret

# Rate Limiting
RATE_LIMIT_REQUESTS_PER_MINUTE=60
RATE_LIMIT_BURST_SIZE=10

# Service Discovery
TELEPHONY_SERVICE_URL=http://telephony-gateway:8081
WEBRTC_SERVICE_URL=http://webrtc-service:8082
AI_PIPELINE_SERVICE_URL=http://ai-pipeline:8083
```

### Routing Configuration
```yaml
routes:
  - path: /api/v1/calls/*
    service: telephony-gateway
    methods: [GET, POST, PUT, DELETE]
    auth_required: true
    rate_limit: 100/minute
    
  - path: /api/v1/webrtc/*
    service: webrtc-service
    methods: [GET, POST]
    auth_required: true
    rate_limit: 50/minute
    
  - path: /api/v1/ai/*
    service: ai-pipeline
    methods: [POST]
    auth_required: true
    rate_limit: 30/minute
```

## API Endpoints

### Authentication
```
POST /auth/login          # User login
POST /auth/refresh        # Token refresh
POST /auth/logout         # User logout
GET  /auth/profile        # User profile
```

### Call Management
```
GET    /api/v1/calls           # List calls
POST   /api/v1/calls           # Create call
GET    /api/v1/calls/{id}      # Get call details
PUT    /api/v1/calls/{id}      # Update call
DELETE /api/v1/calls/{id}      # End call
```

### Agent Configuration
```
GET    /api/v1/agents          # List agents
POST   /api/v1/agents          # Create agent
GET    /api/v1/agents/{id}     # Get agent
PUT    /api/v1/agents/{id}     # Update agent
DELETE /api/v1/agents/{id}     # Delete agent
```

### Analytics
```
GET /api/v1/analytics/calls    # Call analytics
GET /api/v1/analytics/agents   # Agent performance
GET /api/v1/analytics/usage    # Usage statistics
```

## Security

### Authentication Flow
1. Client sends credentials to `/auth/login`
2. Gateway validates credentials with auth service
3. JWT token issued with user claims and permissions
4. Subsequent requests include JWT in Authorization header
5. Gateway validates JWT and extracts user context

### Rate Limiting
- Token bucket algorithm with Redis backend
- Per-user and per-endpoint limits
- Configurable burst capacity
- 429 status code for rate limit exceeded
- Retry-After header for client guidance

### Security Headers
```
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000
Content-Security-Policy: default-src 'self'
```

## Performance

### Optimization Strategies
- Connection pooling for database and Redis
- HTTP/2 support for improved performance
- Response compression (gzip)
- Caching of frequently accessed data
- Async processing for non-critical operations

### Monitoring Metrics
- Request latency (P50, P95, P99)
- Request rate (requests per second)
- Error rate (4xx, 5xx responses)
- Authentication success/failure rate
- Rate limiting hits
- Service health checks

### Performance Targets
- Response time: < 50ms (P95) for routing
- Throughput: 10,000+ requests per second
- Availability: 99.99% uptime
- Memory usage: < 2GB per instance
- CPU usage: < 70% under normal load

## Deployment

### Docker Configuration
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o api-gateway ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/api-gateway .
EXPOSE 8080
CMD ["./api-gateway"]
```

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-gateway
  template:
    metadata:
      labels:
        app: api-gateway
    spec:
      containers:
      - name: api-gateway
        image: agba/api-gateway:latest
        ports:
        - containerPort: 8080
        env:
        - name: PORT
          value: "8080"
        resources:
          requests:
            cpu: 500m
            memory: 1Gi
          limits:
            cpu: 2000m
            memory: 4Gi
```

## Development

### Local Setup
```bash
# Clone repository
git clone https://github.com/ibilola104/Agba.ai.git
cd Agba.ai/services/api-gateway

# Install dependencies
go mod download

# Set environment variables
export POSTGRES_URL=postgresql://localhost:5432/agba_core
export REDIS_URL=redis://localhost:6379

# Run service
go run cmd/main.go
```

### Testing
```bash
# Run unit tests
go test ./...

# Run integration tests
go test -tags=integration ./...

# Run load tests
go test -tags=load ./...

# Generate test coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Monitoring

### Health Checks
- `/health`: Basic health check
- `/health/ready`: Readiness probe
- `/health/live`: Liveness probe

### Metrics Endpoints
- `/metrics`: Prometheus metrics
- `/debug/pprof`: Go profiling endpoints

### Logging
- Structured JSON logging
- Correlation IDs for request tracing
- Log levels: DEBUG, INFO, WARN, ERROR
- Log rotation and retention policies

## Troubleshooting

### Common Issues
1. **High Latency**: Check database connections and Redis performance
2. **Authentication Failures**: Verify JWT secret and OAuth2 configuration
3. **Rate Limiting**: Review rate limit settings and Redis connectivity
4. **Service Discovery**: Check service URLs and network connectivity

### Debug Commands
```bash
# Check service status
curl http://localhost:8080/health

# Test authentication
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test"}'

# Check metrics
curl http://localhost:8080/metrics
```

## Contributing

1. Follow Go coding standards and best practices
2. Write comprehensive unit and integration tests
3. Update API documentation for any endpoint changes
4. Ensure security reviews for authentication changes
5. Performance test any routing or middleware changes

## Support

For API Gateway issues:
- Check logs for error details
- Verify configuration settings
- Test connectivity to downstream services
- Contact the backend development team