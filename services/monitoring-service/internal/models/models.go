package models

import (
	"time"

	"github.com/google/uuid"
)

// MetricType represents the type of metric
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

// AlertSeverity represents the severity of an alert
type AlertSeverity string

const (
	AlertSeverityCritical AlertSeverity = "critical"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityInfo     AlertSeverity = "info"
)

// AlertStatus represents the status of an alert
type AlertStatus string

const (
	AlertStatusFiring      AlertStatus = "firing"
	AlertStatusResolved    AlertStatus = "resolved"
	AlertStatusAcknowledged AlertStatus = "acknowledged"
	AlertStatusSilenced    AlertStatus = "silenced"
)

// ServiceStatus represents the status of a service
type ServiceStatus string

const (
	ServiceStatusHealthy   ServiceStatus = "healthy"
	ServiceStatusUnhealthy ServiceStatus = "unhealthy"
	ServiceStatusDegraded  ServiceStatus = "degraded"
	ServiceStatusUnknown   ServiceStatus = "unknown"
)

// Metric represents a metric data point
type Metric struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Type        MetricType             `json:"type" db:"type"`
	Value       float64                `json:"value" db:"value"`
	Unit        string                 `json:"unit" db:"unit"`
	Labels      map[string]string      `json:"labels" db:"labels"`
	Timestamp   time.Time              `json:"timestamp" db:"timestamp"`
	ServiceName string                 `json:"service_name" db:"service_name"`
	Instance    string                 `json:"instance" db:"instance"`
	Job         string                 `json:"job" db:"job"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}

// MetricSeries represents a time series of metrics
type MetricSeries struct {
	Name      string                 `json:"name"`
	Labels    map[string]string      `json:"labels"`
	Values    []MetricValue          `json:"values"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// MetricValue represents a single metric value with timestamp
type MetricValue struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// Alert represents an alert
type Alert struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	Service     string                 `json:"service" db:"service"`
	Metric      string                 `json:"metric" db:"metric"`
	Condition   string                 `json:"condition" db:"condition"`
	Threshold   float64                `json:"threshold" db:"threshold"`
	Severity    string                 `json:"severity" db:"severity"`
	Status      AlertStatus            `json:"status" db:"status"`
	Enabled     bool                   `json:"enabled" db:"enabled"`
	Labels      map[string]string      `json:"labels" db:"labels"`
	Annotations map[string]string      `json:"annotations" db:"annotations"`
	StartsAt    time.Time              `json:"starts_at" db:"starts_at"`
	EndsAt      *time.Time             `json:"ends_at,omitempty" db:"ends_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	
	// Alert rule information
	RuleID      uuid.UUID              `json:"rule_id" db:"rule_id"`
	RuleName    string                 `json:"rule_name" db:"rule_name"`
	
	// Notification information
	NotifiedAt  *time.Time             `json:"notified_at,omitempty" db:"notified_at"`
	AckedAt     *time.Time             `json:"acked_at,omitempty" db:"acked_at"`
	AckedBy     *uuid.UUID             `json:"acked_by,omitempty" db:"acked_by"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty" db:"resolved_at"`
	ResolvedBy  *uuid.UUID             `json:"resolved_by,omitempty" db:"resolved_by"`
	
	// User tracking
	CreatedBy   uuid.UUID              `json:"created_by" db:"created_by"`
	UpdatedBy   uuid.UUID              `json:"updated_by" db:"updated_by"`
	
	// Additional metadata
	Metadata    map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}

