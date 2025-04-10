package task

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/api/v1/task/mocks"
	"github.com/hftamayo/gotodo/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Setup test environment
func setupTest(t *testing.T) (*gin.Engine, *mocks.MockTaskService) {
    gin.SetMode(gin.TestMode)
    router := gin.New()
    mockService := new(mocks.MockTaskService)
    
    db, _, err := testutils.SetupTestDB()
    if err != nil {
        t.Fatalf("Failed to setup test DB: %v", err)
    }

    handler := &Handler{
        Db:      db,
        Service: mockService,
    }


    router.GET("/tasks", handler.List)
    router.GET("/tasks/:id", handler.ListById)
    router.POST("/tasks", handler.Create)

    return router, mockService
}

func TestHandler_List(t *testing.T) {
    router, mockService := setupTest(t)

    // Setup mock data
    tasks := []*models.Task{
        {Title: "Task 1", Description: "Description 1"},
        {Title: "Task 2", Description: "Description 2"},
    }

    // Setup expectations
    mockService.On("List").Return(tasks, nil)

    // Perform request
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/tasks", nil)
    router.ServeHTTP(w, req)

    // Assert response
    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Equal(t, "OPERATION_SUCCESS", response["resultMessage"])
}

func TestHandler_Create(t *testing.T) {
    router, mockService := setupTest(t)

    // Test task
    task := &models.Task{
        Title:       "New Task",
        Description: "New Description",
    }

    // Setup expectations
    mockService.On("Create", mock.AnythingOfType("*models.Task")).Return(nil)

    // Create request body
    body, _ := json.Marshal(task)

    // Perform request
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    router.ServeHTTP(w, req)

    // Assert response
    assert.Equal(t, http.StatusCreated, w.Code)

    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Equal(t, "OPERATION_SUCCESS", response["resultMessage"])
}

func TestHandler_ListById(t *testing.T) {
    router, mockService := setupTest(t)

    // Setup mock data
    task := &models.Task{
        Title:       "Test Task",
        Description: "Test Description",
    }

    // Setup expectations
    mockService.On("ListById", 1).Return(task, nil)

    // Perform request
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/tasks/1", nil)
    router.ServeHTTP(w, req)

    // Assert response
    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Equal(t, "OPERATION_SUCCESS", response["resultMessage"])
}