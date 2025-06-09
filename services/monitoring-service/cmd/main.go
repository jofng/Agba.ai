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

	"github.com/agba-ai/monitoring-service/internal/alerting"
	"github.com/agba-ai/monitoring-service/internal/config"
	"github.com/agba-ai/monitoring-service/internal/handlers"
	"github.com/agba-ai/monitoring-service/internal/metrics"
	"github.com/agba-ai/monitoring-service/internal/repository"
	"github.com/agba-ai/monitoring-service/internal/service"
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

	// Initialize metrics collector
	metricsCollector := metrics.NewCollector(logger)

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

	// Initialize Prometheus client for querying
	promClient, err := repository.NewPrometheusClient(cfg.Prometheus, logger)
	if err != nil {
		logger.Fatal("Failed to initialize Prometheus client", zap.Error(err))
	}

	// Initialize repositories
	metricsRepo := repository.NewMetricsRepository(db, redisClient, logger)
	alertRepo := repository.NewAlertRepository(db, logger)
	serviceRepo := repository.NewServiceRepository(db, redisClient, logger)

	// Initialize alerting manager
	alertManager := alerting.NewManager(cfg.Alerting, natsConn, logger)

	// Initialize services
	monitoringService := service.NewMonitoringService(
		metricsRepo,
		alertRepo,
		serviceRepo,
		promClient,
		alertManager,
		metricsCollector,
		logger,
	)

	healthService := service.NewHealthService(serviceRepo, logger)
	alertingService := service.NewAlertingService(alertRepo, alertManager, logger)

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

		// Check Prometheus connectivity
		if err := promClient.Health(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"error":  "prometheus connection failed",
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
	healthHandler := handlers.NewHealthHandler(healthService, logger)
	alertingHandler := handlers.NewAlertingHandler(alertingService, logger)

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Metrics routes
		metrics := v1.Group("/metrics")
		{
			metrics.GET("/", monitoringHandler.GetMetrics)
			metrics.GET("/query", monitoringHandler.QueryMetrics)
			metrics.GET("/range", monitoringHandler.QueryRangeMetrics)
			metrics.GET("/services", monitoringHandler.GetServiceMetrics)
			metrics.GET("/services/:service", monitoringHandler.GetServiceMetrics)
			metrics.POST("/custom", monitoringHandler.RecordCustomMetric)
		}

		// Health monitoring routes
		health := v1.Group("/health")
		{
			health.GET("/", healthHandler.GetOverallHealth)
			health.GET("/services", healthHandler.GetServicesHealth)
			health.GET("/services/:service", healthHandler.GetServiceHealth)
			health.POST("/services/:service/check", healthHandler.CheckServiceHealth)
			health.GET("/dependencies", healthHandler.GetDependenciesHealth)
		}

		// Alerting routes
		alerts := v1.Group("/alerts")
		{
			alerts.GET("/", alertingHandler.GetAlerts)
			alerts.GET("/:id", alertingHandler.GetAlert)
			alerts.POST("/", alertingHandler.CreateAlert)
			alerts.PUT("/:id", alertingHandler.UpdateAlert)
			alerts.DELETE("/:id", alertingHandler.DeleteAlert)
			alerts.POST("/:id/acknowledge", alertingHandler.AcknowledgeAlert)
			alerts.POST("/:id/resolve", alertingHandler.ResolveAlert)
			alerts.GET("/rules", alertingHandler.GetAlertRules)
			alerts.POST("/rules", alertingHandler.CreateAlertRule)
			alerts.PUT("/rules/:id", alertingHandler.UpdateAlertRule)
			alerts.DELETE("/rules/:id", alertingHandler.DeleteAlertRule)
		}

		// Dashboard routes
		dashboards := v1.Group("/dashboards")
		{
			dashboards.GET("/", monitoringHandler.GetDashboards)
			dashboards.GET("/:id", monitoringHandler.GetDashboard)
			dashboards.POST("/", monitoringHandler.CreateDashboard)
			dashboards.PUT("/:id", monitoringHandler.UpdateDashboard)
			dashboards.DELETE("/:id", monitoringHandler.DeleteDashboard)
		}

		// System status routes
		status := v1.Group("/status")
		{
			status.GET("/", monitoringHandler.GetSystemStatus)
			status.GET("/summary", monitoringHandler.GetStatusSummary)
			status.GET("/incidents", monitoringHandler.GetIncidents)
			status.POST("/incidents", monitoringHandler.CreateIncident)
			status.PUT("/incidents/:id", monitoringHandler.UpdateIncident)
		}
	}

	// Start background services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start metrics collection
	go monitoringService.StartMetricsCollection(ctx)

	// Start health monitoring
	go healthService.StartHealthMonitoring(ctx)

	// Start alert processing
	go alertingService.StartAlertProcessing(ctx)

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