// AlertRule represents an alerting rule
type AlertRule struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	Query       string                 `json:"query" db:"query"`
	Condition   string                 `json:"condition" db:"condition"`
	Threshold   float64                `json:"threshold" db:"threshold"`
	Duration    time.Duration          `json:"duration" db:"duration"`
	Severity    AlertSeverity          `json:"severity" db:"severity"`
	Labels      map[string]string      `json:"labels" db:"labels"`
	Annotations map[string]string      `json:"annotations" db:"annotations"`
	
	// Rule configuration
	Enabled     bool                   `json:"enabled" db:"enabled"`
	GroupName   string                 `json:"group_name" db:"group_name"`
	Interval    time.Duration          `json:"interval" db:"interval"`
	
	// Notification settings
	NotificationChannels []string      `json:"notification_channels" db:"notification_channels"`
	
	// Metadata
	CreatedBy   uuid.UUID              `json:"created_by" db:"created_by"`
	UpdatedBy   uuid.UUID              `json:"updated_by" db:"updated_by"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	
	// Additional metadata
	Metadata    map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
}

// Service represents a monitored service
type Service struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	Status      ServiceStatus          `json:"status" db:"status"`
	Version     string                 `json:"version" db:"version"`
	Environment string                 `json:"environment" db:"environment"`
	
	// Service endpoints
	HealthEndpoint string             `json:"health_endpoint" db:"health_endpoint"`
	MetricsEndpoint string            `json:"metrics_endpoint" db:"metrics_endpoint"`
	
	// Service metadata
	Labels      map[string]string      `json:"labels" db:"labels"`
	Tags        []string               `json:"tags" db:"tags"`
	
	// Health check configuration
	HealthCheck HealthCheckConfig      `json:"health_check" db:"health_check"`
	
	// Status information
	LastHealthCheck *time.Time         `json:"last_health_check,omitempty" db:"last_health_check"`
	LastSeen        *time.Time         `json:"last_seen,omitempty" db:"last_seen"`
	Uptime          time.Duration      `json:"uptime" db:"uptime"`
	
	// Dependencies
	Dependencies []ServiceDependency   `json:"dependencies" db:"dependencies"`
	
	// Metadata
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
}

// HealthCheckConfig represents health check configuration for a service
type HealthCheckConfig struct {
	Enabled         bool          `json:"enabled"`
	Interval        time.Duration `json:"interval"`
	Timeout         time.Duration `json:"timeout"`
	FailureThreshold int          `json:"failure_threshold"`
	SuccessThreshold int          `json:"success_threshold"`
	Path            string        `json:"path"`
	Method          string        `json:"method"`
	ExpectedStatus  int           `json:"expected_status"`
	Headers         map[string]string `json:"headers,omitempty"`
}

// ServiceDependency represents a service dependency
type ServiceDependency struct {
	ServiceID   uuid.UUID     `json:"service_id"`
	ServiceName string        `json:"service_name"`
	Type        string        `json:"type"` // database, cache, api, etc.
	Critical    bool          `json:"critical"`
	Status      ServiceStatus `json:"status"`
}

// HealthCheck represents a health check result
type HealthCheck struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	ServiceID   uuid.UUID              `json:"service_id" db:"service_id"`
	ServiceName string                 `json:"service_name" db:"service_name"`
	Status      ServiceStatus          `json:"status" db:"status"`
	ResponseTime time.Duration         `json:"response_time" db:"response_time"`
	StatusCode  int                    `json:"status_code" db:"status_code"`
	Message     string                 `json:"message" db:"message"`
	Details     map[string]interface{} `json:"details,omitempty" db:"details"`
	Timestamp   time.Time              `json:"timestamp" db:"timestamp"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}

// Dashboard represents a monitoring dashboard
type Dashboard struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	Type        string                 `json:"type" db:"type"`
	TimeRange   string                 `json:"time_range" db:"time_range"`
	Tags        []string               `json:"tags" db:"tags"`
	Widgets     []Widget               `json:"widgets" db:"widgets"`
	
	// Dashboard configuration
	Config      DashboardConfig        `json:"config" db:"config"`
	Panels      []DashboardPanel       `json:"panels" db:"panels"`
	
	// Access control
	IsPublic    bool                   `json:"is_public" db:"is_public"`
	CreatedBy   uuid.UUID              `json:"created_by" db:"created_by"`
	UpdatedBy   uuid.UUID              `json:"updated_by" db:"updated_by"`
	
	// Metadata
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
}

// DashboardConfig represents dashboard configuration
type DashboardConfig struct {
	TimeRange   TimeRange              `json:"time_range"`
	RefreshRate time.Duration          `json:"refresh_rate"`
	Variables   []DashboardVariable    `json:"variables,omitempty"`
	Layout      DashboardLayout        `json:"layout"`
}

// TimeRange represents a time range for queries
type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// DashboardVariable represents a dashboard variable
type DashboardVariable struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Query   string   `json:"query,omitempty"`
	Options []string `json:"options,omitempty"`
	Default string   `json:"default,omitempty"`
}

// DashboardLayout represents dashboard layout configuration
type DashboardLayout struct {
	Columns int `json:"columns"`
	Rows    int `json:"rows"`
}

