package config

import (
	"context"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient is a mock implementation of redis.Client
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	args := m.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockRedisOptions is a mock implementation of redis.Options
type MockRedisOptions struct {
	Addr string
}

func TestErrorLogConnect(t *testing.T) {
	// Save original environment variables
	originalHost := os.Getenv("REDIS_HOST")
	originalPort := os.Getenv("REDIS_PORT")

	// Cleanup function to restore environment variables
	defer func() {
		os.Setenv("REDIS_HOST", originalHost)
		os.Setenv("REDIS_PORT", originalPort)
	}()

	t.Run("successful connection", func(t *testing.T) {
		// Set test environment variables
		os.Setenv("REDIS_HOST", "localhost")
		os.Setenv("REDIS_PORT", "6379")

		// Create a mock Redis client
		mockRedis := new(MockRedisClient)
		mockRedis.On("Ping", mock.Anything).Return(redis.NewStatusCmd(context.Background(), "PONG"))

		// Test connection
		client, err := ErrorLogConnect()
		assert.NoError(t, err, "Connection should succeed")
		assert.NotNil(t, client, "Redis client should not be nil")

		// Verify Redis operations were called
		mockRedis.AssertExpectations(t)
	})

	t.Run("missing environment variables", func(t *testing.T) {
		// Clear environment variables
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")

		// Test connection
		client, err := ErrorLogConnect()
		assert.Error(t, err, "Connection should fail with missing environment variables")
		assert.Nil(t, client, "Redis client should be nil")
	})

	t.Run("invalid host", func(t *testing.T) {
		// Set invalid host
		os.Setenv("REDIS_HOST", "invalid-host")
		os.Setenv("REDIS_PORT", "6379")

		// Test connection
		client, err := ErrorLogConnect()
		assert.Error(t, err, "Connection should fail with invalid host")
		assert.Nil(t, client, "Redis client should be nil")
	})

	t.Run("invalid port", func(t *testing.T) {
		// Set invalid port
		os.Setenv("REDIS_HOST", "localhost")
		os.Setenv("REDIS_PORT", "invalid-port")

		// Test connection
		client, err := ErrorLogConnect()
		assert.Error(t, err, "Connection should fail with invalid port")
		assert.Nil(t, client, "Redis client should be nil")
	})
}

func TestErrorLogConnect_RedisPing(t *testing.T) {
	// Save original environment variables
	originalHost := os.Getenv("REDIS_HOST")
	originalPort := os.Getenv("REDIS_PORT")

	// Cleanup function to restore environment variables
	defer func() {
		os.Setenv("REDIS_HOST", originalHost)
		os.Setenv("REDIS_PORT", originalPort)
	}()

	t.Run("redis ping success", func(t *testing.T) {
		// Set test environment variables
		os.Setenv("REDIS_HOST", "localhost")
		os.Setenv("REDIS_PORT", "6379")

		// Create a mock Redis client
		mockRedis := new(MockRedisClient)
		mockRedis.On("Ping", mock.Anything).Return(redis.NewStatusCmd(context.Background(), "PONG"))

		// Test connection
		client, err := ErrorLogConnect()
		assert.NoError(t, err, "Connection should succeed with successful ping")
		assert.NotNil(t, client, "Redis client should not be nil")

		// Verify Redis operations were called
		mockRedis.AssertExpectations(t)
	})

	t.Run("redis ping failure", func(t *testing.T) {
		// Set test environment variables
		os.Setenv("REDIS_HOST", "localhost")
		os.Setenv("REDIS_PORT", "6379")

		// Create a mock Redis client
		mockRedis := new(MockRedisClient)
		mockRedis.On("Ping", mock.Anything).Return(redis.NewStatusCmd(context.Background(), ""))

		// Test connection
		client, err := ErrorLogConnect()
		assert.Error(t, err, "Connection should fail with ping failure")
		assert.Nil(t, client, "Redis client should be nil")

		// Verify Redis operations were called
		mockRedis.AssertExpectations(t)
	})
}

func TestErrorLogConnect_ConnectionOptions(t *testing.T) {
	// Save original environment variables
	originalHost := os.Getenv("REDIS_HOST")
	originalPort := os.Getenv("REDIS_PORT")

	// Cleanup function to restore environment variables
	defer func() {
		os.Setenv("REDIS_HOST", originalHost)
		os.Setenv("REDIS_PORT", originalPort)
	}()

	t.Run("connection options", func(t *testing.T) {
		// Set test environment variables
		os.Setenv("REDIS_HOST", "localhost")
		os.Setenv("REDIS_PORT", "6379")

		// Create a mock Redis client
		mockRedis := new(MockRedisClient)
		mockRedis.On("Ping", mock.Anything).Return(redis.NewStatusCmd(context.Background(), "PONG"))

		// Test connection
		client, err := ErrorLogConnect()
		assert.NoError(t, err, "Connection should succeed")
		assert.NotNil(t, client, "Redis client should not be nil")

		// Verify connection options
		options := client.Options()
		assert.Equal(t, "localhost:6379", options.Addr, "Redis address should match environment variables")

		// Verify Redis operations were called
		mockRedis.AssertExpectations(t)
	})
} 