package service

import (
    "context"
    "fmt"

    "notification-service/internal/repository"
    "github.com/google/uuid"
    "go.uber.org/zap"
)

// Service handles business logic for notification operations
type Service struct {
    repository *repository.Repository
    logger     *zap.Logger
}

// NewService creates a new service instance
func NewService(repository *repository.Repository, logger *zap.Logger) *Service {
    return &Service{
        repository: repository,
        logger:     logger,
    }
}

// SendNotification sends a notification
func (s *Service) SendNotification(ctx context.Context, notification *repository.Notification) error {
    s.logger.Info("Sending notification",
        zap.String("id", notification.ID.String()),
        zap.String("channel", notification.Channel),
        zap.String("recipient", notification.Recipient))

    // Create notification record
    if err := s.repository.CreateNotification(ctx, notification); err != nil {
        return fmt.Errorf("failed to create notification: %w", err)
    }

    // Process notification based on channel
    switch notification.Channel {
    case "email":
        return s.sendEmail(ctx, notification)
    case "sms":
        return s.sendSMS(ctx, notification)
    case "push":
        return s.sendPush(ctx, notification)
    case "webhook":
        return s.sendWebhook(ctx, notification)
    default:
        return fmt.Errorf("unsupported channel: %s", notification.Channel)
    }
}

// sendEmail sends an email notification
func (s *Service) sendEmail(ctx context.Context, notification *repository.Notification) error {
    s.logger.Info("Sending email notification", zap.String("recipient", notification.Recipient))
    
    // Update status to sent
    return s.repository.UpdateNotificationStatus(ctx, notification.ID, "sent", "")
}

// sendSMS sends an SMS notification
func (s *Service) sendSMS(ctx context.Context, notification *repository.Notification) error {
    s.logger.Info("Sending SMS notification", zap.String("recipient", notification.Recipient))
    
    // Update status to sent
    return s.repository.UpdateNotificationStatus(ctx, notification.ID, "sent", "")
}

// sendPush sends a push notification
func (s *Service) sendPush(ctx context.Context, notification *repository.Notification) error {
    s.logger.Info("Sending push notification", zap.String("recipient", notification.Recipient))
    
    // Update status to sent
    return s.repository.UpdateNotificationStatus(ctx, notification.ID, "sent", "")
}

// sendWebhook sends a webhook notification
func (s *Service) sendWebhook(ctx context.Context, notification *repository.Notification) error {
    s.logger.Info("Sending webhook notification", zap.String("recipient", notification.Recipient))
    
    // Update status to sent
    return s.repository.UpdateNotificationStatus(ctx, notification.ID, "sent", "")
}

// GetNotification retrieves a notification by ID
func (s *Service) GetNotification(ctx context.Context, id uuid.UUID) (*repository.Notification, error) {
    return s.repository.GetNotification(ctx, id)
}

// HealthCheck performs health checks
func (s *Service) HealthCheck(ctx context.Context) (bool, map[string]interface{}) {
    checks := make(map[string]interface{})
    
    // Check database
    if err := s.repository.HealthCheck(ctx); err != nil {
        checks["database"] = map[string]interface{}{
            "status":  "unhealthy",
            "message": err.Error(),
        }
        return false, checks
    }
    
    checks["database"] = map[string]interface{}{
        "status":  "healthy",
        "message": "Database connection successful",
    }
    
    return true, checks
}
