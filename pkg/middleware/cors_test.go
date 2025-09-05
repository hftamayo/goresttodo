package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/pkg/config"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})
	router.PATCH("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.DELETE("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.OPTIONS("/test", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	return router
}

func setupTestRouterWithCORS(corsMiddleware gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(corsMiddleware)
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})
	router.PATCH("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.DELETE("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.OPTIONS("/test", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	return router
}

func TestCORSMiddleware(t *testing.T) {
	tests := []struct {
		name            string
		origin          string
		method          string
		allowedOrigins  []string
		expectedCode    int
		expectedHeaders map[string]string
		description     string
	}{
		{
			name:   "allowed origin with GET",
			origin: "http://localhost:3000",
			method: "GET",
			allowedOrigins: []string{"http://localhost:3000"},
			expectedCode: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://localhost:3000",
				"Access-Control-Allow-Methods":     "GET, POST, PATCH, DELETE, OPTIONS",
				"Access-Control-Allow-Headers":     "Content-Type, Authorization",
				"Access-Control-Max-Age":          "86400",
				"Access-Control-Allow-Credentials": "true",
			},
			description: "Standard GET request with allowed origin",
		},
		{
			name:   "allowed origin with POST",
			origin: "http://localhost:3000",
			method: "POST",
			allowedOrigins: []string{"http://localhost:3000"},
			expectedCode: http.StatusCreated,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://localhost:3000",
				"Access-Control-Allow-Methods":     "GET, POST, PATCH, DELETE, OPTIONS",
				"Access-Control-Allow-Headers":     "Content-Type, Authorization",
				"Access-Control-Max-Age":          "86400",
				"Access-Control-Allow-Credentials": "true",
			},
			description: "POST request with allowed origin",
		},
		{
			name:   "allowed origin with PATCH",
			origin: "http://localhost:3000",
			method: "PATCH",
			allowedOrigins: []string{"http://localhost:3000"},
			expectedCode: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://localhost:3000",
				"Access-Control-Allow-Methods":     "GET, POST, PATCH, DELETE, OPTIONS",
				"Access-Control-Allow-Headers":     "Content-Type, Authorization",
				"Access-Control-Max-Age":          "86400",
				"Access-Control-Allow-Credentials": "true",
			},
			description: "PATCH request with allowed origin",
		},
		{
			name:   "allowed origin with DELETE",
			origin: "http://localhost:3000",
			method: "DELETE",
			allowedOrigins: []string{"http://localhost:3000"},
			expectedCode: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://localhost:3000",
				"Access-Control-Allow-Methods":     "GET, POST, PATCH, DELETE, OPTIONS",
				"Access-Control-Allow-Headers":     "Content-Type, Authorization",
				"Access-Control-Max-Age":          "86400",
				"Access-Control-Allow-Credentials": "true",
			},
			description: "DELETE request with allowed origin",
		},
		{
			name:   "disallowed origin",
			origin: "http://malicious.com",
			method: "GET",
			allowedOrigins: []string{"http://localhost:3000"},
			expectedCode: http.StatusForbidden,
			expectedHeaders: map[string]string{},
			description: "Request from disallowed origin should be blocked",
		},
		{
			name:   "preflight request with allowed origin",
			origin: "http://localhost:3000",
			method: "OPTIONS",
			allowedOrigins: []string{"http://localhost:3000"},
			expectedCode: http.StatusNoContent,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://localhost:3000",
				"Access-Control-Allow-Methods":     "GET, POST, PATCH, DELETE, OPTIONS",
				"Access-Control-Allow-Headers":     "Content-Type, Authorization",
				"Access-Control-Max-Age":          "86400",
				"Access-Control-Allow-Credentials": "true",
			},
			description: "Preflight OPTIONS request with allowed origin",
		},
		{
			name:   "preflight request with disallowed origin",
			origin: "http://malicious.com",
			method: "OPTIONS",
			allowedOrigins: []string{"http://localhost:3000"},
			expectedCode: http.StatusForbidden,
			expectedHeaders: map[string]string{},
			description: "Preflight OPTIONS request with disallowed origin",
		},
		{
			name:   "no origin header",
			origin: "",
			method: "GET",
			allowedOrigins: []string{"http://localhost:3000"},
			expectedCode: http.StatusForbidden,
			expectedHeaders: map[string]string{},
			description: "Request without Origin header should be blocked",
		},
		{
			name:   "multiple allowed origins - first match",
			origin: "http://app.example.com",
			method: "GET",
			allowedOrigins: []string{
				"http://localhost:3000",
				"http://app.example.com",
				"https://app.example.com",
			},
			expectedCode: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://app.example.com",
				"Access-Control-Allow-Methods":     "GET, POST, PATCH, DELETE, OPTIONS",
				"Access-Control-Allow-Headers":     "Content-Type, Authorization",
				"Access-Control-Max-Age":          "86400",
				"Access-Control-Allow-Credentials": "true",
			},
			description: "Request matching first allowed origin in list",
		},
		{
			name:   "multiple allowed origins - second match",
			origin: "https://app.example.com",
			method: "GET",
			allowedOrigins: []string{
				"http://localhost:3000",
				"http://app.example.com",
				"https://app.example.com",
			},
			expectedCode: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "https://app.example.com",
				"Access-Control-Allow-Methods":     "GET, POST, PATCH, DELETE, OPTIONS",
				"Access-Control-Allow-Headers":     "Content-Type, Authorization",
				"Access-Control-Max-Age":          "86400",
				"Access-Control-Allow-Credentials": "true",
			},
			description: "Request matching second allowed origin in list",
		},
		{
			name:   "case sensitive origin matching",
			origin: "HTTP://LOCALHOST:3000",
			method: "GET",
			allowedOrigins: []string{"http://localhost:3000"},
			expectedCode: http.StatusForbidden,
			expectedHeaders: map[string]string{},
			description: "Origin matching should be case sensitive",
		},
		{
			name:   "empty allowed origins list",
			origin: "http://localhost:3000",
			method: "GET",
			allowedOrigins: []string{},
			expectedCode: http.StatusForbidden,
			expectedHeaders: map[string]string{},
			description: "Request should be blocked when no origins are allowed",
		},
		{
			name:   "wildcard origin not supported",
			origin: "http://any-domain.com",
			method: "GET",
			allowedOrigins: []string{"*"},
			expectedCode: http.StatusForbidden,
			expectedHeaders: map[string]string{},
			description: "Wildcard origins should not be supported for security",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test environment variables
			envVars := &config.EnvVars{
				FeOrigins: tt.allowedOrigins,
				AppPort:   8080, // Add AppPort for debug logging
			}

			// Setup router with CORS middleware
			router := setupTestRouterWithCORS(CORSMiddleware(envVars))

			// Create test request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			// Make request
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedCode, w.Code, "Test: %s", tt.description)

			// Check headers
			for key, expectedValue := range tt.expectedHeaders {
				actualValue := w.Header().Get(key)
				assert.Equal(t, expectedValue, actualValue, "Header %s mismatch for test: %s", key, tt.description)
			}

			// Verify that unauthorized requests don't have CORS headers
			if tt.expectedCode == http.StatusForbidden {
				assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"), 
					"Forbidden requests should not have CORS headers")
				assert.Empty(t, w.Header().Get("Access-Control-Allow-Methods"), 
					"Forbidden requests should not have CORS headers")
				assert.Empty(t, w.Header().Get("Access-Control-Allow-Headers"), 
					"Forbidden requests should not have CORS headers")
			}
		})
	}
}

