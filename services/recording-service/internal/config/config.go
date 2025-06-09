package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all configuration for the Recording Service
type Config struct {
	Environment string `envconfig:"ENVIRONMENT" default:"development"`
	Version     string `envconfig:"VERSION" default:"1.0.0"`

	Server   ServerConfig   `envconfig:"SERVER"`
	Database DatabaseConfig `envconfig:"DATABASE"`
	Redis    RedisConfig    `envconfig:"REDIS"`
	NATS     NATSConfig     `envconfig:"NATS"`

	// Storage configuration
	Storage StorageConfig `envconfig:"STORAGE"`

	// Audio processing configuration
	Audio         AudioConfig         `envconfig:"AUDIO"`
	Transcription TranscriptionConfig `envconfig:"TRANSCRIPTION"`
	Compression   CompressionConfig   `envconfig:"COMPRESSION"`

	// Processing settings
	Processing ProcessingConfig `envconfig:"PROCESSING"`

	// Security and compliance
	Security   SecurityConfig   `envconfig:"SECURITY"`
	Compliance ComplianceConfig `envconfig:"COMPLIANCE"`

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
	DB       int    `envconfig:"DB" default:"5"`
	PoolSize int    `envconfig:"POOL_SIZE" default:"10"`
}

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URL      string `envconfig:"URL" default:"nats://nats-client:4222"`
	Username string `envconfig:"USERNAME" default:"agba_user"`
	Password string `envconfig:"PASSWORD" required:"true"`
}

// StorageConfig holds storage configuration
type StorageConfig struct {
	DefaultProvider string `envconfig:"DEFAULT_PROVIDER" default:"file"` // file, s3, gcs
	
	// File storage configuration
	File FileStorageConfig `envconfig:"FILE"`
	
	// S3 storage configuration
	S3 S3StorageConfig `envconfig:"S3"`
	
	// Google Cloud Storage configuration
	GCS GCSStorageConfig `envconfig:"GCS"`
	
	// General storage settings
	MaxFileSize      int64         `envconfig:"MAX_FILE_SIZE" default:"1073741824"`      // 1GB
	AllowedFormats   []string      `envconfig:"ALLOWED_FORMATS" default:"wav,mp3,flac,ogg,m4a"`
	RetentionPeriod  time.Duration `envconfig:"RETENTION_PERIOD" default:"2160h"`       // 90 days
	ArchivePeriod    time.Duration `envconfig:"ARCHIVE_PERIOD" default:"8760h"`         // 365 days
	CompressionLevel int           `envconfig:"COMPRESSION_LEVEL" default:"6"`
	EnableEncryption bool          `envconfig:"ENABLE_ENCRYPTION" default:"true"`
	EncryptionKey    string        `envconfig:"ENCRYPTION_KEY" required:"true"`
}

// FileStorageConfig holds file storage configuration
type FileStorageConfig struct {
	Enabled   bool   `envconfig:"ENABLED" default:"true"`
	BasePath  string `envconfig:"BASE_PATH" default:"/app/recordings"`
	TempPath  string `envconfig:"TEMP_PATH" default:"/app/temp"`
	ChunkSize int64  `envconfig:"CHUNK_SIZE" default:"1048576"` // 1MB
}

// S3StorageConfig holds S3 storage configuration
type S3StorageConfig struct {
	Enabled         bool   `envconfig:"ENABLED" default:"false"`
	Region          string `envconfig:"REGION" default:"us-east-1"`
	Bucket          string `envconfig:"BUCKET"`
	AccessKeyID     string `envconfig:"ACCESS_KEY_ID"`
	SecretAccessKey string `envconfig:"SECRET_ACCESS_KEY"`
	Endpoint        string `envconfig:"ENDPOINT"` // For S3-compatible services
	UseSSL          bool   `envconfig:"USE_SSL" default:"true"`
	PathStyle       bool   `envconfig:"PATH_STYLE" default:"false"`
}

// GCSStorageConfig holds Google Cloud Storage configuration
type GCSStorageConfig struct {
	Enabled           bool   `envconfig:"ENABLED" default:"false"`
	ProjectID         string `envconfig:"PROJECT_ID"`
	Bucket            string `envconfig:"BUCKET"`
	CredentialsFile   string `envconfig:"CREDENTIALS_FILE"`
	CredentialsJSON   string `envconfig:"CREDENTIALS_JSON"`
}

