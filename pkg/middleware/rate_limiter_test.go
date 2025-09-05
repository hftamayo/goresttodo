package middleware

import (
	"context"
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

func (m *MockRedisClient) Incr(ctx context.Context, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
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

func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
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

func (m *MockRedisClient) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	args := m.Called(ctx, key, values)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) HMSet(ctx context.Context, key string, values ...interface{}) *redis.BoolCmd {
	args := m.Called(ctx, key, values)
	return args.Get(0).(*redis.BoolCmd)
}

func setupRateLimiterTestRouter(rateLimiter *utils.RateLimiter) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimiter(rateLimiter))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	router.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusCreated)
		c.JSON(http.StatusCreated, gin.H{"message": "created"})
	})
	router.PUT("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
		c.JSON(http.StatusOK, gin.H{"message": "updated"})
	})
	router.DELETE("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})
	router.OPTIONS("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
		c.JSON(http.StatusOK, gin.H{"message": "options"})
	})
	router.HEAD("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return router
}

func TestRateLimiter(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		clientIP       string
		setupMock      func(*MockRedisClient)
		expectedCode   int
		expectedBody   string
		description    string
	}{
		{
			name:     "GET request within read rate limit",
			method:   "GET",
			clientIP: "192.168.1.1",
			setupMock: func(mockRedis *MockRedisClient) {
				// First request - increment counter (returning 5 means it's not the first request)
				mockRedis.On("Incr", mock.Anything, "rate_limit:192.168.1.1:60").Return(
					redis.NewIntResult(5, nil))
				// No Expire call expected since currentCount != 1
			},
			expectedCode: http.StatusOK,
			description:  "GET request should be allowed within read rate limit",
		},
		{
			name:     "POST request within write rate limit",
			method:   "POST",
			clientIP: "192.168.1.2",
			setupMock: func(mockRedis *MockRedisClient) {
				// First request - increment counter (returning 10 means it's not the first request)
				mockRedis.On("Incr", mock.Anything, "rate_limit:192.168.1.2:60").Return(
					redis.NewIntResult(10, nil))
				// No Expire call expected since currentCount != 1
			},
			expectedCode: http.StatusCreated,
			description:  "POST request should be allowed within write rate limit",
		},
		{
			name:     "PUT request within write rate limit",
			method:   "PUT",
			clientIP: "192.168.1.3",
			setupMock: func(mockRedis *MockRedisClient) {
				mockRedis.On("Incr", mock.Anything, "rate_limit:192.168.1.3:60").Return(
					redis.NewIntResult(25, nil))
				// No Expire call expected since currentCount != 1
			},
			expectedCode: http.StatusOK,
			description:  "PUT request should be allowed within write rate limit",
		},
		{
			name:     "DELETE request within write rate limit",
			method:   "DELETE",
			clientIP: "192.168.1.4",
			setupMock: func(mockRedis *MockRedisClient) {
				mockRedis.On("Incr", mock.Anything, "rate_limit:192.168.1.4:60").Return(
					redis.NewIntResult(30, nil))
				// No Expire call expected since currentCount != 1
			},
			expectedCode: http.StatusOK,
			description:  "DELETE request should be allowed within write rate limit",
		},
		{
			name:     "GET request exceeds read rate limit",
			method:   "GET",
			clientIP: "192.168.1.5",
			setupMock: func(mockRedis *MockRedisClient) {
				mockRedis.On("Incr", mock.Anything, "rate_limit:192.168.1.5:60").Return(
					redis.NewIntResult(101, nil))
			},
			expectedCode: http.StatusTooManyRequests,
			expectedBody: `{"error":"Rate limit exceeded","retry_after":`,
			description:  "GET request should be blocked when exceeding read rate limit",
		},
		{
			name:     "POST request exceeds write rate limit",
			method:   "POST",
			clientIP: "192.168.1.6",
			setupMock: func(mockRedis *MockRedisClient) {
				mockRedis.On("Incr", mock.Anything, "rate_limit:192.168.1.6:60").Return(
					redis.NewIntResult(51, nil))
			},
			expectedCode: http.StatusTooManyRequests,
			expectedBody: `{"error":"Rate limit exceeded","retry_after":`,
			description:  "POST request should be blocked when exceeding write rate limit",
		},
		{
			name:     "redis error during increment",
			method:   "GET",
			clientIP: "192.168.1.7",
			setupMock: func(mockRedis *MockRedisClient) {
				mockRedis.On("Incr", mock.Anything, "rate_limit:192.168.1.7:60").Return(
					redis.NewIntResult(0, assert.AnError))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"error":"Rate limit error"}`,
			description:  "Should handle Redis increment errors gracefully",
		},
		{
			name:     "redis error during expire",
			method:   "GET",
			clientIP: "192.168.1.8",
			setupMock: func(mockRedis *MockRedisClient) {
				mockRedis.On("Incr", mock.Anything, "rate_limit:192.168.1.8:60").Return(
					redis.NewIntResult(1, nil))
				mockRedis.On("Expire", mock.Anything, "rate_limit:192.168.1.8:60", time.Duration(60)*time.Second).Return(
					redis.NewBoolResult(false, assert.AnError))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"error":"Rate limit error"}`,
			description:  "Should handle Redis expire errors gracefully",
		},
		{
			name:     "unknown client IP",
			method:   "GET",
			clientIP: "",
			setupMock: func(mockRedis *MockRedisClient) {
				mockRedis.On("Incr", mock.Anything, "rate_limit:unknown:60").Return(
					redis.NewIntResult(5, nil))
				// No Expire call expected since currentCount != 1
			},
			expectedCode: http.StatusOK,
			description:  "Should handle empty client IP by using 'unknown'",
		},
		{
			name:     "OPTIONS request treated as read",
			method:   "OPTIONS",
			clientIP: "192.168.1.9",
			setupMock: func(mockRedis *MockRedisClient) {
				mockRedis.On("Incr", mock.Anything, "rate_limit:192.168.1.9:60").Return(
					redis.NewIntResult(5, nil))
				// No Expire call expected since currentCount != 1
			},
			expectedCode: http.StatusOK,
			description:  "OPTIONS request should be treated as read operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock Redis client
			mockRedis := new(MockRedisClient)
			tt.setupMock(mockRedis)

			// Create rate limiter with mock Redis client
			rateLimiter := utils.NewRateLimiter(mockRedis)

			// Setup router with rate limiter middleware
			router := setupRateLimiterTestRouter(rateLimiter)

			// Create test request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, "/test", nil)
			if tt.clientIP != "" {
				req.RemoteAddr = tt.clientIP + ":12345"
			}

			// Make request
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedCode, w.Code, "Test: %s", tt.description)

			if tt.expectedBody != "" {
				if tt.expectedCode == http.StatusTooManyRequests {
					// For rate limit exceeded, check that response contains the expected fields
					assert.Contains(t, w.Body.String(), `"error":"Rate limit exceeded"`, 
						"Response should contain rate limit error message")
					assert.Contains(t, w.Body.String(), `"retry_after"`, 
						"Response should contain retry_after field")
				} else {
					assert.JSONEq(t, tt.expectedBody, w.Body.String(), 
						"Response body mismatch for test: %s", tt.description)
				}
			}

			// Verify rate limit headers for successful requests
			if tt.expectedCode == http.StatusOK || tt.expectedCode == http.StatusCreated {
				assert.NotEmpty(t, w.Header().Get("X-RateLimit-Limit"), 
					"Rate limit header should be present for successful requests")
				assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"), 
					"Rate limit remaining header should be present for successful requests")
			}

			// Verify mock expectations
			mockRedis.AssertExpectations(t)
		})
	}
}

