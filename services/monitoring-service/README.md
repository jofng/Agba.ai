# Monitoring Service

The Monitoring Service provides comprehensive monitoring, alerting, and observability for the Agba Voice AI Platform. It aggregates metrics, manages alerts, monitors service health, and provides real-time insights into system performance.

## Overview

The Monitoring Service offers centralized monitoring capabilities with:
- **Real-time metrics collection** from all platform services
- **Intelligent alerting system** with multiple notification channels
- **Service health monitoring** with dependency tracking
- **Custom dashboards** for visualization and analysis
- **Incident management** for tracking and resolving issues
- **Prometheus integration** for metrics storage and querying
- **Kubernetes service discovery** for automatic monitoring setup

## Architecture

### Technology Stack
- **Language**: Go (Golang) 1.21
- **Framework**: Gin HTTP framework for REST API
- **Database**: PostgreSQL for persistent storage of alerts, incidents, and metadata
- **Cache**: Redis for metrics caching and real-time data
- **Messaging**: NATS for real-time notifications and event streaming
- **Metrics**: Prometheus for metrics collection and querying
- **Service Discovery**: Kubernetes API for automatic service discovery
- **Monitoring**: Comprehensive observability with structured logging

### Key Components
1. **Metrics Collector**: Collects and aggregates metrics from all services
2. **Alert Manager**: Processes alert rules and manages alert lifecycle
3. **Health Monitor**: Monitors service health and dependencies
4. **Notification Engine**: Sends alerts via multiple channels (Slack, email, webhooks)
5. **Dashboard Manager**: Creates and manages monitoring dashboards
6. **Incident Tracker**: Tracks and manages system incidents
7. **Service Discovery**: Automatically discovers and monitors new services

## Features

### Metrics Collection & Aggregation
- **Real-time metrics collection** from all platform services
- **Custom metrics support** for application-specific monitoring
- **Metrics aggregation** with configurable retention policies
- **High-resolution metrics** for detailed performance analysis
- **Batch processing** for efficient metrics storage
- **Automatic service discovery** via Kubernetes API

### Intelligent Alerting System
- **Rule-based alerting** with flexible condition definitions
- **Multi-severity alerts** (critical, warning, info)
- **Alert lifecycle management** (firing, acknowledged, resolved)
- **Alert grouping and deduplication** to reduce noise
- **Escalation policies** for critical alerts
- **Silence management** for maintenance windows

### Service Health Monitoring
- **Comprehensive health checks** for all services
- **Dependency tracking** and impact analysis
- **Service status dashboard** with real-time updates
- **Uptime monitoring** and SLA tracking
- **Performance monitoring** with response time tracking
- **Automated failure detection** and recovery

### Multi-Channel Notifications
- **Slack integration** for team notifications
- **Email notifications** with customizable templates
- **Webhook support** for external system integration
- **PagerDuty integration** for on-call management
- **SMS notifications** for critical alerts
- **Custom notification channels** via plugins

## Configuration

### Environment Variables

#### Server Configuration
```bash
AGBA_MONITORING_ENVIRONMENT=production           # Environment (development/production)
AGBA_MONITORING_VERSION=1.0.0                   # Service version
AGBA_MONITORING_SERVER_PORT=8080                # HTTP server port
AGBA_MONITORING_SERVER_READ_TIMEOUT=30s         # Request read timeout
AGBA_MONITORING_SERVER_WRITE_TIMEOUT=30s        # Response write timeout
AGBA_MONITORING_SERVER_IDLE_TIMEOUT=120s        # Connection idle timeout
```

#### Database Configuration
```bash
AGBA_MONITORING_DATABASE_DRIVER=postgres        # Database driver
AGBA_MONITORING_DATABASE_DSN=postgresql://user:pass@host:5432/db  # Database connection string
AGBA_MONITORING_DATABASE_MAX_OPEN_CONNS=25      # Maximum open connections
AGBA_MONITORING_DATABASE_MAX_IDLE_CONNS=5       # Maximum idle connections
AGBA_MONITORING_DATABASE_CONN_MAX_LIFETIME=300s # Connection maximum lifetime
```

#### Redis Configuration
```bash
AGBA_MONITORING_REDIS_ADDRESS=redis-master:6379 # Redis server address
AGBA_MONITORING_REDIS_PASSWORD=your_password    # Redis password
AGBA_MONITORING_REDIS_DB=2                      # Redis database number
AGBA_MONITORING_REDIS_POOL_SIZE=10              # Connection pool size
```

#### NATS Configuration
```bash
AGBA_MONITORING_NATS_URL=nats://nats-client:4222 # NATS server URL
AGBA_MONITORING_NATS_USERNAME=agba_user         # NATS username
AGBA_MONITORING_NATS_PASSWORD=your_password     # NATS password
```

#### Prometheus Configuration
```bash
AGBA_MONITORING_PROMETHEUS_URL=http://prometheus:9090  # Prometheus server URL
AGBA_MONITORING_PROMETHEUS_TIMEOUT=30s          # Query timeout
AGBA_MONITORING_PROMETHEUS_MAX_CONNECTIONS=10   # Maximum connections
AGBA_MONITORING_PROMETHEUS_RETENTION_PERIOD=720h # Data retention period
```

## API Endpoints

### Metrics Management
```
GET    /api/v1/metrics                           # Get current metrics
GET    /api/v1/metrics/query                     # Query metrics with PromQL
GET    /api/v1/metrics/range                     # Query metrics over time range
GET    /api/v1/metrics/services                  # Get service-specific metrics
GET    /api/v1/metrics/services/{service}        # Get metrics for specific service
POST   /api/v1/metrics/custom                    # Record custom metric
```

### Health Monitoring
```
GET    /api/v1/health                            # Get overall system health
GET    /api/v1/health/services                   # Get all services health status
GET    /api/v1/health/services/{service}         # Get specific service health
POST   /api/v1/health/services/{service}/check   # Trigger health check
GET    /api/v1/health/dependencies               # Get dependency health status
```

### Alert Management
```
GET    /api/v1/alerts                            # List all alerts
GET    /api/v1/alerts/{id}                       # Get specific alert
POST   /api/v1/alerts                            # Create new alert
PUT    /api/v1/alerts/{id}                       # Update alert
DELETE /api/v1/alerts/{id}                       # Delete alert
POST   /api/v1/alerts/{id}/acknowledge           # Acknowledge alert
POST   /api/v1/alerts/{id}/resolve               # Resolve alert
```

## Development

### Local Development Setup
```bash
# Clone repository
git clone https://github.com/ibilola104/Agba.ai.git
cd Agba.ai/services/monitoring-service

# Install dependencies
go mod download

# Set required environment variables
export AGBA_MONITORING_DATABASE_DSN=postgresql://localhost:5432/agba_core
export AGBA_MONITORING_REDIS_ADDRESS=localhost:6379
export AGBA_MONITORING_REDIS_PASSWORD=your_redis_password
export AGBA_MONITORING_NATS_URL=nats://localhost:4222
export AGBA_MONITORING_NATS_PASSWORD=your_nats_password
export AGBA_MONITORING_PROMETHEUS_URL=http://localhost:9090

# Run service
go run cmd/main.go  # or python main.py
```

## Deployment

Standard containerized deployment with Kubernetes.

## Contributing

Follow platform development standards and best practices.

## Support

Contact the development team for monitoring related issues.
