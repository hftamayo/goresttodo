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

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func TestNewCache(t *testing.T) {
	mockClient := new(MockRedisClient)
	cache := NewCache(mockClient)
	
	assert.NotNil(t, cache)
	assert.Equal(t, mockClient, cache.RedisClient)
}

func TestCache_Set(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		value         interface{}
		expiration    time.Duration
		setupMock     func(*MockRedisClient)
		expectedError bool
	}{
		{
			name:       "successful set",
			key:        "test-key",
			value:      map[string]string{"test": "value"},
			expiration: time.Hour,
			setupMock: func(m *MockRedisClient) {
				cmd := redis.NewStatusCmd(context.Background())
				cmd.SetVal("OK")
				m.On("Set", mock.Anything, "test-key", mock.Anything, time.Hour).Return(cmd)
			},
			expectedError: false,
		},
		{
			name:       "failed set",
			key:        "test-key",
			value:      make(chan int), // Invalid JSON value
			expiration: time.Hour,
			setupMock: func(m *MockRedisClient) {
				// No mock setup needed as it should fail before reaching Redis
			},
			expectedError: true,
		},
		{
			name:       "redis error",
			key:        "test-key",
			value:      "test-value",
			expiration: time.Hour,
			setupMock: func(m *MockRedisClient) {
				cmd := redis.NewStatusCmd(context.Background())
				cmd.SetErr(redis.ErrClosed)
				m.On("Set", mock.Anything, "test-key", mock.Anything, time.Hour).Return(cmd)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockRedisClient)
			tt.setupMock(mockClient)
			
			cache := NewCache(mockClient)
			err := cache.Set(tt.key, tt.value, tt.expiration)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestCache_Get(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		value         string
		setupMock     func(*MockRedisClient)
		expectedError bool
		expectedValue interface{}
	}{
		{
			name:  "successful get",
			key:   "test-key",
			value: `{"test":"value"}`,
			setupMock: func(m *MockRedisClient) {
				cmd := redis.NewStringCmd(context.Background())
				cmd.SetVal(`{"test":"value"}`)
				m.On("Get", mock.Anything, "test-key").Return(cmd)
			},
			expectedError: false,
			expectedValue: map[string]string{"test": "value"},
		},
		{
			name:  "key not found",
			key:   "test-key",
			setupMock: func(m *MockRedisClient) {
				cmd := redis.NewStringCmd(context.Background())
				cmd.SetErr(redis.Nil)
				m.On("Get", mock.Anything, "test-key").Return(cmd)
			},
			expectedError: true,
		},
		{
			name:  "invalid json",
			key:   "test-key",
			value: `invalid json`,
			setupMock: func(m *MockRedisClient) {
				cmd := redis.NewStringCmd(context.Background())
				cmd.SetVal(`invalid json`)
				m.On("Get", mock.Anything, "test-key").Return(cmd)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockRedisClient)
			tt.setupMock(mockClient)
			
			cache := NewCache(mockClient)
			var result map[string]string
			err := cache.Get(tt.key, &result)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, result)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestCache_Delete(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		setupMock     func(*MockRedisClient)
		expectedError bool
	}{
		{
			name: "successful delete",
			key:  "test-key",
			setupMock: func(m *MockRedisClient) {
				cmd := redis.NewIntCmd(context.Background())
				cmd.SetVal(1)
				m.On("Del", mock.Anything, []string{"test-key"}).Return(cmd)
			},
			expectedError: false,
		},
		{
			name: "failed delete",
			key:  "test-key",
			setupMock: func(m *MockRedisClient) {
				cmd := redis.NewIntCmd(context.Background())
				cmd.SetErr(redis.ErrClosed)
				m.On("Del", mock.Anything, []string{"test-key"}).Return(cmd)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockRedisClient)
			tt.setupMock(mockClient)
			
			cache := NewCache(mockClient)
			err := cache.Delete(tt.key)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
} 