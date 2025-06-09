package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeEmail   NotificationType = "email"
	NotificationTypeSMS     NotificationType = "sms"
	NotificationTypePush    NotificationType = "push"
	NotificationTypeWebhook NotificationType = "webhook"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusPending    NotificationStatus = "pending"
	NotificationStatusQueued     NotificationStatus = "queued"
	NotificationStatusSending    NotificationStatus = "sending"
	NotificationStatusSent       NotificationStatus = "sent"
	NotificationStatusDelivered  NotificationStatus = "delivered"
	NotificationStatusFailed     NotificationStatus = "failed"
	NotificationStatusRetrying   NotificationStatus = "retrying"
	NotificationStatusCancelled  NotificationStatus = "cancelled"
	NotificationStatusBounced    NotificationStatus = "bounced"
	NotificationStatusComplaint  NotificationStatus = "complaint"
)

// NotificationPriority represents the priority of a notification
type NotificationPriority string

const (
	NotificationPriorityHigh   NotificationPriority = "high"
	NotificationPriorityNormal NotificationPriority = "normal"
	NotificationPriorityLow    NotificationPriority = "low"
)

// TemplateType represents the type of template
type TemplateType string

const (
	TemplateTypeEmail   TemplateType = "email"
	TemplateTypeSMS     TemplateType = "sms"
	TemplateTypePush    TemplateType = "push"
	TemplateTypeWebhook TemplateType = "webhook"
)

// SubscriptionStatus represents the status of a subscription
type SubscriptionStatus string

const (
	SubscriptionStatusActive      SubscriptionStatus = "active"
	SubscriptionStatusUnsubscribed SubscriptionStatus = "unsubscribed"
	SubscriptionStatusBounced     SubscriptionStatus = "bounced"
	SubscriptionStatusComplaint   SubscriptionStatus = "complaint"
)

