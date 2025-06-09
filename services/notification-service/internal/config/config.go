package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all configuration for the Notification Service
type Config struct {
	Environment string `envconfig:"ENVIRONMENT" default:"development"`
	Version     string `envconfig:"VERSION" default:"1.0.0"`

	Server   ServerConfig   `envconfig:"SERVER"`
	Database DatabaseConfig `envconfig:"DATABASE"`
	Redis    RedisConfig    `envconfig:"REDIS"`
	NATS     NATSConfig     `envconfig:"NATS"`

	// Notification providers
	Email   EmailConfig   `envconfig:"EMAIL"`
	SMS     SMSConfig     `envconfig:"SMS"`
	Push    PushConfig    `envconfig:"PUSH"`
	Webhook WebhookConfig `envconfig:"WEBHOOK"`

	// Processing settings
	Processing ProcessingConfig `envconfig:"PROCESSING"`

	// Monitoring
	Monitoring MonitoringConfig `envconfig:"MONITORING"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port         int           `envconfig:"PORT" default:"8080"`
	ReadTimeout  time.Duration `envconfig:"READ_TIMEOUT" default:"30s"`
	WriteTimeout time.Duration `envconfig:"WRITE_TIMEOUT" default:"30s"`
	IdleTimeout  time.Duration `envconfig:"IDLE_TIMEOUT" default:"120s"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver          string        `envconfig:"DRIVER" default:"postgres"`
	DSN             string        `envconfig:"DSN" required:"true"`
	MaxOpenConns    int           `envconfig:"MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns    int           `envconfig:"MAX_IDLE_CONNS" default:"5"`
	ConnMaxLifetime time.Duration `envconfig:"CONN_MAX_LIFETIME" default:"300s"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Address  string `envconfig:"ADDRESS" default:"redis-master:6379"`
	Password string `envconfig:"PASSWORD" required:"true"`
	DB       int    `envconfig:"DB" default:"4"`
	PoolSize int    `envconfig:"POOL_SIZE" default:"10"`
}

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URL      string `envconfig:"URL" default:"nats://nats-client:4222"`
	Username string `envconfig:"USERNAME" default:"agba_user"`
	Password string `envconfig:"PASSWORD" required:"true"`
}

// EmailConfig holds email provider configuration
type EmailConfig struct {
	Enabled  bool   `envconfig:"ENABLED" default:"true"`
	Provider string `envconfig:"PROVIDER" default:"smtp"` // smtp, sendgrid, ses, mailgun

	// SMTP configuration
	SMTP SMTPConfig `envconfig:"SMTP"`

	// SendGrid configuration
	SendGrid SendGridConfig `envconfig:"SENDGRID"`

	// AWS SES configuration
	SES SESConfig `envconfig:"SES"`

	// Mailgun configuration
	Mailgun MailgunConfig `envconfig:"MAILGUN"`

	// General email settings
	FromEmail    string        `envconfig:"FROM_EMAIL" default:"noreply@agba.ai"`
	FromName     string        `envconfig:"FROM_NAME" default:"Agba.ai"`
	ReplyTo      string        `envconfig:"REPLY_TO"`
	ReturnPath   string        `envconfig:"RETURN_PATH"`
	Timeout      time.Duration `envconfig:"TIMEOUT" default:"30s"`
	MaxRetries   int           `envconfig:"MAX_RETRIES" default:"3"`
	RetryDelay   time.Duration `envconfig:"RETRY_DELAY" default:"60s"`
	RateLimit    int           `envconfig:"RATE_LIMIT" default:"100"` // emails per minute
	TrackOpens   bool          `envconfig:"TRACK_OPENS" default:"true"`
	TrackClicks  bool          `envconfig:"TRACK_CLICKS" default:"true"`
	Unsubscribe  bool          `envconfig:"UNSUBSCRIBE" default:"true"`
}

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host     string `envconfig:"HOST" default:"smtp.gmail.com"`
	Port     int    `envconfig:"PORT" default:"587"`
	Username string `envconfig:"USERNAME"`
	Password string `envconfig:"PASSWORD"`
	UseTLS   bool   `envconfig:"USE_TLS" default:"true"`
}

