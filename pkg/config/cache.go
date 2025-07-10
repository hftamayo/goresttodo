package config

import (
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/hftamayo/gotodo/pkg/utils"
)

// CacheConfig holds cache configuration
type CacheConfig struct {
	Host            string
	Port            string
	DB              int
	Password        string
	MaxRetries      int
	MinRetryBackoff time.Duration
	MaxRetryBackoff time.Duration
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	PoolSize        int
	MinIdleConns    int
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		Host:            getEnvOrDefault("REDIS_HOST", "localhost"),
		Port:            getEnvOrDefault("REDIS_PORT", "6379"),
		DB:              getEnvAsIntOrDefault("REDIS_DB", 0),
		Password:        getEnvOrDefault("REDIS_PASSWORD", ""),
		MaxRetries:      getEnvAsIntOrDefault("REDIS_MAX_RETRIES", 3),
		MinRetryBackoff: getEnvAsDurationOrDefault("REDIS_MIN_RETRY_BACKOFF", 8*time.Millisecond),
		MaxRetryBackoff: getEnvAsDurationOrDefault("REDIS_MAX_RETRY_BACKOFF", 512*time.Millisecond),
		DialTimeout:     getEnvAsDurationOrDefault("REDIS_DIAL_TIMEOUT", 5*time.Second),
		ReadTimeout:     getEnvAsDurationOrDefault("REDIS_READ_TIMEOUT", 3*time.Second),
		WriteTimeout:    getEnvAsDurationOrDefault("REDIS_WRITE_TIMEOUT", 3*time.Second),
		PoolSize:        getEnvAsIntOrDefault("REDIS_POOL_SIZE", 10),
		MinIdleConns:    getEnvAsIntOrDefault("REDIS_MIN_IDLE_CONNS", 5),
	}
}

// SetupCache creates a new cache instance with configuration
func SetupCache(config *CacheConfig) (*utils.Cache, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:            config.Host + ":" + config.Port,
		DB:              config.DB,
		Password:        config.Password,
		MaxRetries:      config.MaxRetries,
		MinRetryBackoff: config.MinRetryBackoff,
		MaxRetryBackoff: config.MaxRetryBackoff,
		DialTimeout:     config.DialTimeout,
		ReadTimeout:     config.ReadTimeout,
		WriteTimeout:    config.WriteTimeout,
		PoolSize:        config.PoolSize,
		MinIdleConns:    config.MinIdleConns,
	})

	// Test the connection
	ctx := redisClient.Context()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return utils.NewCache(redisClient), nil
}

// SetupCacheWithDefaults creates a cache instance with default configuration
func SetupCacheWithDefaults() (*utils.Cache, error) {
	config := DefaultCacheConfig()
	return SetupCache(config)
}

// Legacy function for backward compatibility
func SetupCacheLegacy(redisClient *redis.Client) *utils.Cache {
	return utils.NewCache(redisClient)
}

// Helper functions for environment variable handling
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
