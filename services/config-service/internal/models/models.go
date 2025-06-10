package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ConfigType represents the type of configuration
type ConfigType string

const (
	ConfigTypeAgent        ConfigType = "agent"
	ConfigTypeSystem       ConfigType = "system"
	ConfigTypeOrganization ConfigType = "organization"
	ConfigTypeTemplate     ConfigType = "template"
)

// ConfigStatus represents the status of a configuration
type ConfigStatus string

const (
	ConfigStatusDraft     ConfigStatus = "draft"
	ConfigStatusActive    ConfigStatus = "active"
	ConfigStatusInactive  ConfigStatus = "inactive"
	ConfigStatusArchived  ConfigStatus = "archived"
)

// Configuration represents a configuration entity
type Configuration struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	Service        string          `json:"service" db:"service"`
	Environment    string          `json:"environment" db:"environment"`
	Type           ConfigType      `json:"type" db:"type"`
	EntityID       uuid.UUID       `json:"entity_id" db:"entity_id"` // Agent ID, Organization ID, etc.
	Name           string          `json:"name" db:"name"`
	Description    string          `json:"description" db:"description"`
	Version        int             `json:"version" db:"version"`
	Status         ConfigStatus    `json:"status" db:"status"`
	Data           json.RawMessage `json:"data" db:"data"`
	Schema         json.RawMessage `json:"schema,omitempty" db:"schema"`
	Tags           []string        `json:"tags" db:"tags"`
	Metadata       map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedBy      uuid.UUID       `json:"created_by" db:"created_by"`
	UpdatedBy      uuid.UUID       `json:"updated_by" db:"updated_by"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
	ActivatedAt    *time.Time      `json:"activated_at,omitempty" db:"activated_at"`
	DeactivatedAt  *time.Time      `json:"deactivated_at,omitempty" db:"deactivated_at"`
	Checksum       string          `json:"checksum" db:"checksum"`
	Encrypted      bool            `json:"encrypted" db:"encrypted"`
	Compressed     bool            `json:"compressed" db:"compressed"`
}

// ConfigurationVersion represents a version of a configuration
type ConfigurationVersion struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	ConfigurationID uuid.UUID       `json:"configuration_id" db:"configuration_id"`
	Version         int             `json:"version" db:"version"`
	Data            json.RawMessage `json:"data" db:"data"`
	Schema          json.RawMessage `json:"schema,omitempty" db:"schema"`
	ChangeLog       string          `json:"change_log" db:"change_log"`
	CreatedBy       uuid.UUID       `json:"created_by" db:"created_by"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	Checksum        string          `json:"checksum" db:"checksum"`
	Encrypted       bool            `json:"encrypted" db:"encrypted"`
	Compressed      bool            `json:"compressed" db:"compressed"`
}

// ConfigTemplate represents a configuration template
type ConfigTemplate struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	Name        string          `json:"name" db:"name"`
	Description string          `json:"description" db:"description"`
	Category    string          `json:"category" db:"category"`
	Type        ConfigType      `json:"type" db:"type"`
	Template    json.RawMessage `json:"template" db:"template"`
	Schema      json.RawMessage `json:"schema" db:"schema"`
	Defaults    json.RawMessage `json:"defaults,omitempty" db:"defaults"`
	Variables   json.RawMessage `json:"variables,omitempty" db:"variables"`
	Tags        []string        `json:"tags" db:"tags"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedBy   uuid.UUID       `json:"created_by" db:"created_by"`
	UpdatedBy   uuid.UUID       `json:"updated_by" db:"updated_by"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
	IsPublic    bool            `json:"is_public" db:"is_public"`
	UsageCount  int             `json:"usage_count" db:"usage_count"`
}

