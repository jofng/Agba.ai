# AI/ML Pipeline Service

The AI/ML Pipeline Service is the core intelligence component of the Agba Voice AI Platform. It orchestrates speech-to-text, natural language understanding, large language model processing, and text-to-speech to enable natural voice conversations.

## Overview

This service provides:
- Multi-provider STT/TTS integration
- LLM orchestration and prompt management
- Natural Language Understanding (NLU)
- Conversation context management
- Real-time audio processing
- Model performance monitoring

## Architecture

### Technology Stack
- **Language**: Python 3.11+
- **Framework**: FastAPI for async APIs
- **AI/ML Libraries**: Transformers, SpeechRecognition, spaCy
- **Audio Processing**: pydub, librosa
- **Async Processing**: Celery with Redis
- **Model Management**: Hugging Face Hub, MLflow

### Key Components
1. **STT Orchestrator**: Multi-provider speech recognition
2. **TTS Manager**: Text-to-speech synthesis
3. **LLM Gateway**: Large language model integration
4. **NLU Engine**: Intent recognition and entity extraction
5. **Conversation Manager**: Context and state management
6. **Audio Processor**: Real-time audio enhancement

## Features

### Speech-to-Text (STT)
- Multi-provider support (Deepgram, Google, AssemblyAI, Whisper)
- Real-time streaming transcription
- Speaker diarization
- Language detection
- Custom vocabulary and domain adaptation

### Text-to-Speech (TTS)
- Multi-provider support (ElevenLabs, PlayHT, Google, Azure)
- Voice cloning capabilities
- Emotion and prosody control
- SSML support
- Real-time synthesis

### Large Language Models
- OpenAI GPT integration
- Anthropic Claude support
- Google Gemini integration
- Custom model deployment
- Prompt engineering and templates

### Natural Language Understanding
- Intent classification
- Entity extraction
- Sentiment analysis
- Topic modeling
- Conversation flow management

## Configuration

### Environment Variables
```bash
# Server Configuration
PORT=8083
HOST=0.0.0.0
ENV=production

# Database
POSTGRES_URL=postgresql://user:pass@localhost:5432/agba_core
REDIS_URL=redis://localhost:6379

# STT Providers
DEEPGRAM_API_KEY=your-deepgram-key
GOOGLE_CLOUD_CREDENTIALS=path/to/credentials.json
ASSEMBLYAI_API_KEY=your-assemblyai-key
OPENAI_API_KEY=your-openai-key

# TTS Providers
ELEVENLABS_API_KEY=your-elevenlabs-key
PLAYHT_API_KEY=your-playht-key
AZURE_SPEECH_KEY=your-azure-key

# LLM Providers
OPENAI_API_KEY=your-openai-key
ANTHROPIC_API_KEY=your-anthropic-key
GOOGLE_AI_API_KEY=your-google-ai-key

# Model Configuration
DEFAULT_STT_PROVIDER=deepgram
DEFAULT_TTS_PROVIDER=elevenlabs
DEFAULT_LLM_PROVIDER=openai
```

### Provider Configuration
```yaml
stt_providers:
  deepgram:
    model: nova-2
    language: en-US
    punctuate: true
    diarize: true
    
  google:
    model: latest_long
    language_code: en-US
    enable_automatic_punctuation: true
    
  assemblyai:
    model: best
    language_code: en_us
    speaker_labels: true

tts_providers:
  elevenlabs:
    model: eleven_multilingual_v2
    voice_id: 21m00Tcm4TlvDq8ikWAM
    stability: 0.5
    similarity_boost: 0.8
    
  playht:
    voice: en-US-JennyNeural
    speed: 1.0
    pitch: 0
    
llm_providers:
  openai:
    model: gpt-4-turbo-preview
    temperature: 0.7
    max_tokens: 1000
    
  anthropic:
    model: claude-3-sonnet-20240229
    max_tokens: 1000
```

## API Endpoints

### Speech Processing
```
POST /api/v1/stt/transcribe        # Transcribe audio file
POST /api/v1/stt/stream           # Start streaming transcription
POST /api/v1/tts/synthesize       # Synthesize speech
POST /api/v1/tts/stream           # Stream speech synthesis
```

### Conversation Processing
```
POST /api/v1/conversation/process  # Process conversation turn
GET  /api/v1/conversation/{id}     # Get conversation state
PUT  /api/v1/conversation/{id}     # Update conversation
```

### Model Management
```
GET  /api/v1/models               # List available models
POST /api/v1/models/deploy        # Deploy custom model
GET  /api/v1/models/{id}/status   # Get model status
```

## Processing Pipeline

### Conversation Flow
```python
async def process_conversation_turn(audio_data: bytes, context: ConversationContext):
    # 1. Speech-to-Text
    transcript = await stt_service.transcribe(audio_data)
    
    # 2. Natural Language Understanding
    intent, entities = await nlu_service.analyze(transcript)
    
    # 3. Context Management
    context.update(intent, entities, transcript)
    
    # 4. LLM Processing
    response = await llm_service.generate_response(context)
    
    # 5. Text-to-Speech
    audio_response = await tts_service.synthesize(response)
    
    return {
        "transcript": transcript,
        "response": response,
        "audio": audio_response,
        "context": context
    }
```