// AudioConfig holds audio processing configuration
type AudioConfig struct {
	// Format settings
	DefaultFormat    string `envconfig:"DEFAULT_FORMAT" default:"wav"`
	DefaultSampleRate int   `envconfig:"DEFAULT_SAMPLE_RATE" default:"16000"`
	DefaultChannels   int   `envconfig:"DEFAULT_CHANNELS" default:"1"`
	DefaultBitDepth   int   `envconfig:"DEFAULT_BIT_DEPTH" default:"16"`
	
	// Quality settings
	EnableNormalization bool    `envconfig:"ENABLE_NORMALIZATION" default:"true"`
	NormalizationLevel  float64 `envconfig:"NORMALIZATION_LEVEL" default:"-23.0"`
	EnableNoiseReduction bool   `envconfig:"ENABLE_NOISE_REDUCTION" default:"true"`
	NoiseReductionLevel float64 `envconfig:"NOISE_REDUCTION_LEVEL" default:"0.5"`
	
	// Processing settings
	ChunkDuration    time.Duration `envconfig:"CHUNK_DURATION" default:"30s"`
	OverlapDuration  time.Duration `envconfig:"OVERLAP_DURATION" default:"1s"`
	MaxProcessingTime time.Duration `envconfig:"MAX_PROCESSING_TIME" default:"300s"`
	
	// Real-time settings
	BufferSize       int           `envconfig:"BUFFER_SIZE" default:"4096"`
	LatencyTarget    time.Duration `envconfig:"LATENCY_TARGET" default:"100ms"`
	EnableRealtime   bool          `envconfig:"ENABLE_REALTIME" default:"true"`
}

// TranscriptionConfig holds transcription configuration
type TranscriptionConfig struct {
	Enabled  bool   `envconfig:"ENABLED" default:"true"`
	Provider string `envconfig:"PROVIDER" default:"whisper"` // whisper, google, aws, azure
	
	// Whisper configuration
	Whisper WhisperConfig `envconfig:"WHISPER"`
	
	// Google Speech-to-Text configuration
	Google GoogleSTTConfig `envconfig:"GOOGLE"`
	
	// AWS Transcribe configuration
	AWS AWSTranscribeConfig `envconfig:"AWS"`
	
	// Azure Speech Services configuration
	Azure AzureSTTConfig `envconfig:"AZURE"`
	
	// General transcription settings
	Language         string        `envconfig:"LANGUAGE" default:"en-US"`
	EnablePunctuation bool         `envconfig:"ENABLE_PUNCTUATION" default:"true"`
	EnableDiarization bool         `envconfig:"ENABLE_DIARIZATION" default:"true"`
	MaxSpeakers      int           `envconfig:"MAX_SPEAKERS" default:"10"`
	Timeout          time.Duration `envconfig:"TIMEOUT" default:"300s"`
	RetryAttempts    int           `envconfig:"RETRY_ATTEMPTS" default:"3"`
	RetryDelay       time.Duration `envconfig:"RETRY_DELAY" default:"30s"`
}

// WhisperConfig holds Whisper configuration
type WhisperConfig struct {
	ModelPath    string  `envconfig:"MODEL_PATH" default:"/app/models/whisper"`
	ModelSize    string  `envconfig:"MODEL_SIZE" default:"base"`
	Temperature  float64 `envconfig:"TEMPERATURE" default:"0.0"`
	BeamSize     int     `envconfig:"BEAM_SIZE" default:"5"`
	BestOf       int     `envconfig:"BEST_OF" default:"5"`
	UseGPU       bool    `envconfig:"USE_GPU" default:"false"`
	GPUDevices   []int   `envconfig:"GPU_DEVICES" default:"0"`
}

// GoogleSTTConfig holds Google Speech-to-Text configuration
type GoogleSTTConfig struct {
	CredentialsFile string `envconfig:"CREDENTIALS_FILE"`
	CredentialsJSON string `envconfig:"CREDENTIALS_JSON"`
	ProjectID       string `envconfig:"PROJECT_ID"`
}

// AWSTranscribeConfig holds AWS Transcribe configuration
type AWSTranscribeConfig struct {
	Region          string `envconfig:"REGION" default:"us-east-1"`
	AccessKeyID     string `envconfig:"ACCESS_KEY_ID"`
	SecretAccessKey string `envconfig:"SECRET_ACCESS_KEY"`
}

