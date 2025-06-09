# Telephony Gateway Service

The Telephony Gateway Service handles all PSTN and VoIP call operations for the Agba Voice AI Platform. It manages call routing, state management, and integration with FreeSWITCH for media processing.

## Overview

This service provides:
- Inbound and outbound call handling
- SIP protocol management
- Call state tracking and management
- Integration with FreeSWITCH media server
- Phone number provisioning and management
- Call transfer and conferencing capabilities

## Architecture

### Technology Stack
- **Language**: Go (Golang)
- **SIP Library**: Go SIP library for protocol handling
- **Media Server**: FreeSWITCH integration
- **State Management**: Redis for call state
- **Database**: PostgreSQL for call metadata
- **Messaging**: NATS for event publishing

### Key Components
1. **SIP Handler**: SIP protocol processing
2. **Call Manager**: Call lifecycle management
3. **Media Controller**: FreeSWITCH integration
4. **State Manager**: Call state persistence
5. **Event Publisher**: Real-time event notifications
6. **Number Manager**: Phone number provisioning

## Features

### Call Handling
- Inbound call reception and routing
- Outbound call initiation
- Call transfer (warm and cold)
- Conference call management
- Call recording integration
- Call quality monitoring

### SIP Protocol Support
- SIP INVITE, ACK, BYE, CANCEL handling
- SIP authentication and registration
- SDP negotiation for media streams
- NAT traversal support
- SIP over TLS (SIPS) support

### Media Processing
- Audio codec support (G.711, G.722, Opus)
- DTMF detection and generation
- Echo cancellation
- Noise reduction
- Voice activity detection

### Call State Management
- Real-time call state tracking
- Call duration and billing
- Call quality metrics
- Historical call data
- Call analytics and reporting

## Configuration

### Environment Variables
```bash
# Server Configuration
PORT=8081
HOST=0.0.0.0
ENV=production

# Database
POSTGRES_URL=postgresql://user:pass@localhost:5432/agba_core
REDIS_URL=redis://localhost:6379

# FreeSWITCH Configuration
FREESWITCH_HOST=freeswitch.agba.local
FREESWITCH_PORT=8021
FREESWITCH_PASSWORD=ClueCon

# SIP Configuration
SIP_DOMAIN=agba.ai
SIP_PORT=5060
SIP_TLS_PORT=5061

# Telephony Providers
TWILIO_ACCOUNT_SID=your-twilio-sid
TWILIO_AUTH_TOKEN=your-twilio-token
BANDWIDTH_USER_ID=your-bandwidth-user
BANDWIDTH_API_TOKEN=your-bandwidth-token
```

### SIP Configuration
```yaml
sip:
  domain: agba.ai
  port: 5060
  tls_port: 5061
  transport: [UDP, TCP, TLS]
  
  authentication:
    realm: agba.ai
    algorithm: MD5
    
  codecs:
    - G711U
    - G711A
    - G722
    - OPUS
    
  rtp:
    port_range: "10000-20000"
    timeout: 30
```

## API Endpoints

### Call Management
```
POST   /api/v1/calls              # Initiate outbound call
GET    /api/v1/calls              # List active calls
GET    /api/v1/calls/{id}         # Get call details
PUT    /api/v1/calls/{id}/transfer # Transfer call
PUT    /api/v1/calls/{id}/hold    # Hold/unhold call
DELETE /api/v1/calls/{id}         # End call
```

### Phone Numbers
```
GET    /api/v1/numbers            # List phone numbers
POST   /api/v1/numbers            # Purchase number
GET    /api/v1/numbers/{number}   # Get number details
PUT    /api/v1/numbers/{number}   # Update number config
DELETE /api/v1/numbers/{number}   # Release number
```

### Call Analytics
```
GET /api/v1/analytics/calls       # Call statistics
GET /api/v1/analytics/quality     # Call quality metrics
GET /api/v1/analytics/usage       # Usage reports
```

## Call Flow

### Inbound Call Flow
1. SIP INVITE received from carrier
2. Call authenticated and authorized
3. Call routed to appropriate agent
4. Media session established
5. Call state tracked in Redis
6. Events published to NATS
7. Call ended with BYE message

