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

	"config-service/internal/config"
	"config-service/internal/handlers"
	"config-service/internal/repository"
	"config-service/internal/service"
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
	dbRepo := repository.NewDatabaseRepo(db, logger)
	redisRepo := repository.NewRedisRepo(redisClient, logger)

	// Initialize services
	configService := service.NewConfigService(dbRepo, redisRepo, natsConn, logger)

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
	configHandler := handlers.NewConfigHandler(configService, logger)

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Configuration routes
		configs := v1.Group("/configs")
		{
			configs.POST("", configHandler.CreateConfig)
			configs.GET("/:id", configHandler.GetConfig)
			configs.GET("", configHandler.GetConfigs)
			configs.PUT("/:id", configHandler.UpdateConfig)
			configs.DELETE("/:id", configHandler.DeleteConfig)
			configs.GET("/:id/history", configHandler.GetConfigHistory)
			configs.POST("/:id/validate", configHandler.ValidateConfig)
		}

		// Configuration values routes
		values := v1.Group("/values")
		{
			values.GET("/:service/:environment/:key", configHandler.GetConfigValue)
			values.PUT("/:service/:environment/:key", configHandler.SetConfigValue)
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