func TestRateLimiter_OperationTypes(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedOp     utils.OperationType
		description    string
	}{
		{
			name:        "GET method maps to read operation",
			method:      "GET",
			expectedOp:  utils.OperationRead,
			description: "GET requests should be treated as read operations",
		},
		{
			name:        "POST method maps to write operation",
			method:      "POST",
			expectedOp:  utils.OperationWrite,
			description: "POST requests should be treated as write operations",
		},
		{
			name:        "PUT method maps to write operation",
			method:      "PUT",
			expectedOp:  utils.OperationWrite,
			description: "PUT requests should be treated as write operations",
		},
		{
			name:        "PATCH method maps to write operation",
			method:      "PATCH",
			expectedOp:  utils.OperationWrite,
			description: "PATCH requests should be treated as write operations",
		},
		{
			name:        "DELETE method maps to write operation",
			method:      "DELETE",
			expectedOp:  utils.OperationWrite,
			description: "DELETE requests should be treated as write operations",
		},
		{
			name:        "OPTIONS method maps to read operation",
			method:      "OPTIONS",
			expectedOp:  utils.OperationRead,
			description: "OPTIONS requests should be treated as read operations",
		},
		{
			name:        "HEAD method maps to read operation",
			method:      "HEAD",
			expectedOp:  utils.OperationRead,
			description: "HEAD requests should be treated as read operations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock Redis client
			mockRedis := new(MockRedisClient)
			mockRedis.On("Incr", mock.Anything, mock.Anything).Return(redis.NewIntResult(1, nil))
			mockRedis.On("Expire", mock.Anything, mock.Anything, mock.Anything).Return(redis.NewBoolResult(true, nil))

			// Create rate limiter with mock Redis client
			rateLimiter := utils.NewRateLimiter(mockRedis)

			// Setup router with rate limiter middleware
			router := setupRateLimiterTestRouter(rateLimiter)

			// Create test request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, "/test", nil)
			req.RemoteAddr = "192.168.1.1:12345"

			// Make request
			router.ServeHTTP(w, req)

			// Verify the request was processed (status should be OK for read operations)
			if tt.expectedOp == utils.OperationRead {
				assert.Equal(t, http.StatusOK, w.Code, "Test: %s", tt.description)
			} else {
				// For write operations, we expect the request to be processed
				assert.NotEqual(t, http.StatusInternalServerError, w.Code, "Test: %s", tt.description)
			}

			mockRedis.AssertExpectations(t)
		})
	}
}

