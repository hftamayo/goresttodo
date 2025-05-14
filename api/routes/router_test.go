package routes

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/v1/errorlog"
	"github.com/hftamayo/gotodo/api/v1/health"
	"github.com/hftamayo/gotodo/api/v1/task"
	"github.com/hftamayo/gotodo/pkg/utils"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockGinEngine is a mock implementation of gin.Engine
type MockGinEngine struct {
	mock.Mock
}

func (m *MockGinEngine) Group(relativePath string, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	args := m.Called(relativePath, handlers)
	return args.Get(0).(*gin.RouterGroup)
}

func (m *MockGinEngine) GET(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	args := m.Called(relativePath, handlers)
	return args.Get(0).(gin.IRoutes)
}

// MockGormDB is a mock implementation of gorm.DB
type MockGormDB struct {
	mock.Mock
}

// MockRedisClient is a mock implementation of redis.Client
type MockRedisClient struct {
	mock.Mock
}

// MockCache is a mock implementation of utils.Cache
type MockCache struct {
	mock.Mock
}

func TestSetupRouter(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*MockGinEngine, *MockGormDB, *MockRedisClient, *MockCache)
	}{
		{
			name: "setup router",
			setupMock: func(mockEngine *MockGinEngine, mockDB *MockGormDB, mockRedis *MockRedisClient, mockCache *MockCache) {
				mockEngine.On("Group", "/tasks/task", mock.Anything).Return(mockEngine)
				mockEngine.On("GET", "/list", mock.Anything).Return(mockEngine)
				mockEngine.On("GET", "/:id", mock.Anything).Return(mockEngine)
				mockEngine.On("POST", "", mock.Anything).Return(mockEngine)
				mockEngine.On("PATCH", "/:id", mock.Anything).Return(mockEngine)
				mockEngine.On("PATCH", "/:id/done", mock.Anything).Return(mockEngine)
				mockEngine.On("DELETE", "/:id", mock.Anything).Return(mockEngine)
				mockEngine.On("GET", "/tasks/healthcheck/app", mock.Anything).Return(mockEngine)
				mockEngine.On("GET", "/tasks/healthcheck/db", mock.Anything).Return(mockEngine)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEngine := new(MockGinEngine)
			mockDB := new(MockGormDB)
			mockRedis := new(MockRedisClient)
			mockCache := new(MockCache)
			if tt.setupMock != nil {
				tt.setupMock(mockEngine, mockDB, mockRedis, mockCache)
			}
			SetupRouter(mockEngine, mockDB, mockRedis, mockCache)
			mockEngine.AssertExpectations(t)
		})
	}
} 