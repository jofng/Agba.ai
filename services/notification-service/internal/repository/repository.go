package repository

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "time"

    "github.com/google/uuid"
    "go.uber.org/zap"
)

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