// ConfigSchema represents a configuration schema
type ConfigSchema struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	Type        ConfigType      `json:"type" db:"type"`
	Version     string          `json:"version" db:"version"`
	Schema      json.RawMessage `json:"schema" db:"schema"`
	Description string          `json:"description" db:"description"`
	CreatedBy   uuid.UUID       `json:"created_by" db:"created_by"`
	UpdatedBy   uuid.UUID       `json:"updated_by" db:"updated_by"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
	IsActive    bool            `json:"is_active" db:"is_active"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	ConfigurationID uuid.UUID       `json:"configuration_id" db:"configuration_id"`
	Action          string          `json:"action" db:"action"`
	UserID          uuid.UUID       `json:"user_id" db:"user_id"`
	UserEmail       string          `json:"user_email" db:"user_email"`
	IPAddress       string          `json:"ip_address" db:"ip_address"`
	UserAgent       string          `json:"user_agent" db:"user_agent"`
	OldValue        json.RawMessage `json:"old_value,omitempty" db:"old_value"`
	NewValue        json.RawMessage `json:"new_value,omitempty" db:"new_value"`
	Changes         json.RawMessage `json:"changes,omitempty" db:"changes"`
	Timestamp       time.Time       `json:"timestamp" db:"timestamp"`
	Success         bool            `json:"success" db:"success"`
	ErrorMessage    string          `json:"error_message,omitempty" db:"error_message"`
}

// AgentConfig represents agent-specific configuration
type AgentConfig struct {
	// Basic settings
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=500"`
	Language    string `json:"language" validate:"required,oneof=en es fr de it pt"`
	Timezone    string `json:"timezone" validate:"required"`

	// Voice settings
	Voice VoiceConfig `json:"voice" validate:"required"`

	// Conversation settings
	Conversation ConversationConfig `json:"conversation" validate:"required"`

	// AI/LLM settings
	AI AIConfig `json:"ai" validate:"required"`

	// Telephony settings
	Telephony TelephonyConfig `json:"telephony" validate:"required"`

	// Analytics settings
	Analytics AnalyticsConfig `json:"analytics"`

	// Webhook settings
	Webhooks []WebhookConfig `json:"webhooks" validate:"dive"`

	// Custom settings
	Custom map[string]interface{} `json:"custom,omitempty"`
}

// VoiceConfig represents voice configuration
type VoiceConfig struct {
	Provider    string  `json:"provider" validate:"required,oneof=elevenlabs playht openai"`
	VoiceID     string  `json:"voice_id" validate:"required"`
	Speed       float64 `json:"speed" validate:"min=0.5,max=2.0"`
	Pitch       float64 `json:"pitch" validate:"min=0.5,max=2.0"`
	Volume      float64 `json:"volume" validate:"min=0.1,max=1.0"`
	Stability   float64 `json:"stability" validate:"min=0.0,max=1.0"`
	Clarity     float64 `json:"clarity" validate:"min=0.0,max=1.0"`
	StyleExaggeration float64 `json:"style_exaggeration" validate:"min=0.0,max=1.0"`
}

// ConversationConfig represents conversation configuration
type ConversationConfig struct {
	MaxTurns            int     `json:"max_turns" validate:"min=1,max=100"`
	TimeoutSeconds      int     `json:"timeout_seconds" validate:"min=30,max=3600"`
	SilenceTimeoutMs    int     `json:"silence_timeout_ms" validate:"min=1000,max=10000"`
	InterruptionEnabled bool    `json:"interruption_enabled"`
	BackchannelEnabled  bool    `json:"backchannel_enabled"`
	EndpointingSensitivity float64 `json:"endpointing_sensitivity" validate:"min=0.0,max=1.0"`
	NoiseSuppressionLevel  string  `json:"noise_suppression_level" validate:"oneof=low medium high"`
}