// DashboardPanel represents a dashboard panel
type DashboardPanel struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Type        string                 `json:"type"` // graph, stat, table, etc.
	Position    PanelPosition          `json:"position"`
	Size        PanelSize              `json:"size"`
	Query       string                 `json:"query"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// PanelPosition represents panel position on dashboard
type PanelPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// PanelSize represents panel size
type PanelSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Incident represents a system incident
type Incident struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	Title       string                 `json:"title" db:"title"`
	Description string                 `json:"description" db:"description"`
	Severity    AlertSeverity          `json:"severity" db:"severity"`
	Status      IncidentStatus         `json:"status" db:"status"`
	
	// Incident details
	AffectedServices []uuid.UUID       `json:"affected_services" db:"affected_services"`
	RootCause       string             `json:"root_cause" db:"root_cause"`
	Resolution      string             `json:"resolution" db:"resolution"`
	
	// Timeline
	StartedAt   time.Time              `json:"started_at" db:"started_at"`
	DetectedAt  time.Time              `json:"detected_at" db:"detected_at"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty" db:"resolved_at"`
	
	// Assignment
	AssignedTo  *uuid.UUID             `json:"assigned_to,omitempty" db:"assigned_to"`
	CreatedBy   uuid.UUID              `json:"created_by" db:"created_by"`
	UpdatedBy   uuid.UUID              `json:"updated_by" db:"updated_by"`
	
	// Metadata
	Labels      map[string]string      `json:"labels" db:"labels"`
	Tags        []string               `json:"tags" db:"tags"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
}

// IncidentStatus represents the status of an incident
type IncidentStatus string

const (
	IncidentStatusOpen       IncidentStatus = "open"
	IncidentStatusInProgress IncidentStatus = "in_progress"
	IncidentStatusResolved   IncidentStatus = "resolved"
	IncidentStatusClosed     IncidentStatus = "closed"
)

// SystemStatus represents overall system status
type SystemStatus struct {
	OverallStatus string                 `json:"overall_status"`
	Status        ServiceStatus          `json:"status"`
	Message       string                 `json:"message"`
	Services      []ServiceHealth        `json:"services"`
	Incidents     []IncidentSummary      `json:"incidents"`
	Metrics       SystemMetrics          `json:"metrics"`
	ActiveAlerts  int                    `json:"active_alerts"`
	Uptime        time.Duration          `json:"uptime"`
	Version       string                 `json:"version"`
	Environment   string                 `json:"environment"`
	Timestamp     time.Time              `json:"timestamp"`
	LastUpdated   time.Time              `json:"last_updated"`
}

// ServiceStatusSummary represents a summary of service status
type ServiceStatusSummary struct {
	Name        string        `json:"name"`
	Status      ServiceStatus `json:"status"`
	Uptime      time.Duration `json:"uptime"`
	LastCheck   time.Time     `json:"last_check"`
	ResponseTime time.Duration `json:"response_time,omitempty"`
}

// IncidentSummary represents a summary of an incident
type IncidentSummary struct {
	ID          uuid.UUID      `json:"id"`
	Title       string         `json:"title"`
	Severity    AlertSeverity  `json:"severity"`
	Status      IncidentStatus `json:"status"`
	StartedAt   time.Time      `json:"started_at"`
	ResolvedAt  *time.Time     `json:"resolved_at,omitempty"`
}

// SystemMetrics represents system-wide metrics
type SystemMetrics struct {
	TotalServices    int     `json:"total_services"`
	HealthyServices  int     `json:"healthy_services"`
	UnhealthyServices int    `json:"unhealthy_services"`
	OverallUptime    float64 `json:"overall_uptime"`
	ActiveAlerts     int     `json:"active_alerts"`
	CriticalAlerts   int     `json:"critical_alerts"`
	OpenIncidents    int     `json:"open_incidents"`
}

// NotificationChannel represents a notification channel
type NotificationChannel struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Type        string                 `json:"type" db:"type"` // slack, email, webhook, pagerduty
	Config      map[string]interface{} `json:"config" db:"config"`
	Enabled     bool                   `json:"enabled" db:"enabled"`
	
	// Filtering
	Severities  []AlertSeverity        `json:"severities" db:"severities"`
	Services    []string               `json:"services" db:"services"`
	Labels      map[string]string      `json:"labels" db:"labels"`
	
	// Metadata
	CreatedBy   uuid.UUID              `json:"created_by" db:"created_by"`
	UpdatedBy   uuid.UUID              `json:"updated_by" db:"updated_by"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// QueryRequest represents a metrics query request
