package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all configuration for the User Management Service
type Config struct {
	Environment string `envconfig:"ENVIRONMENT" default:"development"`
	Version     string `envconfig:"VERSION" default:"1.0.0"`

	Server   ServerConfig   `envconfig:"SERVER"`
	Database DatabaseConfig `envconfig:"DATABASE"`
	Redis    RedisConfig    `envconfig:"REDIS"`
	NATS     NATSConfig     `envconfig:"NATS"`
	Auth     AuthConfig     `envconfig:"AUTH"`

	// User management settings
	UserManagement UserManagementConfig `envconfig:"USER_MANAGEMENT"`

	// Email settings
	Email EmailConfig `envconfig:"EMAIL"`

	// File storage settings
	Storage StorageConfig `envconfig:"STORAGE"`

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
	DB       int    `envconfig:"DB" default:"3"`
	PoolSize int    `envconfig:"POOL_SIZE" default:"10"`
}

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URL      string `envconfig:"URL" default:"nats://nats-client:4222"`
	Username string `envconfig:"USERNAME" default:"agba_user"`
	Password string `envconfig:"PASSWORD" required:"true"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	// JWT settings
	JWTSecret           string        `envconfig:"JWT_SECRET" required:"true"`
	JWTIssuer           string        `envconfig:"JWT_ISSUER" default:"agba.ai"`
	JWTAudience         string        `envconfig:"JWT_AUDIENCE" default:"agba.ai"`
	AccessTokenTTL      time.Duration `envconfig:"ACCESS_TOKEN_TTL" default:"15m"`
	RefreshTokenTTL     time.Duration `envconfig:"REFRESH_TOKEN_TTL" default:"168h"` // 7 days
	
	// Session settings
	SessionTTL          time.Duration `envconfig:"SESSION_TTL" default:"24h"`
	MaxSessions         int           `envconfig:"MAX_SESSIONS" default:"5"`
	
	// Password settings
	PasswordMinLength   int           `envconfig:"PASSWORD_MIN_LENGTH" default:"8"`
	PasswordMaxLength   int           `envconfig:"PASSWORD_MAX_LENGTH" default:"128"`
	PasswordRequireUpper bool         `envconfig:"PASSWORD_REQUIRE_UPPER" default:"true"`
	PasswordRequireLower bool         `envconfig:"PASSWORD_REQUIRE_LOWER" default:"true"`
	PasswordRequireDigit bool         `envconfig:"PASSWORD_REQUIRE_DIGIT" default:"true"`
	PasswordRequireSpecial bool       `envconfig:"PASSWORD_REQUIRE_SPECIAL" default:"true"`
	
	// Account lockout settings
	MaxLoginAttempts    int           `envconfig:"MAX_LOGIN_ATTEMPTS" default:"5"`
	LockoutDuration     time.Duration `envconfig:"LOCKOUT_DURATION" default:"30m"`
	
	// Two-factor authentication
	TwoFactorEnabled    bool          `envconfig:"TWO_FACTOR_ENABLED" default:"false"`
	TwoFactorIssuer     string        `envconfig:"TWO_FACTOR_ISSUER" default:"Agba.ai"`
	
	// OAuth2 settings
	OAuth2Enabled       bool          `envconfig:"OAUTH2_ENABLED" default:"true"`
	OAuth2Providers     []OAuth2Provider `envconfig:"OAUTH2_PROVIDERS"`
	
	// Email verification
	EmailVerificationRequired bool    `envconfig:"EMAIL_VERIFICATION_REQUIRED" default:"true"`
	EmailVerificationTTL     time.Duration `envconfig:"EMAIL_VERIFICATION_TTL" default:"24h"`
	
	// Password reset
	PasswordResetTTL    time.Duration `envconfig:"PASSWORD_RESET_TTL" default:"1h"`
	
	// Rate limiting
	RateLimitEnabled    bool          `envconfig:"RATE_LIMIT_ENABLED" default:"true"`
	LoginRateLimit      int           `envconfig:"LOGIN_RATE_LIMIT" default:"10"` // per minute
	RegistrationRateLimit int         `envconfig:"REGISTRATION_RATE_LIMIT" default:"5"` // per hour
}

// OAuth2Provider represents an OAuth2 provider configuration
type OAuth2Provider struct {
	Name         string `envconfig:"NAME"`
	ClientID     string `envconfig:"CLIENT_ID"`
	ClientSecret string `envconfig:"CLIENT_SECRET"`
	RedirectURL  string `envconfig:"REDIRECT_URL"`
	Scopes       []string `envconfig:"SCOPES"`
	AuthURL      string `envconfig:"AUTH_URL"`
	TokenURL     string `envconfig:"TOKEN_URL"`
	UserInfoURL  string `envconfig:"USER_INFO_URL"`
}

// UserManagementConfig holds user management settings
type UserManagementConfig struct {
	DefaultRole             string        `envconfig:"DEFAULT_ROLE" default:"user"`
	AllowSelfRegistration   bool          `envconfig:"ALLOW_SELF_REGISTRATION" default:"true"`
	RequireEmailVerification bool         `envconfig:"REQUIRE_EMAIL_VERIFICATION" default:"true"`
	AllowUsernameLogin      bool          `envconfig:"ALLOW_USERNAME_LOGIN" default:"true"`
	AllowEmailLogin         bool          `envconfig:"ALLOW_EMAIL_LOGIN" default:"true"`
	UsernameMinLength       int           `envconfig:"USERNAME_MIN_LENGTH" default:"3"`
	UsernameMaxLength       int           `envconfig:"USERNAME_MAX_LENGTH" default:"50"`
	ProfilePictureMaxSize   int64         `envconfig:"PROFILE_PICTURE_MAX_SIZE" default:"5242880"` // 5MB
	InvitationTTL           time.Duration `envconfig:"INVITATION_TTL" default:"168h"` // 7 days
	UserDataRetention       time.Duration `envconfig:"USER_DATA_RETENTION" default:"2160h"` // 90 days
	AuditLogRetention       time.Duration `envconfig:"AUDIT_LOG_RETENTION" default:"2160h"` // 90 days
}

// EmailConfig holds email configuration
type EmailConfig struct {
	Enabled     bool   `envconfig:"ENABLED" default:"true"`
	Provider    string `envconfig:"PROVIDER" default:"smtp"` // smtp, sendgrid, ses
	SMTPHost    string `envconfig:"SMTP_HOST"`
	SMTPPort    int    `envconfig:"SMTP_PORT" default:"587"`
	SMTPUser    string `envconfig:"SMTP_USER"`
	SMTPPass    string `envconfig:"SMTP_PASS"`
	FromEmail   string `envconfig:"FROM_EMAIL" default:"noreply@agba.ai"`
	FromName    string `envconfig:"FROM_NAME" default:"Agba.ai"`
	
	// Email templates
	Templates EmailTemplates `envconfig:"TEMPLATES"`
}

// EmailTemplates holds email template configuration
type EmailTemplates struct {
	WelcomeTemplate         string `envconfig:"WELCOME_TEMPLATE" default:"welcome"`
	VerificationTemplate    string `envconfig:"VERIFICATION_TEMPLATE" default:"verification"`
	PasswordResetTemplate   string `envconfig:"PASSWORD_RESET_TEMPLATE" default:"password_reset"`
	InvitationTemplate      string `envconfig:"INVITATION_TEMPLATE" default:"invitation"`
	PasswordChangedTemplate string `envconfig:"PASSWORD_CHANGED_TEMPLATE" default:"password_changed"`
}

// StorageConfig holds file storage configuration
type StorageConfig struct {
	Provider    string `envconfig:"PROVIDER" default:"local"` // local, s3, gcs
	LocalPath   string `envconfig:"LOCAL_PATH" default:"./uploads"`
	
	// S3 configuration
	S3Bucket    string `envconfig:"S3_BUCKET"`
	S3Region    string `envconfig:"S3_REGION"`
	S3AccessKey string `envconfig:"S3_ACCESS_KEY"`
	S3SecretKey string `envconfig:"S3_SECRET_KEY"`
	S3Endpoint  string `envconfig:"S3_ENDPOINT"`
	
	// GCS configuration
	GCSBucket      string `envconfig:"GCS_BUCKET"`
	GCSCredentials string `envconfig:"GCS_CREDENTIALS"`
	
	// General settings
	MaxFileSize    int64  `envconfig:"MAX_FILE_SIZE" default:"10485760"` // 10MB
	AllowedTypes   []string `envconfig:"ALLOWED_TYPES" default:"image/jpeg,image/png,image/gif"`
	CDNBaseURL     string `envconfig:"CDN_BASE_URL"`
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
	err := envconfig.Process("AGBA_USER", &cfg)
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