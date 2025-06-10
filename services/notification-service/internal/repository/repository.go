package repository

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "time"

    "notification-service/internal/config"
    "github.com/google/uuid"
    "github.com/go-redis/redis/v8"
    "github.com/nats-io/nats.go"
    "go.uber.org/zap"
    _ "github.com/lib/pq" // PostgreSQL driver
)

// Database connection constructors

// NewDatabase creates a new database connection
func NewDatabase(cfg config.DatabaseConfig, logger *zap.Logger) (*sql.DB, error) {
    db, err := sql.Open(cfg.Driver, cfg.DSN)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    db.SetMaxOpenConns(cfg.MaxOpenConns)
    db.SetMaxIdleConns(cfg.MaxIdleConns)
    db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    logger.Info("Database connection established",
        zap.String("driver", cfg.Driver),
        zap.Int("max_open_conns", cfg.MaxOpenConns),
        zap.Int("max_idle_conns", cfg.MaxIdleConns),
    )

    return db, nil
}

// NewRedis creates a new Redis client
func NewRedis(cfg config.RedisConfig, logger *zap.Logger) (*redis.Client, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     cfg.Address,
        Password: cfg.Password,
        DB:       cfg.DB,
        PoolSize: cfg.PoolSize,
    })

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := client.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("failed to connect to Redis: %w", err)
    }

    logger.Info("Redis connection established",
        zap.String("address", cfg.Address),
        zap.Int("db", cfg.DB),
        zap.Int("pool_size", cfg.PoolSize),
    )

    return client, nil
}

// NewNATS creates a new NATS connection
func NewNATS(cfg config.NATSConfig, logger *zap.Logger) (*nats.Conn, error) {
    opts := []nats.Option{
        nats.Name("notification-service"),
        nats.UserInfo(cfg.Username, cfg.Password),
        nats.ReconnectWait(time.Second),
        nats.MaxReconnects(-1),
    }

    nc, err := nats.Connect(cfg.URL, opts...)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to NATS: %w", err)
    }

    logger.Info("NATS connection established",
        zap.String("url", cfg.URL),
        zap.String("username", cfg.Username),
    )

    return nc, nil
}

// Repository interfaces and implementations

// NotificationRepository interface
type NotificationRepository interface {
    CreateNotification(ctx context.Context, notification *Notification) error
    GetNotification(ctx context.Context, id uuid.UUID) (*Notification, error)
    UpdateNotification(ctx context.Context, notification *Notification) error
    DeleteNotification(ctx context.Context, id uuid.UUID) error
    ListNotifications(ctx context.Context, limit, offset int) ([]*Notification, error)
}

// TemplateRepository interface
type TemplateRepository interface {
    CreateTemplate(ctx context.Context, template *Template) error
    GetTemplate(ctx context.Context, id uuid.UUID) (*Template, error)
    UpdateTemplate(ctx context.Context, template *Template) error
    DeleteTemplate(ctx context.Context, id uuid.UUID) error
    ListTemplates(ctx context.Context, limit, offset int) ([]*Template, error)
}

// SubscriptionRepository interface
type SubscriptionRepository interface {
    CreateSubscription(ctx context.Context, subscription *Subscription) error
    GetSubscription(ctx context.Context, id uuid.UUID) (*Subscription, error)
    UpdateSubscription(ctx context.Context, subscription *Subscription) error
    DeleteSubscription(ctx context.Context, id uuid.UUID) error
    ListSubscriptions(ctx context.Context, limit, offset int) ([]*Subscription, error)
}

// AuditRepository interface
type AuditRepository interface {
    CreateAuditLog(ctx context.Context, log *AuditLog) error
    GetAuditLog(ctx context.Context, id uuid.UUID) (*AuditLog, error)
    ListAuditLogs(ctx context.Context, limit, offset int) ([]*AuditLog, error)
}

// Repository implementations

// Repository handles data operations for notification service
type Repository struct {
    db     *sql.DB
    logger *zap.Logger
}

