# Analytics Service

The Analytics Service provides real-time analytics, sentiment analysis, topic extraction, and business intelligence for the Agba Voice AI Platform. It processes conversation data to generate actionable insights and performance metrics.

## Overview

This service provides:
- Real-time conversation analytics
- Sentiment analysis and emotion detection
- Topic extraction and categorization
- Performance metrics and KPIs
- Business intelligence dashboards
- Predictive analytics

## Architecture

### Technology Stack
- **Language**: Python 3.11+
- **Framework**: FastAPI for APIs
- **Analytics**: Pandas, NumPy, scikit-learn
- **NLP**: spaCy, NLTK, Transformers
- **Time Series**: TimescaleDB, InfluxDB
- **Streaming**: Apache Kafka, NATS
- **Visualization**: Plotly, Matplotlib

### Key Components
1. **Event Processor**: Real-time event ingestion
2. **Sentiment Analyzer**: Emotion and sentiment detection
3. **Topic Extractor**: Conversation topic identification
4. **Metrics Calculator**: KPI computation
5. **Report Generator**: Automated reporting
6. **Dashboard API**: Real-time dashboard data

## Features

### Real-time Analytics
- Live conversation monitoring
- Real-time sentiment tracking
- Performance metric calculation
- Alert generation
- Trend analysis

### Sentiment Analysis
- Emotion detection (joy, anger, sadness, fear)
- Sentiment scoring (-1 to +1)
- Customer satisfaction prediction
- Agent performance evaluation
- Escalation prediction

### Topic Extraction
- Automatic topic identification
- Conversation categorization
- Keyword extraction
- Intent clustering
- Content analysis

### Business Intelligence
- Call volume analytics
- Revenue attribution
- Customer journey analysis
- Agent productivity metrics
- Operational efficiency reports

## Configuration

### Environment Variables
```bash
# Server Configuration
PORT=8084
HOST=0.0.0.0
ENV=production

# Database
POSTGRES_URL=postgresql://user:pass@localhost:5432/agba_analytics
TIMESCALEDB_URL=postgresql://user:pass@localhost:5432/agba_timeseries
REDIS_URL=redis://localhost:6379

# Streaming
KAFKA_BROKERS=kafka1:9092,kafka2:9092
NATS_URL=nats://nats:4222

# ML Models
SENTIMENT_MODEL=cardiffnlp/twitter-roberta-base-sentiment-latest
TOPIC_MODEL=all-MiniLM-L6-v2
EMOTION_MODEL=j-hartmann/emotion-english-distilroberta-base
```

## API Endpoints

### Real-time Analytics
```
GET  /api/v1/analytics/realtime     # Real-time metrics
GET  /api/v1/analytics/sentiment    # Sentiment trends
GET  /api/v1/analytics/topics       # Topic analysis
POST /api/v1/analytics/events       # Submit analytics event
```

### Reports
```
GET  /api/v1/reports/daily          # Daily reports
GET  /api/v1/reports/weekly         # Weekly reports
GET  /api/v1/reports/custom         # Custom date range
POST /api/v1/reports/generate       # Generate report
```

### Dashboards
```
GET /api/v1/dashboards/overview     # Overview dashboard
GET /api/v1/dashboards/agents       # Agent performance
GET /api/v1/dashboards/customers    # Customer analytics
```

## Analytics Pipeline

### Event Processing
```python
async def process_conversation_event(event: ConversationEvent):
    # Extract conversation data
    conversation_id = event.conversation_id
    transcript = event.transcript
    metadata = event.metadata
    
    # Sentiment analysis
    sentiment = await sentiment_analyzer.analyze(transcript)
    
    # Topic extraction
    topics = await topic_extractor.extract(transcript)
    
    # Store analytics data
    await analytics_store.save({
        'conversation_id': conversation_id,
        'sentiment': sentiment,
        'topics': topics,
        'timestamp': event.timestamp
    })
    
    # Update real-time metrics
    await metrics_updater.update(sentiment, topics)
```