// Notification represents a notification in the system
type Notification struct {
	ID          uuid.UUID                 `json:"id" db:"id"`
	Type        NotificationType          `json:"type" db:"type"`
	Status      NotificationStatus        `json:"status" db:"status"`
	Priority    NotificationPriority      `json:"priority" db:"priority"`
	
	// Recipient information
	RecipientID   *uuid.UUID              `json:"recipient_id,omitempty" db:"recipient_id"`
	RecipientType string                  `json:"recipient_type" db:"recipient_type"` // user, organization, group
	To            string                  `json:"to" db:"to"`                         // email, phone, device_token, webhook_url
	
	// Content
	Subject     string                    `json:"subject,omitempty" db:"subject"`
	Body        string                    `json:"body" db:"body"`
	HTMLBody    string                    `json:"html_body,omitempty" db:"html_body"`
	
	// Template information
	TemplateID   *uuid.UUID               `json:"template_id,omitempty" db:"template_id"`
	TemplateData map[string]interface{}   `json:"template_data,omitempty" db:"template_data"`
	
	// Metadata
	Headers     map[string]string         `json:"headers,omitempty" db:"headers"`
	Attachments []Attachment              `json:"attachments,omitempty" db:"attachments"`
	Tags        []string                  `json:"tags,omitempty" db:"tags"`
	
	// Scheduling
	ScheduledAt *time.Time                `json:"scheduled_at,omitempty" db:"scheduled_at"`
	ExpiresAt   *time.Time                `json:"expires_at,omitempty" db:"expires_at"`
	
	// Processing information
	Provider      string                  `json:"provider,omitempty" db:"provider"`
	ProviderID    string                  `json:"provider_id,omitempty" db:"provider_id"`
	Attempts      int                     `json:"attempts" db:"attempts"`
	MaxAttempts   int                     `json:"max_attempts" db:"max_attempts"`
	LastAttemptAt *time.Time              `json:"last_attempt_at,omitempty" db:"last_attempt_at"`
	NextRetryAt   *time.Time              `json:"next_retry_at,omitempty" db:"next_retry_at"`
	
	// Delivery information
	SentAt        *time.Time              `json:"sent_at,omitempty" db:"sent_at"`
	DeliveredAt   *time.Time              `json:"delivered_at,omitempty" db:"delivered_at"`
	OpenedAt      *time.Time              `json:"opened_at,omitempty" db:"opened_at"`
	ClickedAt     *time.Time              `json:"clicked_at,omitempty" db:"clicked_at"`
	
	// Error information
	ErrorMessage  string                  `json:"error_message,omitempty" db:"error_message"`
	ErrorCode     string                  `json:"error_code,omitempty" db:"error_code"`
	
	// Tracking
	TrackingID    string                  `json:"tracking_id,omitempty" db:"tracking_id"`
	CampaignID    *uuid.UUID              `json:"campaign_id,omitempty" db:"campaign_id"`
	
	// Metadata
	Metadata      map[string]interface{}  `json:"metadata,omitempty" db:"metadata"`
	CreatedAt     time.Time               `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time               `json:"updated_at" db:"updated_at"`
}

// Attachment represents a file attachment
type Attachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Content     []byte `json:"content,omitempty"`
	URL         string `json:"url,omitempty"`
	Size        int64  `json:"size"`
}

// Template represents a notification template
type Template struct {
	ID          uuid.UUID                 `json:"id" db:"id"`
	Name        string                    `json:"name" db:"name"`
	Type        TemplateType              `json:"type" db:"type"`
	Subject     string                    `json:"subject,omitempty" db:"subject"`
	Body        string                    `json:"body" db:"body"`
	HTMLBody    string                    `json:"html_body,omitempty" db:"html_body"`
	
	// Template configuration
	Variables   []TemplateVariable        `json:"variables" db:"variables"`
	Defaults    map[string]interface{}    `json:"defaults,omitempty" db:"defaults"`
	
	// Localization
	Language    string                    `json:"language" db:"language"`
	Localizations map[string]TemplateLocalization `json:"localizations,omitempty" db:"localizations"`
	
	// Metadata
	Description string                    `json:"description,omitempty" db:"description"`
	Tags        []string                  `json:"tags,omitempty" db:"tags"`
	Version     int                       `json:"version" db:"version"`
	IsActive    bool                      `json:"is_active" db:"is_active"`
	
	// Ownership
	CreatedBy   uuid.UUID                 `json:"created_by" db:"created_by"`
	UpdatedBy   uuid.UUID                 `json:"updated_by" db:"updated_by"`
	
	// Metadata
	Metadata    map[string]interface{}    `json:"metadata,omitempty" db:"metadata"`
	CreatedAt   time.Time                 `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at" db:"updated_at"`
}

// TemplateVariable represents a template variable
type TemplateVariable struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`        // string, number, boolean, date, array, object
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Description string      `json:"description,omitempty"`
	Validation  string      `json:"validation,omitempty"` // regex or validation rule
}

// TemplateLocalization represents template localization
type TemplateLocalization struct {
	Subject  string `json:"subject,omitempty"`
	Body     string `json:"body"`
	HTMLBody string `json:"html_body,omitempty"`
}

// Subscription represents a notification subscription
type Subscription struct {
	ID          uuid.UUID                 `json:"id" db:"id"`
	UserID      uuid.UUID                 `json:"user_id" db:"user_id"`
	Type        NotificationType          `json:"type" db:"type"`
	Channel     string                    `json:"channel" db:"channel"` // email address, phone number, device token
	Status      SubscriptionStatus        `json:"status" db:"status"`
	
	// Subscription preferences
	Categories  []string                  `json:"categories" db:"categories"`
	Frequency   string                    `json:"frequency" db:"frequency"` // immediate, daily, weekly, monthly
	TimeZone    string                    `json:"timezone" db:"timezone"`
	QuietHours  QuietHours                `json:"quiet_hours,omitempty" db:"quiet_hours"`
	
	// Verification
	Verified      bool                    `json:"verified" db:"verified"`
	VerifiedAt    *time.Time              `json:"verified_at,omitempty" db:"verified_at"`
	VerificationToken string              `json:"-" db:"verification_token"`
	
	// Unsubscribe information
	UnsubscribedAt *time.Time             `json:"unsubscribed_at,omitempty" db:"unsubscribed_at"`
	UnsubscribeToken string               `json:"-" db:"unsubscribe_token"`
	UnsubscribeReason string              `json:"unsubscribe_reason,omitempty" db:"unsubscribe_reason"`
	
	// Metadata
	Metadata    map[string]interface{}    `json:"metadata,omitempty" db:"metadata"`
	CreatedAt   time.Time                 `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at" db:"updated_at"`
}

