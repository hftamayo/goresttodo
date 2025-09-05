package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/stretchr/testify/assert"
)

func setupTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})
}

func TestValidateUserMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		user           models.User
		expectedCode   int
		expectedError  string
		description    string
	}{
		{
			name: "valid user with all fields",
			user: models.User{
				FullName: "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			expectedCode: http.StatusOK,
			description:  "Complete valid user data should pass validation",
		},
		{
			name: "valid user with minimum password length",
			user: models.User{
				FullName: "Jane Smith",
				Email:    "jane@example.com",
				Password: "123456",
			},
			expectedCode: http.StatusOK,
			description:  "User with exactly 6 character password should pass",
		},
		{
			name: "missing full name",
			user: models.User{
				Email:    "john@example.com",
				Password: "password123",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "name is required",
			description:  "User without full name should be rejected",
		},
		{
			name: "empty full name",
			user: models.User{
				FullName: "",
				Email:    "john@example.com",
				Password: "password123",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "name is required",
			description:  "User with empty full name should be rejected",
		},
		{
			name: "whitespace only full name",
			user: models.User{
				FullName: "   ",
				Email:    "john@example.com",
				Password: "password123",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "name is required",
			description:  "User with whitespace-only full name should be rejected",
		},
		{
			name: "missing email",
			user: models.User{
				FullName: "John Doe",
				Password: "password123",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "email is required",
			description:  "User without email should be rejected",
		},
		{
			name: "empty email",
			user: models.User{
				FullName: "John Doe",
				Email:    "",
				Password: "password123",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "email is required",
			description:  "User with empty email should be rejected",
		},
		{
			name: "missing password",
			user: models.User{
				FullName: "John Doe",
				Email:    "john@example.com",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "password is required",
			description:  "User without password should be rejected",
		},
		{
			name: "empty password",
			user: models.User{
				FullName: "John Doe",
				Email:    "john@example.com",
				Password: "",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "password is required",
			description:  "User with empty password should be rejected",
		},
		{
			name: "password too short",
			user: models.User{
				FullName: "John Doe",
				Email:    "john@example.com",
				Password: "12345",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "password must be at least 6 characters long",
			description:  "Password with less than 6 characters should be rejected",
		},
		{
			name: "password exactly 5 characters",
			user: models.User{
				FullName: "John Doe",
				Email:    "john@example.com",
				Password: "12345",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "password must be at least 6 characters long",
			description:  "Password with exactly 5 characters should be rejected",
		},
		{
			name: "invalid email format",
			user: models.User{
				FullName: "John Doe",
				Email:    "invalid-email",
				Password: "password123",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "email is invalid",
			description:  "Email without @ symbol should be rejected",
		},
		{
			name: "email with only @",
			user: models.User{
				FullName: "John Doe",
				Email:    "@",
				Password: "password123",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "email is invalid",
			description:  "Email with only @ symbol should be rejected",
		},
		{
			name: "email starting with @",
			user: models.User{
				FullName: "John Doe",
				Email:    "@example.com",
				Password: "password123",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "email is invalid",
			description:  "Email starting with @ should be rejected",
		},
		{
			name: "email ending with @",
			user: models.User{
				FullName: "John Doe",
				Email:    "john@",
				Password: "password123",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "email is invalid",
			description:  "Email ending with @ should be rejected",
		},
		{
			name: "empty user struct",
			user: models.User{},
			expectedCode:  http.StatusBadRequest,
			expectedError: "name is required",
			description:  "Completely empty user should be rejected",
		},
		{
			name: "multiple validation errors - name and email missing",
			user: models.User{
				Password: "password123",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "name is required",
			description:  "First validation error should be returned",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test request
			body, _ := json.Marshal(tt.user)
			req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create middleware chain
			handler := ValidateUserMiddleware(setupTestHandler())

			// Make request
			handler.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedCode, w.Code, "Test: %s", tt.description)

			if tt.expectedError != "" {
				assert.Equal(t, tt.expectedError, strings.TrimSpace(w.Body.String()), 
					"Error message mismatch for test: %s", tt.description)
			} else {
				assert.Equal(t, "success", w.Body.String(), 
					"Expected success response for test: %s", tt.description)
			}
		})
	}
}

func TestValidateUserMiddleware_InvalidJSON(t *testing.T) {
	tests := []struct {
		name          string
		requestBody   string
		expectedCode  int
		description   string
	}{
		{
			name:         "malformed JSON",
			requestBody:  `{"fullname": "John Doe", "email": "john@example.com", "password": "password123"`,
			expectedCode: http.StatusBadRequest,
			description:  "Request with malformed JSON should be rejected",
		},
		{
			name:         "empty request body",
			requestBody:  "",
			expectedCode: http.StatusBadRequest,
			description:  "Request with empty body should be rejected",
		},
		{
			name:         "null JSON",
			requestBody:  "null",
			expectedCode: http.StatusBadRequest,
			description:  "Request with null JSON should be rejected",
		},
		{
			name:         "invalid JSON structure",
			requestBody:  `{"invalid": "structure"}`,
			expectedCode: http.StatusBadRequest,
			description:  "Request with invalid structure should be rejected at JSON decoding",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler := ValidateUserMiddleware(setupTestHandler())

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code, "Test: %s", tt.description)
		})
	}
}

func TestValidateUserMiddleware_HTTPMethods(t *testing.T) {
	validUser := models.User{
		FullName: "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}

	for _, method := range methods {
		t.Run(fmt.Sprintf("method_%s", method), func(t *testing.T) {
			body, _ := json.Marshal(validUser)
			req := httptest.NewRequest(method, "/test", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler := ValidateUserMiddleware(setupTestHandler())

			handler.ServeHTTP(w, req)

			// All methods should work the same way
			assert.Equal(t, http.StatusOK, w.Code, "Method %s should work", method)
		})
	}
}

func TestValidateUser(t *testing.T) {
	tests := []struct {
		name          string
		user          models.User
		expectedError string
		description   string
	}{
		{
			name: "valid user with all fields",
			user: models.User{
				FullName: "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			expectedError: "",
			description:   "Complete valid user should pass validation",
		},
		{
			name: "valid user with minimum password",
			user: models.User{
				FullName: "Jane Smith",
				Email:    "jane@example.com",
				Password: "123456",
			},
			expectedError: "",
			description:   "User with exactly 6 character password should pass",
		},
		{
			name: "valid user with long password",
			user: models.User{
				FullName: "Admin User",
				Email:    "admin@example.com",
				Password: "very-long-password-with-special-chars!@#$%^&*()",
			},
			expectedError: "",
			description:   "User with long password should pass",
		},
		{
			name: "valid user with special characters in name",
			user: models.User{
				FullName: "José María O'Connor-Smith",
				Email:    "jose@example.com",
				Password: "password123",
			},
			expectedError: "",
			description:   "User with special characters in name should pass",
		},
		{
			name: "valid user with complex email",
			user: models.User{
				FullName: "Test User",
				Email:    "test.user+tag@subdomain.example.co.uk",
				Password: "password123",
			},
			expectedError: "",
			description:   "User with complex email should pass",
		},
		{
			name: "missing full name",
			user: models.User{
				Email:    "john@example.com",
				Password: "password123",
			},
			expectedError: "name is required",
			description:   "User without full name should fail validation",
		},
		{
			name: "empty full name",
			user: models.User{
				FullName: "",
				Email:    "john@example.com",
				Password: "password123",
			},
			expectedError: "name is required",
			description:   "User with empty full name should fail validation",
		},
		{
			name: "whitespace only full name",
			user: models.User{
				FullName: "   ",
				Email:    "john@example.com",
				Password: "password123",
			},
			expectedError: "name is required",
			description:   "User with whitespace-only full name should fail validation",
		},
		{
			name: "missing email",
			user: models.User{
				FullName: "John Doe",
				Password: "password123",
			},
			expectedError: "email is required",
			description:   "User without email should fail validation",
		},
		{
			name: "empty email",
			user: models.User{
				FullName: "John Doe",
				Email:    "",
				Password: "password123",
			},
			expectedError: "email is required",
			description:   "User with empty email should fail validation",
		},
		{
			name: "missing password",
			user: models.User{
				FullName: "John Doe",
				Email:    "john@example.com",
			},
			expectedError: "password is required",
			description:   "User without password should fail validation",
		},
		{
			name: "empty password",
			user: models.User{
				FullName: "John Doe",
				Email:    "john@example.com",
				Password: "",
			},
			expectedError: "password is required",
			description:   "User with empty password should fail validation",
		},
		{
			name: "password too short",
			user: models.User{
				FullName: "John Doe",
				Email:    "john@example.com",
				Password: "12345",
			},
			expectedError: "password must be at least 6 characters long",
			description:   "Password with less than 6 characters should fail validation",
		},
		{
			name: "password exactly 5 characters",
			user: models.User{
				FullName: "John Doe",
				Email:    "john@example.com",
				Password: "12345",
			},
			expectedError: "password must be at least 6 characters long",
			description:   "Password with exactly 5 characters should fail validation",
		},
		{
			name: "invalid email format",
			user: models.User{
				FullName: "John Doe",
				Email:    "invalid-email",
				Password: "password123",
			},
			expectedError: "email is invalid",
			description:   "Email without @ symbol should fail validation",
		},
		{
			name: "email with only @",
			user: models.User{
				FullName: "John Doe",
				Email:    "@",
				Password: "password123",
			},
			expectedError: "email is invalid",
			description:   "Email with only @ symbol should fail validation",
		},
		{
			name: "email starting with @",
			user: models.User{
				FullName: "John Doe",
				Email:    "@example.com",
				Password: "password123",
			},
			expectedError: "email is invalid",
			description:   "Email starting with @ should fail validation",
		},
		{
			name: "email ending with @",
			user: models.User{
				FullName: "John Doe",
				Email:    "john@",
				Password: "password123",
			},
			expectedError: "email is invalid",
			description:   "Email ending with @ should fail validation",
		},
		{
			name: "empty user struct",
			user: models.User{},
			expectedError: "name is required",
			description:   "Completely empty user should fail validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUser(tt.user)
			
			if tt.expectedError != "" {
				assert.Error(t, err, "Test: %s", tt.description)
				assert.Equal(t, tt.expectedError, err.Error(), "Test: %s", tt.description)
			} else {
				assert.NoError(t, err, "Test: %s", tt.description)
			}
		})
	}
}

func TestValidateUser_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		user          models.User
		expectedError string
		description   string
	}{
		{
			name: "very long full name",
			user: models.User{
				FullName: strings.Repeat("a", 100),
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedError: "",
			description:   "Very long full name should pass validation",
		},
		{
			name: "very long email",
			user: models.User{
				FullName: "Test User",
				Email:    strings.Repeat("a", 50) + "@example.com",
				Password: "password123",
			},
			expectedError: "",
			description:   "Very long email should pass validation",
		},
		{
			name: "very long password",
			user: models.User{
				FullName: "Test User",
				Email:    "test@example.com",
				Password: strings.Repeat("a", 1000),
			},
			expectedError: "",
			description:   "Very long password should pass validation",
		},
		{
			name: "unicode characters in name",
			user: models.User{
				FullName: "José María O'Connor-Smith 测试",
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedError: "",
			description:   "Unicode characters in name should pass validation",
		},
		{
			name: "unicode characters in email",
			user: models.User{
				FullName: "Test User",
				Email:    "test@exámple.com",
				Password: "password123",
			},
			expectedError: "",
			description:   "Unicode characters in email should pass validation",
		},
		{
			name: "unicode characters in password",
			user: models.User{
				FullName: "Test User",
				Email:    "test@example.com",
				Password: "pásswörd测试123",
			},
			expectedError: "",
			description:   "Unicode characters in password should pass validation",
		},
		{
			name: "null bytes in name",
			user: models.User{
				FullName: string([]byte{0}) + "John Doe",
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedError: "",
			description:   "Null bytes in name should pass validation",
		},
		{
			name: "null bytes in email",
			user: models.User{
				FullName: "Test User",
				Email:    string([]byte{0}) + "test@example.com",
				Password: "password123",
			},
			expectedError: "",
			description:   "Null bytes in email should pass validation",
		},
		{
			name: "null bytes in password",
			user: models.User{
				FullName: "Test User",
				Email:    "test@example.com",
				Password: string([]byte{0}) + "password123",
			},
			expectedError: "",
			description:   "Null bytes in password should pass validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUser(tt.user)
			
			if tt.expectedError != "" {
				assert.Error(t, err, "Test: %s", tt.description)
				assert.Equal(t, tt.expectedError, err.Error(), "Test: %s", tt.description)
			} else {
				assert.NoError(t, err, "Test: %s", tt.description)
			}
		})
	}
}

func TestValidateUser_Performance(t *testing.T) {
	validUser := models.User{
		FullName: "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	// Test that validation doesn't take too long
	err := validateUser(validUser)
	assert.NoError(t, err, "Validation should complete successfully")

	// Test with large data
	largeUser := models.User{
		FullName: strings.Repeat("a", 1000),
		Email:    strings.Repeat("a", 500) + "@example.com",
		Password: strings.Repeat("a", 1000),
	}

	err = validateUser(largeUser)
	assert.NoError(t, err, "Validation with large data should complete successfully")
}

func BenchmarkValidateUser(b *testing.B) {
	user := models.User{
		FullName: "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateUser(user)
	}
}

func BenchmarkValidateUser_Invalid(b *testing.B) {
	user := models.User{
		FullName: "",
		Email:    "invalid-email",
		Password: "123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateUser(user)
	}
}

func BenchmarkValidateUserMiddleware(b *testing.B) {
	user := models.User{
		FullName: "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	body, _ := json.Marshal(user)
	handler := ValidateUserMiddleware(setupTestHandler())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
} 