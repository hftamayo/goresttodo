package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/hftamayo/gotodo/pkg/utils"
)

// CacheInterface defines the contract for cache operations
// This interface matches what the task service expects
type CacheInterface interface {
	Get(key string, dest interface{}) error
	Set(key string, value interface{}, ttl time.Duration) error
	SetWithTags(key string, value interface{}, ttl time.Duration, tags ...string) error
	Delete(key string) error
	InvalidateByTags(tags ...string) error
}

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

// NewCache creates a new cache instance that implements CacheInterface
// This is the preferred way to create cache instances for the task service
func NewCache(config *CacheConfig) (CacheInterface, error) {
	cache, err := SetupCache(config)
	if err != nil {
		return nil, err
	}
	return cache, nil
}

// NewCacheWithDefaults creates a cache instance with default configuration
// This is the preferred way to create cache instances for the task service
func NewCacheWithDefaults() (CacheInterface, error) {
	config := DefaultCacheConfig()
	return NewCache(config)
}

// NewMemoryCache creates a memory-based cache for testing
// This implements CacheInterface but stores data in memory instead of Redis
func NewMemoryCache() CacheInterface {
	return &MemoryCache{
		data: make(map[string]interface{}),
		ttl:  make(map[string]time.Time),
	}
}

// MemoryCache implements CacheInterface using in-memory storage
// This is useful for testing and development without Redis
type MemoryCache struct {
	data map[string]interface{}
	ttl  map[string]time.Time
}

func (m *MemoryCache) Get(key string, dest interface{}) error {
	// Check if key exists and is not expired
	if value, exists := m.data[key]; exists {
		if ttl, hasTTL := m.ttl[key]; hasTTL && time.Now().After(ttl) {
			// Key has expired, remove it
			delete(m.data, key)
			delete(m.ttl, key)
			return fmt.Errorf("key not found or expired")
		}
		
		// For simplicity, we'll just copy the value
		// In a real implementation, you'd want proper serialization
		if destPtr, ok := dest.(*interface{}); ok {
			*destPtr = value
			return nil
		}
		return fmt.Errorf("destination type not supported")
	}
	return fmt.Errorf("key not found")
}

func (m *MemoryCache) Set(key string, value interface{}, ttl time.Duration) error {
	m.data[key] = value
	if ttl > 0 {
		m.ttl[key] = time.Now().Add(ttl)
	}
	return nil
}

func (m *MemoryCache) SetWithTags(key string, value interface{}, ttl time.Duration, tags ...string) error {
	// For memory cache, we'll just set the value
	// Tag management would require additional complexity
	return m.Set(key, value, ttl)
}

func (m *MemoryCache) Delete(key string) error {
	delete(m.data, key)
	delete(m.ttl, key)
	return nil
}

func (m *MemoryCache) InvalidateByTags(tags ...string) error {
	// For memory cache, we'll just clear all data
	// In a real implementation, you'd want proper tag management
	m.data = make(map[string]interface{})
	m.ttl = make(map[string]time.Time)
	return nil
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
