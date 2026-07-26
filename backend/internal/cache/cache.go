package cache

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/config"
)

// Cache wraps the Redis client.
type Cache struct {
	Client  *redis.Client
	config  config.RedisConfig
	logger  *zap.Logger
	status  string
	lastErr error
}

// New creates a new Redis cache connection.
func New(cfg config.RedisConfig, logger *zap.Logger) (*Cache, error) {
	if !cfg.IsConfigured() {
		return nil, fmt.Errorf("redis is not configured")
	}

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		TLSConfig:    cfg.TLSConfig(),
		DialTimeout:  time.Duration(cfg.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	cache := &Cache{
		Client: client,
		config: cfg,
		logger: logger,
		status: "not_configured",
	}

	logger.Info("Redis configuration",
		zap.String("redis_host", cfg.Host),
		zap.Int("redis_port", cfg.Port),
		zap.Bool("tls_enabled", cfg.TLSEnabled),
		zap.Bool("password_present", cfg.Password != ""),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		cache.setError(err)
		logger.Error("redis ping failed",
			zap.String("host", cfg.Host),
			zap.Int("port", cfg.Port),
			zap.Bool("tls_enabled", cfg.TLSEnabled),
			zap.Bool("password_present", cfg.Password != ""),
			zap.String("reason", cache.status),
			zap.Error(err),
		)
		return cache, nil
	}

	cache.status = "healthy"
	logger.Info("redis connection established",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.Int("db", cfg.DB),
	)

	return cache, nil
}

func (c *Cache) setError(err error) {
	c.lastErr = err
	c.status = classifyRedisError(err)
}

func classifyRedisError(err error) string {
	if err == nil {
		return "healthy"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "timeout"
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "auth") || strings.Contains(msg, "password") || strings.Contains(msg, "noauth") || strings.Contains(msg, "wrongpass"):
		return "authentication_failed"
	case strings.Contains(msg, "handshake") || strings.Contains(msg, "tls"):
		return "tls_handshake_failed"
	case strings.Contains(msg, "connection refused") || strings.Contains(msg, "connect: connection refused"):
		return "connection_refused"
	case strings.Contains(msg, "no such host") || strings.Contains(msg, "lookup") || strings.Contains(msg, "nodename"):
		return "dns_lookup_failed"
	case strings.Contains(msg, "eof"):
		return "eof"
	case strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline"):
		return "timeout"
	default:
		return "unreachable"
	}
}

func (c *Cache) HealthStatus() string {
	if c == nil {
		return "not_configured"
	}
	if c.status == "" {
		return "not_configured"
	}
	return c.status
}

func (c *Cache) LastError() error {
	if c == nil {
		return nil
	}
	return c.lastErr
}

func (c *Cache) DiagnosticSummary() map[string]interface{} {
	if c == nil {
		return map[string]interface{}{
			"connected": false,
			"status":    "not_configured",
		}
	}
	return map[string]interface{}{
		"connected":    c.status == "healthy",
		"status":       c.status,
		"host":         c.config.Host,
		"port":         c.config.Port,
		"tls":          c.config.TLSEnabled,
		"password_set": c.config.Password != "",
	}
}

// Ping checks if Redis is reachable.
func (c *Cache) Ping(ctx context.Context) error {
	err := c.Client.Ping(ctx).Err()
	if err != nil {
		c.setError(err)
		return err
	}
	c.status = "healthy"
	c.lastErr = nil
	return nil
}

// Close closes the Redis connection.
func (c *Cache) Close() {
	if c.Client != nil {
		if err := c.Client.Close(); err != nil {
			c.logger.Error("failed to close redis connection", zap.Error(err))
		} else {
			c.logger.Info("redis connection closed")
		}
	}
}

// Set stores a value in Redis with an expiration time.
func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.Client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value from Redis.
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	return c.Client.Get(ctx, key).Result()
}

// Del deletes one or more keys from Redis.
func (c *Cache) Del(ctx context.Context, keys ...string) error {
	return c.Client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists in Redis.
func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// Expire sets a timeout on a key.
func (c *Cache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.Client.Expire(ctx, key, expiration).Err()
}

// Incr increments the integer value of a key by one.
func (c *Cache) Incr(ctx context.Context, key string) (int64, error) {
	return c.Client.Incr(ctx, key).Result()
}

// TTL returns the remaining time to live of a key.
func (c *Cache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.Client.TTL(ctx, key).Result()
}

// FlushAll removes all keys from the current database.
func (c *Cache) FlushAll(ctx context.Context) error {
	return c.Client.FlushAll(ctx).Err()
}
