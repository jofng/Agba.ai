package repository

import (
	"context"
	"fmt"

	"github.com/agba-ai/config-service/internal/config"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// Redis represents the Redis client
type Redis struct {
	client *redis.Client
	logger *zap.Logger
}

// NewRedis creates a new Redis client
func NewRedis(cfg config.RedisConfig, logger *zap.Logger) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis connection established",
		zap.String("address", cfg.Address),
		zap.Int("db", cfg.DB),
		zap.Int("pool_size", cfg.PoolSize),
	)

	return &Redis{
		client: client,
		logger: logger,
	}, nil
}

// Client returns the underlying Redis client
func (r *Redis) Client() *redis.Client {
	return r.client
}

// Ping tests the Redis connection
func (r *Redis) Ping(ctx context.Context) *redis.StatusCmd {
	return r.client.Ping(ctx)
}

// Close closes the Redis connection
func (r *Redis) Close() error {
	return r.client.Close()
}

// Health returns the health status of Redis
func (r *Redis) Health(ctx context.Context) map[string]interface{} {
	health := map[string]interface{}{
		"status": "healthy",
	}

	// Check if we can ping Redis
	if err := r.Ping(ctx).Err(); err != nil {
		health["status"] = "unhealthy"
		health["error"] = err.Error()
		return health
	}

	// Get Redis info
	info, err := r.client.Info(ctx).Result()
	if err != nil {
		health["warning"] = "could not get Redis info"
	} else {
		health["info"] = info
	}

	return health
}