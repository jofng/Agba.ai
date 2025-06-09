package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/agba-ai/monitoring-service/internal/models"
	"github.com/agba-ai/monitoring-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// MonitoringHandler handles monitoring-related HTTP requests
type MonitoringHandler struct {
	monitoringService *service.MonitoringService
	logger            *zap.Logger
}

// NewMonitoringHandler creates a new monitoring handler
func NewMonitoringHandler(monitoringService *service.MonitoringService, logger *zap.Logger) *MonitoringHandler {
	return &MonitoringHandler{
		monitoringService: monitoringService,
		logger:            logger,
	}
}

// GetMetrics retrieves system metrics
func (h *MonitoringHandler) GetMetrics(c *gin.Context) {
	service := c.Query("service")
	timeRange := c.DefaultQuery("range", "1h")
	
	metrics, err := h.monitoringService.GetMetrics(c.Request.Context(), service, timeRange)
	if err != nil {
		h.logger.Error("Failed to get metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve metrics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetServiceHealth retrieves service health status
func (h *MonitoringHandler) GetServiceHealth(c *gin.Context) {
	serviceName := c.Param("service")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Service name required"})
		return
	}

	health, err := h.monitoringService.GetServiceHealth(c.Request.Context(), serviceName)
	if err != nil {
		h.logger.Error("Failed to get service health", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve service health"})
		return
	}

	c.JSON(http.StatusOK, health)
}

// GetAllServicesHealth retrieves health status for all services
func (h *MonitoringHandler) GetAllServicesHealth(c *gin.Context) {
	healthStatus, err := h.monitoringService.GetAllServicesHealth(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get all services health", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve services health"})
		return
	}

	c.JSON(http.StatusOK, healthStatus)
}

// CreateAlert creates a new alert rule
func (h *MonitoringHandler) CreateAlert(c *gin.Context) {
	var req models.CreateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	alert, err := h.monitoringService.CreateAlert(c.Request.Context(), &req, userID)
	if err != nil {
		h.logger.Error("Failed to create alert", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create alert"})
		return
	}

	c.JSON(http.StatusCreated, alert)
}

// GetAlerts retrieves alert rules
func (h *MonitoringHandler) GetAlerts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	service := c.Query("service")
	status := c.Query("status")

	alerts, total, err := h.monitoringService.GetAlerts(c.Request.Context(), page, limit, service, status)
	if err != nil {
		h.logger.Error("Failed to get alerts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}

// UpdateAlert updates an existing alert rule
func (h *MonitoringHandler) UpdateAlert(c *gin.Context) {
	alertID := c.Param("id")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Alert ID required"})
		return
	}

	id, err := uuid.Parse(alertID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	var req models.UpdateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	alert, err := h.monitoringService.UpdateAlert(c.Request.Context(), id, &req, userID)
	if err != nil {
		h.logger.Error("Failed to update alert", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update alert"})
		return
	}

	c.JSON(http.StatusOK, alert)
}

// DeleteAlert deletes an alert rule
func (h *MonitoringHandler) DeleteAlert(c *gin.Context) {
	alertID := c.Param("id")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Alert ID required"})
		return
	}

	id, err := uuid.Parse(alertID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	err = h.monitoringService.DeleteAlert(c.Request.Context(), id, userID)
	if err != nil {
		h.logger.Error("Failed to delete alert", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete alert"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetAlertHistory retrieves alert history
func (h *MonitoringHandler) GetAlertHistory(c *gin.Context) {
	alertID := c.Param("id")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Alert ID required"})
		return
	}

	id, err := uuid.Parse(alertID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	history, total, err := h.monitoringService.GetAlertHistory(c.Request.Context(), id, page, limit)
	if err != nil {
		h.logger.Error("Failed to get alert history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve alert history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// GetDashboard retrieves dashboard data
func (h *MonitoringHandler) GetDashboard(c *gin.Context) {
	dashboardType := c.DefaultQuery("type", "overview")
	timeRange := c.DefaultQuery("range", "1h")

	dashboard, err := h.monitoringService.GetDashboard(c.Request.Context(), dashboardType, timeRange)
	if err != nil {
		h.logger.Error("Failed to get dashboard", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve dashboard"})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

// GetLogs retrieves system logs
func (h *MonitoringHandler) GetLogs(c *gin.Context) {
	service := c.Query("service")
	level := c.Query("level")
	since := c.Query("since")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	logs, total, err := h.monitoringService.GetLogs(c.Request.Context(), service, level, since, page, limit)
	if err != nil {
		h.logger.Error("Failed to get logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetPerformanceMetrics retrieves performance metrics
func (h *MonitoringHandler) GetPerformanceMetrics(c *gin.Context) {
	service := c.Query("service")
	metric := c.Query("metric")
	timeRange := c.DefaultQuery("range", "1h")

	metrics, err := h.monitoringService.GetPerformanceMetrics(c.Request.Context(), service, metric, timeRange)
	if err != nil {
		h.logger.Error("Failed to get performance metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve performance metrics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetSystemStatus retrieves overall system status
func (h *MonitoringHandler) GetSystemStatus(c *gin.Context) {
	status, err := h.monitoringService.GetSystemStatus(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get system status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve system status"})
		return
	}

	c.JSON(http.StatusOK, status)
}

// AlertHandler handles alert-specific operations
type AlertHandler struct {
	monitoringService *service.MonitoringService
	logger            *zap.Logger
}

// NewAlertHandler creates a new alert handler
func NewAlertHandler(monitoringService *service.MonitoringService, logger *zap.Logger) *AlertHandler {
	return &AlertHandler{
		monitoringService: monitoringService,
		logger:            logger,
	}
}

// TriggerAlert manually triggers an alert for testing
func (h *AlertHandler) TriggerAlert(c *gin.Context) {
	alertID := c.Param("id")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Alert ID required"})
		return
	}

	id, err := uuid.Parse(alertID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	var req models.TriggerAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err = h.monitoringService.TriggerAlert(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.Error("Failed to trigger alert", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to trigger alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert triggered successfully"})
}

// AcknowledgeAlert acknowledges an active alert
func (h *AlertHandler) AcknowledgeAlert(c *gin.Context) {
	alertID := c.Param("id")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Alert ID required"})
		return
	}

	id, err := uuid.Parse(alertID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	var req models.AcknowledgeAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err = h.monitoringService.AcknowledgeAlert(c.Request.Context(), id, userID, &req)
	if err != nil {
		h.logger.Error("Failed to acknowledge alert", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to acknowledge alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert acknowledged successfully"})
}

// ResolveAlert resolves an active alert
func (h *AlertHandler) ResolveAlert(c *gin.Context) {
	alertID := c.Param("id")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Alert ID required"})
		return
	}

	id, err := uuid.Parse(alertID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	var req models.ResolveAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err = h.monitoringService.ResolveAlert(c.Request.Context(), id, userID, &req)
	if err != nil {
		h.logger.Error("Failed to resolve alert", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert resolved successfully"})
}

// GetActiveAlerts retrieves all active alerts
func (h *AlertHandler) GetActiveAlerts(c *gin.Context) {
	service := c.Query("service")
	severity := c.Query("severity")

	alerts, err := h.monitoringService.GetActiveAlerts(c.Request.Context(), service, severity)
	if err != nil {
		h.logger.Error("Failed to get active alerts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve active alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// HealthHandler handles health check operations
type HealthHandler struct {
	monitoringService *service.MonitoringService
	logger            *zap.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(monitoringService *service.MonitoringService, logger *zap.Logger) *HealthHandler {
	return &HealthHandler{
		monitoringService: monitoringService,
		logger:            logger,
	}
}

// Health returns the health status of the monitoring service
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "monitoring-service",
	})
}

// Ready returns the readiness status of the monitoring service
func (h *HealthHandler) Ready(c *gin.Context) {
	ready, checks := h.monitoringService.HealthCheck(c.Request.Context())
	
	status := http.StatusOK
	if !ready {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"status": map[string]interface{}{
			"ready": ready,
		},
		"checks": checks,
	})
}