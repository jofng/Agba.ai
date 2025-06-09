package gateway

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RateLimiter handles rate limiting using Redis
type RateLimiter struct {
	config *RateLimiterConfig
	redis  *redis.Client
	logger *zap.Logger
	stop   chan struct{}
}

// RateLimiterConfig holds rate limiter configuration
type RateLimiterConfig struct {
	RedisClient     *redis.Client
	DefaultRPM      int
	DefaultRPH      int
	BurstMultiplier int
	CleanupInterval time.Duration
	KeyPrefix       string
	SlidingWindow   bool
	BlockDuration   time.Duration
	Logger          *zap.Logger
}

// RateLimitResult represents the result of a rate limit check
type RateLimitResult struct {
	Allowed       bool
	Limit         int
	Remaining     int
	ResetTime     time.Time
	RetryAfter    time.Duration
	IsBlocked     bool
	BlockedUntil  time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *RateLimiterConfig) (*RateLimiter, error) {
	rl := &RateLimiter{
		config: config,
		redis:  config.RedisClient,
		logger: config.Logger,
		stop:   make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl, nil
}

// CheckLimit checks if a request is within rate limits
func (rl *RateLimiter) CheckLimit(ctx context.Context, key string, limit, window int) (*RateLimitResult, error) {
	now := time.Now()
	windowStart := now.Truncate(time.Duration(window) * time.Second)
	
	// Check if key is blocked
	blockKey := rl.config.KeyPrefix + "block:" + key
	blocked, err := rl.redis.Get(ctx, blockKey).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to check block status: %w", err)
	}

	if blocked != "" {
		blockedUntil, _ := time.Parse(time.RFC3339, blocked)
		if now.Before(blockedUntil) {
			return &RateLimitResult{
				Allowed:      false,
				Limit:        limit,
				Remaining:    0,
				ResetTime:    windowStart.Add(time.Duration(window) * time.Second),
				RetryAfter:   blockedUntil.Sub(now),
				IsBlocked:    true,
				BlockedUntil: blockedUntil,
			}, nil
		}
		// Block expired, remove it
		rl.redis.Del(ctx, blockKey)
	}

	if rl.config.SlidingWindow {
		return rl.checkSlidingWindow(ctx, key, limit, window, now)
	}
	return rl.checkFixedWindow(ctx, key, limit, window, windowStart)
}

// checkSlidingWindow implements sliding window rate limiting
func (rl *RateLimiter) checkSlidingWindow(ctx context.Context, key string, limit, window int, now time.Time) (*RateLimitResult, error) {
	redisKey := rl.config.KeyPrefix + key
	windowStart := now.Add(-time.Duration(window) * time.Second)

	pipe := rl.redis.Pipeline()
	
	// Remove old entries
	pipe.ZRemRangeByScore(ctx, redisKey, "0", strconv.FormatInt(windowStart.Unix(), 10))
	
	// Count current entries
	countCmd := pipe.ZCard(ctx, redisKey)
	
	// Add current request
	pipe.ZAdd(ctx, redisKey, &redis.Z{
		Score:  float64(now.Unix()),
		Member: fmt.Sprintf("%d", now.UnixNano()),
	})
	
	// Set expiration
	pipe.Expire(ctx, redisKey, time.Duration(window)*time.Second)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute sliding window check: %w", err)
	}

	count := int(countCmd.Val())
	remaining := limit - count - 1
	if remaining < 0 {
		remaining = 0
	}

	allowed := count < limit
	if !allowed {
		// Block the key if it exceeds burst limit
		burstLimit := limit * rl.config.BurstMultiplier
		if count >= burstLimit {
			rl.blockKey(ctx, key, now.Add(rl.config.BlockDuration))
		}
	}

	return &RateLimitResult{
		Allowed:   allowed,
		Limit:     limit,
		Remaining: remaining,
		ResetTime: now.Add(time.Duration(window) * time.Second),
	}, nil
}

