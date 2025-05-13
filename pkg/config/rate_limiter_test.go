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

func (m *MockRedisClient) Incr(ctx context.Context, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestSetupRateLimiter(t *testing.T) {
	// Create a mock Redis client
	mockRedis := new(MockRedisClient)

	// Test parameters
	limit := 10
	window := 1 * time.Minute

	t.Run("successful rate limiter setup", func(t *testing.T) {
		// Setup rate limiter
		limiter := SetupRateLimiter(mockRedis, limit, window)

		// Verify rate limiter is not nil
		assert.NotNil(t, limiter, "Rate limiter should not be nil")

		// Verify rate limiter type
		_, ok := limiter.(*utils.RateLimiter)
		assert.True(t, ok, "Rate limiter should be of type *utils.RateLimiter")
	})

	t.Run("rate limiter with nil redis client", func(t *testing.T) {
		// Setup rate limiter with nil Redis client
		limiter := SetupRateLimiter(nil, limit, window)

		// Verify rate limiter is not nil
		assert.NotNil(t, limiter, "Rate limiter should not be nil even with nil Redis client")

		// Verify rate limiter type
		_, ok := limiter.(*utils.RateLimiter)
		assert.True(t, ok, "Rate limiter should be of type *utils.RateLimiter even with nil Redis client")
	})

	t.Run("rate limiter with zero limit", func(t *testing.T) {
		// Setup rate limiter with zero limit
		limiter := SetupRateLimiter(mockRedis, 0, window)

		// Verify rate limiter is not nil
		assert.NotNil(t, limiter, "Rate limiter should not be nil with zero limit")

		// Verify rate limiter type
		_, ok := limiter.(*utils.RateLimiter)
		assert.True(t, ok, "Rate limiter should be of type *utils.RateLimiter with zero limit")
	})

	t.Run("rate limiter with zero window", func(t *testing.T) {
		// Setup rate limiter with zero window
		limiter := SetupRateLimiter(mockRedis, limit, 0)

		// Verify rate limiter is not nil
		assert.NotNil(t, limiter, "Rate limiter should not be nil with zero window")

		// Verify rate limiter type
		_, ok := limiter.(*utils.RateLimiter)
		assert.True(t, ok, "Rate limiter should be of type *utils.RateLimiter with zero window")
	})
}

func TestRateLimiterIntegration(t *testing.T) {
	// Create a mock Redis client
	mockRedis := new(MockRedisClient)

	// Test parameters
	limit := 10
	window := 1 * time.Minute
	key := "test-key"

	// Setup rate limiter
	limiter := SetupRateLimiter(mockRedis, limit, window)

	ctx := context.Background()

	t.Run("rate limiter operations", func(t *testing.T) {
		// Mock Redis operations for successful rate limiting
		mockRedis.On("Incr", ctx, key).Return(redis.NewIntCmd(ctx, 1))
		mockRedis.On("Expire", ctx, key, window).Return(redis.NewBoolCmd(ctx, true))

		// Test Allow operation
		allowed, err := limiter.Allow(ctx, key)
		assert.NoError(t, err, "Allow operation should not return error")
		assert.True(t, allowed, "Request should be allowed within limit")

		// Verify Redis operations were called
		mockRedis.AssertExpectations(t)
	})

	t.Run("rate limiter exceeded limit", func(t *testing.T) {
		// Mock Redis operations for exceeded limit
		mockRedis.On("Incr", ctx, key).Return(redis.NewIntCmd(ctx, int64(limit+1)))
		mockRedis.On("Expire", ctx, key, window).Return(redis.NewBoolCmd(ctx, true))

		// Test Allow operation
		allowed, err := limiter.Allow(ctx, key)
		assert.NoError(t, err, "Allow operation should not return error")
		assert.False(t, allowed, "Request should not be allowed when limit is exceeded")

		// Verify Redis operations were called
		mockRedis.AssertExpectations(t)
	})
}

func TestRateLimiterErrorHandling(t *testing.T) {
	// Create a mock Redis client
	mockRedis := new(MockRedisClient)

	// Test parameters
	limit := 10
	window := 1 * time.Minute
	key := "test-key"

	// Setup rate limiter
	limiter := SetupRateLimiter(mockRedis, limit, window)

	ctx := context.Background()

	t.Run("redis incr error", func(t *testing.T) {
		// Mock Redis Incr operation to return error
		mockRedis.On("Incr", ctx, key).Return(redis.NewIntCmd(ctx, 0))

		// Test Allow operation
		allowed, err := limiter.Allow(ctx, key)
		assert.Error(t, err, "Allow operation should return error when Redis Incr fails")
		assert.False(t, allowed, "Request should not be allowed when Redis Incr fails")

		// Verify Redis operations were called
		mockRedis.AssertExpectations(t)
	})

	t.Run("redis expire error", func(t *testing.T) {
		// Mock Redis operations
		mockRedis.On("Incr", ctx, key).Return(redis.NewIntCmd(ctx, 1))
		mockRedis.On("Expire", ctx, key, window).Return(redis.NewBoolCmd(ctx, false))

		// Test Allow operation
		allowed, err := limiter.Allow(ctx, key)
		assert.Error(t, err, "Allow operation should return error when Redis Expire fails")
		assert.False(t, allowed, "Request should not be allowed when Redis Expire fails")

		// Verify Redis operations were called
		mockRedis.AssertExpectations(t)
	})
} 