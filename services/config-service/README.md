# Configuration Service

The Configuration Service manages dynamic configuration for agents, system settings, and organizational preferences in the Agba Voice AI Platform. It provides versioning, validation, and real-time configuration updates with comprehensive audit trails.

## Overview

The Configuration Service provides centralized management of all platform configurations with:
- Dynamic configuration updates without service restarts
- Version control and rollback capabilities with automatic versioning
- Configuration validation and schema enforcement
- Real-time notifications for configuration changes via NATS
- Template-based configuration management for rapid deployment
- Comprehensive audit logging for compliance and debugging
- Encryption and compression for sensitive configurations

## Architecture

### Technology Stack
- **Language**: Go (Golang) 1.21
- **Framework**: Gin HTTP framework for REST API
- **Database**: PostgreSQL with JSONB for configuration storage
- **Cache**: Redis for configuration caching and performance
- **Messaging**: NATS for real-time configuration change notifications
- **Validation**: JSON Schema validation with custom business rules
- **Monitoring**: Prometheus metrics and structured logging

### Key Components
1. **Configuration Manager**: Core configuration CRUD operations with versioning
2. **Version Controller**: Automatic versioning, rollback, and change tracking
3. **Validator**: Schema-based validation with custom business rules
4. **Template Engine**: Configuration template management and instantiation
5. **Audit Logger**: Comprehensive change tracking and compliance logging
6. **Cache Manager**: Redis-based configuration caching with TTL management
7. **Notification Service**: Real-time configuration change notifications

## Features

### Configuration Management
- **Agent-specific configurations**: Voice, AI, telephony, and conversation settings
- **System-wide configuration**: Global platform settings and feature flags
- **Organization-level configuration**: Multi-tenant configuration inheritance
- **Environment-specific configurations**: Development, staging, production isolation
- **Configuration templates**: Pre-built templates for rapid agent deployment
- **Dynamic updates**: Hot-reloading without service restarts

### Version Control & Rollback
- **Automatic versioning**: Every configuration change creates a new version
- **Point-in-time snapshots**: Complete configuration state at any version
- **One-click rollback**: Instant rollback to any previous version
- **Change history**: Detailed diff tracking between versions
- **Audit trails**: Complete change history with user attribution
- **Version retention**: Configurable version retention policies

### Validation & Schema Management
- **JSON Schema validation**: Strict schema enforcement for all configurations
- **Custom validation rules**: Business logic validation beyond schema
- **Configuration dependency checking**: Validate inter-configuration dependencies
- **Real-time validation**: Immediate feedback during configuration editing
- **Schema evolution**: Managed schema updates with migration support
- **Validation levels**: Configurable validation strictness (strict, moderate, lenient)

### Real-time Updates & Notifications
- **NATS-based notifications**: Real-time configuration change events
- **Selective propagation**: Target specific services or agents for updates
- **Conflict resolution**: Automatic handling of concurrent configuration updates
- **Event streaming**: Configuration change event streams for external systems
- **Webhook integration**: HTTP callbacks for configuration changes

## API Endpoints

### Agent Configuration Management
```
GET    /api/v1/agents/{id}/config                    # Get current agent configuration
PUT    /api/v1/agents/{id}/config                    # Update agent configuration
GET    /api/v1/agents/{id}/config/versions           # List all configuration versions
GET    /api/v1/agents/{id}/config/versions/{version} # Get specific version
POST   /api/v1/agents/{id}/config/rollback/{version} # Rollback to specific version
POST   /api/v1/agents/{id}/config/validate           # Validate configuration without saving
```

### System Configuration Management
```
GET    /api/v1/system/config                         # Get system configuration
PUT    /api/v1/system/config                         # Update system configuration
GET    /api/v1/system/config/versions                # List system config versions
GET    /api/v1/system/config/versions/{version}      # Get specific system config version
POST   /api/v1/system/config/rollback/{version}      # Rollback system configuration
```

### Organization Configuration Management
```
GET    /api/v1/organizations/{id}/config             # Get organization configuration
PUT    /api/v1/organizations/{id}/config             # Update organization configuration
GET    /api/v1/organizations/{id}/config/versions    # List organization config versions
```

