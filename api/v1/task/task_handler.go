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
    var query PagePaginationQuery
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
    if query.Page < 1 {
        query.Page = 1
    }
    if query.Limit < 1 || query.Limit > MaxLimit {
        query.Limit = DefaultLimit
    }
    if query.Order == "" {
        query.Order = DefaultOrder
    }

    // Get tasks from service
    tasks, totalCount, err := h.service.ListByPage(query.Page, query.Limit, query.Order)
    if err != nil {
        h.errorLogService.LogError("Task_list", err)
        c.JSON(http.StatusInternalServerError, NewErrorResponse(
            http.StatusInternalServerError,
            utils.OperationFailed,
            err.Error(),
        ))
        return
    }

    // Build response
    response := h.buildListResponse(tasks, totalCount, query.Page, query.Limit, query.Order)

    if listResponse, ok := response.Data.(TaskListResponse); ok {
        setEtagHeader(c, listResponse.ETag)
        c.Header(headerLastModified, listResponse.LastModified)
    }    

    addCacheHeaders(c, false)

    c.JSON(http.StatusOK, response)
}

func (h *Handler) ListById(c *gin.Context) {
    // Parse the ID from the URL parameter
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

    cacheKey := fmt.Sprintf(taskCacheKey, id)
    var cachedResponse TaskOperationResponse
    if err := h.cache.Get(cacheKey, &cachedResponse); err == nil {
        if taskData, ok := cachedResponse.Data.(*TaskResponse); ok {
            // Use the generateTaskETag for consistent ETag generation
            taskETag := generateTaskETag(&models.Task{
                Model:       gorm.Model{ID: uint(taskData.ID)},
                Title:       taskData.Title,
                Description: taskData.Description,
                Done:        taskData.Done,
                Owner:       taskData.Owner,
            })
            
            // Check if client's cached version is still valid
            if ifNoneMatch := c.GetHeader("If-None-Match"); ifNoneMatch != "" && 
                (ifNoneMatch == taskETag || ifNoneMatch == "W/"+taskETag) {
                c.Status(http.StatusNotModified)
                return
            }
            
            setEtagHeader(c, taskETag)
            c.Header(headerLastModified, taskData.UpdatedAt.UTC().Format(http.TimeFormat))
            addCacheHeaders(c, false)
            
            c.JSON(http.StatusOK, cachedResponse)
        } else {
            // Invalid cache data type, log and continue with fresh data
            h.errorLogService.LogError("Task_listbyid_cache_type", 
                fmt.Errorf("unexpected type for cached data"))
        }
        return
    }

    // Continue with database retrieval if not in cache
    task, err := h.service.ListById(id)
    if err != nil {
        // Error handling code...
        return
    }

    if task == nil {
        // Not found handling...
        return
    }

    // Use the new generateTaskETag function
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

    // Cache the response
    if err := h.cache.Set(cacheKey, response, time.Duration(response.CacheTTL)*time.Second); err != nil {
        h.errorLogService.LogError("Task_list_cache_set", err)
    }        

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
        Description: createRequest.Description,
        Done:        false,
        Owner:       createRequest.Owner,
    }

    createdTask, err := h.service.Create(task)
    if err != nil {
        h.errorLogService.LogError("Task_create", err)
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

    // Invalidate all task-related caches
    if err := h.cache.InvalidateByTags(taskServiceListRef); err != nil {
        h.errorLogService.LogError("Task_create_cache_invalidation_tags", err)
    }

    // Also directly invalidate first page cache keys since they're most affected by new items
    if err := h.cache.Delete("tasks_page_1_*"); err != nil {
        h.errorLogService.LogError("Task_create_cache_invalidation_page1", err)
    }
    
    // Small delay to ensure DB consistency before next read
    time.Sleep(time.Millisecond * 50)    

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
        CacheTTL:      30,
    }

    // 1. Invalidate by tags
    if err := h.cache.InvalidateByTags(taskServiceListRef, fmt.Sprintf(errTaskReference, id)); err != nil {
        h.errorLogService.LogError("Task_update_cache_invalidation_tags", err)
    }
    
    // 2. Also directly invalidate specific item cache
    if err := h.cache.Delete(fmt.Sprintf(taskCacheKey, id)); err != nil {
        h.errorLogService.LogError("Task_update_cache_invalidation_item", err)
    }
    
    // 3. Invalidate all page caches since we don't know which pages this task appears on
    if err := h.cache.Delete(taskPageCacheName); err != nil {
        h.errorLogService.LogError("Task_update_cache_invalidation_pages", err)
    }
    
    // 4. Small delay to ensure DB consistency before next read
    time.Sleep(time.Millisecond * 50)

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
        CacheTTL:      30,
    }

    // 1. Invalidate by tags (if your cache.InvalidateByTags is robust)
    if err := h.cache.InvalidateByTags(taskServiceListRef, fmt.Sprintf(errTaskReference, id)); err != nil {
        h.errorLogService.LogError("Task_done_cache_invalidation_tags", err)
    }
    
    // 2. Also directly invalidate specific cache entries
    if err := h.cache.Delete("tasks_cursor_*"); err != nil {
        h.errorLogService.LogError("Task_done_cache_invalidation_cursor", err)
    }

    if err := h.cache.Delete(taskPageCacheName); err != nil {
        h.errorLogService.LogError("Task_done_cache_invalidation_pages", err)
    }

    if err := h.cache.Delete(fmt.Sprintf(taskCacheKey, updatedTask.ID)); err != nil {
        h.errorLogService.LogError("Task_done_cache_invalidation_item", err)
    }
    
    // 3. Small delay to ensure DB consistency before next read
    time.Sleep(time.Millisecond * 50)

    addCacheHeaders(c, true)
    c.JSON(http.StatusOK, response)
}

func (h *Handler) Delete(c *gin.Context) {
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

    // 1. Invalidate by tags
    if err := h.cache.InvalidateByTags(taskServiceListRef, fmt.Sprintf(errTaskReference, id)); err != nil {
        h.errorLogService.LogError("Task_delete_cache_invalidation_tags", err)
    }
    
    // 2. Also directly invalidate specific cache entries
    if err := h.cache.Delete(fmt.Sprintf(taskCacheKey, id)); err != nil {
        h.errorLogService.LogError("Task_delete_cache_invalidation_item", err)
    }
    
    // 3. Invalidate all page caches
    if err := h.cache.Delete(taskPageCacheName); err != nil {
        h.errorLogService.LogError("Task_delete_cache_invalidation_pages", err)
    }
    
    // 4. Invalidate cursor-based list caches
    if err := h.cache.Delete("tasks_cursor_*"); err != nil {
        h.errorLogService.LogError("Task_delete_cache_invalidation_cursor", err)
    }
    
    // 5. Small delay to ensure DB consistency before next read
    time.Sleep(time.Millisecond * 50)


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