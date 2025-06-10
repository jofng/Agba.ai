package service

import (
    "context"
    "fmt"

    "notification-service/internal/repository"
    "notification-service/internal/providers"
    "github.com/google/uuid"
    "github.com/nats-io/nats.go"
    "go.uber.org/zap"
)

// Service interfaces

// NotificationService interface
type NotificationService interface {
    SendNotification(ctx context.Context, notification *repository.Notification) error
    GetNotification(ctx context.Context, id uuid.UUID) (*repository.Notification, error)
    StartNotificationProcessor(ctx context.Context)
    StartRetryProcessor(ctx context.Context)
    StartCleanupProcessor(ctx context.Context)
    StartNATSConsumer(ctx context.Context)
}

// TemplateService interface
type TemplateService interface {
    CreateTemplate(ctx context.Context, template *repository.Template) error
    GetTemplate(ctx context.Context, id uuid.UUID) (*repository.Template, error)
    UpdateTemplate(ctx context.Context, template *repository.Template) error
    DeleteTemplate(ctx context.Context, id uuid.UUID) error
    ListTemplates(ctx context.Context, limit, offset int) ([]*repository.Template, error)
}

// SubscriptionService interface
type SubscriptionService interface {
    CreateSubscription(ctx context.Context, subscription *repository.Subscription) error
    GetSubscription(ctx context.Context, id uuid.UUID) (*repository.Subscription, error)
    UpdateSubscription(ctx context.Context, subscription *repository.Subscription) error
    DeleteSubscription(ctx context.Context, id uuid.UUID) error
    ListSubscriptions(ctx context.Context, limit, offset int) ([]*repository.Subscription, error)
}

// AuditService interface
type AuditService interface {
    CreateAuditLog(ctx context.Context, log *repository.AuditLog) error
    GetAuditLog(ctx context.Context, id uuid.UUID) (*repository.AuditLog, error)
    ListAuditLogs(ctx context.Context, limit, offset int) ([]*repository.AuditLog, error)
}

// Service implementations

// notificationService implements NotificationService
type notificationService struct {
    notificationRepo repository.NotificationRepository
    templateRepo     repository.TemplateRepository
    subscriptionRepo repository.SubscriptionRepository
    auditRepo        repository.AuditRepository
    emailProvider    providers.EmailProvider
    smsProvider      providers.SMSProvider
    pushProvider     providers.PushProvider
    webhookProvider  providers.WebhookProvider
    natsConn         *nats.Conn
    logger           *zap.Logger
}

// NewNotificationService creates a new notification service
func NewNotificationService(
    notificationRepo repository.NotificationRepository,
    templateRepo repository.TemplateRepository,
    subscriptionRepo repository.SubscriptionRepository,
    auditRepo repository.AuditRepository,
    emailProvider providers.EmailProvider,
    smsProvider providers.SMSProvider,
    pushProvider providers.PushProvider,
    webhookProvider providers.WebhookProvider,
    natsConn *nats.Conn,
    logger *zap.Logger,
) NotificationService {
    return &notificationService{
        notificationRepo: notificationRepo,
        templateRepo:     templateRepo,
        subscriptionRepo: subscriptionRepo,
        auditRepo:        auditRepo,
        emailProvider:    emailProvider,
        smsProvider:      smsProvider,
        pushProvider:     pushProvider,
        webhookProvider:  webhookProvider,
        natsConn:         natsConn,
        logger:           logger,
    }
}

// templateService implements TemplateService
type templateService struct {
    templateRepo repository.TemplateRepository
    logger       *zap.Logger
}

// NewTemplateService creates a new template service
func NewTemplateService(templateRepo repository.TemplateRepository, logger *zap.Logger) TemplateService {
    return &templateService{
        templateRepo: templateRepo,
        logger:       logger,
    }
}

// subscriptionService implements SubscriptionService
type subscriptionService struct {
    subscriptionRepo repository.SubscriptionRepository
    logger           *zap.Logger
}

// NewSubscriptionService creates a new subscription service
func NewSubscriptionService(subscriptionRepo repository.SubscriptionRepository, logger *zap.Logger) SubscriptionService {
    return &subscriptionService{
        subscriptionRepo: subscriptionRepo,
        logger:           logger,
    }
}

// auditService implements AuditService
type auditService struct {
    auditRepo repository.AuditRepository
    logger    *zap.Logger
}

// NewAuditService creates a new audit service
func NewAuditService(auditRepo repository.AuditRepository, logger *zap.Logger) AuditService {
    return &auditService{
        auditRepo: auditRepo,
        logger:    logger,
    }
}

// Legacy service (keeping for backward compatibility)

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

// NotificationService implementation methods

func (s *notificationService) SendNotification(ctx context.Context, notification *repository.Notification) error {
    // Mock implementation
    s.logger.Info("Sending notification (mock)", zap.String("id", notification.ID.String()))
    return nil
}

