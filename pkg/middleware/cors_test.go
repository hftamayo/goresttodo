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
	return router
}

func TestCORSMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		method         string
		allowedOrigins []string
		expectedCode   int
		expectedHeaders map[string]string
	}{
		{
			name:   "allowed origin",
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
		},
		{
			name:   "disallowed origin",
			origin: "http://malicious.com",
			method: "GET",
			allowedOrigins: []string{"http://localhost:3000"},
			expectedCode: http.StatusForbidden,
			expectedHeaders: map[string]string{},
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
		},
		{
			name:   "preflight request with disallowed origin",
			origin: "http://malicious.com",
			method: "OPTIONS",
			allowedOrigins: []string{"http://localhost:3000"},
			expectedCode: http.StatusForbidden,
			expectedHeaders: map[string]string{},
		},
		{
			name:   "no origin header",
			origin: "",
			method: "GET",
			allowedOrigins: []string{"http://localhost:3000"},
			expectedCode: http.StatusForbidden,
			expectedHeaders: map[string]string{},
		},
		{
			name:   "multiple allowed origins",
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test environment variables
			envVars := &config.EnvVars{
				FeOrigins: tt.allowedOrigins,
			}

			// Setup router with CORS middleware
			router := setupTestRouter()
			router.Use(CORSMiddleware(envVars))

			// Create test request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			// Make request
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedCode, w.Code)

			// Check headers
			for key, expectedValue := range tt.expectedHeaders {
				assert.Equal(t, expectedValue, w.Header().Get(key), "Header %s mismatch", key)
			}

			// Verify that unauthorized requests don't have CORS headers
			if tt.expectedCode == http.StatusForbidden {
				assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
				assert.Empty(t, w.Header().Get("Access-Control-Allow-Methods"))
				assert.Empty(t, w.Header().Get("Access-Control-Allow-Headers"))
			}
		})
	}
} 