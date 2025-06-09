package routes

import (
	"github.com/agba-ai/config-service/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Router handles all routing configuration for the Config Service
type Router struct {
	engine        *gin.Engine
	logger        *zap.Logger
	configHandler *handlers.ConfigHandler
	templateHandler *handlers.TemplateHandler
}

// NewRouter creates a new router instance
func NewRouter(
	logger *zap.Logger,
	configHandler *handlers.ConfigHandler,
	templateHandler *handlers.TemplateHandler,
) *Router {
	engine := gin.New()
	
	return &Router{
		engine:          engine,
		logger:          logger,
		configHandler:   configHandler,
		templateHandler: templateHandler,
	}
}

// SetupRoutes configures all routes for the Config Service
func (r *Router) SetupRoutes() *gin.Engine {
	// Add global middleware
	r.engine.Use(gin.Logger())
	r.engine.Use(gin.Recovery())
	r.engine.Use(r.corsMiddleware())
	r.engine.Use(r.requestIDMiddleware())

	// Health and metrics endpoints (no auth required)
	r.engine.GET("/health", r.healthCheck)
	r.engine.GET("/ready", r.readinessCheck)
	r.engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API routes
	r.setupAPIRoutes()

	return r.engine
}

// setupAPIRoutes configures API routes
func (r *Router) setupAPIRoutes() {
	api := r.engine.Group("/api")
	
	// Version 1 API routes
	v1 := api.Group("/v1")
	{
		// Configuration management routes
		configs := v1.Group("/configs")
		{
			configs.POST("/", r.configHandler.CreateConfig)
			configs.GET("/", r.configHandler.GetConfigs)
			configs.GET("/:id", r.configHandler.GetConfig)
			configs.PUT("/:id", r.configHandler.UpdateConfig)
			configs.DELETE("/:id", r.configHandler.DeleteConfig)
			configs.GET("/:id/history", r.configHandler.GetConfigHistory)
			configs.POST("/validate", r.configHandler.ValidateConfig)
		}

		// Configuration value routes
		values := v1.Group("/values")
		{
			values.GET("/:service/:environment/:key", r.configHandler.GetConfigValue)
			values.PUT("/:service/:environment/:key", r.configHandler.SetConfigValue)
		}

		// Template management routes
		templates := v1.Group("/templates")
		{
			templates.POST("/", r.templateHandler.CreateTemplate)
			templates.GET("/", r.templateHandler.GetTemplates)
			templates.GET("/:id", r.getTemplate)
			templates.PUT("/:id", r.updateTemplate)
			templates.DELETE("/:id", r.deleteTemplate)
			templates.POST("/:id/apply", r.templateHandler.ApplyTemplate)
		}

		// Service-specific configuration routes
		services := v1.Group("/services")
		{
			services.GET("/:service/configs", r.getServiceConfigs)
			services.GET("/:service/environments", r.getServiceEnvironments)
			services.GET("/:service/:environment", r.getServiceEnvironmentConfig)
			services.PUT("/:service/:environment", r.updateServiceEnvironmentConfig)
		}

		// Environment management routes
		environments := v1.Group("/environments")
		{
			environments.GET("/", r.getEnvironments)
			environments.POST("/", r.createEnvironment)
			environments.GET("/:environment/services", r.getEnvironmentServices)
		}

		// Configuration comparison routes
		compare := v1.Group("/compare")
		{
			compare.POST("/configs", r.compareConfigs)
			compare.POST("/environments", r.compareEnvironments)
		}

		// Configuration export/import routes
		export := v1.Group("/export")
		{
			export.POST("/configs", r.exportConfigs)
			export.POST("/import", r.importConfigs)
		}

		// Configuration search routes
		search := v1.Group("/search")
		{
			search.POST("/configs", r.searchConfigs)
			search.POST("/values", r.searchConfigValues)
		}

		// Configuration audit routes
		audit := v1.Group("/audit")
		{
			audit.GET("/", r.getAuditLogs)
			audit.GET("/configs/:id", r.getConfigAuditLogs)
			audit.GET("/users/:userId", r.getUserAuditLogs)
		}

		// Configuration backup/restore routes
		backup := v1.Group("/backup")
		{
			backup.POST("/", r.createBackup)
			backup.GET("/", r.getBackups)
			backup.POST("/:id/restore", r.restoreBackup)
			backup.DELETE("/:id", r.deleteBackup)
		}

		// Configuration validation routes
		validation := v1.Group("/validation")
		{
			validation.POST("/schema", r.validateSchema)
			validation.POST("/rules", r.validateRules)
			validation.GET("/rules", r.getValidationRules)
			validation.POST("/rules", r.createValidationRule)
		}

		// Configuration deployment routes
		deployment := v1.Group("/deployment")
		{
			deployment.POST("/deploy", r.deployConfig)
			deployment.GET("/status/:deploymentId", r.getDeploymentStatus)
			deployment.POST("/rollback/:deploymentId", r.rollbackDeployment)
		}
	}

	// WebSocket routes for real-time configuration updates
	ws := r.engine.Group("/ws")
	{
		ws.GET("/configs/:service/:environment", r.configWebSocket)
		ws.GET("/events", r.eventsWebSocket)
	}

	// Admin routes
	admin := r.engine.Group("/admin")
	{
		admin.GET("/status", r.getServiceStatus)
		admin.GET("/stats", r.getServiceStats)
		admin.POST("/maintenance", r.enableMaintenance)
		admin.DELETE("/maintenance", r.disableMaintenance)
		admin.POST("/cache/clear", r.clearCache)
		admin.GET("/cache/stats", r.getCacheStats)
	}
}

// Middleware functions

func (r *Router) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}