func (s *notificationService) GetNotification(ctx context.Context, id uuid.UUID) (*repository.Notification, error) {
    // Mock implementation
    s.logger.Info("Getting notification (mock)", zap.String("id", id.String()))
    return s.notificationRepo.GetNotification(ctx, id)
}

func (s *notificationService) StartNotificationProcessor(ctx context.Context) {
    // Mock implementation
    s.logger.Info("Starting notification processor (mock)")
    // In a real implementation, this would start a background goroutine
    // to process notifications from a queue
}

func (s *notificationService) StartRetryProcessor(ctx context.Context) {
    // Mock implementation
    s.logger.Info("Starting retry processor (mock)")
    // In a real implementation, this would start a background goroutine
    // to retry failed notifications
}

func (s *notificationService) StartCleanupProcessor(ctx context.Context) {
    // Mock implementation
    s.logger.Info("Starting cleanup processor (mock)")
    // In a real implementation, this would start a background goroutine
    // to clean up old notifications
}

func (s *notificationService) StartNATSConsumer(ctx context.Context) {
    // Mock implementation
    s.logger.Info("Starting NATS consumer (mock)")
    // In a real implementation, this would start consuming messages from NATS
}

// TemplateService implementation methods

func (s *templateService) CreateTemplate(ctx context.Context, template *repository.Template) error {
    s.logger.Info("Creating template (mock)", zap.String("id", template.ID.String()))
    return s.templateRepo.CreateTemplate(ctx, template)
}

func (s *templateService) GetTemplate(ctx context.Context, id uuid.UUID) (*repository.Template, error) {
    s.logger.Info("Getting template (mock)", zap.String("id", id.String()))
    return s.templateRepo.GetTemplate(ctx, id)
}

func (s *templateService) UpdateTemplate(ctx context.Context, template *repository.Template) error {
    s.logger.Info("Updating template (mock)", zap.String("id", template.ID.String()))
    return s.templateRepo.UpdateTemplate(ctx, template)
}

func (s *templateService) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
    s.logger.Info("Deleting template (mock)", zap.String("id", id.String()))
    return s.templateRepo.DeleteTemplate(ctx, id)
}

func (s *templateService) ListTemplates(ctx context.Context, limit, offset int) ([]*repository.Template, error) {
    s.logger.Info("Listing templates (mock)", zap.Int("limit", limit), zap.Int("offset", offset))
    return s.templateRepo.ListTemplates(ctx, limit, offset)
}

// SubscriptionService implementation methods

func (s *subscriptionService) CreateSubscription(ctx context.Context, subscription *repository.Subscription) error {
    s.logger.Info("Creating subscription (mock)", zap.String("id", subscription.ID.String()))
    return s.subscriptionRepo.CreateSubscription(ctx, subscription)
}

func (s *subscriptionService) GetSubscription(ctx context.Context, id uuid.UUID) (*repository.Subscription, error) {
    s.logger.Info("Getting subscription (mock)", zap.String("id", id.String()))
    return s.subscriptionRepo.GetSubscription(ctx, id)
}

func (s *subscriptionService) UpdateSubscription(ctx context.Context, subscription *repository.Subscription) error {
    s.logger.Info("Updating subscription (mock)", zap.String("id", subscription.ID.String()))
    return s.subscriptionRepo.UpdateSubscription(ctx, subscription)
}

func (s *subscriptionService) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
    s.logger.Info("Deleting subscription (mock)", zap.String("id", id.String()))
    return s.subscriptionRepo.DeleteSubscription(ctx, id)
}

func (s *subscriptionService) ListSubscriptions(ctx context.Context, limit, offset int) ([]*repository.Subscription, error) {
    s.logger.Info("Listing subscriptions (mock)", zap.Int("limit", limit), zap.Int("offset", offset))
    return s.subscriptionRepo.ListSubscriptions(ctx, limit, offset)
}

// AuditService implementation methods

func (s *auditService) CreateAuditLog(ctx context.Context, log *repository.AuditLog) error {
    s.logger.Info("Creating audit log (mock)", zap.String("id", log.ID.String()))
    return s.auditRepo.CreateAuditLog(ctx, log)
}

func (s *auditService) GetAuditLog(ctx context.Context, id uuid.UUID) (*repository.AuditLog, error) {
    s.logger.Info("Getting audit log (mock)", zap.String("id", id.String()))
    return s.auditRepo.GetAuditLog(ctx, id)
}

func (s *auditService) ListAuditLogs(ctx context.Context, limit, offset int) ([]*repository.AuditLog, error) {
    s.logger.Info("Listing audit logs (mock)", zap.Int("limit", limit), zap.Int("offset", offset))
    return s.auditRepo.ListAuditLogs(ctx, limit, offset)
}