// AIConfig represents AI/LLM configuration
type AIConfig struct {
	Provider         string                 `json:"provider" validate:"required,oneof=openai anthropic google"`
	Model            string                 `json:"model" validate:"required"`
	Temperature      float64                `json:"temperature" validate:"min=0.0,max=2.0"`
	MaxTokens        int                    `json:"max_tokens" validate:"min=1,max=4096"`
	TopP             float64                `json:"top_p" validate:"min=0.0,max=1.0"`
	FrequencyPenalty float64                `json:"frequency_penalty" validate:"min=-2.0,max=2.0"`
	PresencePenalty  float64                `json:"presence_penalty" validate:"min=-2.0,max=2.0"`
	SystemPrompt     string                 `json:"system_prompt" validate:"required,min=10,max=2000"`
	Functions        []FunctionConfig       `json:"functions" validate:"dive"`
	Guardrails       GuardrailsConfig       `json:"guardrails"`
	Context          map[string]interface{} `json:"context,omitempty"`
}

// FunctionConfig represents function calling configuration
type FunctionConfig struct {
	Name        string                 `json:"name" validate:"required"`
	Description string                 `json:"description" validate:"required"`
	Parameters  map[string]interface{} `json:"parameters" validate:"required"`
	Endpoint    string                 `json:"endpoint" validate:"required,url"`
	Method      string                 `json:"method" validate:"required,oneof=GET POST PUT DELETE"`
	Headers     map[string]string      `json:"headers,omitempty"`
	Timeout     int                    `json:"timeout" validate:"min=1,max=300"`
}

// GuardrailsConfig represents AI guardrails configuration
type GuardrailsConfig struct {
	ContentFiltering    bool     `json:"content_filtering"`
	ProfanityFilter     bool     `json:"profanity_filter"`
	PIIDetection        bool     `json:"pii_detection"`
	ToxicityThreshold   float64  `json:"toxicity_threshold" validate:"min=0.0,max=1.0"`
	BlockedTopics       []string `json:"blocked_topics"`
	AllowedDomains      []string `json:"allowed_domains"`
	MaxResponseLength   int      `json:"max_response_length" validate:"min=1,max=1000"`
}

// TelephonyConfig represents telephony configuration
type TelephonyConfig struct {
	Provider           string            `json:"provider" validate:"required"`
	PhoneNumbers       []string          `json:"phone_numbers" validate:"dive,e164"`
	CallRecording      bool              `json:"call_recording"`
	CallTranscription  bool              `json:"call_transcription"`
	DTMFEnabled        bool              `json:"dtmf_enabled"`
	TransferEnabled    bool              `json:"transfer_enabled"`
	ConferenceEnabled  bool              `json:"conference_enabled"`
	VoicemailEnabled   bool              `json:"voicemail_enabled"`
	CallForwarding     CallForwardingConfig `json:"call_forwarding"`
	BusinessHours      BusinessHoursConfig  `json:"business_hours"`
}

// CallForwardingConfig represents call forwarding configuration
type CallForwardingConfig struct {
	Enabled     bool     `json:"enabled"`
	Numbers     []string `json:"numbers" validate:"dive,e164"`
	Conditions  []string `json:"conditions" validate:"dive,oneof=busy no_answer unavailable"`
	RingTimeout int      `json:"ring_timeout" validate:"min=5,max=60"`
}

// BusinessHoursConfig represents business hours configuration
type BusinessHoursConfig struct {
	Enabled   bool                    `json:"enabled"`
	Timezone  string                  `json:"timezone" validate:"required"`
	Schedule  map[string]DaySchedule  `json:"schedule"`
	Holidays  []Holiday               `json:"holidays"`
}

// DaySchedule represents a day's schedule
type DaySchedule struct {
	Open   string `json:"open" validate:"required"`   // HH:MM format
	Close  string `json:"close" validate:"required"`  // HH:MM format
	Closed bool   `json:"closed"`
}

