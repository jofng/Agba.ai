package providers

import (
	"context"
	"time"
)

// EmailProvider defines the interface for email providers
type EmailProvider interface {
	SendEmail(ctx context.Context, req *EmailRequest) (*EmailResponse, error)
	SendBulkEmails(ctx context.Context, requests []*EmailRequest) ([]*EmailResponse, error)
	HealthCheck() error
	GetProviderName() string
	GetRateLimit() int
}

// SMSProvider defines the interface for SMS providers
type SMSProvider interface {
	SendSMS(ctx context.Context, req *SMSRequest) (*SMSResponse, error)
	SendBulkSMS(ctx context.Context, requests []*SMSRequest) ([]*SMSResponse, error)
	HealthCheck() error
	GetProviderName() string
	GetRateLimit() int
}

// PushProvider defines the interface for push notification providers
type PushProvider interface {
	SendPushNotification(ctx context.Context, req *PushRequest) (*PushResponse, error)
	SendBulkPushNotifications(ctx context.Context, requests []*PushRequest) ([]*PushResponse, error)
	HealthCheck() error
	GetProviderName() string
	GetRateLimit() int
}

// WebhookProvider defines the interface for webhook providers
type WebhookProvider interface {
	SendWebhook(ctx context.Context, req *WebhookRequest) (*WebhookResponse, error)
	SendBulkWebhooks(ctx context.Context, requests []*WebhookRequest) ([]*WebhookResponse, error)
	HealthCheck() error
	GetProviderName() string
	GetRateLimit() int
}

// EmailRequest represents an email notification request
type EmailRequest struct {
	ID          string            `json:"id"`
	To          []string          `json:"to"`
	CC          []string          `json:"cc,omitempty"`
	BCC         []string          `json:"bcc,omitempty"`
	From        string            `json:"from"`
	FromName    string            `json:"from_name,omitempty"`
	ReplyTo     string            `json:"reply_to,omitempty"`
	Subject     string            `json:"subject"`
	Body        string            `json:"body"`
	HTMLBody    string            `json:"html_body,omitempty"`
	Attachments []Attachment      `json:"attachments,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	TemplateID  string            `json:"template_id,omitempty"`
	Variables   map[string]string `json:"variables,omitempty"`
	TrackOpens  bool              `json:"track_opens"`
	TrackClicks bool              `json:"track_clicks"`
	Priority    string            `json:"priority,omitempty"` // high, normal, low
	ScheduledAt *time.Time        `json:"scheduled_at,omitempty"`
}

// EmailResponse represents an email notification response
type EmailResponse struct {
	ID          string            `json:"id"`
	MessageID   string            `json:"message_id"`
	Status      string            `json:"status"` // sent, failed, queued
	Error       string            `json:"error,omitempty"`
	Provider    string            `json:"provider"`
	SentAt      time.Time         `json:"sent_at"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Cost        float64           `json:"cost,omitempty"`
}

// SMSRequest represents an SMS notification request
type SMSRequest struct {
	ID          string            `json:"id"`
	To          string            `json:"to"`
	From        string            `json:"from"`
	Body        string            `json:"body"`
	MediaURLs   []string          `json:"media_urls,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Priority    string            `json:"priority,omitempty"` // high, normal, low
	ScheduledAt *time.Time        `json:"scheduled_at,omitempty"`
}

// SMSResponse represents an SMS notification response
type SMSResponse struct {
	ID        string            `json:"id"`
	MessageID string            `json:"message_id"`
	Status    string            `json:"status"` // sent, failed, queued, delivered
	Error     string            `json:"error,omitempty"`
	Provider  string            `json:"provider"`
	SentAt    time.Time         `json:"sent_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Cost      float64           `json:"cost,omitempty"`
	Segments  int               `json:"segments"`
}

// PushRequest represents a push notification request
type PushRequest struct {
	ID           string            `json:"id"`
	DeviceTokens []string          `json:"device_tokens"`
	Title        string            `json:"title"`
	Body         string            `json:"body"`
	Icon         string            `json:"icon,omitempty"`
	Sound        string            `json:"sound,omitempty"`
	Badge        int               `json:"badge,omitempty"`
	Data         map[string]string `json:"data,omitempty"`
	Tags         []string          `json:"tags,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	Priority     string            `json:"priority,omitempty"` // high, normal, low
	TTL          time.Duration     `json:"ttl,omitempty"`
	ScheduledAt  *time.Time        `json:"scheduled_at,omitempty"`
	Platform     string            `json:"platform,omitempty"` // ios, android, web
}

// PushResponse represents a push notification response
type PushResponse struct {
	ID               string            `json:"id"`
	MessageID        string            `json:"message_id"`
	Status           string            `json:"status"` // sent, failed, queued
	Error            string            `json:"error,omitempty"`
	Provider         string            `json:"provider"`
	SentAt           time.Time         `json:"sent_at"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	SuccessCount     int               `json:"success_count"`
	FailureCount     int               `json:"failure_count"`
	InvalidTokens    []string          `json:"invalid_tokens,omitempty"`
	CanonicalTokens  map[string]string `json:"canonical_tokens,omitempty"`
}

// WebhookRequest represents a webhook notification request
type WebhookRequest struct {
	ID          string            `json:"id"`
	URL         string            `json:"url"`
	Method      string            `json:"method"` // GET, POST, PUT, PATCH, DELETE
	Headers     map[string]string `json:"headers,omitempty"`
	Body        string            `json:"body,omitempty"`
	ContentType string            `json:"content_type,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Priority    string            `json:"priority,omitempty"` // high, normal, low
	Timeout     time.Duration     `json:"timeout,omitempty"`
	ScheduledAt *time.Time        `json:"scheduled_at,omitempty"`
	Secret      string            `json:"secret,omitempty"` // For HMAC signing
}

// WebhookResponse represents a webhook notification response
type WebhookResponse struct {
	ID           string            `json:"id"`
	Status       string            `json:"status"` // sent, failed, queued
	StatusCode   int               `json:"status_code,omitempty"`
	ResponseBody string            `json:"response_body,omitempty"`
	Error        string            `json:"error,omitempty"`
	Provider     string            `json:"provider"`
	SentAt       time.Time         `json:"sent_at"`
	Duration     time.Duration     `json:"duration"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Content     []byte `json:"content"`
	ContentID   string `json:"content_id,omitempty"` // For inline attachments
}

// DeliveryStatus represents the delivery status of a notification
type DeliveryStatus struct {
	ID          string            `json:"id"`
	Status      string            `json:"status"`
	DeliveredAt *time.Time        `json:"delivered_at,omitempty"`
	FailedAt    *time.Time        `json:"failed_at,omitempty"`
	Error       string            `json:"error,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ProviderStats represents provider statistics
type ProviderStats struct {
	Provider      string        `json:"provider"`
	TotalSent     int64         `json:"total_sent"`
	TotalFailed   int64         `json:"total_failed"`
	SuccessRate   float64       `json:"success_rate"`
	AverageLatency time.Duration `json:"average_latency"`
	LastUsed      time.Time     `json:"last_used"`
	RateLimit     int           `json:"rate_limit"`
	Cost          float64       `json:"cost"`
}