// SendGridConfig holds SendGrid configuration
type SendGridConfig struct {
	APIKey string `envconfig:"API_KEY"`
}

// SESConfig holds AWS SES configuration
type SESConfig struct {
	Region          string `envconfig:"REGION" default:"us-east-1"`
	AccessKeyID     string `envconfig:"ACCESS_KEY_ID"`
	SecretAccessKey string `envconfig:"SECRET_ACCESS_KEY"`
}

// MailgunConfig holds Mailgun configuration
type MailgunConfig struct {
	Domain    string `envconfig:"DOMAIN"`
	APIKey    string `envconfig:"API_KEY"`
	PublicKey string `envconfig:"PUBLIC_KEY"`
}

// SMSConfig holds SMS provider configuration
type SMSConfig struct {
	Enabled  bool   `envconfig:"ENABLED" default:"true"`
	Provider string `envconfig:"PROVIDER" default:"twilio"` // twilio, aws_sns, nexmo

	// Twilio configuration
	Twilio TwilioConfig `envconfig:"TWILIO"`

	// AWS SNS configuration
	SNS SNSConfig `envconfig:"SNS"`

	// Nexmo configuration
	Nexmo NexmoConfig `envconfig:"NEXMO"`

	// General SMS settings
	FromNumber   string        `envconfig:"FROM_NUMBER"`
	Timeout      time.Duration `envconfig:"TIMEOUT" default:"30s"`
	MaxRetries   int           `envconfig:"MAX_RETRIES" default:"3"`
	RetryDelay   time.Duration `envconfig:"RETRY_DELAY" default:"60s"`
	RateLimit    int           `envconfig:"RATE_LIMIT" default:"50"` // SMS per minute
	MaxLength    int           `envconfig:"MAX_LENGTH" default:"160"`
	EnableUnicode bool         `envconfig:"ENABLE_UNICODE" default:"true"`
}

// TwilioConfig holds Twilio configuration
type TwilioConfig struct {
	AccountSID string `envconfig:"ACCOUNT_SID"`
	AuthToken  string `envconfig:"AUTH_TOKEN"`
}

// SNSConfig holds AWS SNS configuration
type SNSConfig struct {
	Region          string `envconfig:"REGION" default:"us-east-1"`
	AccessKeyID     string `envconfig:"ACCESS_KEY_ID"`
	SecretAccessKey string `envconfig:"SECRET_ACCESS_KEY"`
}

// NexmoConfig holds Nexmo configuration
type NexmoConfig struct {
	APIKey    string `envconfig:"API_KEY"`
	APISecret string `envconfig:"API_SECRET"`
}

// PushConfig holds push notification configuration
type PushConfig struct {
	Enabled  bool   `envconfig:"ENABLED" default:"true"`
	Provider string `envconfig:"PROVIDER" default:"fcm"` // fcm, apns

	// Firebase Cloud Messaging configuration
	FCM FCMConfig `envconfig:"FCM"`

	// Apple Push Notification Service configuration
	APNS APNSConfig `envconfig:"APNS"`

	// General push settings
	Timeout      time.Duration `envconfig:"TIMEOUT" default:"30s"`
	MaxRetries   int           `envconfig:"MAX_RETRIES" default:"3"`
	RetryDelay   time.Duration `envconfig:"RETRY_DELAY" default:"60s"`
	RateLimit    int           `envconfig:"RATE_LIMIT" default:"1000"` // notifications per minute
	BatchSize    int           `envconfig:"BATCH_SIZE" default:"100"`
	TTL          time.Duration `envconfig:"TTL" default:"24h"`
}

// FCMConfig holds Firebase Cloud Messaging configuration
type FCMConfig struct {
	ServerKey   string `envconfig:"SERVER_KEY"`
	ProjectID   string `envconfig:"PROJECT_ID"`
	Credentials string `envconfig:"CREDENTIALS"` // JSON credentials file path
}

