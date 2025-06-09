package models

import (
	"time"

	"github.com/google/uuid"
)

// ServiceHealth represents the health status of a backend service
type ServiceHealth struct {
	ServiceName string    `json:"service_name"`
	Healthy     bool      `json:"healthy"`
	LastCheck   time.Time `json:"last_check"`
	Error       string    `json:"error,omitempty"`
	ResponseTime time.Duration `json:"response_time"`
}

// GatewayStatus represents the overall status of the API Gateway
type GatewayStatus struct {
	Status      string          `json:"status"`
	Version     string          `json:"version"`
	Environment string          `json:"environment"`
	Uptime      time.Duration   `json:"uptime"`
	Services    []ServiceHealth `json:"services"`
	Timestamp   time.Time       `json:"timestamp"`
}

// RateLimitInfo represents rate limiting information
type RateLimitInfo struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	ResetTime time.Time `json:"reset_time"`
	RetryAfter time.Duration `json:"retry_after,omitempty"`
}

// APIError represents a standardized API error response
type APIError struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Details   interface{} `json:"details,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// ProxyRequest represents a request to be proxied to a backend service
type ProxyRequest struct {
	ServiceName string            `json:"service_name"`
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Headers     map[string]string `json:"headers"`
	Body        interface{}       `json:"body,omitempty"`
	QueryParams map[string]string `json:"query_params,omitempty"`
}

// ProxyResponse represents a response from a backend service
type ProxyResponse struct {
	StatusCode  int               `json:"status_code"`
	Headers     map[string]string `json:"headers"`
	Body        interface{}       `json:"body"`
	Duration    time.Duration     `json:"duration"`
	ServiceName string            `json:"service_name"`
}

// AuthContext represents authentication context
type AuthContext struct {
	UserID       uuid.UUID `json:"user_id"`
	Email        string    `json:"email"`
	Roles        []string  `json:"roles"`
	Permissions  []string  `json:"permissions"`
	Organization string    `json:"organization"`
	TokenType    string    `json:"token_type"` // jwt, api_key
	ExpiresAt    time.Time `json:"expires_at"`
}

// RouteInfo represents information about an API route
type RouteInfo struct {
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	Description string   `json:"description"`
	Service     string   `json:"service,omitempty"`
	AuthRequired bool    `json:"auth_required"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// MetricsData represents metrics information
type MetricsData struct {
	RequestCount    int64         `json:"request_count"`
	ErrorCount      int64         `json:"error_count"`
	AverageLatency  time.Duration `json:"average_latency"`
	P95Latency      time.Duration `json:"p95_latency"`
	P99Latency      time.Duration `json:"p99_latency"`
	RateLimitHits   int64         `json:"rate_limit_hits"`
	ActiveConnections int         `json:"active_connections"`
	Timestamp       time.Time     `json:"timestamp"`
}

// ServiceMetrics represents metrics for a specific service
type ServiceMetrics struct {
	ServiceName string      `json:"service_name"`
	Metrics     MetricsData `json:"metrics"`
	Health      ServiceHealth `json:"health"`
}

// GatewayMetrics represents overall gateway metrics
type GatewayMetrics struct {
	Gateway  MetricsData      `json:"gateway"`
	Services []ServiceMetrics `json:"services"`
	Period   string           `json:"period"`
	From     time.Time        `json:"from"`
	To       time.Time        `json:"to"`
}