type QueryRequest struct {
	Query     string                 `json:"query"`
	Start     time.Time              `json:"start,omitempty"`
	End       time.Time              `json:"end,omitempty"`
	Step      time.Duration          `json:"step,omitempty"`
	Timeout   time.Duration          `json:"timeout,omitempty"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// QueryResponse represents a metrics query response
type QueryResponse struct {
	Status string        `json:"status"`
	Data   QueryData     `json:"data"`
	Error  string        `json:"error,omitempty"`
}

// QueryData represents query result data
type QueryData struct {
	ResultType string         `json:"result_type"`
	Result     []MetricSeries `json:"result"`
}

// AlertNotification represents an alert notification
type AlertNotification struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	AlertID     uuid.UUID              `json:"alert_id" db:"alert_id"`
	ChannelID   uuid.UUID              `json:"channel_id" db:"channel_id"`
	ChannelType string                 `json:"channel_type" db:"channel_type"`
	Status      NotificationStatus     `json:"status" db:"status"`
	Message     string                 `json:"message" db:"message"`
	SentAt      *time.Time             `json:"sent_at,omitempty" db:"sent_at"`
	Error       string                 `json:"error,omitempty" db:"error"`
	Retries     int                    `json:"retries" db:"retries"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "pending"
	NotificationStatusSent    NotificationStatus = "sent"
	NotificationStatusFailed  NotificationStatus = "failed"
	NotificationStatusRetrying NotificationStatus = "retrying"
)

// MetricsResponse represents a response containing metrics data
type MetricsResponse struct {
	Service   string    `json:"service"`
	TimeRange string    `json:"time_range"`
	Metrics   []Metric  `json:"metrics"`
	Timestamp time.Time `json:"timestamp"`
}

// ServiceHealth represents the health status of a service
type ServiceHealth struct {
	ServiceName  string        `json:"service_name"`
	Status       string        `json:"status"`
	LastCheck    time.Time     `json:"last_check"`
	ResponseTime time.Duration `json:"response_time"`
	Checks       []HealthCheck `json:"checks"`
}

// SystemHealth represents the overall system health
type SystemHealth struct {
	OverallStatus string          `json:"overall_status"`
	Services      []ServiceHealth `json:"services"`
	Timestamp     time.Time       `json:"timestamp"`
}

// CreateAlertRequest represents a request to create an alert
type CreateAlertRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Service     string  `json:"service" validate:"required"`
	Metric      string  `json:"metric" validate:"required"`
	Condition   string  `json:"condition" validate:"required"`
	Threshold   float64 `json:"threshold" validate:"required"`
	Severity    string  `json:"severity" validate:"required"`
}

// UpdateAlertRequest represents a request to update an alert
type UpdateAlertRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Service     string  `json:"service"`
	Metric      string  `json:"metric"`
	Condition   string  `json:"condition"`
	Threshold   float64 `json:"threshold"`
	Severity    string  `json:"severity"`
	Enabled     bool    `json:"enabled"`
}

// AlertEvent represents an alert event in history
type AlertEvent struct {
	ID        uuid.UUID `json:"id"`
	AlertID   uuid.UUID `json:"alert_id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Value     float64   `json:"value"`
	Threshold float64   `json:"threshold"`
	Timestamp time.Time `json:"timestamp"`
}

// Widget represents a dashboard widget
type Widget struct {
	ID    string                 `json:"id"`
	Type  string                 `json:"type"`
	Title string                 `json:"title"`
	Data  map[string]interface{} `json:"data"`
}

// LogEntry represents a log entry
type LogEntry struct {
	ID        uuid.UUID              `json:"id"`
	Service   string                 `json:"service"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// PerformanceMetrics represents performance metrics data
type PerformanceMetrics struct {
	Service    string     `json:"service"`
	Metric     string     `json:"metric"`
	TimeRange  string     `json:"time_range"`
	DataPoints []DataPoint `json:"data_points"`
	Statistics Statistics `json:"statistics"`
}

// DataPoint represents a single data point in a time series
type DataPoint struct {
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

// Statistics represents statistical data for metrics
type Statistics struct {
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Average float64 `json:"average"`
	P95     float64 `json:"p95"`
	P99     float64 `json:"p99"`
}

// TriggerAlertRequest represents a request to trigger an alert
type TriggerAlertRequest struct {
	Value   float64 `json:"value" validate:"required"`
	Message string  `json:"message"`
}

// AcknowledgeAlertRequest represents a request to acknowledge an alert
type AcknowledgeAlertRequest struct {
	Comment string `json:"comment"`
}

// ResolveAlertRequest represents a request to resolve an alert
type ResolveAlertRequest struct {
	Resolution string `json:"resolution" validate:"required"`
}

// ActiveAlert represents an active alert
type ActiveAlert struct {
	ID          uuid.UUID  `json:"id"`
	AlertID     uuid.UUID  `json:"alert_id"`
	Name        string     `json:"name"`
	Service     string     `json:"service"`
	Severity    string     `json:"severity"`
	Message     string     `json:"message"`
	Value       float64    `json:"value,omitempty"`
	Threshold   float64    `json:"threshold,omitempty"`
	TriggeredAt time.Time  `json:"triggered_at"`
}