package utils

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient is a mock implementation of redis.Client
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Incr(ctx context.Context, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func TestNewRateLimiter(t *testing.T) {
	mockClient := new(MockRedisClient)
	limiter := NewRateLimiter(mockClient)

	assert.NotNil(t, limiter)
	assert.Equal(t, mockClient, limiter.RedisClient)
	assert.Equal(t, time.Minute, limiter.Window)
	assert.Equal(t, 100, limiter.operationLimits[OperationRead])
	assert.Equal(t, 30, limiter.operationLimits[OperationWrite])
	assert.Equal(t, 200, limiter.operationLimits[OperationPrefetch])
}

func TestRateLimiter_SetLimitForOperation(t *testing.T) {
	mockClient := new(MockRedisClient)
	limiter := NewRateLimiter(mockClient)

	// Test setting new limit
	limiter.SetLimitForOperation(OperationRead, 150)
	assert.Equal(t, 150, limiter.operationLimits[OperationRead])

	// Test setting limit for new operation type
	limiter.SetLimitForOperation("custom", 75)
	assert.Equal(t, 75, limiter.operationLimits["custom"])
}

func TestRateLimiter_GetLimitForOperation(t *testing.T) {
	mockClient := new(MockRedisClient)
	limiter := NewRateLimiter(mockClient)

	// Test getting existing limit
	assert.Equal(t, 100, limiter.GetLimitForOperation(OperationRead))

	// Test getting non-existent limit (should return read limit as default)
	assert.Equal(t, 100, limiter.GetLimitForOperation("non-existent"))
}

func TestRateLimiter_Allow(t *testing.T) {
	tests := []struct {
		name          string
		clientID      string
		op            OperationType
		setupMock     func(*MockRedisClient)
		expectedAllow bool
		expectedLimit int64
		expectedError bool
	}{
		{
			name:     "read operation within limit",
			clientID: "test-client",
			op:       OperationRead,
			setupMock: func(m *MockRedisClient) {
				incrCmd := redis.NewIntCmd(context.Background())
				incrCmd.SetVal(1)
				m.On("Incr", mock.Anything, "rate_limit:test-client:60").Return(incrCmd)

				expireCmd := redis.NewBoolCmd(context.Background())
				expireCmd.SetVal(true)
				m.On("Expire", mock.Anything, "rate_limit:test-client:60", time.Duration(60)*time.Second).Return(expireCmd)
			},
			expectedAllow: true,
			expectedLimit: 100,
			expectedError: false,
		},
		{
			name:     "write operation within limit",
			clientID: "test-client",
			op:       OperationWrite,
			setupMock: func(m *MockRedisClient) {
				incrCmd := redis.NewIntCmd(context.Background())
				incrCmd.SetVal(1)
				m.On("Incr", mock.Anything, "rate_limit:test-client:60").Return(incrCmd)

				expireCmd := redis.NewBoolCmd(context.Background())
				expireCmd.SetVal(true)
				m.On("Expire", mock.Anything, "rate_limit:test-client:60", time.Duration(60)*time.Second).Return(expireCmd)
			},
			expectedAllow: true,
			expectedLimit: 50,
			expectedError: false,
		},
		{
			name:     "prefetch operation within limit",
			clientID: "test-client",
			op:       OperationPrefetch,
			setupMock: func(m *MockRedisClient) {
				incrCmd := redis.NewIntCmd(context.Background())
				incrCmd.SetVal(1)
				m.On("Incr", mock.Anything, "rate_limit:test-client:60").Return(incrCmd)

				expireCmd := redis.NewBoolCmd(context.Background())
				expireCmd.SetVal(true)
				m.On("Expire", mock.Anything, "rate_limit:test-client:60", time.Duration(60)*time.Second).Return(expireCmd)
			},
			expectedAllow: true,
			expectedLimit: 200,
			expectedError: false,
		},
		{
			name:     "exceeded limit",
			clientID: "test-client",
			op:       OperationRead,
			setupMock: func(m *MockRedisClient) {
				incrCmd := redis.NewIntCmd(context.Background())
				incrCmd.SetVal(101)
				m.On("Incr", mock.Anything, "rate_limit:test-client:60").Return(incrCmd)
			},
			expectedAllow: false,
			expectedLimit: 100,
			expectedError: false,
		},
		{
			name:     "unknown operation type",
			clientID: "test-client",
			op:       "unknown",
			setupMock: func(m *MockRedisClient) {
				// No mock setup needed
			},
			expectedAllow: false,
			expectedLimit: 0,
			expectedError: true,
		},
		{
			name:     "redis error",
			clientID: "test-client",
			op:       OperationRead,
			setupMock: func(m *MockRedisClient) {
				incrCmd := redis.NewIntCmd(context.Background())
				incrCmd.SetErr(redis.ErrClosed)
				m.On("Incr", mock.Anything, "rate_limit:test-client:60").Return(incrCmd)
			},
			expectedAllow: false,
			expectedLimit: 0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockRedisClient)
			tt.setupMock(mockClient)
			
			limiter := NewRateLimiter(mockClient)
			allowed, limit, retryTime, err := limiter.Allow(tt.clientID, tt.op)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedAllow, allowed)
			assert.Equal(t, tt.expectedLimit, limit)
			if !tt.expectedAllow && !tt.expectedError {
				assert.NotZero(t, retryTime)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestRateLimiter_AllowOperation(t *testing.T) {
	tests := []struct {
		name          string
		identifier    string
		config        *RateLimitConfig
		setupMock     func(*MockRedisClient)
		expectedAllow bool
		expectedLimit int64
		expectedError bool
	}{
		{
			name:       "valid config within limit",
			identifier: "test-client",
			config: &RateLimitConfig{
				MaxRequests: 50,
				Window:      60,
			},
			setupMock: func(m *MockRedisClient) {
				incrCmd := redis.NewIntCmd(context.Background())
				incrCmd.SetVal(1)
				m.On("Incr", mock.Anything, "rate_limit:test-client:60").Return(incrCmd)

				expireCmd := redis.NewBoolCmd(context.Background())
				expireCmd.SetVal(true)
				m.On("Expire", mock.Anything, "rate_limit:test-client:60", time.Duration(60)*time.Second).Return(expireCmd)
			},
			expectedAllow: true,
			expectedLimit: 50,
			expectedError: false,
		},
		{
			name:       "nil config",
			identifier: "test-client",
			config:     nil,
			setupMock: func(m *MockRedisClient) {
				// No mock setup needed
			},
			expectedAllow: false,
			expectedLimit: 0,
			expectedError: true,
		},
		{
			name:       "invalid max requests",
			identifier: "test-client",
			config: &RateLimitConfig{
				MaxRequests: 0,
				Window:      60,
			},
			setupMock: func(m *MockRedisClient) {
				// No mock setup needed
			},
			expectedAllow: false,
			expectedLimit: 0,
			expectedError: true,
		},
		{
			name:       "invalid window",
			identifier: "test-client",
			config: &RateLimitConfig{
				MaxRequests: 50,
				Window:      0,
			},
			setupMock: func(m *MockRedisClient) {
				// No mock setup needed
			},
			expectedAllow: false,
			expectedLimit: 0,
			expectedError: true,
		},
		{
			name:       "exceeded limit",
			identifier: "test-client",
			config: &RateLimitConfig{
				MaxRequests: 50,
				Window:      60,
			},
			setupMock: func(m *MockRedisClient) {
				incrCmd := redis.NewIntCmd(context.Background())
				incrCmd.SetVal(51)
				m.On("Incr", mock.Anything, "rate_limit:test-client:60").Return(incrCmd)
			},
			expectedAllow: false,
			expectedLimit: 50,
			expectedError: false,
		},
		{
			name:       "redis error",
			identifier: "test-client",
			config: &RateLimitConfig{
				MaxRequests: 50,
				Window:      60,
			},
			setupMock: func(m *MockRedisClient) {
				incrCmd := redis.NewIntCmd(context.Background())
				incrCmd.SetErr(redis.ErrClosed)
				m.On("Incr", mock.Anything, "rate_limit:test-client:60").Return(incrCmd)
			},
			expectedAllow: false,
			expectedLimit: 0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockRedisClient)
			tt.setupMock(mockClient)
			
			limiter := NewRateLimiter(mockClient)
			allowed, limit, retryTime, err := limiter.AllowOperation(tt.identifier, tt.config)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedAllow, allowed)
			assert.Equal(t, tt.expectedLimit, limit)
			if !tt.expectedAllow && !tt.expectedError {
				assert.NotZero(t, retryTime)
			}

			mockClient.AssertExpectations(t)
		})
	}
} 