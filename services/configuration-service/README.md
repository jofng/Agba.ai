# Configuration Service

The Configuration Service manages all agent configurations, workflow definitions, templates, and system settings for the Agba Voice AI Platform. It provides centralized configuration management with version control and validation.

## Overview

This service provides:
- Agent configuration management
- Workflow definition storage
- Template management system
- Configuration version control
- Validation and testing
- Environment-specific configurations

## Architecture

### Technology Stack
- **Language**: Go (Golang)
- **Framework**: Gin HTTP framework
- **Database**: PostgreSQL for configuration storage
- **Cache**: Redis for configuration caching
- **Validation**: JSON Schema validation
- **Version Control**: Git-like versioning system

### Key Components
1. **Config Manager**: Configuration CRUD operations
2. **Version Controller**: Configuration versioning
3. **Validator**: Configuration validation
4. **Template Engine**: Template management
5. **Cache Manager**: Configuration caching
6. **Backup Service**: Configuration backup and restore

## Features

### Agent Configuration
- AI agent personality settings
- Voice and speech parameters
- Conversation flow definitions
- Integration configurations
- Performance tuning parameters

### Workflow Management
- Visual workflow definitions
- Decision tree configurations
- Conditional logic rules
- External system integrations
- Error handling workflows

### Template System
- Pre-built agent templates
- Industry-specific configurations
- Use case templates
- Custom template creation
- Template marketplace

### Version Control
- Configuration versioning
- Change tracking
- Rollback capabilities
- Diff visualization
- Approval workflows

## Configuration

### Environment Variables
```bash
# Server Configuration
PORT=8085
HOST=0.0.0.0
ENV=production

# Database
POSTGRES_URL=postgresql://user:pass@localhost:5432/agba_core
REDIS_URL=redis://localhost:6379

# Validation
SCHEMA_VALIDATION_ENABLED=true
STRICT_VALIDATION=true

# Backup
BACKUP_ENABLED=true
BACKUP_INTERVAL=24h
BACKUP_RETENTION=30d
```

## API Endpoints

### Agent Configuration
```
GET    /api/v1/agents                    # List agents
POST   /api/v1/agents                    # Create agent
GET    /api/v1/agents/{id}               # Get agent config
PUT    /api/v1/agents/{id}               # Update agent config
DELETE /api/v1/agents/{id}               # Delete agent
POST   /api/v1/agents/{id}/validate      # Validate config
POST   /api/v1/agents/{id}/deploy        # Deploy config
```

### Workflow Management
```
GET    /api/v1/workflows                 # List workflows
POST   /api/v1/workflows                 # Create workflow
GET    /api/v1/workflows/{id}            # Get workflow
PUT    /api/v1/workflows/{id}            # Update workflow
DELETE /api/v1/workflows/{id}            # Delete workflow
POST   /api/v1/workflows/{id}/test       # Test workflow
```

### Templates
```
GET    /api/v1/templates                 # List templates
GET    /api/v1/templates/{id}            # Get template
POST   /api/v1/templates/{id}/apply      # Apply template
GET    /api/v1/templates/categories      # Template categories
```

### Version Control
```
GET    /api/v1/configs/{id}/versions     # List versions
GET    /api/v1/configs/{id}/versions/{v} # Get specific version
POST   /api/v1/configs/{id}/rollback     # Rollback to version
GET    /api/v1/configs/{id}/diff         # Compare versions
```

## Configuration Schema

### Agent Configuration
```json
{
  "agent_id": "agent-123",
  "name": "Customer Support Agent",
  "description": "AI agent for customer support",
  "version": "1.0.0",
  "personality": {
    "tone": "friendly",
    "formality": "casual",
    "empathy_level": 0.8,
    "patience_level": 0.9
  },
  "voice": {
    "provider": "elevenlabs",
    "voice_id": "21m00Tcm4TlvDq8ikWAM",
    "speed": 1.0,
    "pitch": 0,
    "stability": 0.5
  },
  "llm": {
    "provider": "openai",
    "model": "gpt-4-turbo-preview",
    "temperature": 0.7,
    "max_tokens": 1000,
    "system_prompt": "You are a helpful customer support agent..."
  },
  "workflow": {
    "greeting": "Hello! How can I help you today?",
    "max_turns": 20,
    "timeout": 300,
    "fallback_message": "I'm sorry, I didn't understand that."
  },
  "integrations": {
    "crm": {
      "enabled": true,
      "provider": "salesforce",
      "config": {...}
    }
  }
}
```

### Workflow Definition
```json
{
  "workflow_id": "support-workflow-v1",
  "name": "Customer Support Workflow",
  "version": "1.0.0",
  "nodes": [
    {
      "id": "greeting",
      "type": "message",
      "content": "Hello! How can I help you today?",
      "next": "intent_detection"
    },
    {
      "id": "intent_detection",
      "type": "nlu",
      "intents": ["billing", "technical", "general"],
      "next": {
        "billing": "billing_flow",
        "technical": "technical_flow",
        "default": "general_flow"
      }
    }
  ],
  "variables": {
    "customer_id": null,
    "issue_type": null,
    "resolution_status": "pending"
  }
}
```

