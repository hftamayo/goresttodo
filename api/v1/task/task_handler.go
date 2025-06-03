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
    // cacheKey := fmt.Sprintf("tasks_cursor_%s_limit_%d_order_%s", query.Cursor, query.Limit, query.Order)
    // var cachedResponse TaskOperationResponse
    // if err := h.cache.Get(cacheKey, &cachedResponse); err == nil {
    //     if cachedData, ok := cachedResponse.Data.(TaskListResponse); ok {
    //         etag := cachedData.ETag
            
    //         // Check if client's cached version is still valid
    //         if ifNoneMatch := c.GetHeader("If-None-Match"); ifNoneMatch != "" && 
    //             (ifNoneMatch == etag || ifNoneMatch == "W/"+etag) {
    //             c.Status(http.StatusNotModified)
    //             return
    //         }

    //         setEtagHeader(c, cachedData.ETag)
    //         c.Header("Last-Modified", cachedData.LastModified)
    //         addCacheHeaders(c, false)
            
    //         c.JSON(http.StatusOK, cachedResponse)
    //     } else {
    //         // Invalid cache data type, log and continue with fresh data
    //         h.errorLogService.LogError("Task_list_cache_type", 
    //             fmt.Errorf("unexpected type for cached data"))
    //     }
    //     return
    // }

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
    // if err := h.cache.Set(cacheKey, response, utils.DefaultCacheTime); err != nil {
    //     h.errorLogService.LogError("Task_list_cache_set", err)
    // }    

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

    cacheKey := fmt.Sprintf("task_%d", id)
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
    c.Header("Last-Modified", lastModified)    

    // Cache the response
    if err := h.cache.Set(cacheKey, response, time.Duration(response.CacheTTL)*time.Second); err != nil {
        h.errorLogService.LogError("Task_list_cache_set", err)
    }        

    addCacheHeaders(c, false)

    c.JSON(http.StatusOK, response)
}

func (h *Handler) ListByPage(c *gin.Context) {
    _, forceFresh := c.GetQuery("_t")

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

    cacheKey := fmt.Sprintf("tasks_page_%d_limit_%d_order_%s", query.Page, query.Limit, query.Order)    

    // Try to get from cache only if not forcing fresh data
    if forceFresh {
        c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
        c.Header("Pragma", "no-cache")
        c.Header("Expires", "0")
    } else {
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
                return
            }
        }
    }

    // Get data from service
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

    // If no tasks found, return empty response
    if len(tasks) == 0 {
        response := TaskOperationResponse{
            Code:          http.StatusOK,
            ResultMessage: utils.OperationSuccess,
            Data: TaskListResponse{
                Tasks: []*TaskResponse{},
                Pagination: PaginationMeta{
                    Limit:       query.Limit,
                    TotalCount:  0,
                    CurrentPage: query.Page,
                    TotalPages:  0,
                    Order:       query.Order,
                    IsFirstPage: true,
                    IsLastPage:  true,
                    HasMore:     false,
                    HasPrev:     false,
                },
                ETag:          generateETag([]*models.Task{}),
                LastModified:  time.Now().UTC().Format(http.TimeFormat),
            },
            Timestamp:     time.Now().Unix(),
            CacheTTL:      30,
        }

        // Set headers
        setEtagHeader(c, response.Data.(TaskListResponse).ETag)
        c.Header("Last-Modified", response.Data.(TaskListResponse).LastModified)

        // Cache the empty response
        if !forceFresh {
            if err := h.cache.SetWithTags(cacheKey, response, utils.DefaultCacheTime, 
                "tasks:list", fmt.Sprintf("tasks:page:%d", query.Page)); err != nil {
                h.errorLogService.LogError("Task_list_by_page_cache_set", err)
            }
        }

        addCacheHeaders(c, false)
        c.JSON(http.StatusOK, response)
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
        CacheTTL:      30,
    }

    // Set headers
    setEtagHeader(c, etag)
    c.Header("Last-Modified", lastModified)

    // Cache the response with tags
    if !forceFresh {
        if err := h.cache.SetWithTags(cacheKey, response, utils.DefaultCacheTime, 
            "tasks:list", fmt.Sprintf("tasks:page:%d", query.Page)); err != nil {
            h.errorLogService.LogError("Task_list_by_page_cache_set", err)
        }
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
    if err := h.cache.InvalidateByTags("tasks:list"); err != nil {
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
    if err := h.cache.InvalidateByTags("tasks:list", fmt.Sprintf("task:%d", id)); err != nil {
        h.errorLogService.LogError("Task_update_cache_invalidation_tags", err)
    }
    
    // 2. Also directly invalidate specific item cache
    if err := h.cache.Delete(fmt.Sprintf("task_%d", id)); err != nil {
        h.errorLogService.LogError("Task_update_cache_invalidation_item", err)
    }
    
    // 3. Invalidate all page caches since we don't know which pages this task appears on
    if err := h.cache.Delete("tasks_page_*"); err != nil {
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
    if err := h.cache.InvalidateByTags("tasks:list", fmt.Sprintf("task:%d", id)); err != nil {
        h.errorLogService.LogError("Task_done_cache_invalidation_tags", err)
    }
    
    // 2. Also directly invalidate specific cache entries
    if err := h.cache.Delete("tasks_cursor_*"); err != nil {
        h.errorLogService.LogError("Task_done_cache_invalidation_cursor", err)
    }

    if err := h.cache.Delete("tasks_page_*"); err != nil {
        h.errorLogService.LogError("Task_done_cache_invalidation_pages", err)
    }

    if err := h.cache.Delete(fmt.Sprintf("task_%d", updatedTask.ID)); err != nil {
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
    if err := h.cache.InvalidateByTags("tasks:list", fmt.Sprintf("task:%d", id)); err != nil {
        h.errorLogService.LogError("Task_delete_cache_invalidation_tags", err)
    }
    
    // 2. Also directly invalidate specific cache entries
    if err := h.cache.Delete(fmt.Sprintf("task_%d", id)); err != nil {
        h.errorLogService.LogError("Task_delete_cache_invalidation_item", err)
    }
    
    // 3. Invalidate all page caches
    if err := h.cache.Delete("tasks_page_*"); err != nil {
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
        c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
        c.Header("Pragma", "no-cache")
        c.Header("Expires", "0")
    } else {
        // For GET - allow short caching
        if strings.Contains(c.Request.URL.Path, "/list") {
            c.Header("Cache-Control", "no-cache, max-age=0")
        } else {
            // Individual item endpoints can cache longer
            c.Header("Cache-Control", "private, max-age=60")
        }
        c.Header("Vary", "Authorization")
    }
}