// QuietHours represents quiet hours configuration
type QuietHours struct {
	Enabled   bool   `json:"enabled"`
	StartTime string `json:"start_time"` // HH:MM format
	EndTime   string `json:"end_time"`   // HH:MM format
	Days      []int  `json:"days"`       // 0=Sunday, 1=Monday, etc.
}

// Device represents a push notification device
type Device struct {
	ID          uuid.UUID                 `json:"id" db:"id"`
	UserID      uuid.UUID                 `json:"user_id" db:"user_id"`
	Token       string                    `json:"token" db:"token"`
	Platform    string                    `json:"platform" db:"platform"` // ios, android, web
	AppVersion  string                    `json:"app_version,omitempty" db:"app_version"`
	OSVersion   string                    `json:"os_version,omitempty" db:"os_version"`
	DeviceModel string                    `json:"device_model,omitempty" db:"device_model"`
	
	// Status
	IsActive    bool                      `json:"is_active" db:"is_active"`
	LastUsedAt  *time.Time                `json:"last_used_at,omitempty" db:"last_used_at"`
	
	// Metadata
	Metadata    map[string]interface{}    `json:"metadata,omitempty" db:"metadata"`
	CreatedAt   time.Time                 `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at" db:"updated_at"`
}

// Campaign represents a notification campaign
type Campaign struct {
	ID          uuid.UUID                 `json:"id" db:"id"`
	Name        string                    `json:"name" db:"name"`
	Description string                    `json:"description,omitempty" db:"description"`
	Type        NotificationType          `json:"type" db:"type"`
	
	// Campaign configuration
	TemplateID  uuid.UUID                 `json:"template_id" db:"template_id"`
	Recipients  []CampaignRecipient       `json:"recipients" db:"recipients"`
	Filters     CampaignFilters           `json:"filters,omitempty" db:"filters"`
	
	// Scheduling
	ScheduledAt *time.Time                `json:"scheduled_at,omitempty" db:"scheduled_at"`
	StartedAt   *time.Time                `json:"started_at,omitempty" db:"started_at"`
	CompletedAt *time.Time                `json:"completed_at,omitempty" db:"completed_at"`
	
	// Status
	Status      string                    `json:"status" db:"status"` // draft, scheduled, running, completed, cancelled
	
	// Statistics
	TotalRecipients int                   `json:"total_recipients" db:"total_recipients"`
	SentCount       int                   `json:"sent_count" db:"sent_count"`
	DeliveredCount  int                   `json:"delivered_count" db:"delivered_count"`
	FailedCount     int                   `json:"failed_count" db:"failed_count"`
	OpenedCount     int                   `json:"opened_count" db:"opened_count"`
	ClickedCount    int                   `json:"clicked_count" db:"clicked_count"`
	
	// Ownership
	CreatedBy   uuid.UUID                 `json:"created_by" db:"created_by"`
	
	// Metadata
	Metadata    map[string]interface{}    `json:"metadata,omitempty" db:"metadata"`
	CreatedAt   time.Time                 `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at" db:"updated_at"`
}

