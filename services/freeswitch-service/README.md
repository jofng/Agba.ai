# FreeSWITCH Service

The FreeSWITCH Service is a core component of the Agba Voice AI Platform that provides comprehensive telephony capabilities including SIP call handling, media processing, call routing, and integration with other platform services.

## 🎯 Overview

This service acts as the telephony backbone of the Agba AI platform, managing:
- Inbound and outbound SIP calls
- Call routing and transfer operations
- Media processing and recording integration
- Real-time call monitoring and analytics
- Integration with AI/ML services for voice processing

## 🏗️ Architecture

### Core Components

- **FreeSWITCH Manager**: Manages FreeSWITCH Event Socket connections and operations
- **Call Service**: Business logic for call handling and management
- **Integration Layer**: Communication with other platform services
- **HTTP Handlers**: RESTful API endpoints for call operations
- **Configuration Manager**: FreeSWITCH and service configuration management

### Technology Stack

- **Go 1.21**: Service implementation
- **FreeSWITCH**: Core telephony platform
- **Gin**: HTTP web framework
- **PostgreSQL**: Call data and configuration storage
- **Redis**: Caching and session management
- **NATS**: Event streaming and messaging
- **Docker**: Containerization
- **Kubernetes**: Orchestration and deployment

## 🚀 Features

### Call Management
- **Outbound Calls**: Initiate calls through SIP trunks
- **Inbound Calls**: Handle incoming calls with intelligent routing
- **Call Transfer**: Blind and attended call transfers
- **Call Hold/Resume**: Put calls on hold and resume
- **Call Recording**: Start, stop, pause, and resume call recordings
- **Call Analytics**: Real-time call metrics and statistics

### SIP Integration
- **Multiple SIP Profiles**: Internal, external, and WebRTC profiles
- **SIP Trunk Support**: Integration with carrier SIP trunks
- **Codec Support**: OPUS, G.722, PCMU, PCMA, GSM
- **DTMF Handling**: RFC2833 DTMF support
- **NAT Traversal**: Automatic NAT handling

### AI Integration
- **Voice AI Routing**: Route calls to AI voice assistants
- **Recording Integration**: Automatic call recording for AI processing
- **Real-time Events**: Stream call events for AI analysis
- **Custom Dialplan**: AI-optimized call routing logic

### Security
- **TLS Support**: Secure SIP communications
- **Authentication**: SIP authentication and authorization
- **Access Control**: Network-based access control lists
- **Encryption**: Media encryption support

## 📋 API Endpoints

### Call Operations

#### Originate Call
```http
POST /api/v1/calls/originate
Content-Type: application/json

{
  "caller_id": "+1234567890",
  "called_number": "+0987654321",
  "user_id": "uuid",
  "trunk_id": "agba_trunk",
  "recording": true,
  "variables": {
    "custom_var": "value"
  }
}
```

#### Hangup Call
```http
POST /api/v1/calls/{call_id}/hangup
```

#### Transfer Call
```http
POST /api/v1/calls/transfer
Content-Type: application/json

{
  "call_id": "uuid",
  "destination": "+1234567890",
  "type": "blind"
}
```

#### Hold/Unhold Call
```http
POST /api/v1/calls/hold
Content-Type: application/json

{
  "call_id": "uuid",
  "hold": true
}
```

#### Control Recording
```http
POST /api/v1/calls/recording
Content-Type: application/json

{
  "call_id": "uuid",
  "action": "start"
}
```

### Information Endpoints

#### Get Call Information
```http
GET /api/v1/calls/{call_id}
```

#### Get Active Calls
```http
GET /api/v1/calls/active
```

#### Get Call Statistics
```http
GET /api/v1/calls/stats
```

#### Health Check
```http
GET /health
```

#### Metrics
```http
GET /api/v1/metrics
```

## ⚙️ Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_NAME` | Database name | `agba_freeswitch` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | - |
| `REDIS_HOST` | Redis host | `localhost` |
| `REDIS_PORT` | Redis port | `6379` |
| `REDIS_PASSWORD` | Redis password | - |
| `NATS_URL` | NATS server URL | `nats://localhost:4222` |
| `FREESWITCH_HOST` | FreeSWITCH host | `localhost` |
| `FREESWITCH_EVENT_SOCKET_PORT` | Event socket port | `8021` |
| `FREESWITCH_PASSWORD` | Event socket password | `ClueCon` |

### FreeSWITCH Configuration

The service includes comprehensive FreeSWITCH configuration files:

- **vars.xml**: Global variables and settings
- **dialplan.xml**: Call routing and AI integration logic
- **sip_profiles.xml**: SIP profile configurations

### Service Integration

The service integrates with other platform services:

- **Recording Service**: Call recording and storage
- **Monitoring Service**: Metrics and health monitoring
- **User Management Service**: Authentication and authorization
- **Config Service**: Dynamic configuration management
- **Notification Service**: Call status notifications

## 🔧 Development

### Prerequisites

