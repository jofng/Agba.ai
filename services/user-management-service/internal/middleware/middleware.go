package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"user-management-service/internal/auth"
	"user-management-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AuthMiddleware handles authentication middleware
type AuthMiddleware struct {
	sessionService service.SessionServiceInterface
	authManager    *auth.Manager
	logger         *zap.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(sessionService service.SessionServiceInterface, authManager *auth.Manager, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		sessionService: sessionService,
		authManager:    authManager,
		logger:         logger,
	}
}

// RequireAuth middleware that requires authentication
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization token",
			})
			c.Abort()
			return
		}

		claims, err := m.authManager.ValidateToken(token)
		if err != nil {
			m.logger.Debug("Token validation failed", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			c.Abort()
			return
		}

		// Verify session is still active
		session, err := m.sessionService.GetSession(c.Request.Context(), claims.SessionID)
		if err != nil || session == nil {
			m.logger.Debug("Session not found or expired", zap.String("session_id", claims.SessionID))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "session expired",
			})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("organization_id", claims.OrganizationID)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)
		c.Set("permissions", claims.Permissions)
		c.Set("session_id", claims.SessionID)
		c.Set("claims", claims)

		c.Next()
	}
}

// RequireRole middleware that requires a specific role
func (m *AuthMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "no roles found in context",
			})
			c.Abort()
			return
		}

		userRoles, ok := roles.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "invalid roles format",
			})
			c.Abort()
			return
		}

		if !m.authManager.HasRole(userRoles, requiredRole) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission middleware that requires a specific permission
func (m *AuthMiddleware) RequirePermission(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("permissions")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "no permissions found in context",
			})
			c.Abort()
			return
		}

		userPermissions, ok := permissions.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "invalid permissions format",
			})
			c.Abort()
			return
		}

		if !m.authManager.HasPermission(userPermissions, requiredPermission) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole middleware that requires any of the specified roles
func (m *AuthMiddleware) RequireAnyRole(requiredRoles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "no roles found in context",
			})
			c.Abort()
			return
		}

		userRoles, ok := roles.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "invalid roles format",
			})
			c.Abort()
			return
		}

		if !m.authManager.HasAnyRole(userRoles, requiredRoles) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CORS middleware for handling Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestID middleware adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)

		// Add request ID to context for downstream services
		ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// Logging middleware for structured request logging
func Logging(logger *zap.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Custom log format using zap
		logger.Info("HTTP Request",
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("client_ip", param.ClientIP),
			zap.String("user_agent", param.Request.UserAgent()),
			zap.String("request_id", param.Keys["request_id"].(string)),
		)
		return ""
	})
}

// RateLimiting middleware for API rate limiting
func RateLimiting() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement rate limiting logic with Redis
		// For now, just pass through
		c.Next()
	}
}

// SecurityHeaders middleware adds security headers
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// Timeout middleware for request timeout handling
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// OrganizationAccess middleware that ensures user has access to the organization
func (m *AuthMiddleware) OrganizationAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		userOrgID, exists := c.Get("organization_id")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "no organization found in context",
			})
			c.Abort()
			return
		}

		requestedOrgID := c.Param("orgId")
		if requestedOrgID == "" {
			requestedOrgID = c.Param("id") // fallback for organization routes
		}

		if requestedOrgID != "" && requestedOrgID != userOrgID {
			// Check if user has admin role to access other organizations
			roles, exists := c.Get("roles")
			if !exists {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "access denied",
				})
				c.Abort()
				return
			}

			userRoles, ok := roles.([]string)
			if !ok || !m.authManager.HasRole(userRoles, "admin") {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "access denied",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// extractToken extracts the JWT token from the Authorization header
func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check for Bearer token format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}