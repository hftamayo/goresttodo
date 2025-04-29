package task

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/v1/errorlog"
	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/utils"
	"gorm.io/gorm"
)

const (
    ErrInvalidID = "Invalid ID parameter"
    ErrTaskNotFound = "Task not found"
    ErrInvalidRequest = "Invalid request body"
    ErrInvalidPaginationParams = "Invalid pagination parameters"
)

type Handler struct {
	service         TaskServiceInterface
	errorLogService *errorlog.ErrorLogService
	cache 		 	*utils.Cache
}

func NewHandler(service TaskServiceInterface, errorLogService *errorlog.ErrorLogService, cache *utils.Cache) *Handler {
    if service == nil {
        panic("task service is required")
    }
    if errorLogService == nil {
        panic("error log service is required")
    }
    if cache == nil {
        panic("cache is required")
    }

	return &Handler{
        service:         service,
        errorLogService: errorLogService,
        cache:           cache,
	}
}

func (h *Handler) List(c *gin.Context) {
    var query CursorPaginationQuery
    if err := c.ShouldBindQuery(&query); err != nil {
        h.errorLogService.LogError("Task_list_validation", err)
        c.JSON(http.StatusBadRequest, NewErrorResponse(
            http.StatusBadRequest,
            utils.OperationFailed,
            ErrInvalidPaginationParams,
        ))
        return
    }

    if query.Limit <= 0 {
        query.Limit = defaultLimit
    }
    if query.Limit > maxLimit {
        query.Limit = maxLimit
    }    

    tasks, nextCursor, totalCount, err := h.service.List(query.Cursor, query.Limit)
    if err != nil {
        h.errorLogService.LogError("Task_list", err)
        c.JSON(http.StatusInternalServerError, NewErrorResponse(
            http.StatusInternalServerError,
            utils.OperationFailed,
            "Failed to fetch tasks",
        ))
        return
    }

    response := TaskListResponse{
        Tasks: TasksToResponse(tasks),
        Pagination: struct {
            NextCursor string `json:"nextCursor"`
            Limit     int    `json:"limit"`
            TotalCount int64  `json:"totalCount"`
            HasMore   bool   `json:"hasMore"`
        }{
            NextCursor: nextCursor,
            Limit:     query.Limit,
            TotalCount: totalCount,
            HasMore:   nextCursor != "",
        },
    }

    c.JSON(http.StatusOK, response)
}


func (h *Handler) ListById(c *gin.Context) {
	// Parse the ID from the URL parameter.
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorLogService.LogError("Task_list_by_id", err)
        c.JSON(http.StatusBadRequest, NewErrorResponse(
            http.StatusBadRequest,
            utils.OperationFailed,
            ErrInvalidID,
        ))
        return
    }

	task, err := h.service.ListById(id)
    if err != nil {
        h.errorLogService.LogError("Task_list_by_id", err)
        c.JSON(http.StatusInternalServerError, NewErrorResponse(
            http.StatusInternalServerError,
            utils.OperationFailed,
            "Failed to fetch task",
        ))
        return
    }

    if task == nil {
        c.JSON(http.StatusNotFound, gin.H{
            "code": http.StatusNotFound,
            "resultMessage": utils.OperationFailed,
            "error":        ErrTaskNotFound,
        })
        return
    }

    if task == nil {
        c.JSON(http.StatusNotFound, NewErrorResponse(
            http.StatusNotFound,
            utils.OperationFailed,
            ErrTaskNotFound,
        ))
        return
    }

    response := TaskOperationResponse{
        Code:          http.StatusOK,
        ResultMessage: utils.OperationSuccess,
        Data:          ToTaskResponse(task),
    }
    c.JSON(http.StatusOK, response)
}

