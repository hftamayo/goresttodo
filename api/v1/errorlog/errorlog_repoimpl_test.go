package errorlog

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient implements utils.RedisClientInterface for testing
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

func (m *MockRedisClient) Incr(ctx context.Context, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	args := m.Called(ctx, key, values)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) HMSet(ctx context.Context, key string, values ...interface{}) *redis.BoolCmd {
	args := m.Called(ctx, key, values)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestErrorLogRepositoryImpl_LogError(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		err       error
		setupMock func(*MockRedisClient)
		expectErr bool
	}{
		{
			name:      "success",
			operation: "test_op",
			err:       errors.New("test error"),
			setupMock: func(m *MockRedisClient) {
				cmd := redis.NewBoolCmd(context.Background())
				cmd.SetVal(true)
				m.On("HMSet", mock.Anything, "logs", mock.Anything).Return(cmd)
			},
			expectErr: false,
		},
		{
			name:      "redis error",
			operation: "test_op",
			err:       errors.New("test error"),
			setupMock: func(m *MockRedisClient) {
				cmd := redis.NewBoolCmd(context.Background())
				cmd.SetErr(errors.New("redis failure"))
				m.On("HMSet", mock.Anything, "logs", mock.Anything).Return(cmd)
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRedis := new(MockRedisClient)
			if tt.setupMock != nil {
				tt.setupMock(mockRedis)
			}
			repo := NewErrorLogRepositoryImpl(mockRedis)
			err := repo.LogError(tt.operation, tt.err)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockRedis.AssertExpectations(t)
		})
	}
} 