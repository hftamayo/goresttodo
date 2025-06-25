package task

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/v1/errorlog"
	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockTaskServiceInterface mocks the TaskServiceInterface
type MockTaskServiceInterface struct {
	mock.Mock
}

func (m *MockTaskServiceInterface) List(cursor string, limit int, order string) ([]*models.Task, string, string, int64, error) {
	args := m.Called(cursor, limit, order)
	return args.Get(0).([]*models.Task), args.String(1), args.String(2), args.Get(3).(int64), args.Error(4)
}

func (m *MockTaskServiceInterface) ListByPage(page, limit int, order string) ([]*models.Task, int64, error) {
	args := m.Called(page, limit, order)
	return args.Get(0).([]*models.Task), args.Get(1).(int64), args.Error(2)
}

func (m *MockTaskServiceInterface) ListById(id int) (*models.Task, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskServiceInterface) Create(task *models.Task) (*models.Task, error) {
	args := m.Called(task)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskServiceInterface) Update(id int, task *models.Task) (*models.Task, error) {
	args := m.Called(id, task)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskServiceInterface) MarkAsDone(id int) (*models.Task, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskServiceInterface) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockErrorLogRepository mocks the ErrorLogRepository
type MockErrorLogRepository struct {
	mock.Mock
}

func (m *MockErrorLogRepository) LogError(operation string, err error) error {
	args := m.Called(operation, err)
	return args.Error(0)
}

// MockCache mocks the Cache
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Set(key string, value interface{}, ttl time.Duration) error {
	args := m.Called(key, value, ttl)
	return args.Error(0)
}

