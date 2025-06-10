package routes

import (
	"api-gateway/internal/auth"
	"api-gateway/internal/config"
	"api-gateway/internal/handlers"
	"api-gateway/internal/middleware"
	"api-gateway/internal/proxy"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Router handles all routing configuration
type Router struct {
	engine      *gin.Engine
	config      *config.Config
	logger      *zap.Logger
	authService auth.Service
	proxy       proxy.Service
}

// NewRouter creates a new router instance
func NewRouter(cfg *config.Config, logger *zap.Logger, authService auth.Service, serviceProxy proxy.Service) *Router {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	
	return &Router{
		engine:      engine,
		config:      cfg,
		logger:      logger,
		authService: authService,
		proxy:       serviceProxy,
	}
}

// SetupRoutes configures all routes
func (r *Router) SetupRoutes() *gin.Engine {
	// Add global middleware
	r.engine.Use(gin.Logger())
	r.engine.Use(gin.Recovery())
	r.engine.Use(middleware.CORS(r.config.CORS))
	r.engine.Use(middleware.RequestID())
	// r.engine.Use(middleware.Logging(r.logger)) // TODO: Implement Logging middleware
	// r.engine.Use(middleware.RateLimit(r.config.RateLimit)) // TODO: Implement RateLimit middleware

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(r.logger, r.config)
	_ = handlers.NewMetricsHandler(r.logger)
	adminHandler := handlers.NewAdminHandler(r.logger, r.config)

	// Health and metrics endpoints (no auth required)
	r.engine.GET("/health", healthHandler.Health)
	r.engine.GET("/ready", healthHandler.Ready)
	r.engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Admin endpoints (require admin role)
	admin := r.engine.Group("/admin")
	admin.Use(middleware.Authentication(r.authService))
	// admin.Use(middleware.RequireRole("admin")) // TODO: Implement RequireRole middleware
	{
		admin.GET("/status", adminHandler.GetStatus)
		admin.GET("/routes", adminHandler.GetRoutes)
		admin.GET("/services", r.getServicesStatus)
	}

	// API routes
	r.setupAPIRoutes()

	return r.engine
}

// setupAPIRoutes configures API routes with service proxying
func (r *Router) setupAPIRoutes() {
	api := r.engine.Group("/api")
	
	// Apply authentication middleware to all API routes
	api.Use(middleware.Authentication(r.authService))
	
	// Version 1 API routes
	v1 := api.Group("/v1")
	{
		// User Management Service routes
		userMgmt := v1.Group("/users")
		{
			userMgmt.Any("/*path", func(c *gin.Context) {
				r.proxy.ProxyRequest(c, "user-service")
			})
		}

		// Configuration Service routes
		config := v1.Group("/config")
		{
			config.Any("/*path", func(c *gin.Context) {
				r.proxy.ProxyRequest(c, "config-service")
			})
		}

		// Notification Service routes (TODO: implement notification service)
		// notifications := v1.Group("/notifications")
		// {
		//     notifications.Any("/*path", func(c *gin.Context) {
		//         r.proxy.ProxyRequest(c, "notification-service")
		//     })
		// }

		// Recording Service routes
		recordings := v1.Group("/recordings")
		{
			recordings.Any("/*path", func(c *gin.Context) {
				r.proxy.ProxyRequest(c, "recording-service")
			})
		}

		// Monitoring Service routes (TODO: implement monitoring service)
		// monitoring := v1.Group("/monitoring")
		// monitoring.Use(middleware.RequireRole("admin"))
		// {
		//     monitoring.Any("/*path", func(c *gin.Context) {
		//         r.proxy.ProxyRequest(c, "monitoring-service")
		//     })
		// }

		// Telephony Service routes (for Phase 2)
		// telephony := v1.Group("/telephony")
		// {
		//     telephony.Any("/*path", func(c *gin.Context) {
		//         r.proxy.ProxyRequest(c, "telephony-gateway")
		//     })
		// }

		// WebRTC Service routes (for Phase 2)
		webrtc := v1.Group("/webrtc")
		webrtc.Use(middleware.WebRTCAuth(r.authService))
		{
			webrtc.Any("/*path", func(c *gin.Context) {
				r.proxy.ProxyRequest(c, "webrtc-service")
			})
		}

		// AI Pipeline routes (for Phase 3) - TODO: implement
		// ai := v1.Group("/ai")
		// {
		//     ai.Any("/*path", func(c *gin.Context) {
		//         r.proxy.ProxyRequest(c, "ai-pipeline")
		//     })
		// }

		// Analytics Service routes (for Phase 4)
		analytics := v1.Group("/analytics")
		// analytics.Use(middleware.RequirePermission("analytics:read")) // TODO: implement RequirePermission
		{
			analytics.Any("/*path", func(c *gin.Context) {
				r.proxy.ProxyRequest(c, "analytics-service")
			})
		}

		// Biometrics Service routes (for Phase 4)
		biometrics := v1.Group("/biometrics")
		// biometrics.Use(middleware.RequirePermission("biometrics:access")) // TODO: implement RequirePermission
		{
			biometrics.Any("/*path", func(c *gin.Context) {
				r.proxy.ProxyRequest(c, "biometrics-service")
			})
		}

		// Portal Service routes (for Phase 5) - TODO: implement
		// portal := v1.Group("/portal")
		// {
		//     portal.Any("/*path", func(c *gin.Context) {
		//         r.proxy.ProxyRequest(c, "portal-service")
		//     })
		// }

		// Workflow Engine routes (for Phase 5) - TODO: implement
		// workflow := v1.Group("/workflow")
		// workflow.Use(middleware.RequirePermission("workflow:manage"))
		// {
		//     workflow.Any("/*path", func(c *gin.Context) {
		//         r.proxy.ProxyRequest(c, "workflow-engine")
		//     })
		// }
	}

	// Public API routes (no authentication required)
	public := api.Group("/public")
	{
		// Public health checks for services
		public.GET("/health/:service", r.getServiceHealth)
		
		// Public documentation endpoints
		public.GET("/docs", r.getAPIDocumentation)
		public.GET("/openapi.json", r.getOpenAPISpec)
	}

	// WebSocket routes - TODO: implement WebSocket support
	// ws := r.engine.Group("/ws")
	// ws.Use(middleware.Authentication(r.authService)) // Use regular auth for now
	// {
	//     // Real-time recording WebSocket
	//     ws.GET("/recording/:sessionId", func(c *gin.Context) {
	//         r.proxy.ProxyRequest(c, "recording-service")
	//     })
	//     
	//     // Real-time notifications WebSocket
	//     ws.GET("/notifications", func(c *gin.Context) {
	//         r.proxy.ProxyRequest(c, "notification-service")
	//     })
	//     
	//     // Real-time call events WebSocket (for Phase 2)
	//     ws.GET("/calls/:callId", func(c *gin.Context) {
	//         r.proxy.ProxyRequest(c, "telephony-gateway")
	//     })
	// }
}

// getServicesStatus returns the status of all backend services
func (r *Router) getServicesStatus(c *gin.Context) {
	status := r.proxy.GetServiceStatus()
	c.JSON(200, gin.H{
		"services": status,
		"gateway":  "healthy",
	})
}

// getServiceHealth returns the health of a specific service
func (r *Router) getServiceHealth(c *gin.Context) {
	serviceName := c.Param("service")
	err := r.proxy.HealthCheck(serviceName)
	
	if err != nil {
		c.JSON(503, gin.H{
			"service": serviceName,
			"healthy": false,
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(200, gin.H{
		"service": serviceName,
		"healthy": true,
	})
}

// getAPIDocumentation returns API documentation
func (r *Router) getAPIDocumentation(c *gin.Context) {
	c.JSON(200, gin.H{
		"title":       "Agba Voice AI Platform API",
		"version":     r.config.Version,
		"description": "Comprehensive Voice AI Platform API Gateway",
		"endpoints": gin.H{
			"health":        "/health",
			"metrics":       "/metrics",
			"documentation": "/api/public/docs",
			"openapi":       "/api/public/openapi.json",
		},
		"services": []string{
			"user-management-service",
			"config-service",
			"notification-service",
			"recording-service",
			"monitoring-service",
			"telephony-gateway",
			"webrtc-service",
			"ai-pipeline",
			"analytics-service",
			"biometrics-service",
			"portal-service",
			"workflow-engine",
		},
	})
}

// getOpenAPISpec returns OpenAPI specification
func (r *Router) getOpenAPISpec(c *gin.Context) {
	spec := gin.H{
		"openapi": "3.0.0",
		"info": gin.H{
			"title":       "Agba Voice AI Platform API",
			"version":     r.config.Version,
			"description": "Comprehensive Voice AI Platform API Gateway",
		},
		"servers": []gin.H{
			{"url": "/api/v1", "description": "API Gateway"},
		},
		"paths": gin.H{
			"/health": gin.H{
				"get": gin.H{
					"summary":     "Health Check",
					"description": "Returns the health status of the API Gateway",
					"responses": gin.H{
						"200": gin.H{"description": "Healthy"},
					},
				},
			},
		},
		"components": gin.H{
			"securitySchemes": gin.H{
				"bearerAuth": gin.H{
					"type":   "http",
					"scheme": "bearer",
					"bearerFormat": "JWT",
				},
				"apiKeyAuth": gin.H{
					"type": "apiKey",
					"in":   "header",
					"name": "X-API-Key",
				},
			},
		},
		"security": []gin.H{
			{"bearerAuth": []string{}},
			{"apiKeyAuth": []string{}},
		},
	}
	
	c.JSON(200, spec)
}