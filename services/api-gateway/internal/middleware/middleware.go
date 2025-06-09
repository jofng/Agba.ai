package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/agba-ai/api-gateway/internal/auth"
	"github.com/agba-ai/api-gateway/internal/config"
	"github.com/agba-ai/api-gateway/internal/gateway"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var (
	// Prometheus metrics
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agba_gateway_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "agba_gateway_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	rateLimitHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agba_gateway_rate_limit_hits_total",
			Help: "Total number of rate limit hits",
		},
		[]string{"type", "key"},
	)
)

// Logger middleware for structured logging
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.Info("HTTP Request",
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("client_ip", param.ClientIP),
			zap.String("user_agent", param.Request.UserAgent()),
			zap.String("request_id", param.Request.Header.Get("X-Request-ID")),
		)
		return ""
	})
}

// Recovery middleware with structured logging
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.Error("Panic recovered",
			zap.Any("error", recovered),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("request_id", c.GetString("request_id")),
		)
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

// RequestID middleware adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// CORS middleware for cross-origin requests
func CORS(corsConfig config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		
		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range corsConfig.AllowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", strings.Join(corsConfig.AllowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(corsConfig.AllowedHeaders, ", "))
		c.Header("Access-Control-Max-Age", strconv.Itoa(corsConfig.MaxAge))
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Metrics middleware for Prometheus metrics collection
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		c.Next()

		duration := time.Since(start)
		status := strconv.Itoa(c.Writer.Status())

		requestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		requestDuration.WithLabelValues(c.Request.Method, path).Observe(duration.Seconds())
	}
}

// Authentication middleware for JWT and API key validation
func Authentication(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
				"code":  "MISSING_AUTH_HEADER",
			})
			c.Abort()
			return
		}

		var token string
		var authType string

		// Check for Bearer token (JWT)
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
			authType = "jwt"
		} else if strings.HasPrefix(authHeader, "ApiKey ") {
			token = strings.TrimPrefix(authHeader, "ApiKey ")
			authType = "api_key"
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization format",
				"code":  "INVALID_AUTH_FORMAT",
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := authService.ValidateToken(c.Request.Context(), token, authType)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("organization_id", claims.OrganizationID)
		c.Set("scopes", claims.Scopes)
		c.Set("auth_type", authType)
		c.Set("token", token)

		c.Next()
	}
}

// WebRTCAuth middleware for WebRTC-specific authentication
func WebRTCAuth(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// WebRTC may use different auth mechanisms
		// For now, use the same as regular auth but with different scopes
		Authentication(authService)(c)
		
		if c.IsAborted() {
			return
		}

		// Check for WebRTC-specific scopes
		scopes := c.GetStringSlice("scopes")
		hasWebRTCScope := false
		for _, scope := range scopes {
			if scope == "webrtc:access" || scope == "calls:join" {
				hasWebRTCScope = true
				break
			}
		}

		if !hasWebRTCScope {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions for WebRTC access",
				"code":  "INSUFFICIENT_WEBRTC_PERMISSIONS",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminAuth middleware for admin-only endpoints
func AdminAuth(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First run regular authentication
		Authentication(authService)(c)
		
		if c.IsAborted() {
			return
		}

		// Check for admin scopes
		scopes := c.GetStringSlice("scopes")
		hasAdminScope := false
		for _, scope := range scopes {
			if scope == "admin:read" || scope == "admin:write" {
				hasAdminScope = true
				break
			}
		}

		if !hasAdminScope {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Admin access required",
				"code":  "INSUFFICIENT_ADMIN_PERMISSIONS",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimit middleware for rate limiting
func RateLimit(rateLimiter *gateway.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		var result *gateway.RateLimitResult
		var err error
		var limitKey string

		// Determine rate limiting strategy based on auth type
		authType := c.GetString("auth_type")
		
		switch authType {
		case "api_key":
			token := c.GetString("token")
			result, err = rateLimiter.CheckAPIKey(c.Request.Context(), token, nil)
			limitKey = "api_key:" + token[:8] + "..."
		case "jwt":
			userID := c.GetString("user_id")
			if userID != "" {
				result, err = rateLimiter.CheckUser(c.Request.Context(), userID)
				limitKey = "user:" + userID
			} else {
				// Fall back to IP-based limiting
				result, err = rateLimiter.CheckIP(c.Request.Context(), c.ClientIP())
				limitKey = "ip:" + c.ClientIP()
			}
		default:
			// No auth, use IP-based limiting
			result, err = rateLimiter.CheckIP(c.Request.Context(), c.ClientIP())
			limitKey = "ip:" + c.ClientIP()
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Rate limiting service unavailable",
				"code":  "RATE_LIMIT_ERROR",
			})
			c.Abort()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(result.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetTime.Unix(), 10))

		if !result.Allowed {
			rateLimitHits.WithLabelValues(authType, limitKey).Inc()
			
			if result.IsBlocked {
				c.Header("Retry-After", strconv.FormatInt(int64(result.RetryAfter.Seconds()), 10))
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":        "Rate limit exceeded - temporarily blocked",
					"code":         "RATE_LIMIT_BLOCKED",
					"retry_after":  result.RetryAfter.Seconds(),
					"blocked_until": result.BlockedUntil,
				})
			} else {
				c.Header("Retry-After", strconv.FormatInt(int64(result.RetryAfter.Seconds()), 10))
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":      "Rate limit exceeded",
					"code":       "RATE_LIMIT_EXCEEDED",
					"retry_after": result.RetryAfter.Seconds(),
				})
			}
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireScope middleware to check for specific scopes
func RequireScope(requiredScope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		scopes := c.GetStringSlice("scopes")
		
		hasScope := false
		for _, scope := range scopes {
			if scope == requiredScope {
				hasScope = true
				break
			}
		}

		if !hasScope {
			c.JSON(http.StatusForbidden, gin.H{
				"error":          "Insufficient permissions",
				"code":           "INSUFFICIENT_SCOPE",
				"required_scope": requiredScope,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Timeout middleware to set request timeout
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}