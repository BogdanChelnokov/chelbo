package redis

import (
	"context"
	"fmt"
	"time"

	"chelbo/backend/internal/pkg/config"
	"chelbo/backend/internal/pkg/logger"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client
var Ctx = context.Background()

func Init(cfg *config.RedisConfig) error {
	Client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: 100,
	})

	// Test connection
	_, err := Client.Ping(Ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("✅ Redis connected successfully")
	return nil
}

func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}

// Set stores a key-value pair with expiration
func Set(key string, value interface{}, expiration time.Duration) error {
	return Client.Set(Ctx, key, value, expiration).Err()
}

// Get retrieves a value by key
func Get(key string) (string, error) {
	return Client.Get(Ctx, key).Result()
}

// Delete removes a key
func Delete(key string) error {
	return Client.Del(Ctx, key).Err()
}

// Exists checks if a key exists
func Exists(key string) (bool, error) {
	result, err := Client.Exists(Ctx, key).Result()
	return result > 0, err
}

// Publish publishes a message to a channel
func Publish(channel string, message interface{}) error {
	return Client.Publish(Ctx, channel, message).Err()
}

// Subscribe subscribes to a channel
func Subscribe(channel string) *redis.PubSub {
	return Client.Subscribe(Ctx, channel)
}
