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

	"notification-service/internal/config"
	"notification-service/internal/handlers"
	"notification-service/internal/providers"
	"notification-service/internal/repository"
	"notification-service/internal/service"
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

	// Initialize notification providers
	emailProvider, err := providers.NewEmailProvider(cfg.Email, logger)
	if err != nil {
		logger.Fatal("Failed to initialize email provider", zap.Error(err))
	}

	smsProvider, err := providers.NewSMSProvider(cfg.SMS, logger)
	if err != nil {
		logger.Fatal("Failed to initialize SMS provider", zap.Error(err))
	}

	pushProvider, err := providers.NewPushProvider(cfg.Push, logger)
	if err != nil {
		logger.Fatal("Failed to initialize push provider", zap.Error(err))
	}

	webhookProvider := providers.NewWebhookProvider(cfg.Webhook, logger)

	// Initialize repositories
	notificationRepo := repository.NewNotificationRepository(db, logger)
	templateRepo := repository.NewTemplateRepository(db, logger)
	subscriptionRepo := repository.NewSubscriptionRepository(db, redisClient, logger)
	auditRepo := repository.NewAuditRepository(db, logger)

	// Initialize services
	notificationService := service.NewNotificationService(
		notificationRepo,
		templateRepo,
		subscriptionRepo,
		auditRepo,
		emailProvider,
		smsProvider,
		pushProvider,
		webhookProvider,
		natsConn,
		logger,
	)

	templateService := service.NewTemplateService(templateRepo, logger)
	subscriptionService := service.NewSubscriptionService(subscriptionRepo, logger)
	auditService := service.NewAuditService(auditRepo, logger)

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

		// Check email provider
		if err := emailProvider.HealthCheck(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"error":  "email provider connection failed",
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
	notificationHandler := handlers.NewNotificationHandler(notificationService, logger)
	templateHandler := handlers.NewTemplateHandler(templateService, logger)
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionService, logger)
	auditHandler := handlers.NewAuditHandler(auditService, logger)

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Notification routes
		notifications := v1.Group("/notifications")
		{
			notifications.POST("/", notificationHandler.SendNotification)
			notifications.POST("/bulk", notificationHandler.SendBulkNotifications)
			notifications.GET("/", notificationHandler.GetNotifications)
			notifications.GET("/:id", notificationHandler.GetNotification)
			notifications.PUT("/:id/retry", notificationHandler.RetryNotification)
			notifications.DELETE("/:id", notificationHandler.DeleteNotification)
			notifications.GET("/:id/status", notificationHandler.GetNotificationStatus)
		}

		// Email specific routes
		email := v1.Group("/email")
		{
			email.POST("/send", notificationHandler.SendEmail)
			email.POST("/send-template", notificationHandler.SendTemplateEmail)
			email.POST("/send-bulk", notificationHandler.SendBulkEmails)
			email.GET("/verify/:token", notificationHandler.VerifyEmail)
			email.POST("/bounce", notificationHandler.HandleEmailBounce)
			email.POST("/complaint", notificationHandler.HandleEmailComplaint)
		}

		// SMS specific routes
		sms := v1.Group("/sms")
		{
			sms.POST("/send", notificationHandler.SendSMS)
			sms.POST("/send-bulk", notificationHandler.SendBulkSMS)
			sms.POST("/delivery-report", notificationHandler.HandleSMSDeliveryReport)
		}

		// Push notification routes
		push := v1.Group("/push")
		{
			push.POST("/send", notificationHandler.SendPushNotification)
			push.POST("/send-bulk", notificationHandler.SendBulkPushNotifications)
			push.POST("/register-device", notificationHandler.RegisterDevice)
			push.DELETE("/device/:token", notificationHandler.UnregisterDevice)
		}

		// Webhook routes
		webhooks := v1.Group("/webhooks")
		{
			webhooks.POST("/send", notificationHandler.SendWebhook)
			webhooks.POST("/send-bulk", notificationHandler.SendBulkWebhooks)
		}

		// Template management routes
		templates := v1.Group("/templates")
		{
			templates.GET("/", templateHandler.GetTemplates)
			templates.GET("/:id", templateHandler.GetTemplate)
			templates.POST("/", templateHandler.CreateTemplate)
			templates.PUT("/:id", templateHandler.UpdateTemplate)
			templates.DELETE("/:id", templateHandler.DeleteTemplate)
			templates.POST("/:id/preview", templateHandler.PreviewTemplate)
			templates.POST("/:id/test", templateHandler.TestTemplate)
		}

		// Subscription management routes
		subscriptions := v1.Group("/subscriptions")
		{
			subscriptions.GET("/", subscriptionHandler.GetSubscriptions)
			subscriptions.GET("/:id", subscriptionHandler.GetSubscription)
			subscriptions.POST("/", subscriptionHandler.CreateSubscription)
			subscriptions.PUT("/:id", subscriptionHandler.UpdateSubscription)
			subscriptions.DELETE("/:id", subscriptionHandler.DeleteSubscription)
			subscriptions.POST("/:id/unsubscribe", subscriptionHandler.Unsubscribe)
			subscriptions.GET("/user/:userId", subscriptionHandler.GetUserSubscriptions)
		}

		// Audit routes
		audit := v1.Group("/audit")
		{
			audit.GET("/", auditHandler.GetAuditLogs)
			audit.GET("/:id", auditHandler.GetAuditLog)
			audit.GET("/notifications/:notificationId", auditHandler.GetNotificationAuditLogs)
		}

		// Statistics and analytics routes
		stats := v1.Group("/stats")
		{
			stats.GET("/", notificationHandler.GetStatistics)
			stats.GET("/delivery-rates", notificationHandler.GetDeliveryRates)
			stats.GET("/provider-performance", notificationHandler.GetProviderPerformance)
			stats.GET("/template-usage", notificationHandler.GetTemplateUsage)
		}

		// Admin routes
		admin := v1.Group("/admin")
		{
			admin.GET("/health-check", notificationHandler.HealthCheck)
			admin.POST("/test-providers", notificationHandler.TestProviders)
			admin.GET("/queue-status", notificationHandler.GetQueueStatus)
			admin.POST("/flush-queue", notificationHandler.FlushQueue)
			admin.GET("/failed-notifications", notificationHandler.GetFailedNotifications)
			admin.POST("/retry-failed", notificationHandler.RetryFailedNotifications)
		}
	}

	// Start background services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start notification processing
	go notificationService.StartNotificationProcessor(ctx)

	// Start retry processor
	go notificationService.StartRetryProcessor(ctx)

	// Start cleanup processor
	go notificationService.StartCleanupProcessor(ctx)

	// Start NATS message consumer
	go notificationService.StartNATSConsumer(ctx)

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
		logger.Info("Starting Notification Service server",
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

	logger.Info("Shutting down Notification Service server...")

	// Cancel background services
	cancel()

	// Create a deadline for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown error", zap.Error(err))
	}

	logger.Info("Notification Service server stopped")
}