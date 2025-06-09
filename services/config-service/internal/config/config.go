package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all configuration for the Configuration Service
type Config struct {
	Environment string `envconfig:"ENVIRONMENT" default:"development"`
	Version     string `envconfig:"VERSION" default:"1.0.0"`

	Server   ServerConfig   `envconfig:"SERVER"`
	Database DatabaseConfig `envconfig:"DATABASE"`
	Redis    RedisConfig    `envconfig:"REDIS"`
	NATS     NATSConfig     `envconfig:"NATS"`

	// Configuration management settings
	ConfigManagement ConfigManagementConfig `envconfig:"CONFIG_MANAGEMENT"`

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
	DB       int    `envconfig:"DB" default:"1"`
	PoolSize int    `envconfig:"POOL_SIZE" default:"10"`
}

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URL      string `envconfig:"URL" default:"nats://nats-client:4222"`
	Username string `envconfig:"USERNAME" default:"agba_user"`
	Password string `envconfig:"PASSWORD" required:"true"`
}

// ConfigManagementConfig holds configuration management settings
type ConfigManagementConfig struct {
	MaxVersions           int           `envconfig:"MAX_VERSIONS" default:"50"`
	CacheTTL              time.Duration `envconfig:"CACHE_TTL" default:"300s"`
	ValidationTimeout     time.Duration `envconfig:"VALIDATION_TIMEOUT" default:"30s"`
	NotificationEnabled   bool          `envconfig:"NOTIFICATION_ENABLED" default:"true"`
	AutoBackupEnabled     bool          `envconfig:"AUTO_BACKUP_ENABLED" default:"true"`
	BackupInterval        time.Duration `envconfig:"BACKUP_INTERVAL" default:"24h"`
	EncryptionEnabled     bool          `envconfig:"ENCRYPTION_ENABLED" default:"true"`
	CompressionEnabled    bool          `envconfig:"COMPRESSION_ENABLED" default:"true"`
	AuditRetentionDays    int           `envconfig:"AUDIT_RETENTION_DAYS" default:"90"`
	ConfigValidationLevel string        `envconfig:"CONFIG_VALIDATION_LEVEL" default:"strict"`
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
	err := envconfig.Process("AGBA_CONFIG", &cfg)
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