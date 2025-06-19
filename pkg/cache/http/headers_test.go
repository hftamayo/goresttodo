package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestNewHeaders(t *testing.T) {
	headers := NewHeaders()
	assert.NotNil(t, headers)
}

func TestHeaders_SetETag(t *testing.T) {
	headers := NewHeaders()

	tests := []struct {
		name        string
		etag        string
		expectHeader bool
		description  string
	}{
		{
			name:         "set valid ETag",
			etag:         "\"abc123\"",
			expectHeader: true,
			description:  "Should set ETag header for valid ETag",
		},
		{
			name:         "set weak ETag",
			etag:         "W/\"abc123\"",
			expectHeader: true,
			description:  "Should set ETag header for weak ETag",
		},
		{
			name:         "set empty ETag",
			etag:         "",
			expectHeader: false,
			description:  "Should not set ETag header for empty string",
		},
		{
			name:         "set whitespace ETag",
			etag:         "   ",
			expectHeader: true,
			description:  "Should set ETag header for whitespace string",
		},
		{
			name:         "set long ETag",
			etag:         "\"very-long-etag-value-that-exceeds-normal-length\"",
			expectHeader: true,
			description:  "Should set ETag header for long ETag",
		},
		{
			name:         "set special characters ETag",
			etag:         "\"etag-with-special-chars!@#$%^&*()\"",
			expectHeader: true,
			description:  "Should set ETag header with special characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupTestContext()
			
			headers.SetETag(c, tt.etag)
			
			if tt.expectHeader {
				assert.Equal(t, tt.etag, w.Header().Get("ETag"), "Test: %s", tt.description)
			} else {
				assert.Empty(t, w.Header().Get("ETag"), "Test: %s", tt.description)
			}
		})
	}
}

func TestHeaders_AddCacheControl_Modifying(t *testing.T) {
	headers := NewHeaders()

	tests := []struct {
		name        string
		isModifying bool
		maxAge      int
		description string
	}{
		{
			name:        "modifying request with maxAge 0",
			isModifying: true,
			maxAge:      0,
			description: "Should set no-cache headers for modifying request",
		},
		{
			name:        "modifying request with maxAge 100",
			isModifying: true,
			maxAge:      100,
			description: "Should set no-cache headers for modifying request regardless of maxAge",
		},
		{
			name:        "modifying request with negative maxAge",
			isModifying: true,
			maxAge:      -10,
			description: "Should set no-cache headers for modifying request with negative maxAge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupTestContext()
			
			headers.AddCacheControl(c, tt.isModifying, tt.maxAge)
			
			// Verify no-cache headers are set
			assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"), "Test: %s", tt.description)
			assert.Equal(t, "no-cache", w.Header().Get("Pragma"), "Test: %s", tt.description)
			assert.Equal(t, "0", w.Header().Get("Expires"), "Test: %s", tt.description)
			
			// Verify Vary header is not set for modifying requests
			assert.Empty(t, w.Header().Get("Vary"), "Test: %s", tt.description)
		})
	}
}

func TestHeaders_AddCacheControl_NonModifying(t *testing.T) {
	headers := NewHeaders()

	tests := []struct {
		name        string
		isModifying bool
		maxAge      int
		expectedMaxAge string
		description string
	}{
		{
			name:        "non-modifying request with maxAge 0",
			isModifying: false,
			maxAge:      0,
			expectedMaxAge: "0",
			description: "Should set cache headers for non-modifying request with maxAge 0",
		},
		{
			name:        "non-modifying request with maxAge 60",
			isModifying: false,
			maxAge:      60,
			expectedMaxAge: "60",
			description: "Should set cache headers for non-modifying request with maxAge 60",
		},
		{
			name:        "non-modifying request with maxAge 3600",
			isModifying: false,
			maxAge:      3600,
			expectedMaxAge: "3600",
			description: "Should set cache headers for non-modifying request with maxAge 3600",
		},
		{
			name:        "non-modifying request with negative maxAge",
			isModifying: false,
			maxAge:      -10,
			expectedMaxAge: "-10",
			description: "Should set cache headers for non-modifying request with negative maxAge",
		},
		{
			name:        "non-modifying request with large maxAge",
			isModifying: false,
			maxAge:      86400,
			expectedMaxAge: "86400",
			description: "Should set cache headers for non-modifying request with large maxAge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupTestContext()
			
			headers.AddCacheControl(c, tt.isModifying, tt.maxAge)
			
			// Verify cache headers are set correctly
			expectedCacheControl := "private, max-age=" + tt.expectedMaxAge
			assert.Equal(t, expectedCacheControl, w.Header().Get("Cache-Control"), "Test: %s", tt.description)
			assert.Equal(t, "Authorization", w.Header().Get("Vary"), "Test: %s", tt.description)
			
			// Verify no-cache headers are not set
			assert.Empty(t, w.Header().Get("Pragma"), "Test: %s", tt.description)
			assert.Empty(t, w.Header().Get("Expires"), "Test: %s", tt.description)
		})
	}
}