## Configuration Management

### CRUD Operations
```go
type ConfigService struct {
    db    *gorm.DB
    cache *redis.Client
    validator *Validator
}

func (c *ConfigService) CreateAgent(config *AgentConfig) error {
    // Validate configuration
    if err := c.validator.ValidateAgent(config); err != nil {
        return err
    }
    
    // Save to database
    if err := c.db.Create(config).Error; err != nil {
        return err
    }
    
    // Cache configuration
    return c.cache.Set(config.ID, config, time.Hour).Err()
}

func (c *ConfigService) GetAgent(id string) (*AgentConfig, error) {
    // Try cache first
    if cached, err := c.cache.Get(id).Result(); err == nil {
        var config AgentConfig
        json.Unmarshal([]byte(cached), &config)
        return &config, nil
    }
    
    // Fallback to database
    var config AgentConfig
    err := c.db.Where("id = ?", id).First(&config).Error
    return &config, err
}
```

### Version Control
```go
type VersionController struct {
    db *gorm.DB
}

func (v *VersionController) CreateVersion(configID string, config interface{}) error {
    version := &ConfigVersion{
        ConfigID:  configID,
        Version:   v.getNextVersion(configID),
        Data:      config,
        CreatedAt: time.Now(),
        CreatedBy: getCurrentUser(),
    }
    
    return v.db.Create(version).Error
}

func (v *VersionController) Rollback(configID, version string) error {
    // Get target version
    var targetVersion ConfigVersion
    err := v.db.Where("config_id = ? AND version = ?", configID, version).First(&targetVersion).Error
    if err != nil {
        return err
    }
    
    // Create new version with old data
    return v.CreateVersion(configID, targetVersion.Data)
}
```

### Validation
```go
type Validator struct {
    schemas map[string]*jsonschema.Schema
}

func (v *Validator) ValidateAgent(config *AgentConfig) error {
    schema := v.schemas["agent"]
    
    // Convert to JSON for validation
    data, _ := json.Marshal(config)
    
    // Validate against schema
    result := schema.Validate(bytes.NewReader(data))
    if !result.Valid() {
        return fmt.Errorf("validation failed: %v", result.Errors())
    }
    
    // Custom business logic validation
    return v.validateBusinessRules(config)
}
```

## Template System

### Template Management
```go
type TemplateManager struct {
    db *gorm.DB
    storage FileStorage
}

func (t *TemplateManager) ApplyTemplate(templateID, targetID string, params map[string]interface{}) error {
    // Get template
    template, err := t.GetTemplate(templateID)
    if err != nil {
        return err
    }
    
    // Render template with parameters
    rendered, err := t.renderTemplate(template, params)
    if err != nil {
        return err
    }
    
    // Apply to target configuration
    return t.configService.UpdateConfig(targetID, rendered)
}
```

### Template Categories
- **Customer Support**: Support agent templates
- **Sales**: Sales agent configurations
- **Healthcare**: HIPAA-compliant templates
- **E-commerce**: Shopping assistant templates
- **Education**: Educational assistant templates

## Performance Optimization

### Caching Strategy
- Configuration caching with TTL
- Template caching
- Schema caching
- Version metadata caching

### Database Optimization
- Indexed queries for fast retrieval
- Connection pooling
- Query optimization
- Batch operations

### Validation Optimization
- Schema compilation caching
- Parallel validation
- Incremental validation
- Validation result caching

## Security

### Access Control
- Role-based configuration access
- Environment-specific permissions
- Audit logging for changes
- Approval workflows for production

### Data Protection
- Configuration encryption at rest
- Secure API key storage
- Sensitive data masking
- Backup encryption

### Validation Security
- Input sanitization
- Schema injection prevention
- Safe template rendering
- Resource limit enforcement

## Monitoring

### Configuration Metrics
- Configuration change frequency
- Validation success rate
- Template usage statistics
- Version rollback frequency

### Performance Metrics
- Configuration retrieval time
- Validation processing time
- Cache hit rate
- Database query performance

### Quality Metrics
- Configuration error rate
- Template application success
- User satisfaction scores
- System stability impact

## Deployment

### Docker Configuration
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o configuration-service ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/configuration-service .
COPY --from=builder /app/schemas ./schemas
EXPOSE 8085
CMD ["./configuration-service"]
```

## Development

### Local Setup
```bash
# Install dependencies
go mod download

# Set environment variables
export POSTGRES_URL=postgresql://localhost:5432/agba_core
export REDIS_URL=redis://localhost:6379

# Run migrations
go run migrations/main.go

# Start service
go run cmd/main.go
```

### Testing
```bash
# Unit tests
go test ./...

# Integration tests
go test -tags=integration ./...

# Configuration validation tests
go test -tags=validation ./...
```

## Contributing

1. Follow configuration schema standards
2. Validate all configuration changes
3. Document template modifications
4. Test version control operations
5. Ensure backward compatibility

## Support

For configuration issues:
- Check validation error messages
- Verify schema compliance
- Test configuration changes in staging
- Contact configuration team for assistance