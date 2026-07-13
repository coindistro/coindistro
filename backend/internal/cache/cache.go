package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/config"
)

// Cache wraps the Redis client.
type Cache struct {
	Client *redis.Client
	config config.RedisConfig
	logger *zap.Logger
}

// New creates a new Redis cache connection.
func New(cfg config.RedisConfig, logger *zap.Logger) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  time.Duration(cfg.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	logger.Info("redis connection established",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.Int("db", cfg.DB),
	)

	return &Cache{
		Client: client,
		config: cfg,
		logger: logger,
	}, nil
}

// Ping checks if Redis is reachable.
func (c *Cache) Ping(ctx context.Context) error {
	return c.Client.Ping(ctx).Err()
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
