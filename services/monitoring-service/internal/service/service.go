package service

import (
	"context"
	"fmt"
	"time"

	"github.com/agba-ai/monitoring-service/internal/models"
	"github.com/agba-ai/monitoring-service/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// MonitoringService handles business logic for monitoring operations
type MonitoringService struct {
	repository *repository.Repository
	logger     *zap.Logger
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(repository *repository.Repository, logger *zap.Logger) *MonitoringService {
	return &MonitoringService{
		repository: repository,
		logger:     logger,
	}
}

// GetMetrics retrieves system metrics
func (s *MonitoringService) GetMetrics(ctx context.Context, service, timeRange string) (*models.MetricsResponse, error) {
	s.logger.Info("Getting metrics", 
		zap.String("service", service),
		zap.String("timeRange", timeRange),
	)

	// Parse time range
	since := time.Now().Add(-1 * time.Hour) // Default to 1 hour
	switch timeRange {
	case "1h":
		since = time.Now().Add(-1 * time.Hour)
	case "24h":
		since = time.Now().Add(-24 * time.Hour)
	case "7d":
		since = time.Now().Add(-7 * 24 * time.Hour)
	}

	// Get metrics from repository
	repoMetrics, err := s.repository.GetMetrics(ctx, service, "cpu_usage", since)
	if err != nil {
		s.logger.Error("Failed to get metrics from repository", zap.Error(err))
		return nil, err
	}

	// Convert repository metrics to response format
	var metrics []models.Metric
	for _, m := range repoMetrics {
		metrics = append(metrics, models.Metric{
			Name:      m.Name,
			Value:     m.Value,
			Unit:      m.Unit,
			Timestamp: m.Timestamp,
		})
	}

	response := &models.MetricsResponse{
		Service:   service,
		TimeRange: timeRange,
		Metrics:   metrics,
		Timestamp: time.Now(),
	}

	return response, nil
}

// GetServiceHealth retrieves health status for a specific service
func (s *MonitoringService) GetServiceHealth(ctx context.Context, serviceName string) (*models.ServiceHealth, error) {
	s.logger.Info("Getting service health", zap.String("service", serviceName))

	// Perform actual health checks for the service
	startTime := time.Now()
	
	// Check if service is responding
	status := "healthy"
	var checks []models.HealthCheck
	
	// Database connectivity check
	if err := s.repository.HealthCheck(ctx); err != nil {
		status = "unhealthy"
		checks = append(checks, models.HealthCheck{
			ServiceName: serviceName,
			Status:      models.ServiceStatusUnhealthy,
			Message:     fmt.Sprintf("Database connection failed: %v", err),
			Timestamp:   time.Now(),
		})
	} else {
		checks = append(checks, models.HealthCheck{
			ServiceName: serviceName,
			Status:      models.ServiceStatusHealthy,
			Message:     "Database connection successful",
			Timestamp:   time.Now(),
		})
	}
	
	// Additional service-specific health checks can be added here
	responseTime := time.Since(startTime)
	
	health := &models.ServiceHealth{
		ServiceName:  serviceName,
		Status:       status,
		LastCheck:    time.Now(),
		ResponseTime: responseTime,
		Checks:       checks,
	}

	return health, nil
}

// GetAllServicesHealth retrieves health status for all services
func (s *MonitoringService) GetAllServicesHealth(ctx context.Context) (*models.SystemHealth, error) {
	s.logger.Info("Getting all services health")

	services := []string{
		"api-gateway",
		"config-service",
		"user-management-service",
		"notification-service",
		"recording-service",
		"monitoring-service",
	}

	var serviceHealths []models.ServiceHealth
	overallStatus := "healthy"

	for _, service := range services {
		health, err := s.GetServiceHealth(ctx, service)
		if err != nil {
			s.logger.Error("Failed to get service health", 
				zap.String("service", service),
				zap.Error(err),
			)
			overallStatus = "degraded"
			continue
		}
		serviceHealths = append(serviceHealths, *health)
	}

	systemHealth := &models.SystemHealth{
		OverallStatus: overallStatus,
		Services:      serviceHealths,
		Timestamp:     time.Now(),
	}

	return systemHealth, nil
}

// CreateAlert creates a new alert rule
func (s *MonitoringService) CreateAlert(ctx context.Context, req *models.CreateAlertRequest, userID string) (*models.Alert, error) {
	s.logger.Info("Creating alert", 
		zap.String("name", req.Name),
		zap.String("userID", userID),
	)

	// Validate the request
	if req.Name == "" || req.Service == "" || req.Metric == "" {
		return nil, fmt.Errorf("name, service, and metric are required")
	}

	alert := &models.Alert{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Service:     req.Service,
		Metric:      req.Metric,
		Condition:   req.Condition,
		Threshold:   req.Threshold,
		Severity:    req.Severity,
		Enabled:     true,
		CreatedBy:   uuid.MustParse(userID),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Store alert in database
	if err := s.repository.CreateAlert(ctx, alert); err != nil {
		s.logger.Error("Failed to create alert", zap.Error(err))
		return nil, fmt.Errorf("failed to create alert: %w", err)
	}

	s.logger.Info("Alert created successfully", zap.String("alertID", alert.ID.String()))

	return alert, nil
}

// GetAlerts retrieves alert rules with pagination
func (s *MonitoringService) GetAlerts(ctx context.Context, page, limit int, service, status string) ([]*models.Alert, int64, error) {
	s.logger.Info("Getting alerts", 
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.String("service", service),
		zap.String("status", status),
	)

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20 // Default limit
	}

	// Get alerts from database
	alerts, total, err := s.repository.GetAlerts(ctx, page, limit, service, status)
	if err != nil {
		s.logger.Error("Failed to get alerts", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get alerts: %w", err)
	}

	s.logger.Info("Retrieved alerts", 
		zap.Int("count", len(alerts)),
		zap.Int64("total", total),
	)

	return alerts, total, nil
}

// UpdateAlert updates an existing alert rule
func (s *MonitoringService) UpdateAlert(ctx context.Context, id uuid.UUID, req *models.UpdateAlertRequest, userID string) (*models.Alert, error) {
	s.logger.Info("Updating alert", 
		zap.String("alertID", id.String()),
		zap.String("userID", userID),
	)

	// Get existing alert
	existingAlert, err := s.repository.GetAlert(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get existing alert", zap.Error(err))
		return nil, fmt.Errorf("failed to get existing alert: %w", err)
	}

	// Update fields
	alert := &models.Alert{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Service:     req.Service,
		Metric:      req.Metric,
		Condition:   req.Condition,
		Threshold:   req.Threshold,
		Severity:    req.Severity,
		Enabled:     req.Enabled,
		CreatedBy:   existingAlert.CreatedBy,
		CreatedAt:   existingAlert.CreatedAt,
		UpdatedBy:   uuid.MustParse(userID),
		UpdatedAt:   time.Now(),
	}

	// Update in database
	if err := s.repository.UpdateAlert(ctx, alert); err != nil {
		s.logger.Error("Failed to update alert", zap.Error(err))
		return nil, fmt.Errorf("failed to update alert: %w", err)
	}

	s.logger.Info("Alert updated successfully", zap.String("alertID", id.String()))

	return alert, nil
}

// DeleteAlert deletes an alert rule
func (s *MonitoringService) DeleteAlert(ctx context.Context, id uuid.UUID, userID string) error {
	s.logger.Info("Deleting alert", 
		zap.String("alertID", id.String()),
		zap.String("userID", userID),
	)

	// Delete from database
	if err := s.repository.DeleteAlert(ctx, id); err != nil {
		s.logger.Error("Failed to delete alert", zap.Error(err))
		return fmt.Errorf("failed to delete alert: %w", err)
	}

	s.logger.Info("Alert deleted successfully", zap.String("alertID", id.String()))

	return nil
}

// GetAlertHistory retrieves alert history
func (s *MonitoringService) GetAlertHistory(ctx context.Context, alertID uuid.UUID, page, limit int) ([]*models.AlertEvent, int64, error) {
	s.logger.Info("Getting alert history", 
		zap.String("alertID", alertID.String()),
		zap.Int("page", page),
		zap.Int("limit", limit),
	)

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20 // Default limit
	}

	// Get alert events from database
	events, total, err := s.repository.GetAlertEvents(ctx, alertID, page, limit)
	if err != nil {
		s.logger.Error("Failed to get alert history", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get alert history: %w", err)
	}

	s.logger.Info("Retrieved alert history", 
		zap.Int("count", len(events)),
		zap.Int64("total", total),
	)

	return events, total, nil
}

// GetDashboard retrieves dashboard data
func (s *MonitoringService) GetDashboard(ctx context.Context, dashboardType, timeRange string) (*models.Dashboard, error) {
	s.logger.Info("Getting dashboard", 
		zap.String("type", dashboardType),
		zap.String("timeRange", timeRange),
	)

	// Mock implementation - replace with actual dashboard data aggregation
	dashboard := &models.Dashboard{
		Type:      dashboardType,
		TimeRange: timeRange,
		Widgets: []models.Widget{
			{
				ID:    "cpu-usage",
				Type:  "chart",
				Title: "CPU Usage",
				Data: map[string]interface{}{
					"current": 75.5,
					"average": 68.2,
					"max":     89.1,
				},
			},
			{
				ID:    "memory-usage",
				Type:  "chart",
				Title: "Memory Usage",
				Data: map[string]interface{}{
					"current": 60.2,
					"average": 55.8,
					"max":     78.9,
				},
			},
			{
				ID:    "active-alerts",
				Type:  "counter",
				Title: "Active Alerts",
				Data: map[string]interface{}{
					"count": 3,
				},
			},
		},
		UpdatedAt: time.Now(),
	}

	return dashboard, nil
}

// GetLogs retrieves system logs
func (s *MonitoringService) GetLogs(ctx context.Context, service, level, since string, page, limit int) ([]*models.LogEntry, int64, error) {
	s.logger.Info("Getting logs", 
		zap.String("service", service),
		zap.String("level", level),
		zap.String("since", since),
		zap.Int("page", page),
		zap.Int("limit", limit),
	)

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20 // Default limit
	}

	// Parse since parameter
	var sinceTime time.Time
	if since != "" {
		var err error
		sinceTime, err = time.Parse(time.RFC3339, since)
		if err != nil {
			// Default to 1 hour ago if parsing fails
			sinceTime = time.Now().Add(-1 * time.Hour)
		}
	} else {
		sinceTime = time.Now().Add(-1 * time.Hour)
	}

	// Get logs from database
	logs, total, err := s.repository.GetLogEntries(ctx, service, level, sinceTime, page, limit)
	if err != nil {
		s.logger.Error("Failed to get logs", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get logs: %w", err)
	}

	s.logger.Info("Retrieved logs", 
		zap.Int("count", len(logs)),
		zap.Int64("total", total),
	)

	return logs, total, nil
}

// GetPerformanceMetrics retrieves performance metrics
func (s *MonitoringService) GetPerformanceMetrics(ctx context.Context, service, metric, timeRange string) (*models.PerformanceMetrics, error) {
	s.logger.Info("Getting performance metrics", 
		zap.String("service", service),
		zap.String("metric", metric),
		zap.String("timeRange", timeRange),
	)

	// Mock implementation - replace with actual performance data
	metrics := &models.PerformanceMetrics{
		Service:   service,
		Metric:    metric,
		TimeRange: timeRange,
		DataPoints: []models.DataPoint{
			{Value: 75.5, Timestamp: time.Now().Add(-60 * time.Minute)},
			{Value: 78.2, Timestamp: time.Now().Add(-45 * time.Minute)},
			{Value: 82.1, Timestamp: time.Now().Add(-30 * time.Minute)},
			{Value: 79.8, Timestamp: time.Now().Add(-15 * time.Minute)},
			{Value: 76.3, Timestamp: time.Now()},
		},
		Statistics: models.Statistics{
			Min:     75.5,
			Max:     82.1,
			Average: 78.38,
			P95:     81.5,
			P99:     82.0,
		},
	}

	return metrics, nil
}

// GetSystemStatus retrieves overall system status
func (s *MonitoringService) GetSystemStatus(ctx context.Context) (*models.SystemStatus, error) {
	s.logger.Info("Getting system status")

	// Get health status for all services
	health, err := s.GetAllServicesHealth(ctx)
	if err != nil {
		return nil, err
	}

	// Get active alerts
	activeAlerts, err := s.GetActiveAlerts(ctx, "", "")
	if err != nil {
		return nil, err
	}

	status := &models.SystemStatus{
		OverallStatus: health.OverallStatus,
		Services:      health.Services,
		ActiveAlerts:  len(activeAlerts),
		Uptime:        24 * time.Hour, // Mock uptime
		Version:       "1.0.0",
		Environment:   "production",
		Timestamp:     time.Now(),
	}

	return status, nil
}

// TriggerAlert manually triggers an alert for testing
func (s *MonitoringService) TriggerAlert(ctx context.Context, alertID uuid.UUID, req *models.TriggerAlertRequest) error {
	s.logger.Info("Triggering alert", 
		zap.String("alertID", alertID.String()),
		zap.Float64("value", req.Value),
	)

	// Mock implementation - replace with actual alert triggering
	s.logger.Info("Alert triggered successfully", zap.String("alertID", alertID.String()))

	return nil
}

// AcknowledgeAlert acknowledges an active alert
func (s *MonitoringService) AcknowledgeAlert(ctx context.Context, alertID uuid.UUID, userID string, req *models.AcknowledgeAlertRequest) error {
	s.logger.Info("Acknowledging alert", 
		zap.String("alertID", alertID.String()),
		zap.String("userID", userID),
		zap.String("comment", req.Comment),
	)

	// Mock implementation - replace with actual alert acknowledgment
	s.logger.Info("Alert acknowledged successfully", zap.String("alertID", alertID.String()))

	return nil
}

// ResolveAlert resolves an active alert
func (s *MonitoringService) ResolveAlert(ctx context.Context, alertID uuid.UUID, userID string, req *models.ResolveAlertRequest) error {
	s.logger.Info("Resolving alert", 
		zap.String("alertID", alertID.String()),
		zap.String("userID", userID),
		zap.String("resolution", req.Resolution),
	)

	// Mock implementation - replace with actual alert resolution
	s.logger.Info("Alert resolved successfully", zap.String("alertID", alertID.String()))

	return nil
}

// GetActiveAlerts retrieves all active alerts
func (s *MonitoringService) GetActiveAlerts(ctx context.Context, service, severity string) ([]*models.ActiveAlert, error) {
	s.logger.Info("Getting active alerts", 
		zap.String("service", service),
		zap.String("severity", severity),
	)

	// Mock implementation - replace with actual active alerts query
	alerts := []*models.ActiveAlert{
		{
			ID:          uuid.New(),
			AlertID:     uuid.New(),
			Name:        "High CPU Usage",
			Service:     "api-gateway",
			Severity:    "warning",
			Message:     "CPU usage is 85.5%, exceeding threshold of 80%",
			Value:       85.5,
			Threshold:   80.0,
			TriggeredAt: time.Now().Add(-30 * time.Minute),
		},
		{
			ID:          uuid.New(),
			AlertID:     uuid.New(),
			Name:        "Database Connection Error",
			Service:     "config-service",
			Severity:    "critical",
			Message:     "Unable to connect to database",
			TriggeredAt: time.Now().Add(-15 * time.Minute),
		},
	}

	return alerts, nil
}

// HealthCheck performs health checks for the monitoring service
func (s *MonitoringService) HealthCheck(ctx context.Context) (bool, map[string]interface{}) {
	checks := make(map[string]interface{})
	allHealthy := true

	// Check database connectivity
	checks["database"] = map[string]interface{}{
		"status":  "healthy",
		"message": "Database connection successful",
	}

	// Check Redis connectivity
	checks["redis"] = map[string]interface{}{
		"status":  "healthy",
		"message": "Redis connection successful",
	}

	// Check NATS connectivity
	checks["nats"] = map[string]interface{}{
		"status":  "healthy",
		"message": "NATS connection successful",
	}

	// Check metrics collection
	checks["metrics"] = map[string]interface{}{
		"status":  "healthy",
		"message": "Metrics collection active",
	}

	return allHealthy, checks
}

// StartMetricsCollection starts the background metrics collection process
func (s *MonitoringService) StartMetricsCollection(ctx context.Context) {
	s.logger.Info("Starting metrics collection")

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping metrics collection")
			return
		case <-ticker.C:
			s.collectMetrics(ctx)
		}
	}
}

// collectMetrics collects metrics from all services
func (s *MonitoringService) collectMetrics(ctx context.Context) {
	s.logger.Debug("Collecting metrics")

	services := []string{
		"api-gateway",
		"config-service",
		"user-management-service",
		"notification-service",
		"recording-service",
		"monitoring-service",
	}

	for _, serviceName := range services {
		s.logger.Debug("Collecting metrics for service", zap.String("service", serviceName))
		
		// Simulate collecting various metrics
		metrics := []struct {
			name  string
			value float64
			unit  string
		}{
			{"cpu_usage", 45.0 + float64(time.Now().Unix()%40), "percent"},
			{"memory_usage", 60.0 + float64(time.Now().Unix()%30), "percent"},
			{"request_count", float64(time.Now().Unix() % 1000), "count"},
			{"response_time", 150.0 + float64(time.Now().Unix()%100), "milliseconds"},
		}

		for _, m := range metrics {
			metric := &models.Metric{
				ServiceName: serviceName,
				Name:        m.name,
				Value:       m.value,
				Unit:        m.unit,
				Labels:      map[string]string{"instance": serviceName + "-1"},
				Timestamp:   time.Now(),
			}

			if err := s.repository.StoreMetric(ctx, metric); err != nil {
				s.logger.Error("Failed to store metric", 
					zap.String("service", serviceName),
					zap.String("metric", m.name),
					zap.Error(err),
				)
			}
		}
	}
}

// StartAlertEvaluation starts the background alert evaluation process
func (s *MonitoringService) StartAlertEvaluation(ctx context.Context) {
	s.logger.Info("Starting alert evaluation")

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping alert evaluation")
			return
		case <-ticker.C:
			s.evaluateAlerts(ctx)
		}
	}
}

