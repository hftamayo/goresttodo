package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/hftamayo/gotodo/pkg/utils"
)

func TestSetupCache(t *testing.T) {
	// Test cache setup
	t.Run("successful cache setup", func(t *testing.T) {
		// Skip this test if Redis is not available
		t.Skip("Skipping Redis connection test - requires Redis server")

		// Create a cache config
		config := &CacheConfig{
			Host:         "localhost",
			Port:         "6379",
			DB:           0,
			Password:     "",
			MaxRetries:   3,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolSize:     10,
			MinIdleConns: 5,
		}

		// Setup cache
		cache, err := SetupCache(config)

		// Verify cache is not nil and no error
		assert.NoError(t, err, "SetupCache should not return an error")
		assert.NotNil(t, cache, "Cache should not be nil")

		// Verify cache type
		assert.IsType(t, &utils.Cache{}, cache, "Cache should be of type *utils.Cache")
	})

	t.Run("cache with invalid config", func(t *testing.T) {
		// Setup cache with invalid config (nil)
		cache, err := SetupCache(nil)

		// Verify error is returned
		assert.Error(t, err, "SetupCache should return an error with nil config")
		assert.Nil(t, cache, "Cache should be nil when config is nil")
	})
}

func TestMemoryCache(t *testing.T) {
	// Create a memory cache
	cache := NewMemoryCache()

	// Test Set operation
	key := "test-key"
	value := "test-value"
	expiration := 5 * time.Minute

	err := cache.Set(key, value, expiration)
	assert.NoError(t, err, "Set should not return an error")

	// Test Get operation
	var retrievedValue string
	err = cache.Get(key, &retrievedValue)
	assert.NoError(t, err, "Get should not return an error")
	assert.Equal(t, value, retrievedValue, "Retrieved value should match set value")

	// Test Delete operation
	err = cache.Delete(key)
	assert.NoError(t, err, "Delete should not return an error")

	// Test Get after delete
	err = cache.Get(key, &retrievedValue)
	assert.Error(t, err, "Get after delete should return an error")
}

func TestDefaultCacheConfig(t *testing.T) {
	// Test default cache configuration
	config := DefaultCacheConfig()

	// Verify default values
	assert.Equal(t, "localhost", config.Host, "Default host should be localhost")
	assert.Equal(t, "6379", config.Port, "Default port should be 6379")
	assert.Equal(t, 0, config.DB, "Default DB should be 0")
	assert.Equal(t, "", config.Password, "Default password should be empty")
	assert.Equal(t, 3, config.MaxRetries, "Default max retries should be 3")
	assert.Equal(t, 5*time.Second, config.DialTimeout, "Default dial timeout should be 5 seconds")
	assert.Equal(t, 3*time.Second, config.ReadTimeout, "Default read timeout should be 3 seconds")
	assert.Equal(t, 3*time.Second, config.WriteTimeout, "Default write timeout should be 3 seconds")
	assert.Equal(t, 10, config.PoolSize, "Default pool size should be 10")
	assert.Equal(t, 5, config.MinIdleConns, "Default min idle connections should be 5")
} 