// Holiday represents a holiday
type Holiday struct {
	Date        string `json:"date" validate:"required"`        // YYYY-MM-DD format
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

// AnalyticsConfig represents analytics configuration
type AnalyticsConfig struct {
	Enabled              bool     `json:"enabled"`
	SentimentAnalysis    bool     `json:"sentiment_analysis"`
	TopicExtraction      bool     `json:"topic_extraction"`
	KeywordTracking      []string `json:"keyword_tracking"`
	CallSummary          bool     `json:"call_summary"`
	PerformanceMetrics   bool     `json:"performance_metrics"`
	RealTimeAnalytics    bool     `json:"real_time_analytics"`
	DataRetentionDays    int      `json:"data_retention_days" validate:"min=1,max=365"`
}

// WebhookConfig represents webhook configuration
type WebhookConfig struct {
	Name        string            `json:"name" validate:"required"`
	URL         string            `json:"url" validate:"required,url"`
	Events      []string          `json:"events" validate:"required,dive,oneof=call_started call_ended call_failed agent_response user_input"`
	Headers     map[string]string `json:"headers,omitempty"`
	Secret      string            `json:"secret,omitempty"`
	Timeout     int               `json:"timeout" validate:"min=1,max=300"`
	Retries     int               `json:"retries" validate:"min=0,max=5"`
	Enabled     bool              `json:"enabled"`
}

// SystemConfig represents system-wide configuration
type SystemConfig struct {
	// Global settings
	DefaultLanguage string `json:"default_language" validate:"required"`
	DefaultTimezone string `json:"default_timezone" validate:"required"`
	
	// Rate limiting
	RateLimiting RateLimitingConfig `json:"rate_limiting"`
	
	// Security settings
	Security SecurityConfig `json:"security"`
	
	// Monitoring settings
	Monitoring MonitoringConfig `json:"monitoring"`
	
	// Feature flags
	Features FeatureFlags `json:"features"`
	
	// Integration settings
	Integrations IntegrationsConfig `json:"integrations"`
}

// RateLimitingConfig represents rate limiting configuration
type RateLimitingConfig struct {
	DefaultRPM      int `json:"default_rpm" validate:"min=1"`
	DefaultRPH      int `json:"default_rph" validate:"min=1"`
	BurstMultiplier int `json:"burst_multiplier" validate:"min=1,max=10"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	JWTExpirationHours    int      `json:"jwt_expiration_hours" validate:"min=1,max=168"`
	PasswordMinLength     int      `json:"password_min_length" validate:"min=8,max=128"`
	MFARequired           bool     `json:"mfa_required"`
	AllowedOrigins        []string `json:"allowed_origins"`
	IPWhitelist           []string `json:"ip_whitelist"`
	SessionTimeoutMinutes int      `json:"session_timeout_minutes" validate:"min=5,max=1440"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	MetricsEnabled    bool `json:"metrics_enabled"`
	LogLevel          string `json:"log_level" validate:"oneof=debug info warn error"`
	AlertingEnabled   bool `json:"alerting_enabled"`
	HealthCheckInterval int `json:"health_check_interval" validate:"min=10,max=300"`
}

// FeatureFlags represents feature flags
type FeatureFlags struct {
	WebRTCEnabled       bool `json:"webrtc_enabled"`
	BiometricsEnabled   bool `json:"biometrics_enabled"`
	AdvancedAnalytics   bool `json:"advanced_analytics"`
	VisualIVREnabled    bool `json:"visual_ivr_enabled"`
	SelfServicePortal   bool `json:"self_service_portal"`
	MultiLanguageSupport bool `json:"multi_language_support"`
}

// IntegrationsConfig represents integrations configuration
type IntegrationsConfig struct {
	CRM        CRMIntegration        `json:"crm"`
	Calendar   CalendarIntegration   `json:"calendar"`
	Email      EmailIntegration      `json:"email"`
	SMS        SMSIntegration        `json:"sms"`
	Analytics  AnalyticsIntegration  `json:"analytics"`
}

// Request/Response models for API operations

// CreateConfigRequest represents a request to create a configuration
type CreateConfigRequest struct {
	Service     string          `json:"service" validate:"required"`
	Environment string          `json:"environment" validate:"required"`
	Type        ConfigType      `json:"type" validate:"required"`
	EntityID    uuid.UUID       `json:"entity_id" validate:"required"`
	Name        string          `json:"name" validate:"required"`
	Description string          `json:"description"`
	Data        json.RawMessage `json:"data" validate:"required"`
	Schema      json.RawMessage `json:"schema,omitempty"`
	Tags        []string        `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// UpdateConfigRequest represents a request to update a configuration
type UpdateConfigRequest struct {
	Name        *string         `json:"name,omitempty"`
	Description *string         `json:"description,omitempty"`
	Data        json.RawMessage `json:"data,omitempty"`
	Schema      json.RawMessage `json:"schema,omitempty"`
	Tags        []string        `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Status      *ConfigStatus   `json:"status,omitempty"`
}

// ValidateConfigRequest represents a request to validate a configuration
type ValidateConfigRequest struct {
	Data   json.RawMessage `json:"data" validate:"required"`
	Schema json.RawMessage `json:"schema,omitempty"`
	Type   ConfigType      `json:"type" validate:"required"`
}

// ValidationResult represents the result of configuration validation
type ValidationResult struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// CreateTemplateRequest represents a request to create a configuration template
type CreateTemplateRequest struct {
	Name        string          `json:"name" validate:"required"`
	Description string          `json:"description"`
	Category    string          `json:"category"`
	Type        ConfigType      `json:"type" validate:"required"`
	Schema      json.RawMessage `json:"schema" validate:"required"`
	Defaults    json.RawMessage `json:"defaults,omitempty"`
	Variables   json.RawMessage `json:"variables,omitempty"`
	Tags        []string        `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
	IsPublic    bool            `json:"is_public"`
}

// ApplyTemplateRequest represents a request to apply a template
type ApplyTemplateRequest struct {
	TemplateID  uuid.UUID       `json:"template_id" validate:"required"`
	Service     string          `json:"service" validate:"required"`
	Environment string          `json:"environment" validate:"required"`
	EntityID    uuid.UUID       `json:"entity_id" validate:"required"`
	Variables   json.RawMessage `json:"variables,omitempty"`
	Name        string          `json:"name" validate:"required"`
	Description string          `json:"description"`
}

// CRMIntegration represents CRM integration configuration
type CRMIntegration struct {
	Provider    string            `json:"provider" validate:"oneof=salesforce hubspot pipedrive"`
	Enabled     bool              `json:"enabled"`
	APIKey      string            `json:"api_key,omitempty"`
	APISecret   string            `json:"api_secret,omitempty"`
	BaseURL     string            `json:"base_url,omitempty"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
}

// CalendarIntegration represents calendar integration configuration
type CalendarIntegration struct {
	Provider    string            `json:"provider" validate:"oneof=google outlook"`
	Enabled     bool              `json:"enabled"`
	ClientID    string            `json:"client_id,omitempty"`
	ClientSecret string           `json:"client_secret,omitempty"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
}

// EmailIntegration represents email integration configuration
type EmailIntegration struct {
	Provider    string            `json:"provider" validate:"oneof=sendgrid mailgun ses"`
	Enabled     bool              `json:"enabled"`
	APIKey      string            `json:"api_key,omitempty"`
	FromEmail   string            `json:"from_email" validate:"email"`
	FromName    string            `json:"from_name"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
}

// SMSIntegration represents SMS integration configuration
type SMSIntegration struct {
	Provider    string            `json:"provider" validate:"oneof=twilio messagebird"`
	Enabled     bool              `json:"enabled"`
	APIKey      string            `json:"api_key,omitempty"`
	APISecret   string            `json:"api_secret,omitempty"`
	FromNumber  string            `json:"from_number" validate:"e164"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
}

// AnalyticsIntegration represents analytics integration configuration
type AnalyticsIntegration struct {
	Provider    string            `json:"provider" validate:"oneof=google mixpanel amplitude"`
	Enabled     bool              `json:"enabled"`
	TrackingID  string            `json:"tracking_id,omitempty"`
	APIKey      string            `json:"api_key,omitempty"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
}