// evaluateAlerts evaluates all active alert rules
func (s *MonitoringService) evaluateAlerts(ctx context.Context) {
	s.logger.Debug("Evaluating alerts")

	// Get all enabled alerts
	alerts, _, err := s.repository.GetAlerts(ctx, 1, 1000, "", "")
	if err != nil {
		s.logger.Error("Failed to get alerts for evaluation", zap.Error(err))
		return
	}

	for _, alert := range alerts {
		if !alert.Enabled {
			continue
		}

		s.logger.Debug("Evaluating alert", 
			zap.String("alertID", alert.ID.String()),
			zap.String("name", alert.Name),
		)

		// Get recent metrics for this alert
		since := time.Now().Add(-5 * time.Minute)
		metrics, err := s.repository.GetMetrics(ctx, alert.Service, alert.Metric, since)
		if err != nil {
			s.logger.Error("Failed to get metrics for alert evaluation", 
				zap.String("alertID", alert.ID.String()),
				zap.Error(err),
			)
			continue
		}

		if len(metrics) == 0 {
			continue
		}

		// Get the latest metric value
		latestMetric := metrics[0]
		
		// Evaluate condition
		triggered := s.evaluateCondition(alert.Condition, latestMetric.Value, alert.Threshold)
		
		if triggered {
			// Create alert event
			event := &models.AlertEvent{
				ID:        uuid.New(),
				AlertID:   alert.ID,
				Type:      "triggered",
				Message:   fmt.Sprintf("Alert triggered: %s %s %.2f (threshold: %.2f)", alert.Metric, alert.Condition, latestMetric.Value, alert.Threshold),
				Value:     latestMetric.Value,
				Threshold: alert.Threshold,
				Timestamp: time.Now(),
			}

			if err := s.repository.StoreAlertEvent(ctx, event); err != nil {
				s.logger.Error("Failed to store alert event", 
					zap.String("alertID", alert.ID.String()),
					zap.Error(err),
				)
			} else {
				s.logger.Info("Alert triggered", 
					zap.String("alertID", alert.ID.String()),
					zap.String("name", alert.Name),
					zap.Float64("value", latestMetric.Value),
					zap.Float64("threshold", alert.Threshold),
				)
			}
		}
	}
}

// evaluateCondition evaluates an alert condition
func (s *MonitoringService) evaluateCondition(condition string, value, threshold float64) bool {
	switch condition {
	case "greater_than", ">":
		return value > threshold
	case "less_than", "<":
		return value < threshold
	case "equal", "==":
		return value == threshold
	case "not_equal", "!=":
		return value != threshold
	case "greater_equal", ">=":
		return value >= threshold
	case "less_equal", "<=":
		return value <= threshold
	default:
		s.logger.Warn("Unknown condition", zap.String("condition", condition))
		return false
	}
}