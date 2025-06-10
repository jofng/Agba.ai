package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"freeswitch-service/internal/config"
	"freeswitch-service/internal/freeswitch"
	"freeswitch-service/internal/handlers"
	"freeswitch-service/internal/integration"
	"freeswitch-service/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load configuration: %v", err))
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		panic(fmt.Sprintf("Invalid configuration: %v", err))
	}

	// Initialize logger
	logger := initLogger(cfg.Logging)
	defer logger.Sync()

	logger.Info("Starting FreeSWITCH Service",
		zap.String("version", "1.0.0"),
		zap.String("host", cfg.Server.Host),
		zap.Int("port", cfg.Server.Port))

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize FreeSWITCH manager
	fsManager := freeswitch.NewManager(&cfg.FreeSWITCH, logger)

	// Initialize service integration
	serviceIntegration := integration.NewServiceIntegration(cfg, logger)

	// Initialize FreeSWITCH service
	freeswitchService := service.NewFreeSWITCHService(cfg, fsManager, serviceIntegration, logger)

	// Start FreeSWITCH service
	if err := freeswitchService.Start(ctx); err != nil {
		logger.Fatal("Failed to start FreeSWITCH service", zap.Error(err))
	}

	// Initialize HTTP handlers
	handler := handlers.NewFreeSWITCHHandler(freeswitchService, logger)

	// Setup HTTP server
	router := setupRouter(cfg, logger)
	handler.RegisterRoutes(router)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Start HTTP server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Cancel context to stop background services
	cancel()

	// Shutdown HTTP server with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	// Stop FreeSWITCH service
	if err := freeswitchService.Stop(); err != nil {
		logger.Error("Failed to stop FreeSWITCH service", zap.Error(err))
	}

	logger.Info("Server exited")
}

// initLogger initializes the logger based on configuration
func initLogger(cfg config.LoggingConfig) *zap.Logger {
	var zapConfig zap.Config

	if cfg.Format == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	// Set log level
	switch cfg.Level {
	case "debug":
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	// Configure output
	if cfg.Output == "stdout" {
		zapConfig.OutputPaths = []string{"stdout"}
	} else {
		zapConfig.OutputPaths = []string{cfg.Output}
	}

	logger, err := zapConfig.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	return logger
}

// setupRouter sets up the Gin router with middleware
func setupRouter(cfg *config.Config, logger *zap.Logger) *gin.Engine {
	// Set Gin mode based on log level
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(loggingMiddleware(logger))

	// Add health check endpoint at root
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "freeswitch-service",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	return router
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("HTTP Request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
		)
	}
}