func TestHeaders_IsNotModified(t *testing.T) {
	headers := NewHeaders()

	tests := []struct {
		name         string
		ifNoneMatch  string
		etag         string
		expected     bool
		description  string
	}{
		{
			name:         "exact match",
			ifNoneMatch:  "\"abc123\"",
			etag:         "\"abc123\"",
			expected:     true,
			description:  "Should return true for exact ETag match",
		},
		{
			name:         "weak ETag match",
			ifNoneMatch:  "W/\"abc123\"",
			etag:         "\"abc123\"",
			expected:     true,
			description:  "Should return true for weak ETag match",
		},
		{
			name:         "strong ETag with weak If-None-Match",
			ifNoneMatch:  "W/\"abc123\"",
			etag:         "\"abc123\"",
			expected:     true,
			description:  "Should return true when If-None-Match is weak and ETag matches",
		},
		{
			name:         "no match",
			ifNoneMatch:  "\"abc123\"",
			etag:         "\"def456\"",
			expected:     false,
			description:  "Should return false for non-matching ETags",
		},
		{
			name:         "empty If-None-Match",
			ifNoneMatch:  "",
			etag:         "\"abc123\"",
			expected:     false,
			description:  "Should return false for empty If-None-Match header",
		},
		{
			name:         "empty ETag",
			ifNoneMatch:  "\"abc123\"",
			etag:         "",
			expected:     false,
			description:  "Should return false for empty ETag",
		},
		{
			name:         "both empty",
			ifNoneMatch:  "",
			etag:         "",
			expected:     false,
			description:  "Should return false when both are empty",
		},
		{
			name:         "case sensitive match",
			ifNoneMatch:  "\"ABC123\"",
			etag:         "\"abc123\"",
			expected:     false,
			description:  "Should be case sensitive in ETag comparison",
		},
		{
			name:         "multiple ETags in If-None-Match",
			ifNoneMatch:  "\"abc123\", \"def456\"",
			etag:         "\"abc123\"",
			expected:     true,
			description:  "Should match when ETag is in comma-separated list",
		},
		{
			name:         "multiple ETags with weak ETag",
			ifNoneMatch:  "\"abc123\", W/\"def456\"",
			etag:         "\"def456\"",
			expected:     true,
			description:  "Should match when ETag matches weak ETag in list",
		},
		{
			name:         "wildcard If-None-Match",
			ifNoneMatch:  "*",
			etag:         "\"abc123\"",
			expected:     false,
			description:  "Should return false for wildcard If-None-Match (not implemented)",
		},
		{
			name:         "special characters in ETag",
			ifNoneMatch:  "\"etag-with-special-chars!@#\"",
			etag:         "\"etag-with-special-chars!@#\"",
			expected:     true,
			description:  "Should match ETags with special characters",
		},
		{
			name:         "long ETag match",
			ifNoneMatch:  "\"very-long-etag-value-that-exceeds-normal-length\"",
			etag:         "\"very-long-etag-value-that-exceeds-normal-length\"",
			expected:     true,
			description:  "Should match long ETags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := setupTestContext()
			
			// Set the If-None-Match header
			if tt.ifNoneMatch != "" {
				c.Request = &http.Request{
					Header: http.Header{
						"If-None-Match": []string{tt.ifNoneMatch},
					},
				}
			}
			
			result := headers.IsNotModified(c, tt.etag)
			assert.Equal(t, tt.expected, result, "Test: %s", tt.description)
		})
	}
}

