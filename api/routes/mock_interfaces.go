package routes

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockGinEngine is a mock implementation of gin.Engine
type MockGinEngine struct {
	mock.Mock
}

func (m *MockGinEngine) GET(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	args := m.Called(relativePath, handlers)
	return args.Get(0).(gin.IRoutes)
}

func (m *MockGinEngine) POST(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	args := m.Called(relativePath, handlers)
	return args.Get(0).(gin.IRoutes)
}

func (m *MockGinEngine) PUT(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	args := m.Called(relativePath, handlers)
	return args.Get(0).(gin.IRoutes)
}

func (m *MockGinEngine) DELETE(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	args := m.Called(relativePath, handlers)
	return args.Get(0).(gin.IRoutes)
}

func (m *MockGinEngine) PATCH(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	args := m.Called(relativePath, handlers)
	return args.Get(0).(gin.IRoutes)
}

func (m *MockGinEngine) OPTIONS(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	args := m.Called(relativePath, handlers)
	return args.Get(0).(gin.IRoutes)
}

func (m *MockGinEngine) HEAD(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	args := m.Called(relativePath, handlers)
	return args.Get(0).(gin.IRoutes)
}

func (m *MockGinEngine) Group(relativePath string, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	args := m.Called(relativePath, handlers)
	return args.Get(0).(*gin.RouterGroup)
}

func (m *MockGinEngine) Use(middleware ...gin.HandlerFunc) gin.IRoutes {
	args := m.Called(middleware)
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

// MockTaskHandler is a mock implementation of task.Handler
type MockTaskHandler struct {
	mock.Mock
}

func (m *MockTaskHandler) List(c *gin.Context) {
	m.Called(c)
}

func (m *MockTaskHandler) Get(c *gin.Context) {
	m.Called(c)
}

func (m *MockTaskHandler) Create(c *gin.Context) {
	m.Called(c)
}

func (m *MockTaskHandler) Update(c *gin.Context) {
	m.Called(c)
}

func (m *MockTaskHandler) Delete(c *gin.Context) {
	m.Called(c)
}

func (m *MockTaskHandler) ListById(c *gin.Context) {
	m.Called(c)
}

func (m *MockTaskHandler) Done(c *gin.Context) {
	m.Called(c)
}

// MockGormDB is a mock implementation of gorm.DB
type MockGormDB struct {
	mock.Mock
}

func (m *MockGormDB) Create(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(dest, conds)
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(dest, conds)
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Save(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Delete(value interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(value, conds)
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0).(*gorm.DB)
}

func (m *MockGormDB) AutoMigrate(dst ...interface{}) error {
	args := m.Called(dst)
	return args.Error(0)
}

// MockCache implements config.CacheInterface
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(key string, dest interface{}) error {
	args := m.Called(key, dest)
	return args.Error(0)
}

func (m *MockCache) Set(key string, value interface{}, ttl time.Duration) error {
	args := m.Called(key, value, ttl)
	return args.Error(0)
}

func (m *MockCache) SetWithTags(key string, value interface{}, ttl time.Duration, tags ...string) error {
	args := m.Called(key, value, ttl, tags)
	return args.Error(0)
}

func (m *MockCache) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockCache) InvalidateByTags(tags ...string) error {
	args := m.Called(tags)
	return args.Error(0)
}

// MockErrorLogger implements config.ErrorLogger
type MockErrorLogger struct {
	mock.Mock
}

func (m *MockErrorLogger) LogError(ctx context.Context, service, operation, errorMsg string, metadata map[string]interface{}) error {
	args := m.Called(ctx, service, operation, errorMsg, metadata)
	return args.Error(0)
}

func (m *MockErrorLogger) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockRedisClient is a mock implementation of redis.Client
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) Incr(ctx context.Context, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
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

func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}
