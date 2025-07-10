package task

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	service TaskServiceInterface
}

func NewHandler(service TaskServiceInterface) *Handler {
	if service == nil {
		panic("task service is required")
	}
	return &Handler{service: service}
}

func (h *Handler) List(c *gin.Context) {
	var query PagePaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(
			http.StatusBadRequest,
			utils.OperationFailed,
			ErrInvalidPaginationParams.Error(),
		))
		return
	}

	// Validate and set defaults
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 || query.Limit > MaxLimit {
		query.Limit = DefaultLimit
	}
	if query.Order == "" {
		query.Order = DefaultOrder
	}

	tasks, totalCount, err := h.service.ListByPage(query.Page, query.Limit, query.Order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(
			http.StatusInternalServerError,
			utils.OperationFailed,
			err.Error(),
		))
		return
	}

	response := h.buildListResponse(tasks, totalCount, query.Page, query.Limit, query.Order)
	if listResponse, ok := response.Data.(TaskListResponse); ok {
		setEtagHeader(c, listResponse.ETag)
		c.Header(headerLastModified, listResponse.LastModified)
	}
	addCacheHeaders(c, false)
	c.JSON(http.StatusOK, response)
}

func (h *Handler) ListById(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(
			http.StatusBadRequest,
			utils.OperationFailed,
			ErrInvalidID.Error(),
		))
		return
	}
	task, err := h.service.ListById(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, NewErrorResponse(
				http.StatusNotFound,
				utils.OperationFailed,
				ErrTaskNotFound.Error(),
			))
		} else {
			c.JSON(http.StatusInternalServerError, NewErrorResponse(
				http.StatusInternalServerError,
				utils.OperationFailed,
				err.Error(),
			))
		}
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
	taskETag := generateTaskETag(task)
	response := TaskOperationResponse{
		Code:          http.StatusOK,
		ResultMessage: utils.OperationSuccess,
		Data:          ToTaskResponse(task),
		Timestamp:     time.Now().Unix(),
		CacheTTL:      30,
	}
	lastModified := task.UpdatedAt.UTC().Format(http.TimeFormat)
	c.Header("ETag", taskETag)
	c.Header(headerLastModified, lastModified)
	addCacheHeaders(c, false)
	c.JSON(http.StatusOK, response)
}

// validateListParams validates and parses list parameters from the request
func (h *Handler) validateListParams(c *gin.Context) (int, int, string, error) {
    // Parse page parameter
    pageStr := c.DefaultQuery("page", "1")
    page, err := strconv.Atoi(pageStr)
    if err != nil || page < 1 {
        return 0, 0, "", fmt.Errorf("invalid page parameter")
    }

    // Parse limit parameter
    limitStr := c.DefaultQuery("limit", "10")
    limit, err := strconv.Atoi(limitStr)
    if err != nil || limit < 1 || limit > 100 {
        return 0, 0, "", fmt.Errorf("invalid limit parameter")
    }

    // Validate order parameter
    order := c.DefaultQuery("order", "desc")
    if order != "asc" && order != "desc" {
        return 0, 0, "", fmt.Errorf("invalid order parameter")
    }

    return page, limit, order, nil
}

// buildListResponse creates a paginated list response
func (h *Handler) buildListResponse(tasks []*models.Task, totalCount int64, page, limit int, order string) TaskOperationResponse {
    totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))
    hasMore := page < totalPages
    hasPrev := page > 1

    listResponse := TaskListResponse{
        Tasks: TasksToResponse(tasks),
        Pagination: PaginationMeta{
            Limit:        limit,
            TotalCount:   totalCount,
            CurrentPage:  page,
            TotalPages:   totalPages,
            Order:        order,
            HasMore:      hasMore,
            HasPrev:      hasPrev,
            IsFirstPage:  page == 1,
            IsLastPage:   page == totalPages,
        },
        ETag:         generateETag(tasks),
        LastModified: time.Now().UTC().Format(http.TimeFormat),
    }

    return NewTaskOperationResponse(listResponse)
}

