# WebRTC Service

The WebRTC Service enables browser-based voice communication for the Agba Voice AI Platform. It handles WebRTC signaling, media stream management, and provides real-time voice interaction capabilities through web browsers.

## Overview

This service provides:
- WebRTC signaling server functionality
- Browser-based voice call initiation
- Real-time media stream handling
- STUN/TURN server integration
- WebSocket-based communication
- Cross-browser compatibility

## Architecture

### Technology Stack
- **Language**: Go (Golang)
- **WebRTC Library**: Pion WebRTC
- **WebSocket**: Gorilla WebSocket
- **Signaling**: Custom WebSocket protocol
- **Media**: Real-time audio processing
- **STUN/TURN**: Coturn server integration

### Key Components
1. **Signaling Server**: WebSocket-based signaling
2. **Peer Connection Manager**: WebRTC peer connections
3. **Media Handler**: Audio stream processing
4. **STUN/TURN Client**: NAT traversal support
5. **Session Manager**: Call session management
6. **Browser SDK**: Client-side JavaScript library

## Features

### WebRTC Capabilities
- Peer-to-peer audio communication
- NAT traversal with STUN/TURN
- Adaptive bitrate streaming
- Echo cancellation and noise suppression
- Real-time audio processing
- Cross-browser support (Chrome, Firefox, Safari, Edge)

### Signaling Protocol
- WebSocket-based signaling
- SDP offer/answer exchange
- ICE candidate exchange
- Call state management
- Error handling and recovery

### Media Processing
- Audio codec support (Opus, G.722, G.711)
- Real-time audio enhancement
- Voice activity detection
- Automatic gain control
- Jitter buffer management

## Configuration

### Environment Variables
```bash
# Server Configuration
PORT=8082
HOST=0.0.0.0
ENV=production

# WebRTC Configuration
STUN_SERVERS=stun:stun.l.google.com:19302,stun:stun1.l.google.com:19302
TURN_SERVER=turn:turn.agba.ai:3478
TURN_USERNAME=agba-turn-user
TURN_PASSWORD=agba-turn-pass

# Media Configuration
AUDIO_CODECS=opus,g722,pcmu,pcma
BITRATE_MIN=32000
BITRATE_MAX=128000

# Database
REDIS_URL=redis://localhost:6379
POSTGRES_URL=postgresql://user:pass@localhost:5432/agba_core
```

### WebRTC Configuration
```yaml
webrtc:
  ice_servers:
    - urls: ["stun:stun.l.google.com:19302"]
    - urls: ["turn:turn.agba.ai:3478"]
      username: "agba-turn-user"
      credential: "agba-turn-pass"
  
  media:
    audio:
      codecs: ["opus", "g722", "pcmu"]
      bitrate: 64000
      echo_cancellation: true
      noise_suppression: true
      auto_gain_control: true
  
  connection:
    ice_connection_timeout: 30
    ice_gathering_timeout: 10
    dtls_timeout: 30
```

## API Endpoints

### WebRTC Session Management
```
POST   /api/v1/webrtc/sessions     # Create WebRTC session
GET    /api/v1/webrtc/sessions/{id} # Get session details
DELETE /api/v1/webrtc/sessions/{id} # End session
```

### WebSocket Signaling
```
WS /ws/signaling/{session_id}      # WebSocket signaling endpoint
```

### Media Configuration
```
GET /api/v1/webrtc/config          # Get WebRTC configuration
GET /api/v1/webrtc/ice-servers     # Get ICE servers
```

## WebSocket Protocol

### Message Types
```json
{
  "type": "offer",
  "data": {
    "sdp": "v=0\r\no=...",
    "type": "offer"
  }
}

{
  "type": "answer",
  "data": {
    "sdp": "v=0\r\no=...",
    "type": "answer"
  }
}

{
  "type": "ice-candidate",
  "data": {
    "candidate": "candidate:...",
    "sdpMid": "0",
    "sdpMLineIndex": 0
  }
}

{
  "type": "call-state",
  "data": {
    "state": "connected",
    "timestamp": "2025-06-09T10:00:00Z"
  }
}
```

## Browser SDK

### JavaScript SDK Usage
```javascript
// Initialize WebRTC client
const agbaWebRTC = new AgbaWebRTC({
  apiUrl: 'https://api.agba.ai',
  signalingUrl: 'wss://api.agba.ai/ws/signaling'
});

// Start a call
const call = await agbaWebRTC.startCall({
  agentId: 'agent-123',
  userId: 'user-456'
});

// Handle call events
call.on('connected', () => {
  console.log('Call connected');
});

call.on('disconnected', () => {
  console.log('Call ended');
});

call.on('error', (error) => {
  console.error('Call error:', error);
});

// End the call
await call.end();
```