func (h *Handler) Create(c *gin.Context) {
	var createRequest CreateTaskRequest
	if err := c.ShouldBindJSON(&createRequest); err != nil {
		h.errorLogService.LogError("Task_create", err)
        c.JSON(http.StatusBadRequest, NewErrorResponse(
            http.StatusBadRequest,
            utils.OperationFailed,
            ErrInvalidRequest,
        ))
        return
    }

    task := &models.Task{
        Title:       createRequest.Title,
        Description: createRequest.Description,
        Owner:       createRequest.Owner,
    }

    createdTask, err := h.service.Create(task)
    if err != nil {
        h.errorLogService.LogError("Task_create", err)
        c.JSON(http.StatusInternalServerError, NewErrorResponse(
            http.StatusInternalServerError,
            utils.OperationFailed,
            "Failed to create task",
        ))
        return
    }

    response := TaskOperationResponse{
        Code:          http.StatusCreated,
        ResultMessage: utils.OperationSuccess,
        Data:          ToTaskResponse(createdTask),
    }
    c.JSON(http.StatusCreated, response)
}

func (h *Handler) Update(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        h.errorLogService.LogError("Task_update_id", err)
        c.JSON(http.StatusBadRequest, NewErrorResponse(
            http.StatusBadRequest,
            utils.OperationFailed,
            ErrInvalidID,
        ))
        return
    }

    var updateRequest UpdateTaskRequest
    if err := c.ShouldBindJSON(&updateRequest); err != nil {
		h.errorLogService.LogError("Task_update_binding", err)
        c.JSON(http.StatusBadRequest, NewErrorResponse(
            http.StatusBadRequest,
            utils.OperationFailed,
            ErrInvalidRequest,
        ))
        return
    }

    task := &models.Task{
        Model:       gorm.Model{ID: uint(id)},
        Title:       updateRequest.Title,
        Description: updateRequest.Description,
    }

    updatedTask, err := h.service.Update(id, task)
    if err != nil {
        h.errorLogService.LogError("Task_update", err)
        c.JSON(http.StatusInternalServerError, NewErrorResponse(
            http.StatusInternalServerError,
            utils.OperationFailed,
            "Failed to update task",
        ))
        return
    }

    response := TaskOperationResponse{
        Code:          http.StatusOK,
        ResultMessage: utils.OperationSuccess,
        Data:          ToTaskResponse(updatedTask),
    }
    c.JSON(http.StatusOK, response)
}

func (h *Handler) Done(c *gin.Context) {
	// Parse the ID from the URL parameter.
	id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        h.errorLogService.LogError("Task_done", err)
        c.JSON(http.StatusBadRequest, NewErrorResponse(
            http.StatusBadRequest,
            utils.OperationFailed,
            ErrInvalidID,
        ))
        return
    }

    updatedTask, err := h.service.MarkAsDone(id)
    if err != nil {
        h.errorLogService.LogError("Task_done", err)
        c.JSON(http.StatusInternalServerError, NewErrorResponse(
            http.StatusInternalServerError,
            utils.OperationFailed,
            "Failed to mark task as done",
        ))
        return
    }

    response := TaskOperationResponse{
        Code:          http.StatusOK,
        ResultMessage: utils.OperationSuccess,
        Data:          ToTaskResponse(updatedTask),
    }
    c.JSON(http.StatusOK, response)
}

func (h *Handler) Delete(c *gin.Context) {
	// Parse the ID from the URL parameter.
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorLogService.LogError("Task_delete", err)
        c.JSON(http.StatusBadRequest, NewErrorResponse(
            http.StatusBadRequest,
            utils.OperationFailed,
            ErrInvalidID,
        ))
        return
    }

    if err := h.service.Delete(id); err != nil {
        h.errorLogService.LogError("Task_delete", err)
        c.JSON(http.StatusInternalServerError, NewErrorResponse(
            http.StatusInternalServerError,
            utils.OperationFailed,
            "Failed to delete task",
        ))
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "code": http.StatusOK,
        "resultMessage": utils.OperationSuccess,
    })
}