// NewRepository creates a new repository instance
func NewRepository(db *sql.DB, logger *zap.Logger) *Repository {
    return &Repository{
        db:     db,
        logger: logger,
    }
}

// Repository constructor functions

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *sql.DB, logger *zap.Logger) NotificationRepository {
    return &notificationRepository{
        db:     db,
        logger: logger,
    }
}

// NewTemplateRepository creates a new template repository
func NewTemplateRepository(db *sql.DB, logger *zap.Logger) TemplateRepository {
    return &templateRepository{
        db:     db,
        logger: logger,
    }
}

// NewSubscriptionRepository creates a new subscription repository
func NewSubscriptionRepository(db *sql.DB, redis *redis.Client, logger *zap.Logger) SubscriptionRepository {
    return &subscriptionRepository{
        db:     db,
        redis:  redis,
        logger: logger,
    }
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *sql.DB, logger *zap.Logger) AuditRepository {
    return &auditRepository{
        db:     db,
        logger: logger,
    }
}

// Notification represents a notification record
type Notification struct {
    ID          uuid.UUID              `json:"id"`
    TemplateID  *uuid.UUID             `json:"template_id,omitempty"`
    Recipient   string                 `json:"recipient"`
    Channel     string                 `json:"channel"`
    Subject     string                 `json:"subject,omitempty"`
    Body        string                 `json:"body"`
    Status      string                 `json:"status"`
    ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
    SentAt      *time.Time             `json:"sent_at,omitempty"`
    DeliveredAt *time.Time             `json:"delivered_at,omitempty"`
    FailedAt    *time.Time             `json:"failed_at,omitempty"`
    ErrorMessage string                `json:"error_message,omitempty"`
    RetryCount  int                    `json:"retry_count"`
    Metadata    map[string]interface{} `json:"metadata"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}

// CreateNotification creates a new notification
func (r *Repository) CreateNotification(ctx context.Context, notification *Notification) error {
    query := `
        INSERT INTO notifications (id, template_id, recipient, channel, subject, body, status, scheduled_at, metadata, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

    metadataJSON, err := json.Marshal(notification.Metadata)
    if err != nil {
        return fmt.Errorf("failed to marshal metadata: %w", err)
    }

    now := time.Now()
    notification.CreatedAt = now
    notification.UpdatedAt = now

    _, err = r.db.ExecContext(ctx, query,
        notification.ID, notification.TemplateID, notification.Recipient,
        notification.Channel, notification.Subject, notification.Body,
        notification.Status, notification.ScheduledAt, metadataJSON, now, now)

    if err != nil {
        r.logger.Error("Failed to create notification", zap.Error(err))
        return fmt.Errorf("failed to create notification: %w", err)
    }

    return nil
}

// GetNotification retrieves a notification by ID
func (r *Repository) GetNotification(ctx context.Context, id uuid.UUID) (*Notification, error) {
    query := `
        SELECT id, template_id, recipient, channel, subject, body, status, 
               scheduled_at, sent_at, delivered_at, failed_at, error_message, 
               retry_count, metadata, created_at, updated_at
        FROM notifications WHERE id = $1`

    var notification Notification
    var metadataJSON []byte
    var templateID, scheduledAt, sentAt, deliveredAt, failedAt sql.NullTime

    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &notification.ID, &templateID, &notification.Recipient,
        &notification.Channel, &notification.Subject, &notification.Body,
        &notification.Status, &scheduledAt, &sentAt, &deliveredAt,
        &failedAt, &notification.ErrorMessage, &notification.RetryCount,
        &metadataJSON, &notification.CreatedAt, &notification.UpdatedAt)

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("notification not found: %s", id.String())
        }
        return nil, fmt.Errorf("failed to get notification: %w", err)
    }

    // Handle nullable fields
    if scheduledAt.Valid {
        notification.ScheduledAt = &scheduledAt.Time
    }
    if sentAt.Valid {
        notification.SentAt = &sentAt.Time
    }
    if deliveredAt.Valid {
        notification.DeliveredAt = &deliveredAt.Time
    }
    if failedAt.Valid {
        notification.FailedAt = &failedAt.Time
    }

    // Unmarshal metadata
    if err := json.Unmarshal(metadataJSON, &notification.Metadata); err != nil {
        r.logger.Warn("Failed to unmarshal metadata", zap.Error(err))
        notification.Metadata = make(map[string]interface{})
    }

    return &notification, nil
}

