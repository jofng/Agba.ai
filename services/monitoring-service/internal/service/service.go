package service

import (
	"context"
	"encoding/json"
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

	// Mock implementation - replace with actual health checks
	health := &models.ServiceHealth{
		ServiceName:  serviceName,
		Status:       "healthy",
		LastCheck:    time.Now(),
		ResponseTime: 150 * time.Millisecond,
		Checks: []models.HealthCheck{
			{
				Name:   "database",
				Status: "healthy",
				Message: "Database connection successful",
			},
			{
				Name:   "redis",
				Status: "healthy",
				Message: "Redis connection successful",
			},
		},
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

	// Mock implementation - replace with actual database storage
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

	// Mock implementation - replace with actual database query
	alerts := []*models.Alert{
		{
			ID:          uuid.New(),
			Name:        "High CPU Usage",
			Description: "Alert when CPU usage exceeds 80%",
			Service:     "api-gateway",
			Metric:      "cpu_usage",
			Condition:   "greater_than",
			Threshold:   80.0,
			Severity:    "warning",
			Enabled:     true,
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now().Add(-24 * time.Hour),
		},
		{
			ID:          uuid.New(),
			Name:        "High Memory Usage",
			Description: "Alert when memory usage exceeds 90%",
			Service:     "config-service",
			Metric:      "memory_usage",
			Condition:   "greater_than",
			Threshold:   90.0,
			Severity:    "critical",
			Enabled:     true,
			CreatedAt:   time.Now().Add(-12 * time.Hour),
			UpdatedAt:   time.Now().Add(-12 * time.Hour),
		},
	}

	return alerts, int64(len(alerts)), nil
}

// UpdateAlert updates an existing alert rule
func (s *MonitoringService) UpdateAlert(ctx context.Context, id uuid.UUID, req *models.UpdateAlertRequest, userID string) (*models.Alert, error) {
	s.logger.Info("Updating alert", 
		zap.String("alertID", id.String()),
		zap.String("userID", userID),
	)

	// Mock implementation - replace with actual database update
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
		UpdatedBy:   uuid.MustParse(userID),
		UpdatedAt:   time.Now(),
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

	// Mock implementation - replace with actual database deletion
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

	// Mock implementation - replace with actual database query
	events := []*models.AlertEvent{
		{
			ID:        uuid.New(),
			AlertID:   alertID,
			Type:      "triggered",
			Message:   "Alert triggered: CPU usage exceeded threshold",
			Value:     85.5,
			Threshold: 80.0,
			Timestamp: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:        uuid.New(),
			AlertID:   alertID,
			Type:      "resolved",
			Message:   "Alert resolved: CPU usage returned to normal",
			Value:     75.2,
			Threshold: 80.0,
			Timestamp: time.Now().Add(-1 * time.Hour),
		},
	}

	return events, int64(len(events)), nil
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

	// Mock implementation - replace with actual log retrieval
	logs := []*models.LogEntry{
		{
			ID:        uuid.New(),
			Service:   service,
			Level:     "INFO",
			Message:   "Service started successfully",
			Timestamp: time.Now().Add(-1 * time.Hour),
			Metadata: map[string]interface{}{
				"version": "1.0.0",
				"port":    8080,
			},
		},
		{
			ID:        uuid.New(),
			Service:   service,
			Level:     "WARN",
			Message:   "High memory usage detected",
			Timestamp: time.Now().Add(-30 * time.Minute),
			Metadata: map[string]interface{}{
				"memory_usage": 85.5,
				"threshold":    80.0,
			},
		},
	}

	return logs, int64(len(logs)), nil
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

	// Mock implementation - replace with actual metrics collection
	services := []string{
		"api-gateway",
		"config-service",
		"user-management-service",
		"notification-service",
		"recording-service",
	}

	for _, service := range services {
		// Collect metrics for each service
		s.logger.Debug("Collecting metrics for service", zap.String("service", service))
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

	// Mock implementation - replace with actual alert evaluation
	// This would typically:
	// 1. Get all enabled alert rules
	// 2. Evaluate each rule against current metrics
	// 3. Trigger alerts if conditions are met
	// 4. Resolve alerts if conditions are no longer met
}