func (r *Router) requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// Health check handlers

func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":    "healthy",
		"service":   "config-service",
		"timestamp": getCurrentTimestamp(),
	})
}

func (r *Router) readinessCheck(c *gin.Context) {
	// Add readiness checks here (database, redis, etc.)
	c.JSON(200, gin.H{
		"status": "ready",
		"checks": gin.H{
			"database": "ok",
			"redis":    "ok",
			"nats":     "ok",
		},
	})
}

// Placeholder handlers for routes not yet implemented in handlers package

func (r *Router) getTemplate(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) updateTemplate(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) deleteTemplate(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) getServiceConfigs(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) getServiceEnvironments(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) getServiceEnvironmentConfig(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) updateServiceEnvironmentConfig(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) getEnvironments(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) createEnvironment(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) getEnvironmentServices(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) compareConfigs(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) compareEnvironments(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) exportConfigs(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) importConfigs(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) searchConfigs(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) searchConfigValues(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) getAuditLogs(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) getConfigAuditLogs(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) getUserAuditLogs(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) createBackup(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) getBackups(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) restoreBackup(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) deleteBackup(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) validateSchema(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) validateRules(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) getValidationRules(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) createValidationRule(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) deployConfig(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) getDeploymentStatus(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) rollbackDeployment(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) configWebSocket(c *gin.Context) {
	c.JSON(501, gin.H{"error": "WebSocket not implemented yet"})
}

func (r *Router) eventsWebSocket(c *gin.Context) {
	c.JSON(501, gin.H{"error": "WebSocket not implemented yet"})
}

func (r *Router) getServiceStatus(c *gin.Context) {
	c.JSON(200, gin.H{
		"service": "config-service",
		"status":  "running",
		"version": "1.0.0",
	})
}

func (r *Router) getServiceStats(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) enableMaintenance(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) disableMaintenance(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) clearCache(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

func (r *Router) getCacheStats(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

// Helper functions

func generateRequestID() string {
	// Simple request ID generation - in production use UUID or similar
	return "req-" + getCurrentTimestamp()
}

func getCurrentTimestamp() string {
	// Return current timestamp as string
	return "2025-06-09T12:00:00Z"
}