func TestCORSMiddleware_AdditionalHeaders(t *testing.T) {
	envVars := &config.EnvVars{
		FeOrigins: []string{"http://localhost:3000"},
		AppPort:   8080,
	}

	router := setupTestRouterWithCORS(CORSMiddleware(envVars))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	router.ServeHTTP(w, req)

	// Verify all required CORS headers are present
	expectedHeaders := []string{
		"Access-Control-Allow-Origin",
		"Access-Control-Allow-Methods",
		"Access-Control-Allow-Headers",
		"Access-Control-Max-Age",
		"Access-Control-Allow-Credentials",
	}

	for _, header := range expectedHeaders {
		assert.NotEmpty(t, w.Header().Get(header), "Required CORS header %s should be present", header)
	}

	// Verify specific header values
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PATCH, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type, Authorization", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORSMiddleware_Performance(t *testing.T) {
	envVars := &config.EnvVars{
		FeOrigins: []string{"http://localhost:3000", "http://app.example.com", "https://app.example.com"},
		AppPort:   8080,
	}

	router := setupTestRouterWithCORS(CORSMiddleware(envVars))

	// Test multiple requests to ensure consistent behavior
	origins := []string{"http://localhost:3000", "http://app.example.com", "https://app.example.com"}
	methods := []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"}

	for _, origin := range origins {
		for _, method := range methods {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(method, "/test", nil)
			req.Header.Set("Origin", origin)

			router.ServeHTTP(w, req)

			// All requests should succeed for allowed origins
			if method == "OPTIONS" {
				assert.Equal(t, http.StatusNoContent, w.Code)
			} else {
				assert.NotEqual(t, http.StatusForbidden, w.Code, 
					"Request from allowed origin %s with method %s should not be forbidden", origin, method)
			}
		}
	}
}

