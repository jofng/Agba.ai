# Notification Service

The Notification Service provides comprehensive notification delivery capabilities for the Agba Voice AI Platform. It handles email, SMS, push notifications, and webhooks with advanced features like templating, scheduling, delivery tracking, and multi-provider support.

## Overview

The Notification Service offers centralized notification management with:
- **Multi-Channel Delivery** supporting email, SMS, push notifications, and webhooks
- **Template Management** with dynamic content and localization support
- **Delivery Tracking** with real-time status updates and analytics
- **Provider Integration** with multiple email, SMS, and push notification providers
- **Subscription Management** with user preferences and unsubscribe handling
- **Campaign Management** for bulk notification delivery
- **Queue Processing** with priority handling and retry mechanisms
- **Audit Logging** for compliance and delivery verification

## Architecture

### Technology Stack
- **Language**: Go (Golang) 1.21
- **Framework**: Gin HTTP framework for REST API
- **Database**: PostgreSQL for persistent storage of notifications, templates, and audit logs
- **Cache**: Redis for queue management and rate limiting
- **Messaging**: NATS for real-time event processing and queue management
- **Providers**: Multiple notification providers for redundancy and optimization
- **Templates**: Dynamic template engine with variable substitution and localization

### Key Components
1. **Notification Engine**: Core notification processing and delivery
2. **Template Manager**: Dynamic template rendering and management
3. **Provider Manager**: Multi-provider integration and failover
4. **Queue Processor**: Asynchronous notification processing with priorities
5. **Subscription Manager**: User preference and subscription management
6. **Campaign Manager**: Bulk notification campaign execution
7. **Audit Logger**: Comprehensive delivery tracking and compliance logging

## Features

### Multi-Channel Notification Delivery
- **Email notifications** with HTML/text support and attachments
- **SMS notifications** with international support and delivery reports
- **Push notifications** for mobile and web applications
- **Webhook notifications** for system-to-system communication
- **Template-based messaging** with dynamic content rendering
- **Bulk delivery** with batch processing and rate limiting

### Advanced Template Management
- **Dynamic templates** with variable substitution and conditional logic
- **Multi-language support** with localization and internationalization
- **Template versioning** with rollback capabilities
- **Preview and testing** functionality for template validation
- **Rich content support** including HTML, images, and attachments
- **Template inheritance** and reusable components

## Configuration

### Environment Variables

#### Server Configuration
```bash
AGBA_NOTIFICATION_ENVIRONMENT=production       # Environment (development/production)
AGBA_NOTIFICATION_VERSION=1.0.0               # Service version
AGBA_NOTIFICATION_SERVER_PORT=8080            # HTTP server port
```

#### Database Configuration
```bash
AGBA_NOTIFICATION_DATABASE_DSN=postgresql://user:pass@host:5432/db # Database connection string
AGBA_NOTIFICATION_REDIS_ADDRESS=redis-master:6379 # Redis server address
AGBA_NOTIFICATION_NATS_URL=nats://nats-client:4222 # NATS server URL
```

#### Email Configuration
```bash
AGBA_NOTIFICATION_EMAIL_ENABLED=true          # Enable email notifications
AGBA_NOTIFICATION_EMAIL_PROVIDER=smtp         # Email provider
AGBA_NOTIFICATION_EMAIL_SMTP_HOST=smtp.gmail.com # SMTP server host
AGBA_NOTIFICATION_EMAIL_FROM_EMAIL=noreply@agba.ai # From email address
```

## API Endpoints

### Notification Management
```
POST   /api/v1/notifications                    # Send single notification
POST   /api/v1/notifications/bulk               # Send bulk notifications
GET    /api/v1/notifications                    # List notifications
GET    /api/v1/notifications/{id}               # Get notification details
```

### Email Notifications
```
POST   /api/v1/email/send                       # Send email
POST   /api/v1/email/send-template              # Send templated email
POST   /api/v1/email/send-bulk                  # Send bulk emails
```

### Template Management
```
GET    /api/v1/templates                        # List templates
POST   /api/v1/templates                        # Create template
PUT    /api/v1/templates/{id}                   # Update template
```

## Development

### Local Development Setup
```bash
# Clone repository
git clone https://github.com/ibilola104/Agba.ai.git
cd Agba.ai/services/notification-service

# Install dependencies
go mod download

# Set required environment variables
export AGBA_NOTIFICATION_DATABASE_DSN=postgresql://localhost:5432/agba_core
export AGBA_NOTIFICATION_REDIS_ADDRESS=localhost:6379
export AGBA_NOTIFICATION_NATS_URL=nats://localhost:4222

# Run service
go run cmd/main.go  # or python main.py
```

## Deployment

Standard containerized deployment with Kubernetes.

## Contributing

Follow platform development standards and best practices.

## Support

Contact the development team for notification related issues.
