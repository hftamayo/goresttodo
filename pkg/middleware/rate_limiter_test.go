package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/pkg/utils"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient is a mock implementation of redis.Client
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Incr(ctx interface{}, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) Expire(ctx interface{}, key string, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *MockRedisClient) Get(ctx interface{}, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClient) TxPipeline() redis.Pipeliner {
	args := m.Called()
	return args.Get(0).(redis.Pipeliner)
}

// MockPipeliner is a mock implementation of redis.Pipeliner
type MockPipeliner struct {
	mock.Mock
}

func (m *MockPipeliner) Incr(ctx interface{}, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockPipeliner) Expire(ctx interface{}, key string, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *MockPipeliner) Exec(ctx interface{}) ([]redis.Cmder, error) {
	args := m.Called(ctx)
	return args.Get(0).([]redis.Cmder), args.Error(1)
}

func setupTestRouter(rateLimiter *utils.RateLimiter) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimitMiddleware(rateLimiter))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return router
}

func TestRateLimitMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		clientIP       string
		setupMock      func(*MockRedisClient, *MockPipeliner)
		expectedCode   int
		expectedBody   string
	}{
		{
			name:     "request within rate limit",
			clientIP: "192.168.1.1",
			setupMock: func(mockRedis *MockRedisClient, mockPipe *MockPipeliner) {
				mockRedis.On("TxPipeline").Return(mockPipe)
				mockPipe.On("Incr", mock.Anything, "192.168.1.1").Return(redis.NewIntCmd(mock.Anything))
				mockPipe.On("Expire", mock.Anything, "192.168.1.1", mock.Anything).Return(redis.NewBoolCmd(mock.Anything))
				mockPipe.On("Exec", mock.Anything).Return([]redis.Cmder{}, nil)
				mockRedis.On("Get", mock.Anything, "192.168.1.1").Return(redis.NewStringCmd(mock.Anything, "5"))
			},
			expectedCode: http.StatusOK,
		},
		{
			name:     "request exceeds rate limit",
			clientIP: "192.168.1.2",
			setupMock: func(mockRedis *MockRedisClient, mockPipe *MockPipeliner) {
				mockRedis.On("TxPipeline").Return(mockPipe)
				mockPipe.On("Incr", mock.Anything, "192.168.1.2").Return(redis.NewIntCmd(mock.Anything))
				mockPipe.On("Expire", mock.Anything, "192.168.1.2", mock.Anything).Return(redis.NewBoolCmd(mock.Anything))
				mockPipe.On("Exec", mock.Anything).Return([]redis.Cmder{}, nil)
				mockRedis.On("Get", mock.Anything, "192.168.1.2").Return(redis.NewStringCmd(mock.Anything, "101"))
			},
			expectedCode: http.StatusTooManyRequests,
			expectedBody: `{"code":429,"resultMessage":"RATE_LIMIT_EXCEEDED"}`,
		},
		{
			name:     "redis error",
			clientIP: "192.168.1.3",
			setupMock: func(mockRedis *MockRedisClient, mockPipe *MockPipeliner) {
				mockRedis.On("TxPipeline").Return(mockPipe)
				mockPipe.On("Incr", mock.Anything, "192.168.1.3").Return(redis.NewIntCmd(mock.Anything))
				mockPipe.On("Expire", mock.Anything, "192.168.1.3", mock.Anything).Return(redis.NewBoolCmd(mock.Anything))
				mockPipe.On("Exec", mock.Anything).Return([]redis.Cmder{}, assert.AnError)
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"code":500,"resultMessage":"INTERNAL_SERVER_ERROR"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock objects
			mockRedis := new(MockRedisClient)
			mockPipe := new(MockPipeliner)
			tt.setupMock(mockRedis, mockPipe)

			// Create rate limiter with mock Redis client
			rateLimiter := utils.NewRateLimiter(mockRedis, 100, time.Minute)

			// Setup router with rate limiter middleware
			router := setupTestRouter(rateLimiter)

			// Create test request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.clientIP

			// Make request
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		setupMock      func(*MockRedisClient, *MockPipeliner)
		expectedAllow  bool
		expectedError  bool
	}{
		{
			name: "within rate limit",
			key:  "test-key-1",
			setupMock: func(mockRedis *MockRedisClient, mockPipe *MockPipeliner) {
				mockRedis.On("TxPipeline").Return(mockPipe)
				mockPipe.On("Incr", mock.Anything, "test-key-1").Return(redis.NewIntCmd(mock.Anything))
				mockPipe.On("Expire", mock.Anything, "test-key-1", mock.Anything).Return(redis.NewBoolCmd(mock.Anything))
				mockPipe.On("Exec", mock.Anything).Return([]redis.Cmder{}, nil)
				mockRedis.On("Get", mock.Anything, "test-key-1").Return(redis.NewStringCmd(mock.Anything, "5"))
			},
			expectedAllow: true,
			expectedError: false,
		},
		{
			name: "exceeds rate limit",
			key:  "test-key-2",
			setupMock: func(mockRedis *MockRedisClient, mockPipe *MockPipeliner) {
				mockRedis.On("TxPipeline").Return(mockPipe)
				mockPipe.On("Incr", mock.Anything, "test-key-2").Return(redis.NewIntCmd(mock.Anything))
				mockPipe.On("Expire", mock.Anything, "test-key-2", mock.Anything).Return(redis.NewBoolCmd(mock.Anything))
				mockPipe.On("Exec", mock.Anything).Return([]redis.Cmder{}, nil)
				mockRedis.On("Get", mock.Anything, "test-key-2").Return(redis.NewStringCmd(mock.Anything, "101"))
			},
			expectedAllow: false,
			expectedError: false,
		},
		{
			name: "redis error",
			key:  "test-key-3",
			setupMock: func(mockRedis *MockRedisClient, mockPipe *MockPipeliner) {
				mockRedis.On("TxPipeline").Return(mockPipe)
				mockPipe.On("Incr", mock.Anything, "test-key-3").Return(redis.NewIntCmd(mock.Anything))
				mockPipe.On("Expire", mock.Anything, "test-key-3", mock.Anything).Return(redis.NewBoolCmd(mock.Anything))
				mockPipe.On("Exec", mock.Anything).Return([]redis.Cmder{}, assert.AnError)
			},
			expectedAllow: false,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock objects
			mockRedis := new(MockRedisClient)
			mockPipe := new(MockPipeliner)
			tt.setupMock(mockRedis, mockPipe)

			// Create rate limiter with mock Redis client
			rateLimiter := utils.NewRateLimiter(mockRedis, 100, time.Minute)

			// Test Allow method
			allowed, err := rateLimiter.Allow(tt.key)

			// Assert results
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedAllow, allowed)
		})
	}
} 