- Go 1.21+
- FreeSWITCH 1.10+
- PostgreSQL 13+
- Redis 6+
- NATS 2.9+

### Local Development

1. **Clone the repository**
   ```bash
   git clone https://github.com/agba-ai/agba-platform.git
   cd agba-platform/services/freeswitch-service
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up FreeSWITCH**
   ```bash
   # Install FreeSWITCH (Ubuntu/Debian)
   apt-get update
   apt-get install -y freeswitch freeswitch-mod-commands freeswitch-mod-event-socket
   
   # Copy configuration files
   cp configs/freeswitch/* /etc/freeswitch/
   ```

4. **Set up database**
   ```bash
   createdb agba_freeswitch
   psql agba_freeswitch < database/schema.sql
   ```

5. **Run the service**
   ```bash
   go run cmd/main.go
   ```

### Testing

```bash
# Run unit tests
go test ./...

# Run integration tests
go test -tags=integration ./...

# Run with coverage
go test -cover ./...
```

### Building

```bash
# Build binary
go build -o freeswitch-service cmd/main.go

# Build Docker image
docker build -t agba-ai/freeswitch-service:latest .
```

## 🚀 Deployment

### Kubernetes Deployment

1. **Apply the deployment**
   ```bash
   kubectl apply -f deployments/deployment.yaml
   ```

2. **Verify deployment**
   ```bash
   kubectl get pods -l app=freeswitch-service
   kubectl logs -l app=freeswitch-service
   ```

3. **Check service health**
   ```bash
   kubectl port-forward svc/freeswitch-service 8080:8080
   curl http://localhost:8080/health
   ```

### Docker Compose

```yaml
version: '3.8'
services:
  freeswitch-service:
    image: agba-ai/freeswitch-service:latest
    ports:
      - "8080:8080"
      - "5060:5060/udp"
      - "8021:8021"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
      - NATS_URL=nats://nats:4222
    depends_on:
      - postgres
      - redis
      - nats
```

## 📊 Monitoring

### Health Checks

The service provides comprehensive health checks:

- **HTTP Health Endpoint**: `/health`
- **FreeSWITCH Connection**: Event socket connectivity
- **Database Health**: PostgreSQL connection status
- **Service Integration**: Other service connectivity

### Metrics

Key metrics exposed:

- **Call Metrics**: Total calls, active calls, success rate
- **Performance Metrics**: Call setup time, duration
- **System Metrics**: CPU, memory, network usage
- **FreeSWITCH Metrics**: Channel count, registration status

### Logging

Structured logging with configurable levels:

- **Call Events**: Call start, answer, hangup, transfer
- **System Events**: Service start/stop, errors, warnings
- **Integration Events**: Service communication, failures
- **Performance Events**: Slow operations, timeouts

## 🔒 Security

### Network Security

- **TLS Encryption**: Secure SIP communications
- **Network Policies**: Kubernetes network isolation
- **Firewall Rules**: Restricted port access
- **VPN Integration**: Secure trunk connections

### Authentication

- **SIP Authentication**: Digest authentication
- **API Authentication**: JWT token validation
- **Service Authentication**: mTLS between services
- **Database Security**: Encrypted connections

### Data Protection

- **Call Recording Encryption**: Encrypted storage
- **PII Protection**: Caller ID anonymization
- **Audit Logging**: Comprehensive audit trails
- **Data Retention**: Configurable retention policies

## 🐛 Troubleshooting

### Common Issues

1. **FreeSWITCH Connection Failed**
   ```bash
   # Check FreeSWITCH status
   systemctl status freeswitch
   
   # Check event socket
   telnet localhost 8021
   ```

2. **Call Setup Failures**
   ```bash
   # Check SIP trunk configuration
   fs_cli -x "sofia status"
   
   # Check gateway status
   fs_cli -x "sofia status gateway agba_trunk"
   ```

3. **Recording Issues**
   ```bash
   # Check recording directory permissions
   ls -la /usr/local/freeswitch/recordings/
   
   # Check disk space
   df -h /usr/local/freeswitch/recordings/
   ```

### Debug Mode

Enable debug logging:

```bash
export LOG_LEVEL=debug
./freeswitch-service
```

### FreeSWITCH CLI

Access FreeSWITCH CLI for debugging:

```bash
fs_cli

# Show active calls
show calls

# Show channels
show channels

# Show registrations
sofia status profile internal reg
```

## 📚 Documentation

- [API Documentation](./docs/api.md)
- [Configuration Guide](./docs/configuration.md)
- [Deployment Guide](./docs/deployment.md)
- [Troubleshooting Guide](./docs/troubleshooting.md)
- [FreeSWITCH Integration](./docs/freeswitch.md)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For support and questions:

- **Documentation**: Check the docs directory
- **Issues**: Create a GitHub issue
- **Discussions**: Use GitHub discussions
- **Email**: support@agba.ai

---

**Version**: 1.0.0  
**Last Updated**: June 9, 2025  
**Maintainer**: Agba AI Platform Team