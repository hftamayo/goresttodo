package config

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/hftamayo/gotodo/pkg/utils"
)

// MockRedisClient is a mock implementation of redis.Client
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestSetupCache(t *testing.T) {
	// Create a mock Redis client
	mockRedis := new(MockRedisClient)

	// Test cache setup
	t.Run("successful cache setup", func(t *testing.T) {
		// Setup cache
		cache := SetupCache(mockRedis)

		// Verify cache is not nil
		assert.NotNil(t, cache, "Cache should not be nil")

		// Verify cache type
		_, ok := cache.(*utils.Cache)
		assert.True(t, ok, "Cache should be of type *utils.Cache")
	})

	t.Run("cache with nil redis client", func(t *testing.T) {
		// Setup cache with nil Redis client
		cache := SetupCache(nil)

		// Verify cache is not nil (as the function should handle nil client)
		assert.NotNil(t, cache, "Cache should not be nil even with nil Redis client")

		// Verify cache type
		_, ok := cache.(*utils.Cache)
		assert.True(t, ok, "Cache should be of type *utils.Cache even with nil Redis client")
	})
}

func TestCacheIntegration(t *testing.T) {
	// Create a mock Redis client
	mockRedis := new(MockRedisClient)

	// Setup expectations for Redis operations
	ctx := context.Background()
	key := "test-key"
	value := "test-value"
	expiration := 5 * time.Minute

	// Mock Get operation
	mockRedis.On("Get", ctx, key).Return(redis.NewStringCmd(ctx))

	// Mock Set operation
	mockRedis.On("Set", ctx, key, value, expiration).Return(redis.NewStatusCmd(ctx))

	// Mock Del operation
	mockRedis.On("Del", ctx, []string{key}).Return(redis.NewIntCmd(ctx))

	// Setup cache
	cache := SetupCache(mockRedis)

	t.Run("cache operations", func(t *testing.T) {
		// Test Set operation
		err := cache.Set(ctx, key, value, expiration)
		assert.NoError(t, err, "Set operation should not return error")

		// Test Get operation
		_, err = cache.Get(ctx, key)
		assert.NoError(t, err, "Get operation should not return error")

		// Test Delete operation
		err = cache.Delete(ctx, key)
		assert.NoError(t, err, "Delete operation should not return error")
	})

	// Verify all expected Redis operations were called
	mockRedis.AssertExpectations(t)
}

func TestCacheErrorHandling(t *testing.T) {
	// Create a mock Redis client
	mockRedis := new(MockRedisClient)

	// Setup cache
	cache := SetupCache(mockRedis)

	ctx := context.Background()
	key := "test-key"

	t.Run("get non-existent key", func(t *testing.T) {
		// Mock Get operation to return error
		mockRedis.On("Get", ctx, key).Return(redis.NewStringCmd(ctx))

		// Test Get operation
		_, err := cache.Get(ctx, key)
		assert.Error(t, err, "Get operation should return error for non-existent key")
	})

	t.Run("delete non-existent key", func(t *testing.T) {
		// Mock Del operation to return error
		mockRedis.On("Del", ctx, []string{key}).Return(redis.NewIntCmd(ctx))

		// Test Delete operation
		err := cache.Delete(ctx, key)
		assert.Error(t, err, "Delete operation should return error for non-existent key")
	})

	// Verify all expected Redis operations were called
	mockRedis.AssertExpectations(t)
} 