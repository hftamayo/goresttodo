package config

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient is now defined in mock_redis.go

// MockRedisOptions is a mock implementation of redis.Options
type MockRedisOptions struct {
	Addr string
}

func TestErrorLogConnect(t *testing.T) {
	// Skip this test if Redis is not available
	t.Skip("Skipping Redis connection test - requires Redis server")

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
	// Skip this test if Redis is not available
	t.Skip("Skipping Redis connection test - requires Redis server")

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
		// Skip this test if Redis is not available
		t.Skip("Skipping Redis connection test - requires Redis server")

		// Set test environment variables
		os.Setenv("REDIS_HOST", "localhost")
		os.Setenv("REDIS_PORT", "6379")

		// Test connection
		client, err := ErrorLogConnect()
		assert.NoError(t, err, "Connection should succeed")
		assert.NotNil(t, client, "Redis client should not be nil")

		// Test that we can perform operations with the client
		ctx := context.Background()
		// Test a simple operation that the interface supports
		cmd := client.Set(ctx, "test-key", "test-value", 0)
		assert.NoError(t, cmd.Err(), "Set operation should succeed")
	})
}

func TestMemoryErrorLogger_LogErrorAndGetErrors(t *testing.T) {
	logger := NewMemoryErrorLogger()
	err := logger.LogError(context.Background(), "test-service", "test-op", "something went wrong", map[string]interface{}{"foo": "bar"})
	assert.NoError(t, err)
	errors := logger.GetErrors()
	assert.Len(t, errors, 1)
	assert.Equal(t, "test-service", errors[0]["service"])
	assert.Equal(t, "test-op", errors[0]["operation"])
	assert.Equal(t, "something went wrong", errors[0]["error"])
	assert.Equal(t, map[string]interface{}{"foo": "bar"}, errors[0]["metadata"])
}

func TestMemoryErrorLogger_ThreadSafety(t *testing.T) {
	logger := NewMemoryErrorLogger()
	done := make(chan bool)
	go func() {
		for i := 0; i < 100; i++ {
			logger.LogError(context.Background(), "svc", "op", "err", map[string]interface{}{"i": i})
		}
		done <- true
	}()
	go func() {
		for i := 0; i < 100; i++ {
			logger.GetErrors()
		}
		done <- true
	}()
	<-done
	<-done
	assert.True(t, len(logger.GetErrors()) > 0)
}

func TestNonBlockingErrorLogger_LogError(t *testing.T) {
	memLogger := NewMemoryErrorLogger()
	nbLogger := NewNonBlockingErrorLogger(memLogger)
	err := nbLogger.LogError(context.Background(), "svc", "op", "err", map[string]interface{}{"foo": "bar"})
	assert.NoError(t, err)
	// Wait a moment for goroutine to finish
	time.Sleep(10 * time.Millisecond)
	errors := memLogger.GetErrors()
	assert.Len(t, errors, 1)
	assert.Equal(t, "svc", errors[0]["service"])
} 