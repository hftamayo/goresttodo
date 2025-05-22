package task

import (
	"crypto/sha256"
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
        if cachedData, ok := cachedResponse.Data.(TaskListResponse); ok {
            etag := cachedData.ETag
            
            // Check if client's cached version is still valid
            if ifNoneMatch := c.GetHeader("If-None-Match"); ifNoneMatch != "" && 
                (ifNoneMatch == etag || ifNoneMatch == "W/"+etag) {
                c.Status(http.StatusNotModified)
                return
            }

            setEtagHeader(c, cachedData.ETag)
            c.Header("Last-Modified", cachedData.LastModified)
            addCacheHeaders(c, false)
            
            c.JSON(http.StatusOK, cachedResponse)
        } else {
            // Invalid cache data type, log and continue with fresh data
            h.errorLogService.LogError("Task_list_cache_type", 
                fmt.Errorf("unexpected type for cached data"))
        }
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

    if listResponse, ok := response.Data.(TaskListResponse); ok {
        setEtagHeader(c, listResponse.ETag)
        c.Header("Last-Modified", listResponse.LastModified)
    }    

    // Cache the response
    if err := h.cache.Set(cacheKey, response, utils.DefaultCacheTime); err != nil {
        h.errorLogService.LogError("Task_list_cache_set", err)
        // Continue without failing the request
    }    

    addCacheHeaders(c, false)

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

    cacheKey := fmt.Sprintf("task_%d", id)
    var cachedResponse TaskOperationResponse
    if err := h.cache.Get(cacheKey, &cachedResponse); err == nil {
        if taskData, ok := cachedResponse.Data.(*TaskResponse); ok {
            taskETag := fmt.Sprintf("\"%x\"", sha256.Sum256([]byte(
                fmt.Sprintf("%d-%s-%t-%d", taskData.ID, taskData.Title, taskData.Done, taskData.UpdatedAt.UnixNano()),
            )))
            
            // Check if client's cached version is still valid
            if ifNoneMatch := c.GetHeader("If-None-Match"); ifNoneMatch != "" && 
                (ifNoneMatch == taskETag || ifNoneMatch == "W/"+taskETag) {
                c.Status(http.StatusNotModified)
                return
            }
            
            setEtagHeader(c, taskETag)
            c.Header("Last-Modified", taskData.UpdatedAt.UTC().Format(http.TimeFormat))
            addCacheHeaders(c, false)
            
            c.JSON(http.StatusOK, cachedResponse)
        } else {
            // Invalid cache data type, log and continue with fresh data
            h.errorLogService.LogError("Task_listbyid_cache_type", 
                fmt.Errorf("unexpected type for cached data"))
        }
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

    taskETag := fmt.Sprintf("\"%x\"", sha256.Sum256([]byte(
        fmt.Sprintf("%d-%s-%t-%d", task.ID, task.Title, task.Done, task.UpdatedAt.UnixNano()),
    )))

    response := TaskOperationResponse{
        Code:          http.StatusOK,
        ResultMessage: utils.OperationSuccess,
        Data:          ToTaskResponse(task),
        Timestamp:     time.Now().Unix(),
        CacheTTL:      60,        
    }

    lastModified := task.UpdatedAt.UTC().Format(http.TimeFormat)

    c.Header("ETag", taskETag)
    c.Header("Last-Modified", lastModified)    

    // Cache the response
    if err := h.cache.Set(cacheKey, response, utils.DefaultCacheTime); err != nil {
        h.errorLogService.LogError("Task_list_cache_set", err)
    }        

    addCacheHeaders(c, false)

    c.JSON(http.StatusOK, response)
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

    // Validate and set defaults
    query = validatePagePaginationQuery(query)

    // Try to get from cache
    cacheKey := fmt.Sprintf("tasks_page_%d_limit_%d_order_%s", query.Page, query.Limit, query.Order)
    var cachedResponse TaskOperationResponse
    if err := h.cache.Get(cacheKey, &cachedResponse); err == nil {
        if cachedData, ok := cachedResponse.Data.(TaskListResponse); ok {
            etag := cachedData.ETag
            
            // Check if client's cached version is still valid
            if ifNoneMatch := c.GetHeader("If-None-Match"); ifNoneMatch != "" && 
                (ifNoneMatch == etag || ifNoneMatch == "W/"+etag) {
                c.Status(http.StatusNotModified)
                return
            }
            
            setEtagHeader(c, etag)
            c.Header("Last-Modified", cachedData.LastModified)
            addCacheHeaders(c, false)
            
            c.JSON(http.StatusOK, cachedResponse)
        } else {
            // Invalid cache data type, log and continue with fresh data
            h.errorLogService.LogError("Task_listbypage_cache_type", 
                fmt.Errorf("unexpected type for cached data"))
        }
        return
    }    

    // Get data from service - no need to set defaults again as validatePagePaginationQuery already did this
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
    
    // Generate ETag and LastModified
    etag := generateETag(tasks)
    lastModified := time.Now().UTC().Format(http.TimeFormat)
    
    // Build response
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
                IsFirstPage: query.Page == 1,
                IsLastPage:  query.Page >= totalPages,
                HasMore:     query.Page < totalPages,
                HasPrev:     query.Page > 1,                
            },
            ETag:          etag,
            LastModified:  lastModified,            
        },
        Timestamp:     time.Now().Unix(),
        CacheTTL:      60,
    }

    // Set headers
    setEtagHeader(c, etag)
    c.Header("Last-Modified", lastModified)

    // Cache the response
    if err := h.cache.Set(cacheKey, response, utils.DefaultCacheTime); err != nil {
        h.errorLogService.LogError("Task_list_by_page_cache_set", err)
    }        

    addCacheHeaders(c, false)
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
        Description: createRequest.Description, //this is an optional field
        Done:        false, // Default to false
        Owner:       createRequest.Owner,
    }

    createdTask, err := h.service.Create(task)
    if err != nil {
        h.errorLogService.LogError("Task_create", err)

        statusCode := http.StatusInternalServerError
        errorMsg := "Failed to create task"

        // Check for duplicate title error
        if strings.Contains(err.Error(), "already exists") {
            statusCode = http.StatusBadRequest
            errorMsg = err.Error() // Use the actual error message
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
        CacheTTL:      60,
    }

    if err := h.cache.Delete("tasks_cursor_*"); err != nil {
        h.errorLogService.LogError("Task_create_cache_invalidation", err)
    }

    if err := h.cache.Delete("tasks_page_*"); err != nil {
        h.errorLogService.LogError("Task_create_cache_invalidation", err)
    }
    addCacheHeaders(c, true)

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
        Timestamp:     time.Now().Unix(),
        CacheTTL:      60,
    }
    h.cache.Delete("tasks_cursor_*") 
    h.cache.Delete("tasks_page_*") 
    h.cache.Delete(fmt.Sprintf("task_%d", updatedTask.ID)) // Invalidate cache for the updated task
    
    addCacheHeaders(c, true)

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
        Timestamp:     time.Now().Unix(),
        CacheTTL:      60,
    }
    h.cache.Delete("tasks_cursor_*") 
    h.cache.Delete("tasks_page_*") 
    h.cache.Delete(fmt.Sprintf("task_%d", id))    

    addCacheHeaders(c, true)

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

    h.cache.Delete("tasks_cursor_*")
    h.cache.Delete("tasks_page_*")
    h.cache.Delete(fmt.Sprintf("task_%d", id))

    addCacheHeaders(c, true)

    response := TaskOperationResponse{
        Code:          http.StatusOK,
        ResultMessage: utils.OperationSuccess,
        Data:          ToTaskResponse(updatedTask),
        Timestamp:     time.Now().Unix(),
        CacheTTL:      60,
    }   
    
    h.cache.Delete("tasks_cursor_*") 
    h.cache.Delete("tasks_page_*") 
    h.cache.Delete(fmt.Sprintf("task_%d", id))    

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
        c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
        c.Header("Pragma", "no-cache")
        c.Header("Expires", "0")
    } else {
        // For GET - allow short caching
        c.Header("Cache-Control", "private, max-age=60")
        c.Header("Vary", "Authorization")
    }
}