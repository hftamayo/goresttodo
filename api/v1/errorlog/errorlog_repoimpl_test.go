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

// MockRedisClient implements only the HSet method needed for testing
// and embeds mock.Mock for testify compatibility.
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	args := m.Called(ctx, key, values)
	return args.Get(0).(*redis.IntCmd)
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
				cmd := redis.NewIntCmd(context.Background())
				cmd.SetVal(1)
				m.On("HSet", mock.Anything, "logs", mock.Anything).Return(cmd)
			},
			expectErr: false,
		},
		{
			name:      "redis error",
			operation: "test_op",
			err:       errors.New("test error"),
			setupMock: func(m *MockRedisClient) {
				cmd := redis.NewIntCmd(context.Background())
				cmd.SetErr(errors.New("redis failure"))
				m.On("HSet", mock.Anything, "logs", mock.Anything).Return(cmd)
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