### Configuration Templates
```
GET    /api/v1/templates                             # List all configuration templates
GET    /api/v1/templates/{id}                        # Get template details
POST   /api/v1/templates                             # Create new template
PUT    /api/v1/templates/{id}                        # Update existing template
DELETE /api/v1/templates/{id}                        # Delete template
POST   /api/v1/templates/{id}/instantiate            # Create configuration from template
```

### Configuration Schemas
```
GET    /api/v1/schemas                               # List all configuration schemas
GET    /api/v1/schemas/{type}                        # Get schema for configuration type
PUT    /api/v1/schemas/{type}                        # Update schema definition
POST   /api/v1/schemas/{type}/validate               # Validate data against schema
```

### Audit and Compliance
```
GET    /api/v1/audit                                 # Get audit logs with filtering
GET    /api/v1/audit/{id}                            # Get specific audit log entry
GET    /api/v1/audit/summary                         # Get audit summary statistics
```

## Configuration

### Environment Variables

#### Server Configuration
```bash
AGBA_CONFIG_ENVIRONMENT=production                   # Environment (development/production)
AGBA_CONFIG_VERSION=1.0.0                           # Service version
AGBA_CONFIG_SERVER_PORT=8080                        # HTTP server port
AGBA_CONFIG_SERVER_READ_TIMEOUT=30s                 # Request read timeout
AGBA_CONFIG_SERVER_WRITE_TIMEOUT=30s                # Response write timeout
AGBA_CONFIG_SERVER_IDLE_TIMEOUT=120s                # Connection idle timeout
```

#### Database Configuration
```bash
AGBA_CONFIG_DATABASE_DRIVER=postgres                # Database driver
AGBA_CONFIG_DATABASE_DSN=postgresql://user:pass@host:5432/db  # Database connection string
AGBA_CONFIG_DATABASE_MAX_OPEN_CONNS=25              # Maximum open connections
AGBA_CONFIG_DATABASE_MAX_IDLE_CONNS=5               # Maximum idle connections
AGBA_CONFIG_DATABASE_CONN_MAX_LIFETIME=300s         # Connection maximum lifetime
```

#### Redis Configuration
```bash
AGBA_CONFIG_REDIS_ADDRESS=redis-master:6379         # Redis server address
AGBA_CONFIG_REDIS_PASSWORD=your_password            # Redis password
AGBA_CONFIG_REDIS_DB=1                              # Redis database number
AGBA_CONFIG_REDIS_POOL_SIZE=10                      # Connection pool size
```

#### NATS Configuration
```bash
AGBA_CONFIG_NATS_URL=nats://nats-client:4222        # NATS server URL
AGBA_CONFIG_NATS_USERNAME=agba_user                 # NATS username
AGBA_CONFIG_NATS_PASSWORD=your_password             # NATS password
```

#### Configuration Management Settings
```bash
AGBA_CONFIG_CONFIG_MANAGEMENT_MAX_VERSIONS=50       # Maximum versions to retain
AGBA_CONFIG_CONFIG_MANAGEMENT_CACHE_TTL=300s        # Configuration cache TTL
AGBA_CONFIG_CONFIG_MANAGEMENT_VALIDATION_TIMEOUT=30s # Validation timeout
AGBA_CONFIG_CONFIG_MANAGEMENT_NOTIFICATION_ENABLED=true # Enable notifications
AGBA_CONFIG_CONFIG_MANAGEMENT_AUTO_BACKUP_ENABLED=true  # Enable automatic backups
AGBA_CONFIG_CONFIG_MANAGEMENT_BACKUP_INTERVAL=24h   # Backup interval
AGBA_CONFIG_CONFIG_MANAGEMENT_ENCRYPTION_ENABLED=true   # Enable encryption
AGBA_CONFIG_CONFIG_MANAGEMENT_COMPRESSION_ENABLED=true  # Enable compression
AGBA_CONFIG_CONFIG_MANAGEMENT_AUDIT_RETENTION_DAYS=90   # Audit log retention
AGBA_CONFIG_CONFIG_MANAGEMENT_CONFIG_VALIDATION_LEVEL=strict # Validation level
```

## Development

### Local Development Setup
```bash
# Clone repository
git clone https://github.com/ibilola104/Agba.ai.git
cd Agba.ai/services/config-service

# Install dependencies
go mod download

# Set required environment variables
export AGBA_CONFIG_DATABASE_DSN=postgresql://localhost:5432/agba_core
export AGBA_CONFIG_REDIS_ADDRESS=localhost:6379
export AGBA_CONFIG_REDIS_PASSWORD=your_redis_password
export AGBA_CONFIG_NATS_URL=nats://localhost:4222
export AGBA_CONFIG_NATS_PASSWORD=your_nats_password

# Start the service
go run cmd/main.go
```

