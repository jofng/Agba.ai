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

	"recording-service/internal/config"
	"recording-service/internal/handlers"
	"recording-service/internal/processors"
	"recording-service/internal/repository"
	"recording-service/internal/service"
	"recording-service/internal/storage"
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

	// Initialize storage providers
	fileStorage, err := storage.NewFileStorage(cfg.Storage.File, logger)
	if err != nil {
		logger.Fatal("Failed to initialize file storage", zap.Error(err))
	}

	s3Storage, err := storage.NewS3Storage(cfg.Storage.S3, logger)
	if err != nil {
		logger.Fatal("Failed to initialize S3 storage", zap.Error(err))
	}

	gcsStorage, err := storage.NewGCSStorage(cfg.Storage.GCS, logger)
	if err != nil {
		logger.Fatal("Failed to initialize GCS storage", zap.Error(err))
	}

	// Initialize storage manager
	storageManager := storage.NewStorageManager(fileStorage, s3Storage, gcsStorage, cfg.Storage, logger)

	// Initialize audio processors
	audioProcessor := processors.NewAudioProcessor(cfg.Audio, logger)
	transcriptionProcessor := processors.NewTranscriptionProcessor(cfg.Transcription, logger)
	compressionProcessor := processors.NewCompressionProcessor(cfg.Compression, logger)

	// Initialize repositories
	recordingRepo := repository.NewRecordingRepository(db, logger)
	metadataRepo := repository.NewMetadataRepository(db, logger)
	auditRepo := repository.NewAuditRepository(db, logger)
	sessionRepo := repository.NewSessionRepository(db, redisClient, logger)

	// Initialize services
	recordingService := service.NewRecordingService(
		recordingRepo,
		metadataRepo,
		auditRepo,
		sessionRepo,
		storageManager,
		audioProcessor,
		transcriptionProcessor,
		compressionProcessor,
		natsConn,
		logger,
	)

	metadataService := service.NewMetadataService(metadataRepo, logger)
	auditService := service.NewAuditService(auditRepo, logger)
	sessionService := service.NewSessionService(sessionRepo, logger)
	storageService := service.NewStorageService(storageManager, logger)

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

		// Check storage connectivity
		if err := storageManager.HealthCheck(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"error":  "storage connection failed",
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
	recordingHandler := handlers.NewRecordingHandler(recordingService, logger)
	metadataHandler := handlers.NewMetadataHandler(metadataService, logger)
	auditHandler := handlers.NewAuditHandler(auditService, logger)
	sessionHandler := handlers.NewSessionHandler(sessionService, logger)
	storageHandler := handlers.NewStorageHandler(storageService, logger)

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Recording management routes
		recordings := v1.Group("/recordings")
		{
			recordings.POST("/", recordingHandler.CreateRecording)
			recordings.GET("/", recordingHandler.GetRecordings)
			recordings.GET("/:id", recordingHandler.GetRecording)
			recordings.PUT("/:id", recordingHandler.UpdateRecording)
			recordings.DELETE("/:id", recordingHandler.DeleteRecording)
			recordings.POST("/:id/upload", recordingHandler.UploadRecording)
			recordings.GET("/:id/download", recordingHandler.DownloadRecording)
			recordings.GET("/:id/stream", recordingHandler.StreamRecording)
			recordings.POST("/:id/process", recordingHandler.ProcessRecording)
			recordings.GET("/:id/status", recordingHandler.GetRecordingStatus)
			recordings.POST("/:id/transcribe", recordingHandler.TranscribeRecording)
			recordings.GET("/:id/transcription", recordingHandler.GetTranscription)
		}

		// Real-time recording routes
		realtime := v1.Group("/realtime")
		{
			realtime.POST("/start", recordingHandler.StartRecording)
			realtime.POST("/stop", recordingHandler.StopRecording)
			realtime.POST("/pause", recordingHandler.PauseRecording)
			realtime.POST("/resume", recordingHandler.ResumeRecording)
			realtime.POST("/chunk", recordingHandler.UploadChunk)
			realtime.GET("/status/:sessionId", recordingHandler.GetRealtimeStatus)
		}

		// Batch processing routes
		batch := v1.Group("/batch")
		{
			batch.POST("/process", recordingHandler.BatchProcess)
			batch.POST("/transcribe", recordingHandler.BatchTranscribe)
			batch.POST("/compress", recordingHandler.BatchCompress)
			batch.POST("/migrate", recordingHandler.BatchMigrate)
			batch.GET("/jobs", recordingHandler.GetBatchJobs)
			batch.GET("/jobs/:id", recordingHandler.GetBatchJob)
			batch.DELETE("/jobs/:id", recordingHandler.CancelBatchJob)
		}

		// Metadata management routes
		metadata := v1.Group("/metadata")
		{
			metadata.GET("/", metadataHandler.GetMetadata)
			metadata.GET("/:id", metadataHandler.GetMetadataByID)
			metadata.POST("/", metadataHandler.CreateMetadata)
			metadata.PUT("/:id", metadataHandler.UpdateMetadata)
			metadata.DELETE("/:id", metadataHandler.DeleteMetadata)
			metadata.GET("/recording/:recordingId", metadataHandler.GetRecordingMetadata)
			metadata.POST("/search", metadataHandler.SearchMetadata)
		}

		// Session management routes
		sessions := v1.Group("/sessions")
		{
			sessions.GET("/", sessionHandler.GetSessions)
			sessions.GET("/:id", sessionHandler.GetSession)
			sessions.POST("/", sessionHandler.CreateSession)
			sessions.PUT("/:id", sessionHandler.UpdateSession)
			sessions.DELETE("/:id", sessionHandler.DeleteSession)
			sessions.GET("/:id/recordings", sessionHandler.GetSessionRecordings)
			sessions.POST("/:id/recordings", sessionHandler.AddRecordingToSession)
		}

		// Storage management routes
		storage := v1.Group("/storage")
		{
			storage.GET("/", storageHandler.GetStorageInfo)
			storage.GET("/usage", storageHandler.GetStorageUsage)
			storage.POST("/cleanup", storageHandler.CleanupStorage)
			storage.POST("/migrate", storageHandler.MigrateStorage)
			storage.GET("/providers", storageHandler.GetStorageProviders)
			storage.POST("/test", storageHandler.TestStorageProvider)
		}

		// Audit and compliance routes
		audit := v1.Group("/audit")
		{
			audit.GET("/", auditHandler.GetAuditLogs)
			audit.GET("/:id", auditHandler.GetAuditLog)
			audit.GET("/recording/:recordingId", auditHandler.GetRecordingAuditLogs)
			audit.POST("/export", auditHandler.ExportAuditLogs)
			audit.GET("/compliance", auditHandler.GetComplianceReport)
		}

		// Analytics and reporting routes
		analytics := v1.Group("/analytics")
		{
			analytics.GET("/", recordingHandler.GetAnalytics)
			analytics.GET("/usage", recordingHandler.GetUsageStatistics)
			analytics.GET("/storage", recordingHandler.GetStorageStatistics)
			analytics.GET("/quality", recordingHandler.GetQualityMetrics)
			analytics.GET("/performance", recordingHandler.GetPerformanceMetrics)
		}

		// Admin routes
		admin := v1.Group("/admin")
		{
			admin.GET("/health-check", recordingHandler.HealthCheck)
			admin.POST("/maintenance", recordingHandler.EnableMaintenance)
			admin.DELETE("/maintenance", recordingHandler.DisableMaintenance)
			admin.GET("/system-info", recordingHandler.GetSystemInfo)
			admin.POST("/cleanup", recordingHandler.CleanupOldRecordings)
			admin.GET("/queue-status", recordingHandler.GetQueueStatus)
			admin.POST("/reprocess-failed", recordingHandler.ReprocessFailedRecordings)
		}
	}

	// WebSocket endpoint for real-time recording
	router.GET("/ws/recording/:sessionId", recordingHandler.HandleWebSocket)

	// Start background services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start recording processor
	go recordingService.StartRecordingProcessor(ctx)

	// Start transcription processor
	go recordingService.StartTranscriptionProcessor(ctx)

	// Start cleanup processor
	go recordingService.StartCleanupProcessor(ctx)

	// Start NATS message consumer
	go recordingService.StartNATSConsumer(ctx)

	// Start storage monitor
	go storageService.StartStorageMonitor(ctx)

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
		logger.Info("Starting Recording Service server",
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

	logger.Info("Shutting down Recording Service server...")

	// Cancel background services
	cancel()

	// Create a deadline for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown error", zap.Error(err))
	}

	logger.Info("Recording Service server stopped")
}