package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agba-ai/config-service/internal/config"
	"github.com/agba-ai/config-service/internal/handlers"
	"github.com/agba-ai/config-service/internal/repository"
	"github.com/agba-ai/config-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize database connection
	db, err := repository.NewDatabase(cfg.Database, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize Redis connection
	redisClient, err := repository.NewRedis(cfg.Redis, logger)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize NATS connection
	natsConn, err := repository.NewNATS(cfg.NATS, logger)
	if err != nil {
		logger.Fatal("Failed to connect to NATS", zap.Error(err))
	}
	defer natsConn.Close()

	// Initialize repositories
	configRepo := repository.NewConfigRepository(db, redisClient, logger)
	auditRepo := repository.NewAuditRepository(db, logger)

	// Initialize services
	configService := service.NewConfigService(configRepo, auditRepo, natsConn, logger)
	validationService := service.NewValidationService(logger)

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   cfg.Version,
		})
	})

	router.GET("/ready", func(c *gin.Context) {
		// Check database connectivity
		if err := db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"error":  "database connection failed",
			})
			return
		}

		// Check Redis connectivity
		if err := redisClient.Ping(context.Background()).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"error":  "redis connection failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	})

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Initialize handlers
	configHandler := handlers.NewConfigHandler(configService, validationService, logger)

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Agent configuration routes
		agents := v1.Group("/agents")
		{
			agents.GET("/:id/config", configHandler.GetAgentConfig)
			agents.PUT("/:id/config", configHandler.UpdateAgentConfig)
			agents.GET("/:id/config/versions", configHandler.GetAgentConfigVersions)
			agents.GET("/:id/config/versions/:version", configHandler.GetAgentConfigVersion)
			agents.POST("/:id/config/rollback/:version", configHandler.RollbackAgentConfig)
			agents.POST("/:id/config/validate", configHandler.ValidateAgentConfig)
		}

		// System configuration routes
		system := v1.Group("/system")
		{
			system.GET("/config", configHandler.GetSystemConfig)
			system.PUT("/config", configHandler.UpdateSystemConfig)
			system.GET("/config/versions", configHandler.GetSystemConfigVersions)
			system.GET("/config/versions/:version", configHandler.GetSystemConfigVersion)
			system.POST("/config/rollback/:version", configHandler.RollbackSystemConfig)
		}

		// Organization configuration routes
		orgs := v1.Group("/organizations")
		{
			orgs.GET("/:id/config", configHandler.GetOrganizationConfig)
			orgs.PUT("/:id/config", configHandler.UpdateOrganizationConfig)
			orgs.GET("/:id/config/versions", configHandler.GetOrganizationConfigVersions)
		}

		// Configuration templates
		templates := v1.Group("/templates")
		{
			templates.GET("/", configHandler.ListConfigTemplates)
			templates.GET("/:id", configHandler.GetConfigTemplate)
			templates.POST("/", configHandler.CreateConfigTemplate)
			templates.PUT("/:id", configHandler.UpdateConfigTemplate)
			templates.DELETE("/:id", configHandler.DeleteConfigTemplate)
		}

		// Configuration schemas
		schemas := v1.Group("/schemas")
		{
			schemas.GET("/", configHandler.ListConfigSchemas)
			schemas.GET("/:type", configHandler.GetConfigSchema)
			schemas.PUT("/:type", configHandler.UpdateConfigSchema)
		}

		// Configuration audit
		audit := v1.Group("/audit")
		{
			audit.GET("/", configHandler.GetAuditLogs)
			audit.GET("/:id", configHandler.GetAuditLog)
		}
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting Configuration Service server",
			zap.String("address", server.Addr),
			zap.String("environment", cfg.Environment),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Configuration Service server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown error", zap.Error(err))
	}

	logger.Info("Configuration Service server stopped")
}