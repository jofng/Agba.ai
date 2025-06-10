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

	"user-management-service/internal/auth"
	"user-management-service/internal/config"
	"user-management-service/internal/handlers"
	"user-management-service/internal/middleware"
	"user-management-service/internal/repository"
	"user-management-service/internal/service"
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
	userRepo := repository.NewUserRepository(db, logger)
	organizationRepo := repository.NewOrganizationRepository(db, logger)
	roleRepo := repository.NewRoleRepository(db, logger)
	sessionRepo := repository.NewSessionRepository(redisClient, logger)
	auditRepo := repository.NewAuditRepository(db, logger)

	// Initialize authentication manager
	authManager := auth.NewManager(cfg.Auth, logger)

	// Initialize services
	userService := service.NewUserService(userRepo, organizationRepo, roleRepo, authManager, natsConn, logger)
	organizationService := service.NewOrganizationService(organizationRepo, userRepo, roleRepo, natsConn, logger)
	roleService := service.NewRoleService(roleRepo, userRepo, logger)
	sessionService := service.NewSessionService(sessionRepo, userRepo, authManager, logger)
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
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())
	router.Use(middleware.Logging(logger))

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
	userHandler := handlers.NewUserHandler(userService, logger)
	organizationHandler := handlers.NewOrganizationHandler(organizationService, logger)
	roleHandler := handlers.NewRoleHandler(roleService, logger)
	authHandler := handlers.NewAuthHandler(sessionService, userService, authManager, logger)
	auditHandler := handlers.NewAuditHandler(auditService, logger)

	// Authentication middleware
	authMiddleware := middleware.NewAuthMiddleware(sessionService, authManager, logger)

	// Public routes (no authentication required)
	public := router.Group("/api/v1")
	{
		// Authentication routes
		auth := public.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
			auth.POST("/verify-email", authHandler.VerifyEmail)
			auth.POST("/resend-verification", authHandler.ResendVerification)
		}

		// User registration (public)
		public.POST("/users/register", userHandler.Register)
		public.POST("/organizations/register", organizationHandler.Register)
	}

	// Protected routes (authentication required)
	protected := router.Group("/api/v1")
	protected.Use(authMiddleware.RequireAuth())
	{
		// User management routes
		users := protected.Group("/users")
		{
			users.GET("/", userHandler.ListUsers)
			users.GET("/me", userHandler.GetCurrentUser)
			users.PUT("/me", userHandler.UpdateCurrentUser)
			users.PUT("/me/password", userHandler.ChangePassword)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
			users.POST("/:id/activate", userHandler.ActivateUser)
			users.POST("/:id/deactivate", userHandler.DeactivateUser)
			users.GET("/:id/roles", userHandler.GetUserRoles)
			users.POST("/:id/roles", userHandler.AssignRole)
			users.DELETE("/:id/roles/:roleId", userHandler.RemoveRole)
			users.GET("/:id/permissions", userHandler.GetUserPermissions)
			users.GET("/:id/sessions", userHandler.GetUserSessions)
			users.DELETE("/:id/sessions/:sessionId", userHandler.RevokeSession)
		}

		// Organization management routes
		organizations := protected.Group("/organizations")
		{
			organizations.GET("/", organizationHandler.ListOrganizations)
			organizations.GET("/:id", organizationHandler.GetOrganization)
			organizations.PUT("/:id", organizationHandler.UpdateOrganization)
			organizations.DELETE("/:id", organizationHandler.DeleteOrganization)
			organizations.GET("/:id/users", organizationHandler.GetOrganizationUsers)
			organizations.POST("/:id/users", organizationHandler.AddUser)
			organizations.DELETE("/:id/users/:userId", organizationHandler.RemoveUser)
			organizations.GET("/:id/roles", organizationHandler.GetOrganizationRoles)
			organizations.POST("/:id/invite", organizationHandler.InviteUser)
			organizations.GET("/:id/invitations", organizationHandler.GetInvitations)
			organizations.POST("/:id/invitations/:inviteId/accept", organizationHandler.AcceptInvitation)
			organizations.POST("/:id/invitations/:inviteId/decline", organizationHandler.DeclineInvitation)
		}

		// Role management routes
		roles := protected.Group("/roles")
		{
			roles.GET("/", roleHandler.ListRoles)
			roles.GET("/:id", roleHandler.GetRole)
			roles.POST("/", roleHandler.CreateRole)
			roles.PUT("/:id", roleHandler.UpdateRole)
			roles.DELETE("/:id", roleHandler.DeleteRole)
			roles.GET("/:id/permissions", roleHandler.GetRolePermissions)
			roles.POST("/:id/permissions", roleHandler.AddPermission)
			roles.DELETE("/:id/permissions/:permissionId", roleHandler.RemovePermission)
			roles.GET("/:id/users", roleHandler.GetRoleUsers)
		}

		// Session management routes
		sessions := protected.Group("/sessions")
		{
			sessions.GET("/", authHandler.GetSessions)
			sessions.DELETE("/:id", authHandler.RevokeSession)
			sessions.DELETE("/", authHandler.RevokeAllSessions)
		}

		// Audit routes
		audit := protected.Group("/audit")
		{
			audit.GET("/", auditHandler.GetAuditLogs)
			audit.GET("/:id", auditHandler.GetAuditLog)
			audit.GET("/users/:userId", auditHandler.GetUserAuditLogs)
			audit.GET("/organizations/:orgId", auditHandler.GetOrganizationAuditLogs)
		}

		// Profile management routes
		profile := protected.Group("/profile")
		{
			profile.GET("/", userHandler.GetProfile)
			profile.PUT("/", userHandler.UpdateProfile)
			profile.POST("/avatar", userHandler.UploadAvatar)
			profile.DELETE("/avatar", userHandler.DeleteAvatar)
			profile.GET("/preferences", userHandler.GetPreferences)
			profile.PUT("/preferences", userHandler.UpdatePreferences)
			profile.GET("/activity", userHandler.GetActivity)
		}
	}

	// Admin routes (admin authentication required)
	admin := router.Group("/api/v1/admin")
	admin.Use(authMiddleware.RequireAuth())
	admin.Use(authMiddleware.RequireRole("admin"))
	{
		// System administration
		admin.GET("/stats", userHandler.GetSystemStats)
		admin.GET("/health", userHandler.GetSystemHealth)
		admin.POST("/maintenance", userHandler.EnableMaintenance)
		admin.DELETE("/maintenance", userHandler.DisableMaintenance)

		// User administration
		adminUsers := admin.Group("/users")
		{
			adminUsers.POST("/", userHandler.CreateUser)
			adminUsers.POST("/:id/impersonate", userHandler.ImpersonateUser)
			adminUsers.POST("/:id/force-password-reset", userHandler.ForcePasswordReset)
			adminUsers.POST("/bulk-import", userHandler.BulkImportUsers)
			adminUsers.POST("/bulk-export", userHandler.BulkExportUsers)
		}

		// Organization administration
		adminOrgs := admin.Group("/organizations")
		{
			adminOrgs.POST("/", organizationHandler.CreateOrganization)
			adminOrgs.POST("/:id/suspend", organizationHandler.SuspendOrganization)
			adminOrgs.POST("/:id/unsuspend", organizationHandler.UnsuspendOrganization)
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
		logger.Info("Starting User Management Service server",
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

	logger.Info("Shutting down User Management Service server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown error", zap.Error(err))
	}

	logger.Info("User Management Service server stopped")
}