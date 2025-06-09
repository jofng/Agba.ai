package gateway

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/agba-ai/api-gateway/internal/auth"
	"github.com/agba-ai/api-gateway/internal/config"
	"github.com/agba-ai/api-gateway/internal/proxy"
	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// Gateway represents the main gateway instance
type Gateway struct {
	config      *config.Config
	logger      *zap.Logger
	redisClient *redis.Client
	natsConn    *nats.Conn
	authService *auth.Service
	proxy       *proxy.Service
	rateLimiter *RateLimiter
	httpClient  *http.Client
	ready       bool
	mu          sync.RWMutex
}

// New creates a new Gateway instance
func New(cfg *config.Config, logger *zap.Logger) (*Gateway, error) {
	gw := &Gateway{
		config: cfg,
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}

	// Initialize Redis client
	if err := gw.initRedis(); err != nil {
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}

	// Initialize NATS connection
	if err := gw.initNATS(); err != nil {
		return nil, fmt.Errorf("failed to initialize NATS: %w", err)
	}

	// Initialize authentication service
	if err := gw.initAuth(); err != nil {
		return nil, fmt.Errorf("failed to initialize auth service: %w", err)
	}

	// Initialize proxy service
	if err := gw.initProxy(); err != nil {
		return nil, fmt.Errorf("failed to initialize proxy service: %w", err)
	}

	// Initialize rate limiter
	if err := gw.initRateLimiter(); err != nil {
		return nil, fmt.Errorf("failed to initialize rate limiter: %w", err)
	}

	// Start health checks
	go gw.startHealthChecks()

	gw.setReady(true)
	logger.Info("Gateway initialized successfully")

	return gw, nil
}

// initRedis initializes the Redis client
func (gw *Gateway) initRedis() error {
	gw.redisClient = redis.NewClient(&redis.Options{
		Addr:     gw.config.Redis.Address,
		Password: gw.config.Redis.Password,
		DB:       gw.config.Redis.DB,
		PoolSize: gw.config.Redis.PoolSize,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := gw.redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis connection failed: %w", err)
	}

	gw.logger.Info("Redis connection established")
	return nil
}

// initNATS initializes the NATS connection
func (gw *Gateway) initNATS() error {
	opts := []nats.Option{
		nats.UserInfo(gw.config.NATS.Username, gw.config.NATS.Password),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(-1),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			gw.logger.Warn("NATS disconnected", zap.Error(err))
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			gw.logger.Info("NATS reconnected")
		}),
	}

	nc, err := nats.Connect(gw.config.NATS.URL, opts...)
	if err != nil {
		return fmt.Errorf("NATS connection failed: %w", err)
	}

	gw.natsConn = nc
	gw.logger.Info("NATS connection established")
	return nil
}

// initAuth initializes the authentication service
func (gw *Gateway) initAuth() error {
	authService, err := auth.New(&auth.Config{
		OAuth2ServerURL:    gw.config.Auth.OAuth2ServerURL,
		APIKeyValidatorURL: gw.config.Auth.APIKeyValidatorURL,
		JWTSecret:          gw.config.Auth.JWTSecret,
		TokenCacheTTL:      gw.config.Auth.TokenCacheTTL,
		RedisClient:        gw.redisClient,
		HTTPClient:         gw.httpClient,
		Logger:             gw.logger,
	})
	if err != nil {
		return err
	}

	gw.authService = authService
	gw.logger.Info("Authentication service initialized")
	return nil
}

// initProxy initializes the proxy service
func (gw *Gateway) initProxy() error {
	proxyService, err := proxy.New(&proxy.Config{
		Services:   gw.config.Services,
		HTTPClient: gw.httpClient,
		Logger:     gw.logger,
		NATSConn:   gw.natsConn,
	})
	if err != nil {
		return err
	}

	gw.proxy = proxyService
	gw.logger.Info("Proxy service initialized")
	return nil
}

// initRateLimiter initializes the rate limiter
func (gw *Gateway) initRateLimiter() error {
	rateLimiter, err := NewRateLimiter(&RateLimiterConfig{
		RedisClient:     gw.redisClient,
		DefaultRPM:      gw.config.RateLimit.DefaultRPM,
		DefaultRPH:      gw.config.RateLimit.DefaultRPH,
		BurstMultiplier: gw.config.RateLimit.BurstMultiplier,
		CleanupInterval: gw.config.RateLimit.CleanupInterval,
		KeyPrefix:       gw.config.RateLimit.RedisKeyPrefix,
		SlidingWindow:   gw.config.RateLimit.SlidingWindow,
		BlockDuration:   gw.config.RateLimit.BlockDuration,
		Logger:          gw.logger,
	})
	if err != nil {
		return err
	}

	gw.rateLimiter = rateLimiter
	gw.logger.Info("Rate limiter initialized")
	return nil
}

// startHealthChecks starts periodic health checks for dependencies
func (gw *Gateway) startHealthChecks() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			gw.performHealthChecks()
		}
	}
}

// performHealthChecks checks the health of all dependencies
func (gw *Gateway) performHealthChecks() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	healthy := true

	// Check Redis
	if err := gw.redisClient.Ping(ctx).Err(); err != nil {
		gw.logger.Error("Redis health check failed", zap.Error(err))
		healthy = false
	}

	// Check NATS
	if !gw.natsConn.IsConnected() {
		gw.logger.Error("NATS health check failed - not connected")
		healthy = false
	}

	// Check backend services
	if !gw.proxy.HealthCheck(ctx) {
		gw.logger.Error("Backend services health check failed")
		healthy = false
	}

	gw.setReady(healthy)
}

// GetAuthService returns the authentication service
func (gw *Gateway) GetAuthService() *auth.Service {
	return gw.authService
}

// GetProxy returns the proxy service
func (gw *Gateway) GetProxy() *proxy.Service {
	return gw.proxy
}

// GetRateLimiter returns the rate limiter
func (gw *Gateway) GetRateLimiter() *RateLimiter {
	return gw.rateLimiter
}

// GetRedisClient returns the Redis client
func (gw *Gateway) GetRedisClient() *redis.Client {
	return gw.redisClient
}

// GetNATSConn returns the NATS connection
func (gw *Gateway) GetNATSConn() *nats.Conn {
	return gw.natsConn
}

// IsReady returns whether the gateway is ready to serve requests
func (gw *Gateway) IsReady() bool {
	gw.mu.RLock()
	defer gw.mu.RUnlock()
	return gw.ready
}

// setReady sets the ready status
func (gw *Gateway) setReady(ready bool) {
	gw.mu.Lock()
	defer gw.mu.Unlock()
	gw.ready = ready
}

// Shutdown gracefully shuts down the gateway
func (gw *Gateway) Shutdown(ctx context.Context) error {
	gw.logger.Info("Shutting down gateway...")

	gw.setReady(false)

	// Close NATS connection
	if gw.natsConn != nil {
		gw.natsConn.Close()
	}

	// Close Redis connection
	if gw.redisClient != nil {
		if err := gw.redisClient.Close(); err != nil {
			gw.logger.Error("Error closing Redis connection", zap.Error(err))
		}
	}

	// Shutdown rate limiter
	if gw.rateLimiter != nil {
		gw.rateLimiter.Shutdown()
	}

	gw.logger.Info("Gateway shutdown completed")
	return nil
}