// ListByPage handles the GET /tasks endpoint with pagination
func (h *Handler) ListByPage(c *gin.Context) {
    // Validate and parse parameters
    page, limit, order, err := h.validateListParams(c)
    if err != nil {
        c.JSON(http.StatusBadRequest, NewErrorResponse(
            http.StatusBadRequest,
            "INVALID_PARAMETERS",
            err.Error(),
        ))
        return
    }

    // Get tasks from service
    tasks, totalCount, err := h.service.ListByPage(page, limit, order)
    if err != nil {
        c.JSON(http.StatusInternalServerError, NewErrorResponse(
            http.StatusInternalServerError,
            "INTERNAL_SERVER_ERROR",
            "Failed to list tasks",
        ))
        return
    }

    // Build response using existing DTO
    response := h.buildListResponse(tasks, totalCount, page, limit, order)
    c.JSON(http.StatusOK, response)
}

func (h *Handler) Create(c *gin.Context) {
	var createRequest CreateTaskRequest
	if err := c.ShouldBindJSON(&createRequest); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(
			http.StatusBadRequest,
			utils.OperationFailed,
			ErrInvalidRequest.Error(),
		))
		return
	}
	task := &models.Task{
		Title:       createRequest.Title,
		Description: createRequest.Description,
		Done:        false,
		Owner:       createRequest.Owner,
	}
	createdTask, err := h.service.Create(task)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := "Failed to create task"
		if strings.Contains(err.Error(), "already exists") {
			statusCode = http.StatusBadRequest
			errorMsg = err.Error()
		}
		c.JSON(statusCode, NewErrorResponse(
			statusCode,
			utils.OperationFailed,
			errorMsg,
		))
		return
	}
	response := TaskOperationResponse{
		Code:          http.StatusCreated,
		ResultMessage: utils.OperationSuccess,
		Data:          ToTaskResponse(createdTask),
		Timestamp:     time.Now().Unix(),
		CacheTTL:      30,
	}
	addCacheHeaders(c, true)
	c.JSON(http.StatusCreated, response)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(
			http.StatusBadRequest,
			utils.OperationFailed,
			ErrInvalidID.Error(),
		))
		return
	}
	var updateRequest UpdateTaskRequest
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
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
		Timestamp:     time.Now().Unix(),
		CacheTTL:      30,
	}
	addCacheHeaders(c, true)
	c.JSON(http.StatusOK, response)
}

func (h *Handler) Done(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(
			http.StatusBadRequest,
			utils.OperationFailed,
			ErrInvalidID.Error(),
		))
		return
	}
	updatedTask, err := h.service.MarkAsDone(id)
	if err != nil {
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
		Timestamp:     time.Now().Unix(),
		CacheTTL:      30,
	}
	addCacheHeaders(c, true)
	c.JSON(http.StatusOK, response)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(
			http.StatusBadRequest,
			utils.OperationFailed,
			ErrInvalidID.Error(),
		))
		return
	}
	if err := h.service.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(
			http.StatusInternalServerError,
			utils.OperationFailed,
			"Failed to delete task",
		))
		return
	}
	response := TaskOperationResponse{
		Code:          http.StatusOK,
		ResultMessage: utils.OperationSuccess,
		Data:          nil,
		Timestamp:     time.Now().Unix(),
	}
	addCacheHeaders(c, true)
	c.JSON(http.StatusOK, response)
}

func setEtagHeader(c *gin.Context, etag string) {
    if etag != "" {
        c.Header("ETag", etag)
    }
}

func addCacheHeaders(c *gin.Context, isModifying bool) {
    if isModifying {
        // For POST, PUT, DELETE - tell browser not to cache
        c.Header(headerCacheControl, "no-cache, no-store, must-revalidate")
        c.Header("Pragma", "no-cache")
        c.Header("Expires", "0")
    } else {
        // For GET - allow short caching
        if strings.Contains(c.Request.URL.Path, "/list") {
            c.Header(headerCacheControl, "no-cache, max-age=0")
        } else {
            // Individual item endpoints can cache longer
            c.Header(headerCacheControl, "private, max-age=60")
        }
        c.Header("Vary", "Authorization")
    }
}