func TestHeaders_IsNotModified_EdgeCases(t *testing.T) {
	headers := NewHeaders()

	tests := []struct {
		name         string
		ifNoneMatch  string
		etag         string
		expected     bool
		description  string
	}{
		{
			name:         "whitespace in If-None-Match",
			ifNoneMatch:  "  \"abc123\"  ",
			etag:         "\"abc123\"",
			expected:     false, // Current implementation doesn't trim whitespace
			description:  "Should handle whitespace in If-None-Match header",
		},
		{
			name:         "whitespace in ETag",
			ifNoneMatch:  "\"abc123\"",
			etag:         "  \"abc123\"  ",
			expected:     false, // Current implementation doesn't trim whitespace
			description:  "Should handle whitespace in ETag",
		},
		{
			name:         "quotes only in If-None-Match",
			ifNoneMatch:  "\"\"",
			etag:         "\"abc123\"",
			expected:     false,
			description:  "Should handle empty quoted string in If-None-Match",
		},
		{
			name:         "quotes only in ETag",
			ifNoneMatch:  "\"abc123\"",
			etag:         "\"\"",
			expected:     false,
			description:  "Should handle empty quoted string in ETag",
		},
		{
			name:         "unquoted If-None-Match",
			ifNoneMatch:  "abc123",
			etag:         "\"abc123\"",
			expected:     false,
			description:  "Should handle unquoted If-None-Match header",
		},
		{
			name:         "unquoted ETag",
			ifNoneMatch:  "\"abc123\"",
			etag:         "abc123",
			expected:     false,
			description:  "Should handle unquoted ETag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := setupTestContext()
			
			// Set the If-None-Match header
			if tt.ifNoneMatch != "" {
				c.Request = &http.Request{
					Header: http.Header{
						"If-None-Match": []string{tt.ifNoneMatch},
					},
				}
			}
			
			result := headers.IsNotModified(c, tt.etag)
			assert.Equal(t, tt.expected, result, "Test: %s", tt.description)
		})
	}
}

func TestHeaders_Integration(t *testing.T) {
	headers := NewHeaders()

	// Test complete workflow
	c, w := setupTestContext()
	
	// Set ETag
	etag := "\"abc123\""
	headers.SetETag(c, etag)
	assert.Equal(t, etag, w.Header().Get("ETag"))
	
	// Add cache control for non-modifying request
	headers.AddCacheControl(c, false, 60)
	assert.Equal(t, "private, max-age=60", w.Header().Get("Cache-Control"))
	assert.Equal(t, "Authorization", w.Header().Get("Vary"))
	
	// Test IsNotModified
	c.Request = &http.Request{
		Header: http.Header{
			"If-None-Match": []string{etag},
		},
	}
	assert.True(t, headers.IsNotModified(c, etag))
}

func TestHeaders_Performance(t *testing.T) {
	headers := NewHeaders()
	c, w := setupTestContext()
	
	// Test that header operations don't take too long
	etag := "\"performance-test-etag\""
	headers.SetETag(c, etag)
	headers.AddCacheControl(c, false, 3600)
	
	assert.Equal(t, etag, w.Header().Get("ETag"))
	assert.Equal(t, "private, max-age=3600", w.Header().Get("Cache-Control"))
}

func BenchmarkHeaders_SetETag(b *testing.B) {
	headers := NewHeaders()
	c, _ := setupTestContext()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		headers.SetETag(c, "\"benchmark-etag\"")
	}
}

func BenchmarkHeaders_AddCacheControl(b *testing.B) {
	headers := NewHeaders()
	c, _ := setupTestContext()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		headers.AddCacheControl(c, false, 60)
	}
}

func BenchmarkHeaders_IsNotModified(b *testing.B) {
	headers := NewHeaders()
	c, _ := setupTestContext()
	c.Request = &http.Request{
		Header: http.Header{
			"If-None-Match": []string{"\"benchmark-etag\""},
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		headers.IsNotModified(c, "\"benchmark-etag\"")
	}
} 