### SDK Features
- Promise-based API
- Event-driven architecture
- Automatic reconnection
- Error handling and recovery
- Media device management
- Call quality monitoring

## Integration

### AI Pipeline Integration
```go
// Forward audio to AI pipeline
func (w *WebRTCService) handleAudioTrack(track *webrtc.TrackRemote) {
    for {
        packet, err := track.ReadRTP()
        if err != nil {
            break
        }
        
        // Send audio to AI pipeline
        w.aiPipeline.ProcessAudio(packet.Payload)
    }
}
```

### Call Recording Integration
```go
// Record WebRTC call
func (w *WebRTCService) startRecording(sessionID string) {
    recorder := &CallRecorder{
        SessionID: sessionID,
        Format:    "wav",
        Bitrate:   16000,
    }
    
    w.recordingService.StartRecording(recorder)
}
```

## Security

### Transport Security
- WSS (WebSocket Secure) for signaling
- DTLS for media encryption
- SRTP for audio stream protection
- Certificate validation
- Origin validation

### Access Control
- JWT-based authentication
- Session-based authorization
- Rate limiting for connections
- IP-based restrictions
- CORS configuration

## Performance

### Optimization Strategies
- Connection pooling for WebSocket
- Efficient SDP parsing
- Optimized audio processing
- Memory management for streams
- Concurrent connection handling

### Monitoring Metrics
- Active WebRTC sessions
- Connection success rate
- Media quality metrics
- Signaling latency
- Bandwidth usage

### Performance Targets
- Connection setup: < 2 seconds
- Signaling latency: < 50ms
- Audio latency: < 150ms
- Concurrent sessions: 1,000+ per instance
- Connection success rate: > 98%

## Deployment

### Docker Configuration
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o webrtc-service ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/webrtc-service .
COPY --from=builder /app/static ./static
EXPOSE 8082
CMD ["./webrtc-service"]
```

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webrtc-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: webrtc-service
  template:
    metadata:
      labels:
        app: webrtc-service
    spec:
      containers:
      - name: webrtc-service
        image: agba/webrtc-service:latest
        ports:
        - containerPort: 8082
        env:
        - name: STUN_SERVERS
          value: "stun:stun.l.google.com:19302"
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
# Install dependencies
go mod download

# Set environment variables
export STUN_SERVERS=stun:stun.l.google.com:19302
export REDIS_URL=redis://localhost:6379

# Run service
go run cmd/main.go

# Serve test page
open http://localhost:8082/test
```

### Testing
```bash
# Unit tests
go test ./...

# WebRTC testing
go test -tags=webrtc ./...

# Browser testing
npm run test:browser
```

## Browser Compatibility

### Supported Browsers
- Chrome 80+
- Firefox 75+
- Safari 14+
- Edge 80+
- Mobile Chrome 80+
- Mobile Safari 14+

### Feature Detection
```javascript
// Check WebRTC support
if (!navigator.mediaDevices || !window.RTCPeerConnection) {
  throw new Error('WebRTC not supported');
}

// Check codec support
const codecs = RTCRtpSender.getCapabilities('audio').codecs;
const opusSupported = codecs.some(codec => 
  codec.mimeType === 'audio/opus'
);
```

## Monitoring

### Health Checks
- `/health`: Service health
- `/health/webrtc`: WebRTC stack health
- `/health/turn`: TURN server connectivity

### Metrics
- Active WebRTC connections
- Signaling message rate
- Media quality statistics
- Connection failure rate
- Bandwidth utilization

### Alerts
- High connection failure rate
- TURN server unavailable
- Excessive bandwidth usage
- Memory leaks in connections

## Troubleshooting

### Common Issues
1. **Connection Failures**: Check STUN/TURN configuration
2. **Audio Issues**: Verify codec support and media permissions
3. **NAT Traversal**: Ensure TURN server accessibility
4. **Browser Compatibility**: Check WebRTC API support

### Debug Tools
```bash
# Check WebRTC stats
curl http://localhost:8082/debug/webrtc/stats

# Monitor WebSocket connections
curl http://localhost:8082/debug/websocket/connections

# Test STUN/TURN connectivity
curl http://localhost:8082/debug/ice/test
```

## Contributing

1. Follow WebRTC best practices
2. Test across multiple browsers
3. Ensure proper error handling
4. Document protocol changes
5. Performance test with multiple connections

## Support

For WebRTC issues:
- Check browser console for errors
- Verify STUN/TURN server connectivity
- Test with different network conditions
- Contact WebRTC development team