### Testing
```bash
# Run unit tests
go test ./...

# Run integration tests
go test -tags=integration ./...

# Run tests with coverage
go test -cover ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Test configuration validation
curl -X POST http://localhost:8080/api/v1/agents/123/config/validate \
  -H "Content-Type: application/json" \
  -d @examples/agent-config.json
```

### Building
```bash
# Build binary
go build -o config-service cmd/main.go

# Build Docker image
docker build -t agba/config-service:latest .
```

## Deployment

### Kubernetes Deployment
```bash
# Deploy to Kubernetes
kubectl apply -f deployments/deployment.yaml

# Check deployment status
kubectl get pods -l app=config-service -n agba-system
kubectl logs -l app=config-service -n agba-system

# Scale deployment
kubectl scale deployment config-service --replicas=5 -n agba-system
```

### Database Setup
```bash
# Run database schema setup
kubectl apply -f deployments/deployment.yaml

# Verify database tables
kubectl exec -it postgresql-0 -n agba-data -- \
  psql -U agba_user -d agba_core -c "\dt config*"
```

## Monitoring

### Health Checks
- **`/health`**: Basic service health check
- **`/ready`**: Readiness check including database and Redis connectivity
- **`/metrics`**: Prometheus metrics endpoint for monitoring

### Key Metrics
- Configuration operations per second
- Validation success/failure rates
- Cache hit ratios and performance
- Database query performance
- Version operations and rollback frequency

### Alerting Rules
- Service availability alerts
- Database connectivity alerts
- High error rate alerts (>5% for 5 minutes)
- Validation failure alerts
- Cache performance alerts
- Memory usage alerts (>80%)

## Security

### Access Control & Authentication
- Role-based access control for configuration management
- Organization-level isolation and multi-tenant security
- API key authentication for service-to-service communication
- JWT token validation for user authentication
- IP-based access restrictions

### Data Protection & Encryption
- Configuration encryption at rest with AES-256
- Secure transmission over TLS
- Sensitive data masking in logs
- Configuration backup encryption
- Field-level encryption for sensitive data

## Troubleshooting

### Common Issues

1. **Configuration Validation Failures**
   ```bash
   # Check schema compatibility
   curl http://localhost:8080/api/v1/schemas/agent
   
   # Validate configuration manually
   curl -X POST http://localhost:8080/api/v1/agents/123/config/validate \
     -H "Content-Type: application/json" \
     -d @your-config.json
   ```

2. **Cache Performance Issues**
   ```bash
   # Check Redis connectivity
   kubectl exec -it deployment/config-service -n agba-system -- \
     redis-cli -h redis-master -a <password> ping
   
   # Monitor cache metrics
   curl http://localhost:8080/metrics | grep cache
   ```

3. **Database Connection Problems**
   ```bash
   # Check database connectivity
   kubectl exec -it deployment/config-service -n agba-system -- \
     pg_isready -h pgbouncer -p 5432 -U agba_user
   
   # Check connection pool status
   curl http://localhost:8080/ready
   ```

### Debug Commands
```bash
# Service health check
curl http://localhost:8080/health

# Get service metrics
curl http://localhost:8080/metrics

# Test configuration CRUD
curl http://localhost:8080/api/v1/agents/123/config
curl -X PUT http://localhost:8080/api/v1/agents/123/config \
  -H "Content-Type: application/json" \
  -d '{"name":"test","language":"en"}'

# Check configuration versions
curl http://localhost:8080/api/v1/agents/123/config/versions

# Get audit logs
curl http://localhost:8080/api/v1/audit?limit=10
```

## Contributing

### Development Guidelines
1. Follow Go coding standards and best practices
2. Add comprehensive unit tests for new features
3. Update API documentation for endpoint changes
4. Ensure proper error handling and logging
5. Test configuration validation thoroughly

### Code Review Process
1. Create feature branch from main
2. Implement with comprehensive tests
3. Update documentation
4. Submit pull request with detailed description
5. Address code review feedback
6. Merge after CI/CD validation