func (m *MockCache) Get(key string, dest interface{}) error {
	args := m.Called(key, dest)
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

// TestCache is a test double for utils.Cache that avoids nil pointer dereference
type TestCache struct {
	*utils.Cache
}

// NewTestCache creates a new test cache instance
func NewTestCache() *TestCache {
	return &TestCache{}
}

// Get overrides the Get method to return an error (simulating cache miss)
func (tc *TestCache) Get(key string, dest interface{}) error {
	return errors.New("cache miss")
}

// Set overrides the Set method to do nothing (simulating successful cache set)
func (tc *TestCache) Set(key string, value interface{}, expiration time.Duration) error {
	return nil
}

// Delete overrides the Delete method to do nothing
func (tc *TestCache) Delete(key string) error {
	return nil
}

// InvalidateByTags overrides the InvalidateByTags method to do nothing
func (tc *TestCache) InvalidateByTags(tags ...string) error {
	return nil
}

func TestHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		query          string
		setupMocks     func(*MockTaskServiceInterface, *MockErrorLogRepository)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:  "successful list",
			query: "?page=1&limit=10&order=desc",
			setupMocks: func(mockService *MockTaskServiceInterface, mockErrorLogRepo *MockErrorLogRepository) {
				tasks := []*models.Task{
					{
						Model: gorm.Model{
							ID:        1,
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Title:       "Test Task 1",
						Description: "Test Description 1",
						Done:        false,
						Owner:       1,
					},
				}
				mockService.On("ListByPage", 1, 10, "desc").Return(tasks, int64(1), nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"code":          float64(200),
				"resultMessage": "OPERATION_SUCCESS",
			},
		},
		{
			name:  "invalid pagination parameters - sets defaults",
			query: "?page=0&limit=0",
			setupMocks: func(mockService *MockTaskServiceInterface, mockErrorLogRepo *MockErrorLogRepository) {
				// The List method sets defaults: page=1, limit=10, order=desc
				mockService.On("ListByPage", 1, 10, "desc").Return([]*models.Task{}, int64(0), nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"code":          float64(200),
				"resultMessage": "OPERATION_SUCCESS",
			},
		},
		{
			name:  "service error",
			query: "?page=1&limit=10",
			setupMocks: func(mockService *MockTaskServiceInterface, mockErrorLogRepo *MockErrorLogRepository) {
				mockService.On("ListByPage", 1, 10, "desc").Return([]*models.Task{}, int64(0), errors.New("database error"))
				mockErrorLogRepo.On("LogError", "Task_list", mock.AnythingOfType("*errors.errorString")).Return(nil)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"code":          float64(500),
				"resultMessage": "OPERATION_FAILED",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockTaskServiceInterface{}
			testCache := NewTestCache()
			mockErrorLogRepo := &MockErrorLogRepository{}
			mockErrorLogService := errorlog.NewErrorLogService(mockErrorLogRepo)

			tt.setupMocks(mockService, mockErrorLogRepo)

			handler := NewHandler(mockService, mockErrorLogService, testCache)

			req, _ := http.NewRequest("GET", "/tasks"+tt.query, nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.List(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedBody["code"], response["code"])
			assert.Equal(t, tt.expectedBody["resultMessage"], response["resultMessage"])

			mockService.AssertExpectations(t)
			mockErrorLogRepo.AssertExpectations(t)
		})
	}
}

func TestHandler_ListById(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		taskID         string
		setupMocks     func(*MockTaskServiceInterface, *MockErrorLogRepository)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:   "successful get by id",
			taskID: "1",
			setupMocks: func(mockService *MockTaskServiceInterface, mockErrorLogRepo *MockErrorLogRepository) {
				task := &models.Task{
					Model: gorm.Model{
						ID:        1,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Title:       "Test Task",
					Description: "Test Description",
					Done:        false,
					Owner:       1,
				}
				mockService.On("ListById", 1).Return(task, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"code":          float64(200),
				"resultMessage": "OPERATION_SUCCESS",
			},
		},
		{
			name:   "invalid id parameter",
			taskID: "invalid",
			setupMocks: func(mockService *MockTaskServiceInterface, mockErrorLogRepo *MockErrorLogRepository) {
				mockErrorLogRepo.On("LogError", "Task_list_by_id", mock.AnythingOfType("*strconv.NumError")).Return(nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":          float64(400),
				"resultMessage": "OPERATION_FAILED",
			},
		},
		{
			name:   "task not found",
			taskID: "999",
			setupMocks: func(mockService *MockTaskServiceInterface, mockErrorLogRepo *MockErrorLogRepository) {
				mockService.On("ListById", 999).Return(nil, errors.New("task not found"))
				mockErrorLogRepo.On("LogError", "Task_list_by_id", mock.AnythingOfType("*errors.errorString")).Return(nil)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"code":          float64(500),
				"resultMessage": "OPERATION_FAILED",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockTaskServiceInterface{}
			testCache := NewTestCache()
			mockErrorLogRepo := &MockErrorLogRepository{}
			mockErrorLogService := errorlog.NewErrorLogService(mockErrorLogRepo)

			tt.setupMocks(mockService, mockErrorLogRepo)

			handler := NewHandler(mockService, mockErrorLogService, testCache)

			req, _ := http.NewRequest("GET", "/tasks/"+tt.taskID, nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: tt.taskID}}

			handler.ListById(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedBody["code"], response["code"])
			assert.Equal(t, tt.expectedBody["resultMessage"], response["resultMessage"])

			mockService.AssertExpectations(t)
			mockErrorLogRepo.AssertExpectations(t)
		})
	}
}

func TestHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMocks     func(*MockTaskServiceInterface)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful create",
			requestBody: map[string]interface{}{
				"title":       "New Task",
				"description": "New Description",
				"owner":       1,
			},
			setupMocks: func(mockService *MockTaskServiceInterface) {
				createdTask := &models.Task{
					Model: gorm.Model{
						ID:        1,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Title:       "New Task",
					Description: "New Description",
					Done:        false,
					Owner:       1,
				}
				mockService.On("Create", mock.AnythingOfType("*models.Task")).Return(createdTask, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"code":          float64(201),
				"resultMessage": "OPERATION_SUCCESS",
			},
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"title": "", // Invalid empty title
			},
			setupMocks: func(mockService *MockTaskServiceInterface) {
				// No service calls expected
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":          float64(400),
				"resultMessage": "OPERATION_FAILED",
			},
		},
		{
			name: "service error",
			requestBody: map[string]interface{}{
				"title":       "New Task",
				"description": "New Description",
				"owner":       1,
			},
			setupMocks: func(mockService *MockTaskServiceInterface) {
				mockService.On("Create", mock.AnythingOfType("*models.Task")).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"code":          float64(500),
				"resultMessage": "OPERATION_FAILED",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockTaskServiceInterface{}
			mockCache := &MockCache{}
			mockErrorLogRepo := &MockErrorLogRepository{}
			mockErrorLogService := errorlog.NewErrorLogService(mockErrorLogRepo)

			tt.setupMocks(mockService)

			handler := NewHandler(mockService, mockErrorLogService, mockCache)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.Create(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedBody["code"], response["code"])
			assert.Equal(t, tt.expectedBody["resultMessage"], response["resultMessage"])

			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		taskID         string
		requestBody    map[string]interface{}
		setupMocks     func(*MockTaskServiceInterface)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:   "successful update",
			taskID: "1",
			requestBody: map[string]interface{}{
				"title":       "Updated Task",
				"description": "Updated Description",
			},
			setupMocks: func(mockService *MockTaskServiceInterface) {
				updatedTask := &models.Task{
					Model: gorm.Model{
						ID:        1,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Title:       "Updated Task",
					Description: "Updated Description",
					Done:        false,
					Owner:       1,
				}
				mockService.On("Update", 1, mock.AnythingOfType("*models.Task")).Return(updatedTask, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"code":          float64(200),
				"resultMessage": "OPERATION_SUCCESS",
			},
		},
		{
			name:   "invalid id parameter",
			taskID: "invalid",
			requestBody: map[string]interface{}{
				"title": "Updated Task",
			},
			setupMocks: func(mockService *MockTaskServiceInterface) {
				// No service calls expected
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":          float64(400),
				"resultMessage": "OPERATION_FAILED",
			},
		},
		{
			name:   "service error",
			taskID: "1",
			requestBody: map[string]interface{}{
				"title": "Updated Task",
			},
			setupMocks: func(mockService *MockTaskServiceInterface) {
				mockService.On("Update", 1, mock.AnythingOfType("*models.Task")).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"code":          float64(500),
				"resultMessage": "OPERATION_FAILED",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockTaskServiceInterface{}
			mockCache := &MockCache{}
			mockErrorLogRepo := &MockErrorLogRepository{}
			mockErrorLogService := errorlog.NewErrorLogService(mockErrorLogRepo)

			tt.setupMocks(mockService)

			handler := NewHandler(mockService, mockErrorLogService, mockCache)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("PUT", "/tasks/"+tt.taskID, bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: tt.taskID}}

			handler.Update(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedBody["code"], response["code"])
			assert.Equal(t, tt.expectedBody["resultMessage"], response["resultMessage"])

			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_Done(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		taskID         string
		setupMocks     func(*MockTaskServiceInterface)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:   "successful mark as done",
			taskID: "1",
			setupMocks: func(mockService *MockTaskServiceInterface) {
				updatedTask := &models.Task{
					Model: gorm.Model{
						ID:        1,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Title:       "Test Task",
					Description: "Test Description",
					Done:        true,
					Owner:       1,
				}
				mockService.On("MarkAsDone", 1).Return(updatedTask, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"code":          float64(200),
				"resultMessage": "OPERATION_SUCCESS",
			},
		},
		{
			name:   "invalid id parameter",
			taskID: "invalid",
			setupMocks: func(mockService *MockTaskServiceInterface) {
				// No service calls expected
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":          float64(400),
				"resultMessage": "OPERATION_FAILED",
			},
		},
		{
			name:   "service error",
			taskID: "1",
			setupMocks: func(mockService *MockTaskServiceInterface) {
				mockService.On("MarkAsDone", 1).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"code":          float64(500),
				"resultMessage": "OPERATION_FAILED",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockTaskServiceInterface{}
			mockCache := &MockCache{}
			mockErrorLogRepo := &MockErrorLogRepository{}
			mockErrorLogService := errorlog.NewErrorLogService(mockErrorLogRepo)

			tt.setupMocks(mockService)

			handler := NewHandler(mockService, mockErrorLogService, mockCache)

			req, _ := http.NewRequest("PATCH", "/tasks/"+tt.taskID+"/done", nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: tt.taskID}}

			handler.Done(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedBody["code"], response["code"])
			assert.Equal(t, tt.expectedBody["resultMessage"], response["resultMessage"])

			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		taskID         string
		setupMocks     func(*MockTaskServiceInterface)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:   "successful delete",
			taskID: "1",
			setupMocks: func(mockService *MockTaskServiceInterface) {
				mockService.On("Delete", 1).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"code":          float64(200),
				"resultMessage": "OPERATION_SUCCESS",
			},
		},
		{
			name:   "invalid id parameter",
			taskID: "invalid",
			setupMocks: func(mockService *MockTaskServiceInterface) {
				// No service calls expected
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"code":          float64(400),
				"resultMessage": "OPERATION_FAILED",
			},
		},
		{
			name:   "service error",
			taskID: "1",
			setupMocks: func(mockService *MockTaskServiceInterface) {
				mockService.On("Delete", 1).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"code":          float64(500),
				"resultMessage": "OPERATION_FAILED",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockTaskServiceInterface{}
			mockCache := &MockCache{}
			mockErrorLogRepo := &MockErrorLogRepository{}
			mockErrorLogService := errorlog.NewErrorLogService(mockErrorLogRepo)

			tt.setupMocks(mockService)

			handler := NewHandler(mockService, mockErrorLogService, mockCache)

			req, _ := http.NewRequest("DELETE", "/tasks/"+tt.taskID, nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: tt.taskID}}

			handler.Delete(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedBody["code"], response["code"])
			assert.Equal(t, tt.expectedBody["resultMessage"], response["resultMessage"])

			mockService.AssertExpectations(t)
		})
	}
} 