// checkFixedWindow implements fixed window rate limiting
func (rl *RateLimiter) checkFixedWindow(ctx context.Context, key string, limit, window int, windowStart time.Time) (*RateLimitResult, error) {
	redisKey := rl.config.KeyPrefix + key + ":" + strconv.FormatInt(windowStart.Unix(), 10)
	
	pipe := rl.redis.Pipeline()
	incrCmd := pipe.Incr(ctx, redisKey)
	pipe.Expire(ctx, redisKey, time.Duration(window)*time.Second)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute fixed window check: %w", err)
	}

	count := int(incrCmd.Val())
	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	allowed := count <= limit
	if !allowed {
		// Block the key if it exceeds burst limit
		burstLimit := limit * rl.config.BurstMultiplier
		if count >= burstLimit {
			rl.blockKey(ctx, key, time.Now().Add(rl.config.BlockDuration))
		}
	}

	return &RateLimitResult{
		Allowed:   allowed,
		Limit:     limit,
		Remaining: remaining,
		ResetTime: windowStart.Add(time.Duration(window) * time.Second),
	}, nil
}

// blockKey blocks a key for a specified duration
func (rl *RateLimiter) blockKey(ctx context.Context, key string, until time.Time) {
	blockKey := rl.config.KeyPrefix + "block:" + key
	rl.redis.Set(ctx, blockKey, until.Format(time.RFC3339), rl.config.BlockDuration)
	
	rl.logger.Warn("Rate limit exceeded, blocking key",
		zap.String("key", key),
		zap.Time("blocked_until", until),
	)
}

// CheckAPIKey checks rate limits for an API key
func (rl *RateLimiter) CheckAPIKey(ctx context.Context, apiKey string, customLimits map[string]int) (*RateLimitResult, error) {
	// Get limits for this API key (from custom limits or defaults)
	rpm := rl.config.DefaultRPM
	if customRPM, exists := customLimits["rpm"]; exists {
		rpm = customRPM
	}

	return rl.CheckLimit(ctx, "api:"+apiKey, rpm, 60)
}

// CheckUser checks rate limits for a user
func (rl *RateLimiter) CheckUser(ctx context.Context, userID string) (*RateLimitResult, error) {
	return rl.CheckLimit(ctx, "user:"+userID, rl.config.DefaultRPM, 60)
}

// CheckIP checks rate limits for an IP address
func (rl *RateLimiter) CheckIP(ctx context.Context, ip string) (*RateLimitResult, error) {
	// More restrictive limits for IP-based rate limiting
	return rl.CheckLimit(ctx, "ip:"+ip, rl.config.DefaultRPM/10, 60)
}

// cleanup periodically cleans up expired rate limit data
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.performCleanup()
		case <-rl.stop:
			return
		}
	}
}

// performCleanup removes expired rate limit data
func (rl *RateLimiter) performCleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	now := time.Now()
	cutoff := now.Add(-1 * time.Hour) // Clean up data older than 1 hour

	// Find keys to clean up
	pattern := rl.config.KeyPrefix + "*"
	keys, err := rl.redis.Keys(ctx, pattern).Result()
	if err != nil {
		rl.logger.Error("Failed to get keys for cleanup", zap.Error(err))
		return
	}

	cleaned := 0
	for _, key := range keys {
		// For sliding window, remove old entries
		if rl.config.SlidingWindow {
			removed, err := rl.redis.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(cutoff.Unix(), 10)).Result()
			if err != nil {
				continue
			}
			if removed > 0 {
				cleaned++
			}
		}
	}

	if cleaned > 0 {
		rl.logger.Debug("Rate limiter cleanup completed",
			zap.Int("keys_cleaned", cleaned),
		)
	}
}

// Shutdown stops the rate limiter
func (rl *RateLimiter) Shutdown() {
	close(rl.stop)
}