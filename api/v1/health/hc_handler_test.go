package health

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockDB is a mock implementation of gorm.DB
type MockDB struct {
	mock.Mock
}

func (m *MockDB) DB() (*sql.DB, error) {
	args := m.Called()
	return args.Get(0).(*sql.DB), args.Error(1)
}

// MockSQLDB is a mock implementation of sql.DB
type MockSQLDB struct {
	mock.Mock
}

func (m *MockSQLDB) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func setupTestRouter(handler *HealthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/tasks/healthcheck/app", handler.AppStatus)
	router.GET("/tasks/healthcheck/db", handler.DbStatus)
	return router
}

func TestHealthHandler_AppStatus(t *testing.T) {
	// Create a new handler with a mock DB
	mockDB := new(MockDB)
	handler := NewHealthHandler(mockDB)
	router := setupTestRouter(handler)

	// Test the AppStatus endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tasks/healthcheck/app", nil)
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response AppHealthDetails
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify response fields
	assert.NotEmpty(t, response.Timestamp)
	assert.Greater(t, response.Uptime, 0.0)
	assert.Greater(t, response.MemoryUsage.Total, uint64(0))
	assert.Greater(t, response.MemoryUsage.Free, uint64(0))
	assert.Greater(t, response.StartTime, int64(0))
}

func TestHealthHandler_DbStatus(t *testing.T) {
	tests := []struct {
		name           string
		setupMock     func(*MockDB, *MockSQLDB)
		expectedCode  int
		expectedError string
	}{
		{
			name: "successful database check",
			setupMock: func(mockDB *MockDB, mockSQLDB *MockSQLDB) {
				mockDB.On("DB").Return(mockSQLDB, nil)
				mockSQLDB.On("Ping").Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "database connection error",
			setupMock: func(mockDB *MockDB, mockSQLDB *MockSQLDB) {
				mockDB.On("DB").Return(nil, assert.AnError)
			},
			expectedCode:  http.StatusServiceUnavailable,
			expectedError: "Database connection error: " + assert.AnError.Error(),
		},
		{
			name: "database ping error",
			setupMock: func(mockDB *MockDB, mockSQLDB *MockSQLDB) {
				mockDB.On("DB").Return(mockSQLDB, nil)
				mockSQLDB.On("Ping").Return(assert.AnError)
			},
			expectedCode:  http.StatusServiceUnavailable,
			expectedError: "Database ping failed: " + assert.AnError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock objects
			mockDB := new(MockDB)
			mockSQLDB := new(MockSQLDB)
			tt.setupMock(mockDB, mockSQLDB)

			// Create handler and router
			handler := NewHealthHandler(mockDB)
			router := setupTestRouter(handler)

			// Make request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/tasks/healthcheck/db", nil)
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedError != "" {
				var response DbHealthDetails
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Error)
				assert.Equal(t, "error", response.DatabaseStatus)
			} else {
				var response DbHealthDetails
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "healthy", response.DatabaseStatus)
				assert.Greater(t, response.ConnectionTime, 0.0)
				assert.NotEmpty(t, response.Timestamp)
			}
		})
	}
} 