### Sentiment Analysis
```python
class SentimentAnalyzer:
    def __init__(self):
        self.model = pipeline(
            "sentiment-analysis",
            model="cardiffnlp/twitter-roberta-base-sentiment-latest"
        )
    
    async def analyze(self, text: str) -> SentimentResult:
        result = self.model(text)
        return SentimentResult(
            label=result[0]['label'],
            score=result[0]['score'],
            emotions=await self.extract_emotions(text)
        )
```

### Topic Extraction
```python
class TopicExtractor:
    def __init__(self):
        self.model = SentenceTransformer('all-MiniLM-L6-v2')
        self.clusterer = KMeans(n_clusters=10)
    
    async def extract(self, text: str) -> List[Topic]:
        # Extract embeddings
        embeddings = self.model.encode([text])
        
        # Cluster topics
        cluster = self.clusterer.predict(embeddings)[0]
        
        # Extract keywords
        keywords = self.extract_keywords(text)
        
        return [Topic(
            cluster_id=cluster,
            keywords=keywords,
            confidence=0.85
        )]
```

## Metrics and KPIs

### Call Metrics
- Total call volume
- Average call duration
- Call success rate
- Call abandonment rate
- Peak hour analysis

### Sentiment Metrics
- Average sentiment score
- Sentiment distribution
- Sentiment trends over time
- Customer satisfaction index
- Escalation prediction accuracy

### Agent Metrics
- Agent performance scores
- Average handling time
- Resolution rate
- Customer feedback scores
- Productivity metrics

### Business Metrics
- Revenue per call
- Cost per interaction
- Customer lifetime value
- Conversion rates
- ROI analysis

## Real-time Processing

### Stream Processing
```python
async def process_event_stream():
    async for event in kafka_consumer:
        try:
            # Parse event
            conversation_event = ConversationEvent.parse(event.value)
            
            # Process analytics
            analytics_result = await process_conversation_event(conversation_event)
            
            # Update real-time dashboard
            await dashboard_updater.update(analytics_result)
            
            # Check for alerts
            await alert_manager.check_thresholds(analytics_result)
            
        except Exception as e:
            logger.error(f"Error processing event: {e}")
```

### Alert System
```python
class AlertManager:
    async def check_thresholds(self, analytics_result: AnalyticsResult):
        # Check sentiment threshold
        if analytics_result.sentiment.score < -0.7:
            await self.send_alert(
                type="negative_sentiment",
                severity="high",
                data=analytics_result
            )
        
        # Check call volume threshold
        if analytics_result.call_volume > self.thresholds.max_volume:
            await self.send_alert(
                type="high_volume",
                severity="medium",
                data=analytics_result
            )
```

## Performance Optimization

### Data Processing
- Batch processing for historical data
- Stream processing for real-time data
- Parallel processing for large datasets
- Caching for frequently accessed metrics

### Model Optimization
- Model quantization for faster inference
- Batch prediction for efficiency
- Model caching and preloading
- GPU acceleration for large models

### Database Optimization
- Time-series partitioning
- Indexed queries for fast retrieval
- Materialized views for aggregations
- Connection pooling

## Deployment

### Docker Configuration
```dockerfile
FROM python:3.11-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Download ML models
RUN python -c "from transformers import pipeline; pipeline('sentiment-analysis', model='cardiffnlp/twitter-roberta-base-sentiment-latest')"

COPY . .

EXPOSE 8084
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8084"]
```

## Monitoring

### Performance Metrics
- Event processing rate
- Model inference time
- Database query performance
- Memory and CPU usage
- Alert response time

### Quality Metrics
- Sentiment analysis accuracy
- Topic extraction relevance
- Prediction accuracy
- Data completeness
- Model drift detection

## Security

### Data Protection
- Anonymization of sensitive data
- Encryption of analytics data
- Access control for reports
- Audit logging
- GDPR compliance

### Model Security
- Model versioning and validation
- Input sanitization
- Output filtering
- Secure model storage
- Regular security updates

## Contributing

1. Follow data science best practices
2. Validate model accuracy with test datasets
3. Document analytics methodologies
4. Ensure data privacy compliance
5. Performance test with large datasets

## Support

For analytics issues:
- Check data pipeline logs
- Verify model performance metrics
- Test with sample data
- Contact analytics team for support