### Real-time Streaming
```python
async def stream_conversation(websocket: WebSocket):
    async for audio_chunk in websocket.iter_bytes():
        # Process audio chunk
        partial_transcript = await stt_service.stream_transcribe(audio_chunk)
        
        # Send partial results
        await websocket.send_json({
            "type": "partial_transcript",
            "data": partial_transcript
        })
        
        # Check for end of utterance
        if stt_service.is_utterance_complete():
            # Process complete utterance
            response = await process_complete_utterance()
            await websocket.send_json({
                "type": "response",
                "data": response
            })
```

## Model Integration

### STT Provider Integration
```python
class STTProvider(ABC):
    @abstractmethod
    async def transcribe(self, audio: bytes) -> str:
        pass
    
    @abstractmethod
    async def stream_transcribe(self, audio_stream) -> AsyncIterator[str]:
        pass

class DeepgramSTT(STTProvider):
    async def transcribe(self, audio: bytes) -> str:
        response = await self.client.transcription.prerecorded({
            'buffer': audio,
            'mimetype': 'audio/wav'
        })
        return response['results']['channels'][0]['alternatives'][0]['transcript']
```

### LLM Integration
```python
class LLMProvider(ABC):
    @abstractmethod
    async def generate_response(self, prompt: str, context: dict) -> str:
        pass

class OpenAILLM(LLMProvider):
    async def generate_response(self, prompt: str, context: dict) -> str:
        response = await self.client.chat.completions.create(
            model="gpt-4-turbo-preview",
            messages=[
                {"role": "system", "content": context.get("system_prompt")},
                {"role": "user", "content": prompt}
            ],
            temperature=0.7
        )
        return response.choices[0].message.content
```

## Performance Optimization

### Caching Strategy
- Model response caching
- Audio preprocessing caching
- Conversation context caching
- Provider response caching

### Async Processing
- Concurrent provider requests
- Streaming audio processing
- Background model loading
- Async database operations

### Resource Management
- GPU memory optimization
- Model quantization
- Batch processing
- Connection pooling

## Monitoring

### Performance Metrics
- STT accuracy and latency
- TTS quality and speed
- LLM response time
- End-to-end latency
- Provider availability

### Quality Metrics
- Conversation success rate
- User satisfaction scores
- Intent recognition accuracy
- Response relevance

### System Metrics
- CPU and GPU utilization
- Memory usage
- Request throughput
- Error rates

## Deployment

### Docker Configuration
```dockerfile
FROM python:3.11-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    ffmpeg \
    libsndfile1 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY . .

EXPOSE 8083
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8083"]
```

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ai-pipeline
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ai-pipeline
  template:
    metadata:
      labels:
        app: ai-pipeline
    spec:
      containers:
      - name: ai-pipeline
        image: agba/ai-pipeline:latest
        ports:
        - containerPort: 8083
        env:
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: ai-secrets
              key: openai-api-key
        resources:
          requests:
            cpu: 2000m
            memory: 4Gi
            nvidia.com/gpu: 1
          limits:
            cpu: 8000m
            memory: 16Gi
            nvidia.com/gpu: 1
```

## Development

### Local Setup
```bash
# Create virtual environment
python -m venv venv
source venv/bin/activate

# Install dependencies
pip install -r requirements.txt

# Set environment variables
export OPENAI_API_KEY=your-key
export DEEPGRAM_API_KEY=your-key

# Run service
uvicorn main:app --reload --port 8083
```

### Testing
```bash
# Unit tests
pytest tests/unit/

# Integration tests
pytest tests/integration/

# Performance tests
pytest tests/performance/

# Model accuracy tests
pytest tests/accuracy/
```

## Security

### API Security
- JWT authentication
- Rate limiting per user
- Input validation and sanitization
- Output filtering for sensitive data

### Model Security
- Prompt injection protection
- Content filtering
- Model access controls
- Audit logging

### Data Protection
- Audio data encryption
- Conversation data anonymization
- Secure model storage
- GDPR compliance

## Troubleshooting

### Common Issues
1. **High Latency**: Check provider response times and model loading
2. **Poor Accuracy**: Verify model configuration and training data
3. **Memory Issues**: Monitor GPU memory and model sizes
4. **Provider Failures**: Implement fallback mechanisms

### Debug Tools
```bash
# Check model status
curl http://localhost:8083/debug/models/status

# Test STT accuracy
curl -X POST http://localhost:8083/debug/stt/test \
  -F "audio=@test.wav"

# Monitor performance
curl http://localhost:8083/metrics
```

## Contributing

1. Follow Python best practices and type hints
2. Write comprehensive tests for AI components
3. Document model changes and performance impacts
4. Ensure provider integration stability
5. Test with diverse audio samples and languages

## Support

For AI/ML pipeline issues:
- Check model logs and performance metrics
- Verify provider API keys and quotas
- Test with known good audio samples
- Contact AI/ML development team