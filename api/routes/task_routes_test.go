package routes

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/v1/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGinEngine is a mock implementation of gin.Engine
type MockGinEngine struct {
	mock.Mock
}

func (m *MockGinEngine) Group(relativePath string, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	args := m.Called(relativePath, handlers)
	return args.Get(0).(*gin.RouterGroup)
}

// MockTaskHandler is a mock implementation of task.Handler
type MockTaskHandler struct {
	mock.Mock
}

func (m *MockTaskHandler) List(c *gin.Context) {
	m.Called(c)
}

func (m *MockTaskHandler) ListById(c *gin.Context) {
	m.Called(c)
}

func (m *MockTaskHandler) Create(c *gin.Context) {
	m.Called(c)
}

func (m *MockTaskHandler) Update(c *gin.Context) {
	m.Called(c)
}

func (m *MockTaskHandler) Done(c *gin.Context) {
	m.Called(c)
}

func (m *MockTaskHandler) Delete(c *gin.Context) {
	m.Called(c)
}

func TestSetupTaskRoutes(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*MockGinEngine, *MockTaskHandler)
	}{
		{
			name: "setup task routes",
			setupMock: func(mockEngine *MockGinEngine, mockHandler *MockTaskHandler) {
				mockEngine.On("Group", "/tasks/task", mock.Anything).Return(mockEngine)
				mockEngine.On("GET", "/list", mock.Anything).Return(mockEngine)
				mockEngine.On("GET", "/:id", mock.Anything).Return(mockEngine)
				mockEngine.On("POST", "", mock.Anything).Return(mockEngine)
				mockEngine.On("PATCH", "/:id", mock.Anything).Return(mockEngine)
				mockEngine.On("PATCH", "/:id/done", mock.Anything).Return(mockEngine)
				mockEngine.On("DELETE", "/:id", mock.Anything).Return(mockEngine)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEngine := new(MockGinEngine)
			mockHandler := new(MockTaskHandler)
			if tt.setupMock != nil {
				tt.setupMock(mockEngine, mockHandler)
			}
			SetupTaskRoutes(mockEngine, mockHandler)
			mockEngine.AssertExpectations(t)
		})
	}
} 