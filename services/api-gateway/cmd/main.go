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

	"github.com/agba-ai/api-gateway/internal/config"
	"github.com/agba-ai/api-gateway/internal/gateway"
	"github.com/agba-ai/api-gateway/internal/middleware"
	"github.com/agba-ai/api-gateway/internal/routes"
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

	// Initialize gateway components
	gw, err := gateway.New(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize gateway", zap.Error(err))
	}

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.New()

	// Add middleware
	router.Use(middleware.Logger(logger))
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.CORS(cfg.CORS))
	router.Use(middleware.RequestID())
	router.Use(middleware.Metrics())

	// Health check endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   cfg.Version,
		})
	})

	router.GET("/ready", func(c *gin.Context) {
		if gw.IsReady() {
			c.JSON(http.StatusOK, gin.H{
				"status": "ready",
			})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
			})
		}
	})

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API routes with authentication and rate limiting
	api := router.Group("/api/v1")
	api.Use(middleware.Authentication(gw.GetAuthService()))
	api.Use(middleware.RateLimit(gw.GetRateLimiter()))

	// Register routes
	routes.RegisterCallRoutes(api, gw)
	routes.RegisterAgentRoutes(api, gw)
	routes.RegisterAnalyticsRoutes(api, gw)
	routes.RegisterWebhookRoutes(api, gw)
	routes.RegisterUserRoutes(api, gw)
	routes.RegisterConfigRoutes(api, gw)

	// WebRTC routes (separate group for different auth)
	webrtc := router.Group("/webrtc")
	webrtc.Use(middleware.WebRTCAuth(gw.GetAuthService()))
	routes.RegisterWebRTCRoutes(webrtc, gw)

	// Admin routes (higher privilege required)
	admin := router.Group("/admin")
	admin.Use(middleware.AdminAuth(gw.GetAuthService()))
	routes.RegisterAdminRoutes(admin, gw)

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
		logger.Info("Starting API Gateway server",
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

	logger.Info("Shutting down API Gateway server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown gateway components
	if err := gw.Shutdown(ctx); err != nil {
		logger.Error("Gateway shutdown error", zap.Error(err))
	}

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown error", zap.Error(err))
	}

	logger.Info("API Gateway server stopped")
}