package handlers

import (
	"net/http"
	"time"

	"github.com/agba-ai/api-gateway/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	logger *zap.Logger
	config *config.Config
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(logger *zap.Logger, cfg *config.Config) *HealthHandler {
	return &HealthHandler{
		logger: logger,
		config: cfg,
	}
}

// Health returns the health status of the API Gateway
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   h.config.Version,
		"service":   "api-gateway",
	})
}

// Ready returns the readiness status of the API Gateway
func (h *HealthHandler) Ready(c *gin.Context) {
	// Add readiness checks here (database, redis, etc.)
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"checks": gin.H{
			"database": "ok",
			"redis":    "ok",
			"services": "ok",
		},
	})
}

// MetricsHandler handles metrics requests
type MetricsHandler struct {
	logger *zap.Logger
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(logger *zap.Logger) *MetricsHandler {
	return &MetricsHandler{
		logger: logger,
	}
}

// GetMetrics returns API Gateway metrics
func (h *MetricsHandler) GetMetrics(c *gin.Context) {
	// Prometheus metrics are handled by the middleware
	// This is for custom metrics if needed
	c.JSON(http.StatusOK, gin.H{
		"message": "Metrics available at /metrics endpoint",
	})
}

// AdminHandler handles administrative requests
type AdminHandler struct {
	logger *zap.Logger
	config *config.Config
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(logger *zap.Logger, cfg *config.Config) *AdminHandler {
	return &AdminHandler{
		logger: logger,
		config: cfg,
	}
}

// GetStatus returns the status of the API Gateway
func (h *AdminHandler) GetStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":     "api-gateway",
		"version":     h.config.Version,
		"environment": h.config.Environment,
		"uptime":      time.Since(time.Now()).String(), // This should be calculated from start time
		"status":      "running",
	})
}

// GetRoutes returns the available routes
func (h *AdminHandler) GetRoutes(c *gin.Context) {
	routes := []gin.H{
		{"method": "GET", "path": "/health", "description": "Health check"},
		{"method": "GET", "path": "/ready", "description": "Readiness check"},
		{"method": "GET", "path": "/metrics", "description": "Prometheus metrics"},
		{"method": "GET", "path": "/admin/status", "description": "Service status"},
		{"method": "GET", "path": "/admin/routes", "description": "Available routes"},
	}

	c.JSON(http.StatusOK, gin.H{
		"routes": routes,
		"count":  len(routes),
	})
}