// CampaignRecipient represents a campaign recipient
type CampaignRecipient struct {
	ID       uuid.UUID              `json:"id,omitempty"`
	Type     string                 `json:"type"` // user, email, phone
	Value    string                 `json:"value"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

// CampaignFilters represents campaign recipient filters
type CampaignFilters struct {
	UserSegments    []string               `json:"user_segments,omitempty"`
	UserTags        []string               `json:"user_tags,omitempty"`
	Organizations   []uuid.UUID            `json:"organizations,omitempty"`
	Subscriptions   []NotificationType     `json:"subscriptions,omitempty"`
	LastActiveAfter *time.Time             `json:"last_active_after,omitempty"`
	CreatedAfter    *time.Time             `json:"created_after,omitempty"`
	CustomFilters   map[string]interface{} `json:"custom_filters,omitempty"`
}

// NotificationAudit represents an audit log entry for notifications
type NotificationAudit struct {
	ID             uuid.UUID                 `json:"id" db:"id"`
	NotificationID uuid.UUID                 `json:"notification_id" db:"notification_id"`
	Action         string                    `json:"action" db:"action"`
	Status         NotificationStatus        `json:"status" db:"status"`
	Provider       string                    `json:"provider,omitempty" db:"provider"`
	
	// Request/Response data
	RequestData    map[string]interface{}    `json:"request_data,omitempty" db:"request_data"`
	ResponseData   map[string]interface{}    `json:"response_data,omitempty" db:"response_data"`
	
	// Error information
	ErrorMessage   string                    `json:"error_message,omitempty" db:"error_message"`
	ErrorCode      string                    `json:"error_code,omitempty" db:"error_code"`
	
	// Timing
	Duration       time.Duration             `json:"duration,omitempty" db:"duration"`
	
	// Metadata
	Metadata       map[string]interface{}    `json:"metadata,omitempty" db:"metadata"`
	Timestamp      time.Time                 `json:"timestamp" db:"timestamp"`
}

// NotificationStatistics represents notification statistics
type NotificationStatistics struct {
	TotalNotifications int64                    `json:"total_notifications"`
	SentNotifications  int64                    `json:"sent_notifications"`
	FailedNotifications int64                   `json:"failed_notifications"`
	DeliveryRate       float64                  `json:"delivery_rate"`
	
	// By type
	ByType map[NotificationType]TypeStatistics `json:"by_type"`
	
	// By provider
	ByProvider map[string]ProviderStatistics   `json:"by_provider"`
	
	// Time range
	StartTime time.Time                        `json:"start_time"`
	EndTime   time.Time                        `json:"end_time"`
}

// TypeStatistics represents statistics for a notification type
type TypeStatistics struct {
	Total     int64   `json:"total"`
	Sent      int64   `json:"sent"`
	Failed    int64   `json:"failed"`
	Delivered int64   `json:"delivered"`
	Opened    int64   `json:"opened"`
	Clicked   int64   `json:"clicked"`
	Rate      float64 `json:"rate"`
}

// ProviderStatistics represents statistics for a provider
type ProviderStatistics struct {
	Total        int64         `json:"total"`
	Sent         int64         `json:"sent"`
	Failed       int64         `json:"failed"`
	AvgDuration  time.Duration `json:"avg_duration"`
	SuccessRate  float64       `json:"success_rate"`
	ErrorCodes   map[string]int64 `json:"error_codes"`
}

// SendNotificationRequest represents a request to send a notification
type SendNotificationRequest struct {
	Type        NotificationType          `json:"type" binding:"required"`
	To          string                    `json:"to" binding:"required"`
	Subject     string                    `json:"subject,omitempty"`
	Body        string                    `json:"body,omitempty"`
	HTMLBody    string                    `json:"html_body,omitempty"`
	TemplateID  *uuid.UUID                `json:"template_id,omitempty"`
	TemplateData map[string]interface{}   `json:"template_data,omitempty"`
	Priority    NotificationPriority      `json:"priority,omitempty"`
	ScheduledAt *time.Time                `json:"scheduled_at,omitempty"`
	ExpiresAt   *time.Time                `json:"expires_at,omitempty"`
	Headers     map[string]string         `json:"headers,omitempty"`
	Attachments []Attachment              `json:"attachments,omitempty"`
	Tags        []string                  `json:"tags,omitempty"`
	Metadata    map[string]interface{}    `json:"metadata,omitempty"`
}

// SendBulkNotificationRequest represents a request to send bulk notifications
type SendBulkNotificationRequest struct {
	Type         NotificationType          `json:"type" binding:"required"`
	Recipients   []string                  `json:"recipients" binding:"required"`
	Subject      string                    `json:"subject,omitempty"`
	Body         string                    `json:"body,omitempty"`
	HTMLBody     string                    `json:"html_body,omitempty"`
	TemplateID   *uuid.UUID                `json:"template_id,omitempty"`
	TemplateData map[string]interface{}    `json:"template_data,omitempty"`
	Priority     NotificationPriority      `json:"priority,omitempty"`
	ScheduledAt  *time.Time                `json:"scheduled_at,omitempty"`
	ExpiresAt    *time.Time                `json:"expires_at,omitempty"`
	Headers      map[string]string         `json:"headers,omitempty"`
	Attachments  []Attachment              `json:"attachments,omitempty"`
	Tags         []string                  `json:"tags,omitempty"`
	Metadata     map[string]interface{}    `json:"metadata,omitempty"`
}

// CreateTemplateRequest represents a request to create a template
type CreateTemplateRequest struct {
	Name        string                    `json:"name" binding:"required"`
	Type        TemplateType              `json:"type" binding:"required"`
	Subject     string                    `json:"subject,omitempty"`
	Body        string                    `json:"body" binding:"required"`
	HTMLBody    string                    `json:"html_body,omitempty"`
	Variables   []TemplateVariable        `json:"variables,omitempty"`
	Defaults    map[string]interface{}    `json:"defaults,omitempty"`
	Language    string                    `json:"language,omitempty"`
	Description string                    `json:"description,omitempty"`
	Tags        []string                  `json:"tags,omitempty"`
}

// UpdateTemplateRequest represents a request to update a template
type UpdateTemplateRequest struct {
	Name        *string                   `json:"name,omitempty"`
	Subject     *string                   `json:"subject,omitempty"`
	Body        *string                   `json:"body,omitempty"`
	HTMLBody    *string                   `json:"html_body,omitempty"`
	Variables   []TemplateVariable        `json:"variables,omitempty"`
	Defaults    map[string]interface{}    `json:"defaults,omitempty"`
	Description *string                   `json:"description,omitempty"`
	Tags        []string                  `json:"tags,omitempty"`
	IsActive    *bool                     `json:"is_active,omitempty"`
}

// CreateSubscriptionRequest represents a request to create a subscription
type CreateSubscriptionRequest struct {
	UserID     uuid.UUID            `json:"user_id" binding:"required"`
	Type       NotificationType     `json:"type" binding:"required"`
	Channel    string               `json:"channel" binding:"required"`
	Categories []string             `json:"categories,omitempty"`
	Frequency  string               `json:"frequency,omitempty"`
	TimeZone   string               `json:"timezone,omitempty"`
	QuietHours *QuietHours          `json:"quiet_hours,omitempty"`
}

// UpdateSubscriptionRequest represents a request to update a subscription
type UpdateSubscriptionRequest struct {
	Categories []string    `json:"categories,omitempty"`
	Frequency  *string     `json:"frequency,omitempty"`
	TimeZone   *string     `json:"timezone,omitempty"`
	QuietHours *QuietHours `json:"quiet_hours,omitempty"`
}

// RegisterDeviceRequest represents a request to register a device
type RegisterDeviceRequest struct {
	UserID      uuid.UUID `json:"user_id" binding:"required"`
	Token       string    `json:"token" binding:"required"`
	Platform    string    `json:"platform" binding:"required"`
	AppVersion  string    `json:"app_version,omitempty"`
	OSVersion   string    `json:"os_version,omitempty"`
	DeviceModel string    `json:"device_model,omitempty"`
}

// WebhookPayload represents a webhook payload
type WebhookPayload struct {
	Event     string                 `json:"event"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Signature string                 `json:"signature,omitempty"`
}

// EmailBounceEvent represents an email bounce event
type EmailBounceEvent struct {
	NotificationID uuid.UUID `json:"notification_id"`
	Email          string    `json:"email"`
	BounceType     string    `json:"bounce_type"` // hard, soft, complaint
	Reason         string    `json:"reason"`
	Timestamp      time.Time `json:"timestamp"`
}

// SMSDeliveryReport represents an SMS delivery report
type SMSDeliveryReport struct {
	NotificationID uuid.UUID `json:"notification_id"`
	MessageID      string    `json:"message_id"`
	Status         string    `json:"status"`
	ErrorCode      string    `json:"error_code,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
}