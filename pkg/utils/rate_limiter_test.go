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

func (m *MockRedisClient) TxPipeline() redis.Pipeliner {
	args := m.Called()
	return args.Get(0).(redis.Pipeliner)
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

// MockPipeliner is a mock implementation of redis.Pipeliner
type MockPipeliner struct {
	mock.Mock
}

func (m *MockPipeliner) Incr(ctx context.Context, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockPipeliner) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *MockPipeliner) Exec(ctx context.Context) ([]redis.Cmder, error) {
	args := m.Called(ctx)
	return args.Get(0).([]redis.Cmder), args.Error(1)
}

func TestNewRateLimiter(t *testing.T) {
	mockClient := new(MockRedisClient)
	limit := 10
	window := time.Minute

	limiter := NewRateLimiter(mockClient, limit, window)

	assert.NotNil(t, limiter)
	assert.Equal(t, mockClient, limiter.RedisClient)
	assert.Equal(t, limit, limiter.Limit)
	assert.Equal(t, window, limiter.Window)
}

func TestRateLimiter_Allow(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		limit         int
		window        time.Duration
		setupMock     func(*MockRedisClient, *MockPipeliner)
		expectedAllow bool
		expectedError bool
	}{
		{
			name:   "within limit",
			key:    "test-key",
			limit:  5,
			window: time.Minute,
			setupMock: func(m *MockRedisClient, p *MockPipeliner) {
				// Setup pipeline
				m.On("TxPipeline").Return(p)
				
				// Setup Incr command
				incrCmd := redis.NewIntCmd(context.Background())
				incrCmd.SetVal(1)
				p.On("Incr", mock.Anything, "test-key").Return(incrCmd)
				
				// Setup Expire command
				expireCmd := redis.NewBoolCmd(context.Background())
				expireCmd.SetVal(true)
				p.On("Expire", mock.Anything, "test-key", time.Minute).Return(expireCmd)
				
				// Setup Exec command
				p.On("Exec", mock.Anything).Return([]redis.Cmder{incrCmd, expireCmd}, nil)
				
				// Setup Get command
				getCmd := redis.NewStringCmd(context.Background())
				getCmd.SetVal("1")
				m.On("Get", mock.Anything, "test-key").Return(getCmd)
			},
			expectedAllow: true,
			expectedError: false,
		},
		{
			name:   "exceeded limit",
			key:    "test-key",
			limit:  5,
			window: time.Minute,
			setupMock: func(m *MockRedisClient, p *MockPipeliner) {
				// Setup pipeline
				m.On("TxPipeline").Return(p)
				
				// Setup Incr command
				incrCmd := redis.NewIntCmd(context.Background())
				incrCmd.SetVal(6)
				p.On("Incr", mock.Anything, "test-key").Return(incrCmd)
				
				// Setup Expire command
				expireCmd := redis.NewBoolCmd(context.Background())
				expireCmd.SetVal(true)
				p.On("Expire", mock.Anything, "test-key", time.Minute).Return(expireCmd)
				
				// Setup Exec command
				p.On("Exec", mock.Anything).Return([]redis.Cmder{incrCmd, expireCmd}, nil)
				
				// Setup Get command
				getCmd := redis.NewStringCmd(context.Background())
				getCmd.SetVal("6")
				m.On("Get", mock.Anything, "test-key").Return(getCmd)
			},
			expectedAllow: false,
			expectedError: false,
		},
		{
			name:   "pipeline execution error",
			key:    "test-key",
			limit:  5,
			window: time.Minute,
			setupMock: func(m *MockRedisClient, p *MockPipeliner) {
				// Setup pipeline
				m.On("TxPipeline").Return(p)
				
				// Setup Incr command
				incrCmd := redis.NewIntCmd(context.Background())
				p.On("Incr", mock.Anything, "test-key").Return(incrCmd)
				
				// Setup Expire command
				expireCmd := redis.NewBoolCmd(context.Background())
				p.On("Expire", mock.Anything, "test-key", time.Minute).Return(expireCmd)
				
				// Setup Exec command with error
				p.On("Exec", mock.Anything).Return([]redis.Cmder{}, redis.ErrClosed)
			},
			expectedAllow: false,
			expectedError: true,
		},
		{
			name:   "get count error",
			key:    "test-key",
			limit:  5,
			window: time.Minute,
			setupMock: func(m *MockRedisClient, p *MockPipeliner) {
				// Setup pipeline
				m.On("TxPipeline").Return(p)
				
				// Setup Incr command
				incrCmd := redis.NewIntCmd(context.Background())
				incrCmd.SetVal(1)
				p.On("Incr", mock.Anything, "test-key").Return(incrCmd)
				
				// Setup Expire command
				expireCmd := redis.NewBoolCmd(context.Background())
				expireCmd.SetVal(true)
				p.On("Expire", mock.Anything, "test-key", time.Minute).Return(expireCmd)
				
				// Setup Exec command
				p.On("Exec", mock.Anything).Return([]redis.Cmder{incrCmd, expireCmd}, nil)
				
				// Setup Get command with error
				getCmd := redis.NewStringCmd(context.Background())
				getCmd.SetErr(redis.ErrClosed)
				m.On("Get", mock.Anything, "test-key").Return(getCmd)
			},
			expectedAllow: false,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockRedisClient)
			mockPipeline := new(MockPipeliner)
			tt.setupMock(mockClient, mockPipeline)
			
			limiter := NewRateLimiter(mockClient, tt.limit, tt.window)
			allowed, err := limiter.Allow(tt.key)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedAllow, allowed)

			mockClient.AssertExpectations(t)
			mockPipeline.AssertExpectations(t)
		})
	}
} 