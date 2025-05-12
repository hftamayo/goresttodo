package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/stretchr/testify/assert"
)

func setupTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestValidateUserMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		user           models.User
		expectedCode   int
		expectedError  string
	}{
		{
			name: "valid user",
			user: models.User{
				FullName: "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "missing name",
			user: models.User{
				Email:    "john@example.com",
				Password: "password123",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "name is required",
		},
		{
			name: "missing email",
			user: models.User{
				FullName: "John Doe",
				Password: "password123",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "email is required",
		},
		{
			name: "missing password",
			user: models.User{
				FullName: "John Doe",
				Email:    "john@example.com",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "password is required",
		},
		{
			name: "short password",
			user: models.User{
				FullName: "John Doe",
				Email:    "john@example.com",
				Password: "12345",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "password must be at least 6 characters long",
		},
		{
			name: "invalid email",
			user: models.User{
				FullName: "John Doe",
				Email:    "invalid-email",
				Password: "password123",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "email is invalid",
		},
		{
			name: "empty user",
			user: models.User{},
			expectedCode:  http.StatusBadRequest,
			expectedError: "name is required",
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
			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedError != "" {
				assert.Equal(t, tt.expectedError, w.Body.String())
			}
		})
	}
}

func TestValidateUser(t *testing.T) {
	tests := []struct {
		name          string
		user          models.User
		expectedError string
	}{
		{
			name: "valid user",
			user: models.User{
				FullName: "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			expectedError: "",
		},
		{
			name: "missing name",
			user: models.User{
				Email:    "john@example.com",
				Password: "password123",
			},
			expectedError: "name is required",
		},
		{
			name: "missing email",
			user: models.User{
				FullName: "John Doe",
				Password: "password123",
			},
			expectedError: "email is required",
		},
		{
			name: "missing password",
			user: models.User{
				FullName: "John Doe",
				Email:    "john@example.com",
			},
			expectedError: "password is required",
		},
		{
			name: "short password",
			user: models.User{
				FullName: "John Doe",
				Email:    "john@example.com",
				Password: "12345",
			},
			expectedError: "password must be at least 6 characters long",
		},
		{
			name: "invalid email",
			user: models.User{
				FullName: "John Doe",
				Email:    "invalid-email",
				Password: "password123",
			},
			expectedError: "email is invalid",
		},
		{
			name: "empty user",
			user: models.User{},
			expectedError: "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUser(tt.user)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
} 