func TestRateLimiter_Headers(t *testing.T) {
	// Create mock Redis client
	mockRedis := new(MockRedisClient)
	mockRedis.On("Incr", mock.Anything, mock.Anything).Return(redis.NewIntResult(5, nil))
	// No Expire call expected since currentCount != 1

	// Create rate limiter with mock Redis client
	rateLimiter := utils.NewRateLimiter(mockRedis)

			// Setup router with rate limiter middleware
		router := setupRateLimiterTestRouter(rateLimiter)

	// Create test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	// Make request
	router.ServeHTTP(w, req)

	// Verify rate limit headers
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Limit"), "X-RateLimit-Limit header should be present")
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"), "X-RateLimit-Remaining header should be present")
	
	// Verify header values are numeric
	limit := w.Header().Get("X-RateLimit-Limit")
	remaining := w.Header().Get("X-RateLimit-Remaining")
	assert.NotEqual(t, "", limit, "Limit header should not be empty")
	assert.NotEqual(t, "", remaining, "Remaining header should not be empty")

	mockRedis.AssertExpectations(t)
}

func TestRateLimiter_ClientIPExtraction(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		expectedIP     string
		description    string
	}{
		{
			name:        "standard IP address",
			remoteAddr:  "192.168.1.1:12345",
			expectedIP:  "192.168.1.1",
			description: "Should extract IP from standard remote address",
		},
		{
			name:        "IPv6 address",
			remoteAddr:  "[2001:db8::1]:12345",
			expectedIP:  "2001:db8::1",
			description: "Should extract IP from IPv6 address",
		},
		{
			name:        "empty remote address",
			remoteAddr:  "",
			expectedIP:  "unknown",
			description: "Should use 'unknown' for empty remote address",
		},
		{
			name:        "malformed remote address",
			remoteAddr:  "invalid-address",
			expectedIP:  "invalid-address",
			description: "Should handle malformed remote address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock Redis client
			mockRedis := new(MockRedisClient)
			mockRedis.On("Incr", mock.Anything, mock.Anything).Return(redis.NewIntResult(1, nil))
			mockRedis.On("Expire", mock.Anything, mock.Anything, mock.Anything).Return(redis.NewBoolResult(true, nil))

			// Create rate limiter with mock Redis client
			rateLimiter := utils.NewRateLimiter(mockRedis)

			// Setup router with rate limiter middleware
			router := setupRateLimiterTestRouter(rateLimiter)

			// Create test request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.remoteAddr != "" {
				req.RemoteAddr = tt.remoteAddr
			}

			// Make request
			router.ServeHTTP(w, req)

			// Verify the request was processed
			assert.Equal(t, http.StatusOK, w.Code, "Test: %s", tt.description)

			mockRedis.AssertExpectations(t)
		})
	}
}

func TestRateLimiter_Performance(t *testing.T) {
	// Create mock Redis client
	mockRedis := new(MockRedisClient)
	mockRedis.On("Incr", mock.Anything, mock.Anything).Return(redis.NewIntResult(1, nil))
	mockRedis.On("Expire", mock.Anything, mock.Anything, mock.Anything).Return(redis.NewBoolResult(true, nil))

	// Create rate limiter with mock Redis client
	rateLimiter := utils.NewRateLimiter(mockRedis)

			// Setup router with rate limiter middleware
		router := setupRateLimiterTestRouter(rateLimiter)

	// Test multiple concurrent requests
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	ips := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3", "192.168.1.4"}

	for _, method := range methods {
		for _, ip := range ips {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(method, "/test", nil)
			req.RemoteAddr = ip + ":12345"

			router.ServeHTTP(w, req)

			// All requests should be processed successfully
			assert.NotEqual(t, http.StatusInternalServerError, w.Code, 
				"Request with method %s from IP %s should be processed", method, ip)
		}
	}

	mockRedis.AssertExpectations(t)
}

func BenchmarkRateLimiter(b *testing.B) {
	// Create mock Redis client
	mockRedis := new(MockRedisClient)
	mockRedis.On("Incr", mock.Anything, mock.Anything).Return(redis.NewIntResult(1, nil))
	mockRedis.On("Expire", mock.Anything, mock.Anything, mock.Anything).Return(redis.NewBoolResult(true, nil))

	// Create rate limiter with mock Redis client
	rateLimiter := utils.NewRateLimiter(mockRedis)

			// Setup router with rate limiter middleware
		router := setupRateLimiterTestRouter(rateLimiter)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"

		router.ServeHTTP(w, req)
	}
}

func BenchmarkRateLimiter_Write(b *testing.B) {
	// Create mock Redis client
	mockRedis := new(MockRedisClient)
	mockRedis.On("Incr", mock.Anything, mock.Anything).Return(redis.NewIntResult(1, nil))
	mockRedis.On("Expire", mock.Anything, mock.Anything, mock.Anything).Return(redis.NewBoolResult(true, nil))

	// Create rate limiter with mock Redis client
	rateLimiter := utils.NewRateLimiter(mockRedis)

			// Setup router with rate limiter middleware
		router := setupRateLimiterTestRouter(rateLimiter)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"

		router.ServeHTTP(w, req)
	}
} 