// UpdateNotificationStatus updates the status of a notification
func (r *Repository) UpdateNotificationStatus(ctx context.Context, id uuid.UUID, status string, errorMessage string) error {
    query := `
        UPDATE notifications 
        SET status = $2, error_message = $3, updated_at = $4
        WHERE id = $1`

    _, err := r.db.ExecContext(ctx, query, id, status, errorMessage, time.Now())
    if err != nil {
        r.logger.Error("Failed to update notification status", zap.Error(err))
        return fmt.Errorf("failed to update notification status: %w", err)
    }

    return nil
}

// HealthCheck checks repository health
func (r *Repository) HealthCheck(ctx context.Context) error {
    return r.db.PingContext(ctx)
}

// Additional data structures

// Template represents a notification template
type Template struct {
    ID          uuid.UUID              `json:"id"`
    Name        string                 `json:"name"`
    Subject     string                 `json:"subject,omitempty"`
    Body        string                 `json:"body"`
    Channel     string                 `json:"channel"`
    Variables   []string               `json:"variables"`
    Metadata    map[string]interface{} `json:"metadata"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}

// Subscription represents a notification subscription
type Subscription struct {
    ID          uuid.UUID              `json:"id"`
    UserID      uuid.UUID              `json:"user_id"`
    Channel     string                 `json:"channel"`
    Endpoint    string                 `json:"endpoint"`
    Active      bool                   `json:"active"`
    Preferences map[string]interface{} `json:"preferences"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
    ID            uuid.UUID              `json:"id"`
    NotificationID *uuid.UUID            `json:"notification_id,omitempty"`
    Action        string                 `json:"action"`
    UserID        *uuid.UUID             `json:"user_id,omitempty"`
    Details       map[string]interface{} `json:"details"`
    CreatedAt     time.Time              `json:"created_at"`
}

// Repository implementations

// notificationRepository implements NotificationRepository
type notificationRepository struct {
    db     *sql.DB
    logger *zap.Logger
}

func (r *notificationRepository) CreateNotification(ctx context.Context, notification *Notification) error {
    // Mock implementation
    r.logger.Info("Creating notification (mock)", zap.String("id", notification.ID.String()))
    return nil
}

func (r *notificationRepository) GetNotification(ctx context.Context, id uuid.UUID) (*Notification, error) {
    // Mock implementation
    r.logger.Info("Getting notification (mock)", zap.String("id", id.String()))
    return &Notification{
        ID:        id,
        Recipient: "mock@example.com",
        Channel:   "email",
        Subject:   "Mock Notification",
        Body:      "This is a mock notification",
        Status:    "sent",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }, nil
}

func (r *notificationRepository) UpdateNotification(ctx context.Context, notification *Notification) error {
    // Mock implementation
    r.logger.Info("Updating notification (mock)", zap.String("id", notification.ID.String()))
    return nil
}

func (r *notificationRepository) DeleteNotification(ctx context.Context, id uuid.UUID) error {
    // Mock implementation
    r.logger.Info("Deleting notification (mock)", zap.String("id", id.String()))
    return nil
}

func (r *notificationRepository) ListNotifications(ctx context.Context, limit, offset int) ([]*Notification, error) {
    // Mock implementation
    r.logger.Info("Listing notifications (mock)", zap.Int("limit", limit), zap.Int("offset", offset))
    return []*Notification{}, nil
}