### Outbound Call Flow
1. API request to initiate call
2. SIP INVITE sent to destination
3. Call progress monitored
4. Media session established
5. Call connected to AI agent
6. Call state managed throughout
7. Call termination handling

## Integration

### FreeSWITCH Integration
```go
// FreeSWITCH Event Socket connection
type FreeSWITCHClient struct {
    conn net.Conn
    auth string
}

// Handle call events from FreeSWITCH
func (f *FreeSWITCHClient) HandleEvent(event *Event) {
    switch event.Type {
    case "CHANNEL_CREATE":
        // Handle new call
    case "CHANNEL_ANSWER":
        // Handle call answer
    case "CHANNEL_HANGUP":
        // Handle call end
    }
}
```

### Carrier Integration
- Twilio SIP trunking
- Bandwidth Voice API
- Direct SIP carrier connections
- E.164 number formatting
- International calling support

## Security

### SIP Security
- SIP digest authentication
- TLS encryption for signaling
- SRTP for media encryption
- IP whitelisting for carriers
- Rate limiting for SIP requests

### Access Control
- JWT-based API authentication
- Role-based permissions
- Audit logging for all operations
- Secure credential storage
- Network security policies

## Performance

### Optimization Strategies
- Connection pooling for database
- Redis caching for call state
- Async processing for events
- Load balancing across instances
- Media optimization

### Monitoring Metrics
- Concurrent call capacity
- Call setup success rate
- Call quality metrics (MOS, jitter, packet loss)
- SIP response times
- Media latency

### Performance Targets
- Call setup time: < 3 seconds
- Concurrent calls: 1,000+ per instance
- SIP response time: < 100ms
- Media latency: < 150ms
- Call success rate: > 99%

## Deployment

### Docker Configuration
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o telephony-gateway ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/telephony-gateway .
EXPOSE 8081 5060 5061
CMD ["./telephony-gateway"]
```

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: telephony-gateway
spec:
  replicas: 5
  selector:
    matchLabels:
      app: telephony-gateway
  template:
    metadata:
      labels:
        app: telephony-gateway
    spec:
      containers:
      - name: telephony-gateway
        image: agba/telephony-gateway:latest
        ports:
        - containerPort: 8081
        - containerPort: 5060
        - containerPort: 5061
        resources:
          requests:
            cpu: 1000m
            memory: 2Gi
          limits:
            cpu: 4000m
            memory: 8Gi
```

## Development

### Local Setup
```bash
# Install FreeSWITCH locally
docker run -d --name freeswitch \
  -p 5060:5060/udp \
  -p 8021:8021 \
  drachtio/freeswitch

# Set environment variables
export FREESWITCH_HOST=localhost
export FREESWITCH_PASSWORD=ClueCon

# Run service
go run cmd/main.go
```

### Testing
```bash
# Unit tests
go test ./...

# SIP testing with SIPp
sipp -sf uac.xml localhost:5060

# Load testing
go test -tags=load ./...
```

## Monitoring

### Health Checks
- `/health`: Service health
- `/health/sip`: SIP stack health
- `/health/freeswitch`: FreeSWITCH connectivity

### Metrics
- Active call count
- Call setup rate
- Call failure rate
- SIP message rate
- Media quality metrics

### Alerts
- High call failure rate
- FreeSWITCH connectivity issues
- SIP authentication failures
- Resource exhaustion

## Troubleshooting

### Common Issues
1. **Call Setup Failures**: Check SIP configuration and carrier connectivity
2. **Audio Issues**: Verify codec negotiation and RTP ports
3. **FreeSWITCH Connection**: Check event socket connectivity
4. **High Latency**: Review network configuration and media routing

### Debug Tools
```bash
# Check SIP traffic
tcpdump -i any port 5060

# FreeSWITCH CLI
fs_cli -H localhost -P ClueCon

# Call trace
curl http://localhost:8081/debug/calls/{call-id}
```

## Contributing

1. Follow Go best practices for concurrent programming
2. Test SIP scenarios thoroughly
3. Ensure FreeSWITCH integration stability
4. Document call flow changes
5. Performance test under load

## Support

For telephony issues:
- Check SIP logs and traces
- Verify FreeSWITCH configuration
- Test carrier connectivity
- Contact telephony team for escalation