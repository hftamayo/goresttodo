package task

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/v1/errorlog"
	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTaskService is a mock implementation of TaskServiceInterface
type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) List(cursor string, limit int) ([]*models.Task, string, int64, error) {
	args := m.Called(cursor, limit)
	return args.Get(0).([]*models.Task), args.String(1), args.Get(2).(int64), args.Error(3)
}

func (m *MockTaskService) ListById(id int) (*models.Task, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskService) Create(task *models.Task) (*models.Task, error) {
	args := m.Called(task)
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskService) Update(id int, task *models.Task) (*models.Task, error) {
	args := m.Called(id, task)
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskService) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTaskService) MarkAsDone(id int) (*models.Task, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Task), args.Error(1)
}

// MockErrorLogService is a mock implementation of errorlog.ErrorLogService
type MockErrorLogService struct {
	mock.Mock
}

func (m *MockErrorLogService) LogError(operation string, err error) {
	m.Called(operation, err)
}

func setupTestRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.GET("/tasks", handler.List)
	router.GET("/tasks/:id", handler.ListById)
	router.POST("/tasks", handler.Create)
	router.PUT("/tasks/:id", handler.Update)
	router.PATCH("/tasks/:id/done", handler.Done)
	router.DELETE("/tasks/:id", handler.Delete)
	
	return router
}

func TestHandler_List(t *testing.T) {
	mockService := new(MockTaskService)
	mockErrorLog := new(MockErrorLogService)
	mockCache := &utils.Cache{}
	handler := NewHandler(mockService, mockErrorLog, mockCache)
	router := setupTestRouter(handler)

	tests := []struct {
		name           string
		query         string
		setupMock     func()
		expectedCode  int
		expectedError string
	}{
		{
			name:  "successful list",
			query: "?limit=10",
			setupMock: func() {
				mockService.On("List", "", 10).Return(
					[]*models.Task{},
					"",
					int64(0),
					nil,
				)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:  "invalid limit",
			query: "?limit=0",
			setupMock: func() {
				mockErrorLog.On("LogError", "Task_list_validation", mock.Anything).Return()
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: ErrInvalidPaginationParams,
		},
		{
			name:  "service error",
			query: "?limit=10",
			setupMock: func() {
				mockService.On("List", "", 10).Return(
					nil,
					"",
					int64(0),
					assert.AnError,
				)
				mockErrorLog.On("LogError", "Task_list", mock.Anything).Return()
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to fetch tasks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/tasks"+tt.query, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			
			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Error)
			}
		})
	}
}

func TestHandler_ListById(t *testing.T) {
	mockService := new(MockTaskService)
	mockErrorLog := new(MockErrorLogService)
	mockCache := &utils.Cache{}
	handler := NewHandler(mockService, mockErrorLog, mockCache)
	router := setupTestRouter(handler)

	tests := []struct {
		name           string
		id            string
		setupMock     func()
		expectedCode  int
		expectedError string
	}{
		{
			name: "successful get",
			id:   "1",
			setupMock: func() {
				mockService.On("ListById", 1).Return(&models.Task{}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "invalid id",
			id:   "invalid",
			setupMock: func() {
				mockErrorLog.On("LogError", "Task_list_by_id", mock.Anything).Return()
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: ErrInvalidID,
		},
		{
			name: "task not found",
			id:   "999",
			setupMock: func() {
				mockService.On("ListById", 999).Return(nil, nil)
			},
			expectedCode:  http.StatusNotFound,
			expectedError: ErrTaskNotFound,
		},
		{
			name: "service error",
			id:   "1",
			setupMock: func() {
				mockService.On("ListById", 1).Return(nil, assert.AnError)
				mockErrorLog.On("LogError", "Task_list_by_id", mock.Anything).Return()
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to fetch task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/tasks/"+tt.id, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			
			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Error)
			}
		})
	}
}

func TestHandler_Create(t *testing.T) {
	mockService := new(MockTaskService)
	mockErrorLog := new(MockErrorLogService)
	mockCache := &utils.Cache{}
	handler := NewHandler(mockService, mockErrorLog, mockCache)
	router := setupTestRouter(handler)

	validRequest := CreateTaskRequest{
		Title: "New Task",
		Owner: 1,
	}

	tests := []struct {
		name           string
		request       interface{}
		setupMock     func()
		expectedCode  int
		expectedError string
	}{
		{
			name:    "successful create",
			request: validRequest,
			setupMock: func() {
				mockService.On("Create", mock.Anything).Return(&models.Task{}, nil)
			},
			expectedCode: http.StatusCreated,
		},
		{
			name:    "invalid request",
			request: map[string]interface{}{"invalid": "request"},
			setupMock: func() {
				mockErrorLog.On("LogError", "Task_create", mock.Anything).Return()
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: ErrInvalidRequest,
		},
		{
			name:    "service error",
			request: validRequest,
			setupMock: func() {
				mockService.On("Create", mock.Anything).Return(nil, assert.AnError)
				mockErrorLog.On("LogError", "Task_create", mock.Anything).Return()
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to create task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			
			body, _ := json.Marshal(tt.request)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			
			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Error)
			}
		})
	}
}

