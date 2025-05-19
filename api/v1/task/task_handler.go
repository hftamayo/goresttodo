package task

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/v1/errorlog"
	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/utils"
	"gorm.io/gorm"
)

var (
    ErrInvalidID = errors.New("invalid ID parameter")
    ErrTaskNotFound = errors.New("task not found")
    ErrInvalidRequest = errors.New("invalid request body")
    ErrInvalidPaginationParams = errors.New("invalid pagination parameters")
    ErrInvalidCursor = errors.New("invalid cursor")
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
        cache:          cache,
    }
}

func (h *Handler) List(c *gin.Context) {
    var query CursorPaginationQuery
    if err := c.ShouldBindQuery(&query); err != nil {
        h.errorLogService.LogError("Task_list_validation", err)
        c.JSON(http.StatusBadRequest, NewErrorResponse(
            http.StatusBadRequest,
            utils.OperationFailed,
            ErrInvalidPaginationParams.Error(),
        ))
        return
    }

    // Validate and set defaults
    query = validatePaginationQuery(query)

    // Try to get from cache
    cacheKey := fmt.Sprintf("tasks_cursor_%s_limit_%d_order_%s", query.Cursor, query.Limit, query.Order)
    var cachedResponse TaskOperationResponse
    if err := h.cache.Get(cacheKey, &cachedResponse); err == nil {
        c.JSON(http.StatusOK, cachedResponse)
        return
    }

    tasks, nextCursor, prevCursor, totalCount, err := h.service.List(query.Cursor, query.Limit, query.Order)
    if err != nil {
        h.errorLogService.LogError("Task_list", err)
        statusCode := http.StatusInternalServerError
        if errors.Is(err, ErrInvalidCursor) {
            statusCode = http.StatusBadRequest
        }
        c.JSON(statusCode, NewErrorResponse(
            statusCode,
            utils.OperationFailed,
            err.Error(),
        ))
        return
    }

    // Build response
    response := buildListResponse(tasks, query, nextCursor, prevCursor, totalCount)

    // Cache the response
    h.cache.Set(cacheKey, response, defaultCacheTime)

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
            ErrInvalidID.Error(),
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
        c.JSON(http.StatusNotFound, NewErrorResponse(
            http.StatusNotFound,
            utils.OperationFailed,
            ErrTaskNotFound.Error(),
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
            ErrInvalidRequest.Error(),
        ))
        return
    }

    task := &models.Task{
        Title:       createRequest.Title,
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
            ErrInvalidID.Error(),
        ))
        return
    }

    var updateRequest UpdateTaskRequest
    if err := c.ShouldBindJSON(&updateRequest); err != nil {
		h.errorLogService.LogError("Task_update_binding", err)
        c.JSON(http.StatusBadRequest, NewErrorResponse(
            http.StatusBadRequest,
            utils.OperationFailed,
            ErrInvalidRequest.Error(),
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
            ErrInvalidID.Error(),
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
            ErrInvalidID.Error(),
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

func (h *Handler) ListByPage(c *gin.Context) {
    var query PagePaginationQuery
    if err := c.ShouldBindQuery(&query); err != nil {
        h.errorLogService.LogError("Task_list_by_page_validation", err)
        c.JSON(http.StatusBadRequest, NewErrorResponse(
            http.StatusBadRequest,
            utils.OperationFailed,
            ErrInvalidPaginationParams.Error(),
        ))
        return
    }

    // Set defaults
    if query.Page <= 0 {
        query.Page = 1
    }
    if query.Limit <= 0 {
        query.Limit = DefaultLimit
    }
    if query.Order == "" {
        query.Order = DefaultOrder
    }

    tasks, totalCount, err := h.service.ListByPage(query.Page, query.Limit, query.Order)
    if err != nil {
        h.errorLogService.LogError("Task_list_by_page", err)
        c.JSON(http.StatusInternalServerError, NewErrorResponse(
            http.StatusInternalServerError,
            utils.OperationFailed,
            "Failed to fetch tasks",
        ))
        return
    }

    totalPages := int(math.Ceil(float64(totalCount) / float64(query.Limit)))
    
    response := TaskOperationResponse{
        Code:          http.StatusOK,
        ResultMessage: utils.OperationSuccess,
        Data: TaskListResponse{
            Tasks: TasksToResponse(tasks),
            Pagination: PaginationMeta{
                Limit:       query.Limit,
                TotalCount:  totalCount,
                CurrentPage: query.Page,
                TotalPages:  totalPages,
                Order:       query.Order,
            },
        },
    }

    c.JSON(http.StatusOK, response)
}
