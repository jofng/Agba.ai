package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all configuration for the Monitoring Service
type Config struct {
	Environment string `envconfig:"ENVIRONMENT" default:"development"`
	Version     string `envconfig:"VERSION" default:"1.0.0"`

	Server     ServerConfig     `envconfig:"SERVER"`
	Database   DatabaseConfig   `envconfig:"DATABASE"`
	Redis      RedisConfig      `envconfig:"REDIS"`
	NATS       NATSConfig       `envconfig:"NATS"`
	Prometheus PrometheusConfig `envconfig:"PROMETHEUS"`
	Alerting   AlertingConfig   `envconfig:"ALERTING"`

	// Monitoring settings
	Monitoring MonitoringConfig `envconfig:"MONITORING"`

	// Collection settings
	Collection CollectionConfig `envconfig:"COLLECTION"`
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
	DB       int    `envconfig:"DB" default:"2"`
	PoolSize int    `envconfig:"POOL_SIZE" default:"10"`
}

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URL      string `envconfig:"URL" default:"nats://nats-client:4222"`
	Username string `envconfig:"USERNAME" default:"agba_user"`
	Password string `envconfig:"PASSWORD" required:"true"`
}

// PrometheusConfig holds Prometheus configuration
type PrometheusConfig struct {
	URL             string        `envconfig:"URL" default:"http://prometheus:9090"`
	Username        string        `envconfig:"USERNAME"`
	Password        string        `envconfig:"PASSWORD"`
	Timeout         time.Duration `envconfig:"TIMEOUT" default:"30s"`
	MaxConnections  int           `envconfig:"MAX_CONNECTIONS" default:"10"`
	RetentionPeriod time.Duration `envconfig:"RETENTION_PERIOD" default:"720h"` // 30 days
}

// AlertingConfig holds alerting configuration
type AlertingConfig struct {
	Enabled                bool          `envconfig:"ENABLED" default:"true"`
	DefaultSeverity        string        `envconfig:"DEFAULT_SEVERITY" default:"warning"`
	EvaluationInterval     time.Duration `envconfig:"EVALUATION_INTERVAL" default:"30s"`
	NotificationTimeout    time.Duration `envconfig:"NOTIFICATION_TIMEOUT" default:"10s"`
	MaxRetries             int           `envconfig:"MAX_RETRIES" default:"3"`
	RetryInterval          time.Duration `envconfig:"RETRY_INTERVAL" default:"60s"`
	SilenceDefaultDuration time.Duration `envconfig:"SILENCE_DEFAULT_DURATION" default:"1h"`
	
	// Notification channels
	Slack    SlackConfig    `envconfig:"SLACK"`
	Email    EmailConfig    `envconfig:"EMAIL"`
	Webhook  WebhookConfig  `envconfig:"WEBHOOK"`
	PagerDuty PagerDutyConfig `envconfig:"PAGERDUTY"`
}

// SlackConfig holds Slack notification configuration
type SlackConfig struct {
	Enabled   bool   `envconfig:"ENABLED" default:"false"`
	WebhookURL string `envconfig:"WEBHOOK_URL"`
	Channel   string `envconfig:"CHANNEL" default:"#alerts"`
	Username  string `envconfig:"USERNAME" default:"Agba Monitoring"`
}

// EmailConfig holds email notification configuration
type EmailConfig struct {
	Enabled    bool   `envconfig:"ENABLED" default:"false"`
	SMTPHost   string `envconfig:"SMTP_HOST"`
	SMTPPort   int    `envconfig:"SMTP_PORT" default:"587"`
	Username   string `envconfig:"USERNAME"`
	Password   string `envconfig:"PASSWORD"`
	FromEmail  string `envconfig:"FROM_EMAIL"`
	FromName   string `envconfig:"FROM_NAME" default:"Agba Monitoring"`
}

// WebhookConfig holds webhook notification configuration
type WebhookConfig struct {
	Enabled bool   `envconfig:"ENABLED" default:"false"`
	URL     string `envconfig:"URL"`
	Secret  string `envconfig:"SECRET"`
	Timeout time.Duration `envconfig:"TIMEOUT" default:"10s"`
}

// PagerDutyConfig holds PagerDuty notification configuration
type PagerDutyConfig struct {
	Enabled        bool   `envconfig:"ENABLED" default:"false"`
	IntegrationKey string `envconfig:"INTEGRATION_KEY"`
	ServiceKey     string `envconfig:"SERVICE_KEY"`
}

// MonitoringConfig holds monitoring configuration
type MonitoringConfig struct {
	Enabled           bool          `envconfig:"ENABLED" default:"true"`
	MetricsPath       string        `envconfig:"METRICS_PATH" default:"/metrics"`
	RequestLogging    bool          `envconfig:"REQUEST_LOGGING" default:"true"`
	SlowRequestThresh time.Duration `envconfig:"SLOW_REQUEST_THRESHOLD" default:"1s"`
}

// CollectionConfig holds metrics collection configuration
type CollectionConfig struct {
	Interval              time.Duration `envconfig:"INTERVAL" default:"30s"`
	BatchSize             int           `envconfig:"BATCH_SIZE" default:"100"`
	BufferSize            int           `envconfig:"BUFFER_SIZE" default:"1000"`
	FlushInterval         time.Duration `envconfig:"FLUSH_INTERVAL" default:"60s"`
	RetentionPeriod       time.Duration `envconfig:"RETENTION_PERIOD" default:"2160h"` // 90 days
	HighResolutionPeriod  time.Duration `envconfig:"HIGH_RESOLUTION_PERIOD" default:"24h"`
	MediumResolutionPeriod time.Duration `envconfig:"MEDIUM_RESOLUTION_PERIOD" default:"168h"` // 7 days
	
	// Service discovery
	ServiceDiscovery ServiceDiscoveryConfig `envconfig:"SERVICE_DISCOVERY"`
	
	// Health check settings
	HealthCheck HealthCheckConfig `envconfig:"HEALTH_CHECK"`
}

// ServiceDiscoveryConfig holds service discovery configuration
type ServiceDiscoveryConfig struct {
	Enabled          bool          `envconfig:"ENABLED" default:"true"`
	KubernetesEnabled bool         `envconfig:"KUBERNETES_ENABLED" default:"true"`
	RefreshInterval  time.Duration `envconfig:"REFRESH_INTERVAL" default:"60s"`
	Namespace        string        `envconfig:"NAMESPACE" default:"agba-system"`
	LabelSelector    string        `envconfig:"LABEL_SELECTOR" default:"app"`
}

// HealthCheckConfig holds health check configuration
type HealthCheckConfig struct {
	Enabled         bool          `envconfig:"ENABLED" default:"true"`
	Interval        time.Duration `envconfig:"INTERVAL" default:"30s"`
	Timeout         time.Duration `envconfig:"TIMEOUT" default:"10s"`
	FailureThreshold int          `envconfig:"FAILURE_THRESHOLD" default:"3"`
	SuccessThreshold int          `envconfig:"SUCCESS_THRESHOLD" default:"1"`
	
	// Health check endpoints
	Endpoints []HealthEndpoint `envconfig:"ENDPOINTS"`
}

// HealthEndpoint represents a health check endpoint
type HealthEndpoint struct {
	Name     string `envconfig:"NAME"`
	URL      string `envconfig:"URL"`
	Method   string `envconfig:"METHOD" default:"GET"`
	Timeout  time.Duration `envconfig:"TIMEOUT" default:"10s"`
	Expected int    `envconfig:"EXPECTED" default:"200"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	var cfg Config
	err := envconfig.Process("AGBA_MONITORING", &cfg)
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