func TestCORSMiddleware_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		allowedOrigins []string
		expectedCode   int
		description    string
	}{
		{
			name:           "origin with trailing slash",
			origin:         "http://localhost:3000/",
			allowedOrigins: []string{"http://localhost:3000"},
			expectedCode:   http.StatusForbidden,
			description:    "Origin with trailing slash should not match",
		},
		{
			name:           "origin with path",
			origin:         "http://localhost:3000/path",
			allowedOrigins: []string{"http://localhost:3000"},
			expectedCode:   http.StatusForbidden,
			description:    "Origin with path should not match",
		},
		{
			name:           "origin with query parameters",
			origin:         "http://localhost:3000?param=value",
			allowedOrigins: []string{"http://localhost:3000"},
			expectedCode:   http.StatusForbidden,
			description:    "Origin with query parameters should not match",
		},
		{
			name:           "origin with fragment",
			origin:         "http://localhost:3000#fragment",
			allowedOrigins: []string{"http://localhost:3000"},
			expectedCode:   http.StatusForbidden,
			description:    "Origin with fragment should not match",
		},
		{
			name:           "null byte in origin",
			origin:         string([]byte{0}) + "http://localhost:3000",
			allowedOrigins: []string{"http://localhost:3000"},
			expectedCode:   http.StatusForbidden,
			description:    "Origin with null bytes should be blocked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envVars := &config.EnvVars{
				FeOrigins: tt.allowedOrigins,
				AppPort:   8080,
			}

			router := setupTestRouterWithCORS(CORSMiddleware(envVars))

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", tt.origin)

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code, "Test: %s", tt.description)
		})
	}
}

func BenchmarkCORSMiddleware(b *testing.B) {
	envVars := &config.EnvVars{
		FeOrigins: []string{"http://localhost:3000", "http://app.example.com", "https://app.example.com"},
		AppPort:   8080,
	}

	router := setupTestRouterWithCORS(CORSMiddleware(envVars))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")

		router.ServeHTTP(w, req)
	}
}

func BenchmarkCORSMiddleware_DisallowedOrigin(b *testing.B) {
	envVars := &config.EnvVars{
		FeOrigins: []string{"http://localhost:3000"},
		AppPort:   8080,
	}

	router := setupTestRouterWithCORS(CORSMiddleware(envVars))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://malicious.com")

		router.ServeHTTP(w, req)
	}
} 