// APNSConfig holds Apple Push Notification Service configuration
type APNSConfig struct {
	KeyID      string `envconfig:"KEY_ID"`
	TeamID     string `envconfig:"TEAM_ID"`
	BundleID   string `envconfig:"BUNDLE_ID"`
	KeyFile    string `envconfig:"KEY_FILE"`    // P8 key file path
	Production bool   `envconfig:"PRODUCTION" default:"false"`
}

// WebhookConfig holds webhook configuration
type WebhookConfig struct {
	Enabled    bool          `envconfig:"ENABLED" default:"true"`
	Timeout    time.Duration `envconfig:"TIMEOUT" default:"30s"`
	MaxRetries int           `envconfig:"MAX_RETRIES" default:"3"`
	RetryDelay time.Duration `envconfig:"RETRY_DELAY" default:"60s"`
	RateLimit  int           `envconfig:"RATE_LIMIT" default:"200"` // webhooks per minute
	
	// Security settings
	SigningSecret string `envconfig:"SIGNING_SECRET"`
	VerifySSL     bool   `envconfig:"VERIFY_SSL" default:"true"`
	
	// Headers
	UserAgent string            `envconfig:"USER_AGENT" default:"Agba-Notification-Service/1.0"`
	Headers   map[string]string `envconfig:"HEADERS"`
}

// ProcessingConfig holds notification processing configuration
type ProcessingConfig struct {
	// Queue settings
	QueueSize        int           `envconfig:"QUEUE_SIZE" default:"10000"`
	WorkerCount      int           `envconfig:"WORKER_COUNT" default:"10"`
	BatchSize        int           `envconfig:"BATCH_SIZE" default:"100"`
	ProcessInterval  time.Duration `envconfig:"PROCESS_INTERVAL" default:"5s"`
	
	// Retry settings
	MaxRetries       int           `envconfig:"MAX_RETRIES" default:"3"`
	RetryDelay       time.Duration `envconfig:"RETRY_DELAY" default:"60s"`
	RetryBackoff     string        `envconfig:"RETRY_BACKOFF" default:"exponential"` // linear, exponential
	RetryMaxDelay    time.Duration `envconfig:"RETRY_MAX_DELAY" default:"3600s"`
	
	// Cleanup settings
	CleanupInterval  time.Duration `envconfig:"CLEANUP_INTERVAL" default:"1h"`
	RetentionPeriod  time.Duration `envconfig:"RETENTION_PERIOD" default:"2160h"` // 90 days
	FailedRetention  time.Duration `envconfig:"FAILED_RETENTION" default:"720h"`  // 30 days
	
	// Rate limiting
	GlobalRateLimit  int           `envconfig:"GLOBAL_RATE_LIMIT" default:"1000"` // notifications per minute
	PerUserRateLimit int           `envconfig:"PER_USER_RATE_LIMIT" default:"100"` // notifications per hour
	
	// Priority settings
	EnablePriority   bool          `envconfig:"ENABLE_PRIORITY" default:"true"`
	HighPriorityQueue string       `envconfig:"HIGH_PRIORITY_QUEUE" default:"notifications.high"`
	NormalPriorityQueue string     `envconfig:"NORMAL_PRIORITY_QUEUE" default:"notifications.normal"`
	LowPriorityQueue string        `envconfig:"LOW_PRIORITY_QUEUE" default:"notifications.low"`
	
	// Deduplication
	EnableDeduplication bool        `envconfig:"ENABLE_DEDUPLICATION" default:"true"`
	DeduplicationWindow time.Duration `envconfig:"DEDUPLICATION_WINDOW" default:"300s"` // 5 minutes
}

// MonitoringConfig holds monitoring configuration
type MonitoringConfig struct {
	Enabled           bool   `envconfig:"ENABLED" default:"true"`
	MetricsPath       string `envconfig:"METRICS_PATH" default:"/metrics"`
	RequestLogging    bool   `envconfig:"REQUEST_LOGGING" default:"true"`
	SlowRequestThresh time.Duration `envconfig:"SLOW_REQUEST_THRESHOLD" default:"1s"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	var cfg Config
	err := envconfig.Process("AGBA_NOTIFICATION", &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Add validation logic here
	return nil
}