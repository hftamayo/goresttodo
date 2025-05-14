package routes

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/v1/health"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGinEngine is a mock implementation of gin.Engine
type MockGinEngine struct {
	mock.Mock
}

func (m *MockGinEngine) GET(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	args := m.Called(relativePath, handlers)
	return args.Get(0).(gin.IRoutes)
}

// MockHealthHandler is a mock implementation of health.HealthHandler
type MockHealthHandler struct {
	mock.Mock
}

func (m *MockHealthHandler) AppStatus(c *gin.Context) {
	m.Called(c)
}

func (m *MockHealthHandler) DbStatus(c *gin.Context) {
	m.Called(c)
}

func TestSetupHealthCheckRoutes(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*MockGinEngine, *MockHealthHandler)
	}{
		{
			name: "setup health check routes",
			setupMock: func(mockEngine *MockGinEngine, mockHandler *MockHealthHandler) {
				mockEngine.On("GET", "/tasks/healthcheck/app", mock.Anything).Return(mockEngine)
				mockEngine.On("GET", "/tasks/healthcheck/db", mock.Anything).Return(mockEngine)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEngine := new(MockGinEngine)
			mockHandler := new(MockHealthHandler)
			if tt.setupMock != nil {
				tt.setupMock(mockEngine, mockHandler)
			}
			SetupHealthCheckRoutes(mockEngine, mockHandler)
			mockEngine.AssertExpectations(t)
		})
	}
} 