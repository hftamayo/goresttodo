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

func (m *MockRedisClient) Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
	args := m.Called(ctx, cursor, match, count)
	return args.Get(0).(*redis.ScanCmd)
}

func (m *MockRedisClient) SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	args := m.Called(ctx, key, members)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) SRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	args := m.Called(ctx, key, members)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) SMembers(ctx context.Context, key string) *redis.StringSliceCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (m *MockRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

// TestNewCache tests the NewCache function
func TestNewCache(t *testing.T) {
	// Create a mock that implements RedisClientInterface
	mockClient := &MockRedisClient{}
	
	// Now NewCache should accept our mock since it implements RedisClientInterface
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

func TestCache_DeletePattern(t *testing.T) {
	tests := []struct {
		name          string
		pattern       string
		setupMock     func(*MockRedisClient)
		expectedError bool
	}{
		{
			name:    "successful pattern delete",
			pattern: "test-*",
			setupMock: func(m *MockRedisClient) {
				// Setup scan iterator
				scanCmd := redis.NewScanCmd(context.Background())
				scanCmd.SetVal([]string{"test-1", "test-2"}, 0)
				m.On("Scan", mock.Anything, uint64(0), "test-*", int64(0)).Return(scanCmd)

				// Setup delete commands
				delCmd := redis.NewIntCmd(context.Background())
				delCmd.SetVal(1)
				m.On("Del", mock.Anything, []string{"test-1"}).Return(delCmd)
				m.On("Del", mock.Anything, []string{"test-2"}).Return(delCmd)
			},
			expectedError: false,
		},
		{
			name:    "scan error",
			pattern: "test-*",
			setupMock: func(m *MockRedisClient) {
				scanCmd := redis.NewScanCmd(context.Background())
				scanCmd.SetErr(redis.ErrClosed)
				m.On("Scan", mock.Anything, uint64(0), "test-*", int64(0)).Return(scanCmd)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockRedisClient)
			tt.setupMock(mockClient)
			
			cache := NewCache(mockClient)
			err := cache.DeletePattern(tt.pattern)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestCache_DeleteByPrefix(t *testing.T) {
	tests := []struct {
		name          string
		prefix        string
		setupMock     func(*MockRedisClient)
		expectedError bool
	}{
		{
			name:   "successful prefix delete",
			prefix: "test",
			setupMock: func(m *MockRedisClient) {
				// Setup scan iterator
				scanCmd := redis.NewScanCmd(context.Background())
				scanCmd.SetVal([]string{"test-1", "test-2"}, 0)
				m.On("Scan", mock.Anything, uint64(0), "test*", int64(0)).Return(scanCmd)

				// Setup delete commands
				delCmd := redis.NewIntCmd(context.Background())
				delCmd.SetVal(1)
				m.On("Del", mock.Anything, []string{"test-1"}).Return(delCmd)
				m.On("Del", mock.Anything, []string{"test-2"}).Return(delCmd)
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockRedisClient)
			tt.setupMock(mockClient)
			
			cache := NewCache(mockClient)
			err := cache.DeleteByPrefix(tt.prefix)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestCache_SetWithTags(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		value         interface{}
		expiration    time.Duration
		tags          []string
		setupMock     func(*MockRedisClient)
		expectedError bool
	}{
		{
			name:       "successful set with tags",
			key:        "test-key",
			value:      "test-value",
			expiration: time.Hour,
			tags:       []string{"tag1", "tag2"},
			setupMock: func(m *MockRedisClient) {
				// Setup Set command
				setCmd := redis.NewStatusCmd(context.Background())
				setCmd.SetVal("OK")
				m.On("Set", mock.Anything, "test-key", mock.Anything, time.Hour).Return(setCmd)

				// Setup SAdd commands
				saddCmd := redis.NewIntCmd(context.Background())
				saddCmd.SetVal(1)
				m.On("SAdd", mock.Anything, "tag:tag1", "test-key").Return(saddCmd)
				m.On("SAdd", mock.Anything, "tag:tag2", "test-key").Return(saddCmd)

				// Setup Expire commands
				expireCmd := redis.NewBoolCmd(context.Background())
				expireCmd.SetVal(true)
				m.On("Expire", mock.Anything, "tag:tag1", time.Hour).Return(expireCmd)
				m.On("Expire", mock.Anything, "tag:tag2", time.Hour).Return(expireCmd)
			},
			expectedError: false,
		},
		{
			name:       "set error",
			key:        "test-key",
			value:      make(chan int), // Invalid JSON value
			expiration: time.Hour,
			tags:       []string{"tag1"},
			setupMock: func(m *MockRedisClient) {
				// No mock setup needed as it should fail before reaching Redis
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockRedisClient)
			tt.setupMock(mockClient)
			
			cache := NewCache(mockClient)
			err := cache.SetWithTags(tt.key, tt.value, tt.expiration, tt.tags...)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestCache_InvalidateByTags(t *testing.T) {
	tests := []struct {
		name          string
		tags          []string
		setupMock     func(*MockRedisClient)
		expectedError bool
	}{
		{
			name: "successful tag invalidation",
			tags: []string{"tag1", "tag2"},
			setupMock: func(m *MockRedisClient) {
				// Setup SMembers commands
				smembersCmd1 := redis.NewStringSliceCmd(context.Background())
				smembersCmd1.SetVal([]string{"key1", "key2"})
				m.On("SMembers", mock.Anything, "tag:tag1").Return(smembersCmd1)

				smembersCmd2 := redis.NewStringSliceCmd(context.Background())
				smembersCmd2.SetVal([]string{"key3"})
				m.On("SMembers", mock.Anything, "tag:tag2").Return(smembersCmd2)

				// Setup Del commands
				delCmd := redis.NewIntCmd(context.Background())
				delCmd.SetVal(1)
				m.On("Del", mock.Anything, []string{"key1"}).Return(delCmd)
				m.On("Del", mock.Anything, []string{"key2"}).Return(delCmd)
				m.On("Del", mock.Anything, []string{"key3"}).Return(delCmd)
				m.On("Del", mock.Anything, []string{"tag:tag1"}).Return(delCmd)
				m.On("Del", mock.Anything, []string{"tag:tag2"}).Return(delCmd)
			},
			expectedError: false,
		},
		{
			name: "smembers error",
			tags: []string{"tag1"},
			setupMock: func(m *MockRedisClient) {
				smembersCmd := redis.NewStringSliceCmd(context.Background())
				smembersCmd.SetErr(redis.ErrClosed)
				m.On("SMembers", mock.Anything, "tag:tag1").Return(smembersCmd)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockRedisClient)
			tt.setupMock(mockClient)
			
			cache := NewCache(mockClient)
			err := cache.InvalidateByTags(tt.tags...)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestCache_AddTag(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		tag           string
		setupMock     func(*MockRedisClient)
		expectedError bool
	}{
		{
			name: "successful tag addition",
			key:  "test-key",
			tag:  "test-tag",
			setupMock: func(m *MockRedisClient) {
				saddCmd := redis.NewIntCmd(context.Background())
				saddCmd.SetVal(1)
				m.On("SAdd", mock.Anything, "tag:test-tag", "test-key").Return(saddCmd)
			},
			expectedError: false,
		},
		{
			name: "sadd error",
			key:  "test-key",
			tag:  "test-tag",
			setupMock: func(m *MockRedisClient) {
				saddCmd := redis.NewIntCmd(context.Background())
				saddCmd.SetErr(redis.ErrClosed)
				m.On("SAdd", mock.Anything, "tag:test-tag", "test-key").Return(saddCmd)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockRedisClient)
			tt.setupMock(mockClient)
			
			cache := NewCache(mockClient)
			err := cache.AddTag(tt.key, tt.tag)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestCache_RemoveTag(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		tag           string
		setupMock     func(*MockRedisClient)
		expectedError bool
	}{
		{
			name: "successful tag removal",
			key:  "test-key",
			tag:  "test-tag",
			setupMock: func(m *MockRedisClient) {
				sremCmd := redis.NewIntCmd(context.Background())
				sremCmd.SetVal(1)
				m.On("SRem", mock.Anything, "tag:test-tag", "test-key").Return(sremCmd)
			},
			expectedError: false,
		},
		{
			name: "srem error",
			key:  "test-key",
			tag:  "test-tag",
			setupMock: func(m *MockRedisClient) {
				sremCmd := redis.NewIntCmd(context.Background())
				sremCmd.SetErr(redis.ErrClosed)
				m.On("SRem", mock.Anything, "tag:test-tag", "test-key").Return(sremCmd)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockRedisClient)
			tt.setupMock(mockClient)
			
			cache := NewCache(mockClient)
			err := cache.RemoveTag(tt.key, tt.tag)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
} 