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
	"github.com/hftamayo/gotodo/api/v1/task/mocks"
	"github.com/hftamayo/gotodo/pkg/testutils"
	"github.com/hftamayo/gotodo/pkg/utils"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTaskService struct {
    mock.Mock
}

// Implement all TaskServiceInterface methods
func (m *MockTaskService) List() ([]*models.Task, error) {
    args := m.Called()
    return args.Get(0).([]*models.Task), args.Error(1)
}

func (m *MockTaskService) ListById(id int) (*models.Task, error) {
    args := m.Called(id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskService) Create(task *models.Task) error {
    args := m.Called(task)
    return args.Error(0)
}

func (m *MockTaskService) Update(task *models.Task) error {
    args := m.Called(task)
    return args.Error(0)
}

func (m *MockTaskService) Delete(id int) error {
    args := m.Called(id)
    return args.Error(0)
}

func (m *MockTaskService) Done(id int, done bool) (*models.Task, error) {
    args := m.Called(id, done)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.Task), args.Error(1)
}

// Setup test environment
func setupTest(t *testing.T) (*gin.Engine, *mocks.MockTaskService) {
    gin.SetMode(gin.TestMode)
    router := gin.New()
    mockService := new(mocks.MockTaskService)
    
    db, _, err := testutils.SetupTestDB()
    if err != nil {
        t.Fatalf("Failed to setup test DB: %v", err)
    }

    // Create mock Redis client
    redisClient := redis.NewClient(&redis.Options{})
    
    // Create handler with mocked dependencies
    handler := NewHandler(
        db,
        mockService,
        errorlog.NewErrorLogService(redisClient),
    )


    router.GET("/tasks", handler.List)
    router.GET("/tasks/:id", handler.ListById)
    router.POST("/tasks", handler.Create)
    router.PATCH("/tasks/:id", handler.Update)
    router.PATCH("/tasks/:id/done", handler.Done)
    router.DELETE("/tasks/:id", handler.Delete)

    return router, mockService
}

func TestHandler_List(t *testing.T) {
    router, mockService, _ := setupTest(t)

    tasks := []*models.Task{
        {Title: "Task 1", Description: "Description 1"},
        {Title: "Task 2", Description: "Description 2"},
    }

    mockService.On("List").Return(tasks, nil)

    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/tasks", nil)
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Equal(t, utils.OperationSuccess, response["resultMessage"])
}

func TestHandler_Create(t *testing.T) {
    router, mockService, _ := setupTest(t)

    task := &models.Task{
        Title:       "New Task",
        Description: "New Description",
    }

    mockService.On("Create", mock.AnythingOfType("*models.Task")).Return(nil)

    body, _ := json.Marshal(task)
    w := httptest.NewRecorder()
    req := httptest.NewRequest("POST", "/tasks", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Equal(t, utils.OperationSuccess, response["resultMessage"])
}

func TestHandler_Update(t *testing.T) {
    router, mockService, _ := setupTest(t)

    existingTask := &models.Task{
        Title:       "Existing Task",
        Description: "Existing Description",
        Owner:       1,
    }

    updatedTask := &models.Task{
        Title:       "Updated Task",
        Description: "Updated Description",
    }

    mockService.On("ListById", 1).Return(existingTask, nil)
    mockService.On("Update", mock.AnythingOfType("*models.Task")).Return(nil)

    body, _ := json.Marshal(updatedTask)
    w := httptest.NewRecorder()
    req := httptest.NewRequest("PATCH", "/tasks/1", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Equal(t, utils.OperationSuccess, response["resultMessage"])
}