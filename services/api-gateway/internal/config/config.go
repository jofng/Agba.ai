package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all configuration for the API Gateway
type Config struct {
	Environment string `envconfig:"ENVIRONMENT" default:"development"`
	Version     string `envconfig:"VERSION" default:"1.0.0"`

	Server ServerConfig `envconfig:"SERVER"`
	Auth   AuthConfig   `envconfig:"AUTH"`
	Redis  RedisConfig  `envconfig:"REDIS"`
	NATS   NATSConfig   `envconfig:"NATS"`
	CORS   CORSConfig   `envconfig:"CORS"`

	// Service endpoints
	Services ServiceEndpoints `envconfig:"SERVICES"`

	// Rate limiting
	RateLimit RateLimitConfig `envconfig:"RATE_LIMIT"`

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

// AuthConfig holds authentication configuration
type AuthConfig struct {
	OAuth2ServerURL    string        `envconfig:"OAUTH2_SERVER_URL" default:"http://oauth2-server:8080"`
	APIKeyValidatorURL string        `envconfig:"API_KEY_VALIDATOR_URL" default:"http://api-key-validator:8082"`
	JWTSecret          string        `envconfig:"JWT_SECRET" required:"true"`
	TokenCacheTTL      time.Duration `envconfig:"TOKEN_CACHE_TTL" default:"300s"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Address  string `envconfig:"ADDRESS" default:"redis-master:6379"`
	Password string `envconfig:"PASSWORD" required:"true"`
	DB       int    `envconfig:"DB" default:"0"`
	PoolSize int    `envconfig:"POOL_SIZE" default:"10"`
}

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URL      string `envconfig:"URL" default:"nats://nats-client:4222"`
	Username string `envconfig:"USERNAME" default:"agba_user"`
	Password string `envconfig:"PASSWORD" required:"true"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `envconfig:"ALLOWED_ORIGINS" default:"https://dashboard.agba.ai,https://admin.agba.ai"`
	AllowedMethods []string `envconfig:"ALLOWED_METHODS" default:"GET,POST,PUT,DELETE,OPTIONS"`
	AllowedHeaders []string `envconfig:"ALLOWED_HEADERS" default:"Authorization,Content-Type,X-Requested-With"`
	MaxAge         int      `envconfig:"MAX_AGE" default:"86400"`
}

// ServiceEndpoints holds backend service endpoints
type ServiceEndpoints struct {
	CallService       string `envconfig:"CALL_SERVICE" default:"http://call-service:8080"`
	AgentService      string `envconfig:"AGENT_SERVICE" default:"http://agent-service:8080"`
	AnalyticsService  string `envconfig:"ANALYTICS_SERVICE" default:"http://analytics-service:8080"`
	WebhookService    string `envconfig:"WEBHOOK_SERVICE" default:"http://webhook-service:8080"`
	UserService       string `envconfig:"USER_SERVICE" default:"http://user-service:8080"`
	ConfigService     string `envconfig:"CONFIG_SERVICE" default:"http://config-service:8080"`
	WebRTCService     string `envconfig:"WEBRTC_SERVICE" default:"http://webrtc-service:8080"`
	BiometricsService string `envconfig:"BIOMETRICS_SERVICE" default:"http://biometrics-service:8080"`
	RecordingService  string `envconfig:"RECORDING_SERVICE" default:"http://recording-service:8080"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled            bool          `envconfig:"ENABLED" default:"true"`
	DefaultRPM         int           `envconfig:"DEFAULT_RPM" default:"1000"`
	DefaultRPH         int           `envconfig:"DEFAULT_RPH" default:"10000"`
	BurstMultiplier    int           `envconfig:"BURST_MULTIPLIER" default:"2"`
	CleanupInterval    time.Duration `envconfig:"CLEANUP_INTERVAL" default:"60s"`
	RedisKeyPrefix     string        `envconfig:"REDIS_KEY_PREFIX" default:"rate_limit:"`
	SlidingWindow      bool          `envconfig:"SLIDING_WINDOW" default:"true"`
	BlockDuration      time.Duration `envconfig:"BLOCK_DURATION" default:"300s"`
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
	err := envconfig.Process("AGBA_GATEWAY", &cfg)
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