// templateRepository implements TemplateRepository
type templateRepository struct {
    db     *sql.DB
    logger *zap.Logger
}

func (r *templateRepository) CreateTemplate(ctx context.Context, template *Template) error {
    // Mock implementation
    r.logger.Info("Creating template (mock)", zap.String("id", template.ID.String()))
    return nil
}

func (r *templateRepository) GetTemplate(ctx context.Context, id uuid.UUID) (*Template, error) {
    // Mock implementation
    r.logger.Info("Getting template (mock)", zap.String("id", id.String()))
    return &Template{
        ID:        id,
        Name:      "Mock Template",
        Subject:   "{{subject}}",
        Body:      "Hello {{name}}, this is a mock template",
        Channel:   "email",
        Variables: []string{"subject", "name"},
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }, nil
}

func (r *templateRepository) UpdateTemplate(ctx context.Context, template *Template) error {
    // Mock implementation
    r.logger.Info("Updating template (mock)", zap.String("id", template.ID.String()))
    return nil
}

func (r *templateRepository) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
    // Mock implementation
    r.logger.Info("Deleting template (mock)", zap.String("id", id.String()))
    return nil
}

func (r *templateRepository) ListTemplates(ctx context.Context, limit, offset int) ([]*Template, error) {
    // Mock implementation
    r.logger.Info("Listing templates (mock)", zap.Int("limit", limit), zap.Int("offset", offset))
    return []*Template{}, nil
}

// subscriptionRepository implements SubscriptionRepository
type subscriptionRepository struct {
    db     *sql.DB
    redis  *redis.Client
    logger *zap.Logger
}

func (r *subscriptionRepository) CreateSubscription(ctx context.Context, subscription *Subscription) error {
    // Mock implementation
    r.logger.Info("Creating subscription (mock)", zap.String("id", subscription.ID.String()))
    return nil
}

func (r *subscriptionRepository) GetSubscription(ctx context.Context, id uuid.UUID) (*Subscription, error) {
    // Mock implementation
    r.logger.Info("Getting subscription (mock)", zap.String("id", id.String()))
    return &Subscription{
        ID:       id,
        UserID:   uuid.New(),
        Channel:  "email",
        Endpoint: "user@example.com",
        Active:   true,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }, nil
}

func (r *subscriptionRepository) UpdateSubscription(ctx context.Context, subscription *Subscription) error {
    // Mock implementation
    r.logger.Info("Updating subscription (mock)", zap.String("id", subscription.ID.String()))
    return nil
}

func (r *subscriptionRepository) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
    // Mock implementation
    r.logger.Info("Deleting subscription (mock)", zap.String("id", id.String()))
    return nil
}

func (r *subscriptionRepository) ListSubscriptions(ctx context.Context, limit, offset int) ([]*Subscription, error) {
    // Mock implementation
    r.logger.Info("Listing subscriptions (mock)", zap.Int("limit", limit), zap.Int("offset", offset))
    return []*Subscription{}, nil
}

// auditRepository implements AuditRepository
type auditRepository struct {
    db     *sql.DB
    logger *zap.Logger
}

func (r *auditRepository) CreateAuditLog(ctx context.Context, log *AuditLog) error {
    // Mock implementation
    r.logger.Info("Creating audit log (mock)", zap.String("id", log.ID.String()))
    return nil
}

func (r *auditRepository) GetAuditLog(ctx context.Context, id uuid.UUID) (*AuditLog, error) {
    // Mock implementation
    r.logger.Info("Getting audit log (mock)", zap.String("id", id.String()))
    return &AuditLog{
        ID:        id,
        Action:    "notification_sent",
        CreatedAt: time.Now(),
    }, nil
}

func (r *auditRepository) ListAuditLogs(ctx context.Context, limit, offset int) ([]*AuditLog, error) {
    // Mock implementation
    r.logger.Info("Listing audit logs (mock)", zap.Int("limit", limit), zap.Int("offset", offset))
    return []*AuditLog{}, nil
}
