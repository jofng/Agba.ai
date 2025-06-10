package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agba-ai/monitoring-service/internal/config"
	"github.com/agba-ai/monitoring-service/internal/handlers"
	"github.com/agba-ai/monitoring-service/internal/repository"
	"github.com/agba-ai/monitoring-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	_ "github.com/lib/pq" // PostgreSQL driver
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
	db, err := sql.Open("postgres", cfg.Database.DSN)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}

	// Initialize repository
	repo := repository.NewRepository(db, logger)

	// Initialize service
	monitoringService := service.NewMonitoringService(repo, logger)

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
		healthy, checks := monitoringService.HealthCheck(context.Background())
		status := http.StatusOK
		if !healthy {
			status = http.StatusServiceUnavailable
		}
		
		c.JSON(status, gin.H{
			"status":    map[bool]string{true: "healthy", false: "unhealthy"}[healthy],
			"timestamp": time.Now().UTC(),
			"version":   cfg.Version,
			"checks":    checks,
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

		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	})

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Initialize handlers
	monitoringHandler := handlers.NewMonitoringHandler(monitoringService, logger)
	alertHandler := handlers.NewAlertHandler(monitoringService, logger)

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Metrics routes
		metrics := v1.Group("/metrics")
		{
			metrics.GET("/", monitoringHandler.GetMetrics)
			metrics.GET("/performance", monitoringHandler.GetPerformanceMetrics)
		}

		// Health monitoring routes
		health := v1.Group("/health")
		{
			health.GET("/", monitoringHandler.GetSystemStatus)
			health.GET("/services", monitoringHandler.GetAllServicesHealth)
			health.GET("/services/:service", monitoringHandler.GetServiceHealth)
		}

		// Alerting routes
		alerts := v1.Group("/alerts")
		{
			alerts.GET("/", monitoringHandler.GetAlerts)
			alerts.POST("/", monitoringHandler.CreateAlert)
			alerts.PUT("/:id", monitoringHandler.UpdateAlert)
			alerts.DELETE("/:id", monitoringHandler.DeleteAlert)
			alerts.GET("/:id/history", monitoringHandler.GetAlertHistory)
			alerts.GET("/active", alertHandler.GetActiveAlerts)
			alerts.POST("/:id/acknowledge", alertHandler.AcknowledgeAlert)
			alerts.POST("/:id/resolve", alertHandler.ResolveAlert)
			alerts.POST("/:id/trigger", alertHandler.TriggerAlert)
		}

		// Dashboard routes
		dashboards := v1.Group("/dashboards")
		{
			dashboards.GET("/:type", monitoringHandler.GetDashboard)
		}

		// System status routes
		status := v1.Group("/status")
		{
			status.GET("/", monitoringHandler.GetSystemStatus)
		}

		// Logs routes
		logs := v1.Group("/logs")
		{
			logs.GET("/", monitoringHandler.GetLogs)
		}
	}

	// Start background services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start metrics collection
	go monitoringService.StartMetricsCollection(ctx)

	// Start alert evaluation
	go monitoringService.StartAlertEvaluation(ctx)

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
		logger.Info("Starting Monitoring Service server",
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

	logger.Info("Shutting down Monitoring Service server...")

	// Cancel background services
	cancel()

	// Create a deadline for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown error", zap.Error(err))
	}

	logger.Info("Monitoring Service server stopped")
}