// ConfigUpdate represents a configuration update request
type ConfigUpdate struct {
	Service     string                 `json:"service"`
	Environment string                 `json:"environment"`
	Config      map[string]interface{} `json:"config"`
	UpdatedBy   string                 `json:"updated_by"`
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// LoadBalancerConfig represents load balancer configuration
type LoadBalancerConfig struct {
	Algorithm string   `json:"algorithm"` // round_robin, least_connections, weighted
	Targets   []Target `json:"targets"`
	HealthCheck HealthCheckConfig `json:"health_check"`
}

// Target represents a load balancer target
type Target struct {
	URL    string `json:"url"`
	Weight int    `json:"weight"`
	Active bool   `json:"active"`
}

// HealthCheckConfig represents health check configuration
type HealthCheckConfig struct {
	Path     string        `json:"path"`
	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`
	Retries  int           `json:"retries"`
}

// CircuitBreakerConfig represents circuit breaker configuration
type CircuitBreakerConfig struct {
	MaxFailures int           `json:"max_failures"`
	Timeout     time.Duration `json:"timeout"`
	ResetTime   time.Duration `json:"reset_time"`
}

// RetryConfig represents retry configuration
type RetryConfig struct {
	MaxAttempts int           `json:"max_attempts"`
	BackoffType string        `json:"backoff_type"` // fixed, exponential
	BaseDelay   time.Duration `json:"base_delay"`
	MaxDelay    time.Duration `json:"max_delay"`
}

// ServiceConfig represents configuration for a backend service
type ServiceConfig struct {
	Name            string                 `json:"name"`
	BaseURL         string                 `json:"base_url"`
	Timeout         time.Duration          `json:"timeout"`
	LoadBalancer    LoadBalancerConfig     `json:"load_balancer,omitempty"`
	CircuitBreaker  CircuitBreakerConfig   `json:"circuit_breaker,omitempty"`
	Retry           RetryConfig            `json:"retry,omitempty"`
	Headers         map[string]string      `json:"headers,omitempty"`
	Authentication  string                 `json:"authentication,omitempty"`
	RateLimit       map[string]interface{} `json:"rate_limit,omitempty"`
}

// APIDocumentation represents API documentation
type APIDocumentation struct {
	Title       string      `json:"title"`
	Version     string      `json:"version"`
	Description string      `json:"description"`
	BaseURL     string      `json:"base_url"`
	Routes      []RouteInfo `json:"routes"`
	Services    []string    `json:"services"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// OpenAPISpec represents OpenAPI specification
type OpenAPISpec struct {
	OpenAPI    string                 `json:"openapi"`
	Info       map[string]interface{} `json:"info"`
	Servers    []map[string]string    `json:"servers"`
	Paths      map[string]interface{} `json:"paths"`
	Components map[string]interface{} `json:"components"`
	Security   []map[string][]string  `json:"security"`
}

// RequestLog represents a request log entry
type RequestLog struct {
	RequestID    string            `json:"request_id"`
	Method       string            `json:"method"`
	Path         string            `json:"path"`
	StatusCode   int               `json:"status_code"`
	Duration     time.Duration     `json:"duration"`
	ClientIP     string            `json:"client_ip"`
	UserAgent    string            `json:"user_agent"`
	UserID       string            `json:"user_id,omitempty"`
	Service      string            `json:"service,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
	Error        string            `json:"error,omitempty"`
	Timestamp    time.Time         `json:"timestamp"`
}

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	Type        string            `json:"type"` // auth_failure, rate_limit, suspicious_activity
	Severity    string            `json:"severity"` // low, medium, high, critical
	ClientIP    string            `json:"client_ip"`
	UserID      string            `json:"user_id,omitempty"`
	UserAgent   string            `json:"user_agent,omitempty"`
	Path        string            `json:"path"`
	Details     map[string]interface{} `json:"details"`
	Timestamp   time.Time         `json:"timestamp"`
	RequestID   string            `json:"request_id,omitempty"`
}

// MaintenanceMode represents maintenance mode configuration
type MaintenanceMode struct {
	Enabled   bool      `json:"enabled"`
	Message   string    `json:"message"`
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
	Services  []string  `json:"services,omitempty"` // specific services in maintenance
}

// CacheConfig represents caching configuration
type CacheConfig struct {
	Enabled    bool          `json:"enabled"`
	TTL        time.Duration `json:"ttl"`
	MaxSize    int64         `json:"max_size"`
	KeyPattern string        `json:"key_pattern"`
}

// CompressionConfig represents compression configuration
type CompressionConfig struct {
	Enabled     bool     `json:"enabled"`
	MinSize     int      `json:"min_size"`
	Types       []string `json:"types"`
	Level       int      `json:"level"`
}

// TLSConfig represents TLS configuration
type TLSConfig struct {
	Enabled     bool     `json:"enabled"`
	CertFile    string   `json:"cert_file"`
	KeyFile     string   `json:"key_file"`
	MinVersion  string   `json:"min_version"`
	CipherSuites []string `json:"cipher_suites"`
}