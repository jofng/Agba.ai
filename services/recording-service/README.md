# Recording Service

The Recording Service provides comprehensive call recording, audio storage, and transcription capabilities for the Agba Voice AI Platform. It handles real-time recording, audio processing, secure storage, and compliance features with support for multiple storage providers and audio formats.

## Overview

The Recording Service offers centralized recording management with:
- **Real-time Recording** with WebSocket support for live audio streaming
- **Multi-format Audio Support** including WAV, MP3, FLAC, OGG, M4A, OPUS, and AAC
- **Secure Storage** with encryption and multiple storage providers (File, S3, GCS)
- **Audio Processing** with noise reduction, normalization, and compression
- **Transcription Services** with multiple providers (Whisper, Google STT, AWS, Azure)
- **Compliance Features** with data retention, audit logging, and GDPR support
- **Batch Processing** for bulk operations and background tasks
- **Quality Monitoring** with audio quality metrics and performance tracking

## Architecture

### Technology Stack
- **Language**: Go (Golang) 1.21
- **Framework**: Gin HTTP framework for REST API and WebSocket support
- **Database**: PostgreSQL for persistent storage of recordings, metadata, and audit logs
- **Cache**: Redis for session management and real-time data
- **Messaging**: NATS for real-time event processing and queue management
- **Storage**: Multi-provider support (File system, AWS S3, Google Cloud Storage)
- **Audio Processing**: FFmpeg integration for audio format conversion and processing
- **Transcription**: Multiple AI providers for speech-to-text conversion

### Key Components
1. **Recording Engine**: Core recording capture and management
2. **Storage Manager**: Multi-provider storage abstraction and management
3. **Audio Processor**: Real-time audio processing and format conversion
4. **Transcription Engine**: Speech-to-text processing with multiple providers
5. **Session Manager**: Real-time recording session management
6. **Batch Processor**: Background processing for bulk operations
7. **Audit Logger**: Comprehensive compliance and activity logging

## Features

### Real-time Recording Capabilities
- **Live audio streaming** with WebSocket support for real-time recording
- **Multi-channel recording** with support for conference calls and multiple participants
- **Pause and resume** functionality for flexible recording control
- **Quality monitoring** with real-time audio quality metrics
- **Automatic stop** based on silence detection or maximum duration
- **Chunk-based processing** for efficient memory usage and streaming

### Audio Processing & Quality
- **Format conversion** between WAV, MP3, FLAC, OGG, M4A, OPUS, and AAC
- **Audio normalization** with configurable levels and automatic gain control
- **Noise reduction** with advanced filtering algorithms
- **Compression** with multiple algorithms and quality settings
- **Quality analysis** with noise level, silence ratio, and quality scoring
- **Sample rate conversion** and channel mixing capabilities

### Storage & Security
- **Multi-provider storage** with File system, AWS S3, and Google Cloud Storage
- **Encryption at rest** with AES-256-GCM encryption for all recordings
- **Secure access control** with JWT authentication and API key support
- **Storage optimization** with automatic compression and archival
- **Redundancy support** with multiple storage provider failover
- **Cost optimization** with intelligent storage tier management

## Configuration

### Environment Variables

#### Server Configuration
```bash
AGBA_RECORDING_ENVIRONMENT=production          # Environment (development/production)
AGBA_RECORDING_VERSION=1.0.0                  # Service version
AGBA_RECORDING_SERVER_PORT=8080               # HTTP server port
```

#### Database Configuration
```bash
AGBA_RECORDING_DATABASE_DSN=postgresql://user:pass@host:5432/db # Database connection string
AGBA_RECORDING_REDIS_ADDRESS=redis-master:6379 # Redis server address
AGBA_RECORDING_NATS_URL=nats://nats-client:4222 # NATS server URL
```

#### Storage Configuration
```bash
AGBA_RECORDING_STORAGE_DEFAULT_PROVIDER=file   # Default storage provider
AGBA_RECORDING_STORAGE_FILE_BASE_PATH=/app/recordings # File storage path
AGBA_RECORDING_STORAGE_S3_BUCKET=recordings-bucket # S3 bucket name
AGBA_RECORDING_STORAGE_GCS_BUCKET=recordings-gcs # GCS bucket name
AGBA_RECORDING_STORAGE_ENCRYPTION_KEY=your-32-byte-key # Encryption key
```

#### Audio Processing Configuration
```bash
AGBA_RECORDING_AUDIO_DEFAULT_FORMAT=wav        # Default audio format
AGBA_RECORDING_AUDIO_DEFAULT_SAMPLE_RATE=16000 # Default sample rate
AGBA_RECORDING_AUDIO_ENABLE_NOISE_REDUCTION=true # Enable noise reduction
AGBA_RECORDING_AUDIO_ENABLE_NORMALIZATION=true # Enable audio normalization
```

## API Endpoints

### Recording Management
```
POST   /api/v1/recordings                      # Create new recording
GET    /api/v1/recordings                      # List recordings
GET    /api/v1/recordings/{id}                 # Get recording details
PUT    /api/v1/recordings/{id}                 # Update recording
DELETE /api/v1/recordings/{id}                 # Delete recording
POST   /api/v1/recordings/{id}/upload          # Upload recording file
GET    /api/v1/recordings/{id}/download        # Download recording
GET    /api/v1/recordings/{id}/stream          # Stream recording
```

### Real-time Recording
```
POST   /api/v1/realtime/start                  # Start real-time recording
POST   /api/v1/realtime/stop                   # Stop recording
POST   /api/v1/realtime/pause                  # Pause recording
POST   /api/v1/realtime/resume                 # Resume recording
POST   /api/v1/realtime/chunk                  # Upload audio chunk
GET    /api/v1/realtime/status/{sessionId}     # Get recording status
```

### Transcription Services
```
POST   /api/v1/recordings/{id}/transcribe      # Start transcription
GET    /api/v1/recordings/{id}/transcription   # Get transcription
POST   /api/v1/batch/transcribe                # Batch transcription
```

### Session Management
```
GET    /api/v1/sessions                        # List recording sessions
POST   /api/v1/sessions                        # Create recording session
GET    /api/v1/sessions/{id}                   # Get session details
PUT    /api/v1/sessions/{id}                   # Update session
DELETE /api/v1/sessions/{id}                   # Delete session
```

### Storage Management
```
GET    /api/v1/storage                         # Get storage information
GET    /api/v1/storage/usage                   # Get storage usage statistics
POST   /api/v1/storage/cleanup                 # Cleanup old recordings
POST   /api/v1/storage/migrate                 # Migrate between storage providers
```

## Development

### Local Development Setup
```bash
# Clone repository
git clone https://github.com/ibilola104/Agba.ai.git
cd Agba.ai/services/recording-service

# Install dependencies
go mod download

# Set required environment variables
export AGBA_RECORDING_DATABASE_DSN=postgresql://localhost:5432/agba_core
export AGBA_RECORDING_REDIS_ADDRESS=localhost:6379
export AGBA_RECORDING_NATS_URL=nats://localhost:4222
export AGBA_RECORDING_STORAGE_ENCRYPTION_KEY=your-32-byte-encryption-key

# Run service
go run cmd/main.go  # or python main.py
```

## Deployment

Standard containerized deployment with Kubernetes.

## Contributing

Follow platform development standards and best practices.

## Support

Contact the development team for recording related issues.