func TestHandler_Update(t *testing.T) {
	mockService := new(MockTaskService)
	mockErrorLog := new(MockErrorLogService)
	mockCache := &utils.Cache{}
	handler := NewHandler(mockService, mockErrorLog, mockCache)
	router := setupTestRouter(handler)

	validRequest := UpdateTaskRequest{
		Title:       "Updated Task",
		Description: "Updated Description",
	}

	tests := []struct {
		name           string
		id            string
		request       interface{}
		setupMock     func()
		expectedCode  int
		expectedError string
	}{
		{
			name:    "successful update",
			id:      "1",
			request: validRequest,
			setupMock: func() {
				mockService.On("Update", 1, mock.Anything).Return(&models.Task{}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:    "invalid id",
			id:      "invalid",
			request: validRequest,
			setupMock: func() {
				mockErrorLog.On("LogError", "Task_update_id", mock.Anything).Return()
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: ErrInvalidID,
		},
		{
			name:    "invalid request",
			id:      "1",
			request: map[string]interface{}{"invalid": "request"},
			setupMock: func() {
				mockErrorLog.On("LogError", "Task_update_binding", mock.Anything).Return()
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: ErrInvalidRequest,
		},
		{
			name:    "service error",
			id:      "1",
			request: validRequest,
			setupMock: func() {
				mockService.On("Update", 1, mock.Anything).Return(nil, assert.AnError)
				mockErrorLog.On("LogError", "Task_update", mock.Anything).Return()
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to update task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			
			body, _ := json.Marshal(tt.request)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("PUT", "/tasks/"+tt.id, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			
			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Error)
			}
		})
	}
}

func TestHandler_Done(t *testing.T) {
	mockService := new(MockTaskService)
	mockErrorLog := new(MockErrorLogService)
	mockCache := &utils.Cache{}
	handler := NewHandler(mockService, mockErrorLog, mockCache)
	router := setupTestRouter(handler)

	tests := []struct {
		name           string
		id            string
		setupMock     func()
		expectedCode  int
		expectedError string
	}{
		{
			name: "successful mark as done",
			id:   "1",
			setupMock: func() {
				mockService.On("MarkAsDone", 1).Return(&models.Task{}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "invalid id",
			id:   "invalid",
			setupMock: func() {
				mockErrorLog.On("LogError", "Task_done", mock.Anything).Return()
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: ErrInvalidID,
		},
		{
			name: "service error",
			id:   "1",
			setupMock: func() {
				mockService.On("MarkAsDone", 1).Return(nil, assert.AnError)
				mockErrorLog.On("LogError", "Task_done", mock.Anything).Return()
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to mark task as done",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("PATCH", "/tasks/"+tt.id+"/done", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			
			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Error)
			}
		})
	}
}

func TestHandler_Delete(t *testing.T) {
	mockService := new(MockTaskService)
	mockErrorLog := new(MockErrorLogService)
	mockCache := &utils.Cache{}
	handler := NewHandler(mockService, mockErrorLog, mockCache)
	router := setupTestRouter(handler)

	tests := []struct {
		name           string
		id            string
		setupMock     func()
		expectedCode  int
		expectedError string
	}{
		{
			name: "successful delete",
			id:   "1",
			setupMock: func() {
				mockService.On("Delete", 1).Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "invalid id",
			id:   "invalid",
			setupMock: func() {
				mockErrorLog.On("LogError", "Task_delete", mock.Anything).Return()
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: ErrInvalidID,
		},
		{
			name: "service error",
			id:   "1",
			setupMock: func() {
				mockService.On("Delete", 1).Return(assert.AnError)
				mockErrorLog.On("LogError", "Task_delete", mock.Anything).Return()
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to delete task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", "/tasks/"+tt.id, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			
			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Error)
			}
		})
	}
} 