// AzureSTTConfig holds Azure Speech Services configuration
type AzureSTTConfig struct {
	SubscriptionKey string `envconfig:"SUBSCRIPTION_KEY"`
	Region          string `envconfig:"REGION" default:"eastus"`
	Endpoint        string `envconfig:"ENDPOINT"`
}

// CompressionConfig holds compression configuration
type CompressionConfig struct {
	Enabled       bool   `envconfig:"ENABLED" default:"true"`
	Algorithm     string `envconfig:"ALGORITHM" default:"opus"` // opus, aac, mp3, flac
	Quality       int    `envconfig:"QUALITY" default:"6"`      // 0-10 scale
	Bitrate       int    `envconfig:"BITRATE" default:"64000"`  // bits per second
	EnableVBR     bool   `envconfig:"ENABLE_VBR" default:"true"`
	ComplexityLevel int  `envconfig:"COMPLEXITY_LEVEL" default:"10"`
}

// ProcessingConfig holds processing configuration
type ProcessingConfig struct {
	// Queue settings
	QueueSize        int           `envconfig:"QUEUE_SIZE" default:"1000"`
	WorkerCount      int           `envconfig:"WORKER_COUNT" default:"5"`
	BatchSize        int           `envconfig:"BATCH_SIZE" default:"10"`
	ProcessInterval  time.Duration `envconfig:"PROCESS_INTERVAL" default:"10s"`
	
	// Retry settings
	MaxRetries       int           `envconfig:"MAX_RETRIES" default:"3"`
	RetryDelay       time.Duration `envconfig:"RETRY_DELAY" default:"60s"`
	RetryBackoff     string        `envconfig:"RETRY_BACKOFF" default:"exponential"`
	RetryMaxDelay    time.Duration `envconfig:"RETRY_MAX_DELAY" default:"3600s"`
	
	// Cleanup settings
	CleanupInterval  time.Duration `envconfig:"CLEANUP_INTERVAL" default:"1h"`
	TempFileRetention time.Duration `envconfig:"TEMP_FILE_RETENTION" default:"24h"`
	FailedJobRetention time.Duration `envconfig:"FAILED_JOB_RETENTION" default:"168h"` // 7 days
	
	// Concurrency settings
	MaxConcurrentUploads    int `envconfig:"MAX_CONCURRENT_UPLOADS" default:"10"`
	MaxConcurrentDownloads  int `envconfig:"MAX_CONCURRENT_DOWNLOADS" default:"20"`
	MaxConcurrentProcessing int `envconfig:"MAX_CONCURRENT_PROCESSING" default:"5"`
	
	// Memory settings
	MaxMemoryUsage int64 `envconfig:"MAX_MEMORY_USAGE" default:"2147483648"` // 2GB
	BufferPoolSize int   `envconfig:"BUFFER_POOL_SIZE" default:"100"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	// Encryption settings
	EncryptionAlgorithm string `envconfig:"ENCRYPTION_ALGORITHM" default:"AES-256-GCM"`
	KeyRotationPeriod   time.Duration `envconfig:"KEY_ROTATION_PERIOD" default:"2160h"` // 90 days
	
	// Access control
	EnableAccessControl bool     `envconfig:"ENABLE_ACCESS_CONTROL" default:"true"`
	AllowedOrigins      []string `envconfig:"ALLOWED_ORIGINS" default:"*"`
	AllowedMethods      []string `envconfig:"ALLOWED_METHODS" default:"GET,POST,PUT,DELETE"`
	AllowedHeaders      []string `envconfig:"ALLOWED_HEADERS" default:"*"`
	
	// Authentication
	RequireAuthentication bool   `envconfig:"REQUIRE_AUTHENTICATION" default:"true"`
	JWTSecret            string `envconfig:"JWT_SECRET" required:"true"`
	APIKeyHeader         string `envconfig:"API_KEY_HEADER" default:"X-API-Key"`
	
	// Rate limiting
	EnableRateLimit    bool          `envconfig:"ENABLE_RATE_LIMIT" default:"true"`
	RateLimit          int           `envconfig:"RATE_LIMIT" default:"100"`        // requests per minute
	RateLimitWindow    time.Duration `envconfig:"RATE_LIMIT_WINDOW" default:"1m"`
	BurstLimit         int           `envconfig:"BURST_LIMIT" default:"20"`
	
	// File validation
	EnableFileValidation bool     `envconfig:"ENABLE_FILE_VALIDATION" default:"true"`
	MaxFileSize          int64    `envconfig:"MAX_FILE_SIZE" default:"1073741824"` // 1GB
	AllowedMimeTypes     []string `envconfig:"ALLOWED_MIME_TYPES" default:"audio/wav,audio/mpeg,audio/flac,audio/ogg"`
	ScanForMalware       bool     `envconfig:"SCAN_FOR_MALWARE" default:"true"`
}

// ComplianceConfig holds compliance configuration
type ComplianceConfig struct {
	// Data retention
	EnableDataRetention   bool          `envconfig:"ENABLE_DATA_RETENTION" default:"true"`
	DefaultRetentionPeriod time.Duration `envconfig:"DEFAULT_RETENTION_PERIOD" default:"2160h"` // 90 days
	MaxRetentionPeriod    time.Duration `envconfig:"MAX_RETENTION_PERIOD" default:"26280h"`    // 3 years
	
	// Audit logging
	EnableAuditLogging    bool `envconfig:"ENABLE_AUDIT_LOGGING" default:"true"`
	AuditLogLevel         string `envconfig:"AUDIT_LOG_LEVEL" default:"INFO"`
	AuditLogRetention     time.Duration `envconfig:"AUDIT_LOG_RETENTION" default:"26280h"` // 3 years
	
	// Privacy settings
	EnableDataAnonymization bool `envconfig:"ENABLE_DATA_ANONYMIZATION" default:"false"`
	AnonymizationDelay      time.Duration `envconfig:"ANONYMIZATION_DELAY" default:"720h"` // 30 days
	
	// Compliance standards
	EnableGDPR    bool `envconfig:"ENABLE_GDPR" default:"true"`
	EnableHIPAA   bool `envconfig:"ENABLE_HIPAA" default:"false"`
	EnableSOC2    bool `envconfig:"ENABLE_SOC2" default:"true"`
	EnablePCI     bool `envconfig:"ENABLE_PCI" default:"false"`
	
	// Data export
	EnableDataExport      bool          `envconfig:"ENABLE_DATA_EXPORT" default:"true"`
	ExportFormat          string        `envconfig:"EXPORT_FORMAT" default:"json"`
	ExportRetention       time.Duration `envconfig:"EXPORT_RETENTION" default:"168h"` // 7 days
	MaxExportSize         int64         `envconfig:"MAX_EXPORT_SIZE" default:"10737418240"` // 10GB
}

// MonitoringConfig holds monitoring configuration
type MonitoringConfig struct {
	Enabled           bool   `envconfig:"ENABLED" default:"true"`
	MetricsPath       string `envconfig:"METRICS_PATH" default:"/metrics"`
	RequestLogging    bool   `envconfig:"REQUEST_LOGGING" default:"true"`
	SlowRequestThresh time.Duration `envconfig:"SLOW_REQUEST_THRESHOLD" default:"1s"`
	
	// Performance monitoring
	EnablePerformanceMonitoring bool          `envconfig:"ENABLE_PERFORMANCE_MONITORING" default:"true"`
	PerformanceMetricsInterval  time.Duration `envconfig:"PERFORMANCE_METRICS_INTERVAL" default:"30s"`
	
	// Health checks
	HealthCheckInterval time.Duration `envconfig:"HEALTH_CHECK_INTERVAL" default:"30s"`
	HealthCheckTimeout  time.Duration `envconfig:"HEALTH_CHECK_TIMEOUT" default:"5s"`
	
	// Alerting
	EnableAlerting      bool          `envconfig:"ENABLE_ALERTING" default:"true"`
	AlertingInterval    time.Duration `envconfig:"ALERTING_INTERVAL" default:"60s"`
	ErrorRateThreshold  float64       `envconfig:"ERROR_RATE_THRESHOLD" default:"0.05"`
	LatencyThreshold    time.Duration `envconfig:"LATENCY_THRESHOLD" default:"5s"`
	StorageThreshold    float64       `envconfig:"STORAGE_THRESHOLD" default:"0.85"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	var cfg Config
	err := envconfig.Process("AGBA_RECORDING", &cfg)
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