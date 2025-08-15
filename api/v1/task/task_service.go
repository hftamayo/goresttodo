package task

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/config"
)

type TaskService struct {
	repo       TaskRepository
	cache      CacheInterface
	errorLog   config.ErrorLogger
	config     *TaskServiceConfig
}

var _ TaskServiceInterface = (*TaskService)(nil)

var cachedData struct {
    Tasks      []*models.Task        `json:"tasks"`
    Pagination PaginationMeta  `json:"pagination"`
    TotalCount int64          `json:"totalCount"`
}

// NewTaskService creates a new task service with default configuration
// This is the legacy constructor for backward compatibility
func NewTaskService(repo TaskRepository, cache CacheInterface) TaskServiceInterface {
	config := DefaultTaskServiceConfig()
	config.ErrorLogger = config.NewMemoryErrorLogger()
	if config.AsyncLogging {
		config.ErrorLogger = config.NewNonBlockingErrorLogger(config.ErrorLogger)
	}
	
	return NewTaskServiceWithConfig(repo, cache, config)
}

// NewTaskServiceWithConfig creates a task service with custom configuration
func NewTaskServiceWithConfig(repo TaskRepository, cache CacheInterface, config *TaskServiceConfig) TaskServiceInterface {
	if config == nil {
		config = DefaultTaskServiceConfig()
	}
	
	if config.ErrorLogger == nil {
		config.ErrorLogger = config.NewMemoryErrorLogger()
		if config.AsyncLogging {
			config.ErrorLogger = config.NewNonBlockingErrorLogger(config.ErrorLogger)
		}
	}
	
	return &TaskService{
		repo:     repo,
		cache:    cache,
		errorLog: config.ErrorLogger,
		config:   config,
	}
}

// NewTaskServiceWithErrorLogger creates a task service with a custom error logger
// This is kept for backward compatibility
func NewTaskServiceWithErrorLogger(repo TaskRepository, cache CacheInterface, errorLog config.ErrorLogger) TaskServiceInterface {
	config := DefaultTaskServiceConfig()
	config.ErrorLogger = errorLog
	
	return NewTaskServiceWithConfig(repo, cache, config)
}

func (s *TaskService) List(cursor string, limit int, order string) ([]*models.Task, string, string, int64, error) {
    // Use the validation helper
    query := validatePaginationQuery(CursorPaginationQuery{
        Cursor: cursor,
        Limit:  limit,
        Order:  order,
    })

    // Try to get from cache first if enabled
    if s.config.EnableCache {
        cacheKey := fmt.Sprintf(s.config.CacheKeys.TaskCursorKey, 
            query.Cursor, query.Limit, query.Order)
        if err := s.cache.Get(cacheKey, &cachedData); err == nil {
            return cachedData.Tasks, cachedData.Pagination.NextCursor, 
                cachedData.Pagination.PrevCursor, cachedData.TotalCount, nil
        }
    }

    // Get from repository with extra record for hasMore check
    tasks, nextCursor, prevCursor, err := s.repo.List(query.Limit+1, 
        query.Cursor, query.Order)
    if err != nil {
        s.logError("list", fmt.Sprintf("Failed to list tasks: %v", err), map[string]interface{}{"error": err.Error()})
        return nil, "", "", 0, fmt.Errorf("failed to list tasks: %w", err)
    }

    // Get total count
    totalCount, err := s.repo.GetTotalCount()
    if err != nil {
        s.logError("list", fmt.Sprintf("Failed to get total count: %v", err), map[string]interface{}{"error": err.Error()})
        return nil, "", "", 0, fmt.Errorf("failed to get total count: %w", err)
    }

    // Handle hasMore and slice tasks
    hasMore := len(tasks) > limit
    if hasMore {
        tasks = tasks[:limit] // Remove the extra record
    }

    // Calculate pagination metadata
    totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))
    currentPage := 1
    if cursor != "" {
        // For cursor-based pagination, we don't need to calculate current page
        // as it's not relevant to the user
        currentPage = 0
    }

    cacheData := struct {
        Tasks      []*models.Task  `json:"tasks"`
        Pagination PaginationMeta  `json:"pagination"`
        TotalCount int64          `json:"totalCount"`
    }{
        Tasks: tasks,
        Pagination: PaginationMeta{
            NextCursor:  nextCursor,
            PrevCursor:  prevCursor,
            HasMore:     hasMore,
            Limit:       limit,
            TotalCount:  totalCount,
            CurrentPage: currentPage,
            TotalPages:  totalPages,
            Order:       order,
        },
        TotalCount: totalCount,
    }
    
    // Cache the result if enabled
    if s.config.EnableCache {
        cacheKey := fmt.Sprintf(s.config.CacheKeys.TaskCursorKey, 
            query.Cursor, query.Limit, query.Order)
        if cacheBytes, err := json.Marshal(cacheData); err == nil {
            s.cache.Set(cacheKey, string(cacheBytes), s.config.CacheTTL)
        }
    }

    return tasks, nextCursor, prevCursor, totalCount, nil
}

// validateTaskExistence checks if a task exists and hasn't been modified
func (s *TaskService) validateTaskExistence(id int, cachedTask *models.Task) (*models.Task, error) {
    existingTask, err := s.repo.ListById(id)
    if err != nil {
        return nil, fmt.Errorf("failed to verify task existence: %w", err)
    }

    // If task doesn't exist anymore, invalidate cache and return error
    if existingTask == nil {
        if s.config.EnableCache {
            if err := s.cache.InvalidateByTags(fmt.Sprintf(s.config.CacheKeys.TaskReference, id)); err != nil {
                s.logError("validate-task-existence", 
                    fmt.Sprintf("Failed to invalidate cache for task %d: %v", id, err), 
                    map[string]interface{}{"task_id": id, "error": err.Error()})
            }
        }
        return nil, fmt.Errorf(s.config.ValidationConfig.ErrTaskNotFoundFmt, id)
    }

    // If task exists and hasn't been modified, return cached version
    if cachedTask != nil && cachedTask.UpdatedAt.Equal(existingTask.UpdatedAt) {
        return cachedTask, nil
    }

    // If task has been modified, invalidate cache and continue
    if s.config.EnableCache {
        if err := s.cache.InvalidateByTags(fmt.Sprintf(s.config.CacheKeys.TaskReference, id)); err != nil {
            s.logError("validate-task-existence", 
                fmt.Sprintf("Failed to invalidate cache for task %d: %v", id, err), 
                map[string]interface{}{"task_id": id, "error": err.Error()})
        }
    }

    return existingTask, nil
}

// ListById retrieves a task by its ID
func (s *TaskService) ListById(id int) (*models.Task, error) {
    var task *models.Task

    // Try to get from cache first if enabled
    if s.config.EnableCache {
        cacheKey := fmt.Sprintf(s.config.CacheKeys.TaskKey, id)
        if err := s.cache.Get(cacheKey, &task); err == nil {
            return s.validateTaskExistence(id, task)
        }
    }

    // Get fresh data from repository
    task, err := s.repo.ListById(id)
    if err != nil {
        s.logError("list-by-id", fmt.Sprintf("Failed to get task by id: %v", err), map[string]interface{}{"task_id": id, "error": err.Error()})
        return nil, fmt.Errorf("failed to get task by id: %w", err)
    }

    if task == nil {
        return nil, fmt.Errorf(s.config.ValidationConfig.ErrTaskNotFoundFmt, id)
    }

    // Cache the result with tags if enabled
    if s.config.EnableCache {
        cacheKey := fmt.Sprintf(s.config.CacheKeys.TaskKey, id)
        if err := s.cache.SetWithTags(cacheKey, task, s.config.CacheTTL, 
            fmt.Sprintf(s.config.CacheKeys.TaskReference, id)); err != nil {
            s.logError("list-by-id", 
                fmt.Sprintf("Failed to cache task %d: %v", id, err), 
                map[string]interface{}{"task_id": id, "error": err.Error()})
        }
    }

    return task, nil
}

func (s *TaskService) Create(task *models.Task) (*models.Task, error) {
    if task == nil {
        return nil, fmt.Errorf("invalid task data")
    }

    existingTask, err := s.repo.SearchByTitle(task.Title)
    if err != nil {
        s.logError("create", fmt.Sprintf("Failed to check for duplicate title: %v", err), map[string]interface{}{"error": err.Error()})
        return nil, fmt.Errorf("failed to check for duplicate title: %w", err)
    }

    if existingTask != nil {
        return nil, fmt.Errorf("task with title %s already exists", task.Title)
    }

    createdTask, err := s.repo.Create(task)
    if err != nil {
        s.logError("create", fmt.Sprintf("Failed to create task: %v", err), map[string]interface{}{"error": err.Error()})
        return nil, fmt.Errorf("failed to create task: %w", err)
    }

    // Invalidate all task-related caches if enabled
    if s.config.EnableCache {
        if err := s.cache.InvalidateByTags(s.config.CacheKeys.TaskListRef); err != nil {
            s.logError("create", 
                fmt.Sprintf("Failed to invalidate cache for task list: %v", err), 
                map[string]interface{}{"error": err.Error()})
        }
    }

    return createdTask, nil
}

func (s *TaskService) Update(id int, task *models.Task) (*models.Task, error) {
    if task == nil {
        return nil, fmt.Errorf("invalid task data")
    }

    if int(task.ID) != id {
        return nil, fmt.Errorf("inconsistent task ID between URL and body")
    }    

    existingTask, err := s.repo.ListById(id)
    if err != nil {
        s.logError("update", fmt.Sprintf("Failed to verify task existence: %v", err), map[string]interface{}{"task_id": id, "error": err.Error()})
        return nil, fmt.Errorf("failed to verify task existence: %w", err)
    }

    if existingTask == nil {
        return nil, fmt.Errorf(s.config.ValidationConfig.ErrTaskNotFoundFmt, task.ID)
    }

    // Prevent modification of immutable fields
    task.ID = uint(id)
    task.CreatedAt = existingTask.CreatedAt    

    updatedTask, err := s.repo.Update(id, task)
    if err != nil {
        s.logError("update", fmt.Sprintf("Failed to update task: %v", err), map[string]interface{}{"task_id": id, "error": err.Error()})
        return nil, fmt.Errorf("failed to update task: %w", err)
    }

    // Invalidate all task-related caches if enabled
    if s.config.EnableCache {
        if err := s.cache.InvalidateByTags(s.config.CacheKeys.TaskListRef, fmt.Sprintf(s.config.CacheKeys.TaskReference, id)); err != nil {
            s.logError("update", 
                fmt.Sprintf("Failed to invalidate cache for task %d update: %v", id, err), 
                map[string]interface{}{"task_id": id, "error": err.Error()})
        }
    }

    return updatedTask, nil
}

func (s *TaskService) MarkAsDone(id int) (*models.Task, error) {
    existingTask, err := s.repo.ListById(id)
    if err != nil {
        s.logError("mark-as-done", fmt.Sprintf("Failed to get task: %v", err), map[string]interface{}{"task_id": id, "error": err.Error()})
        return nil, fmt.Errorf("failed to get task: %w", err)
    }

    if existingTask == nil {
        return nil, fmt.Errorf(s.config.ValidationConfig.ErrTaskNotFoundFmt, id)
    }

    updatedTask, err := s.repo.MarkAsDone(id)
    if err != nil {
        s.logError("mark-as-done", fmt.Sprintf("Failed to mark task as done: %v", err), map[string]interface{}{"task_id": id, "error": err.Error()})
        return nil, fmt.Errorf("failed to mark task as done: %w", err)
    }

    // Invalidate all task-related caches if enabled
    if s.config.EnableCache {
        if err := s.cache.InvalidateByTags(s.config.CacheKeys.TaskListRef, fmt.Sprintf(s.config.CacheKeys.TaskReference, id)); err != nil {
            s.logError("mark-as-done", 
                fmt.Sprintf("Failed to invalidate cache for task %d mark as done: %v", id, err), 
                map[string]interface{}{"task_id": id, "error": err.Error()})
        }
    }

    return updatedTask, nil
}

func (s *TaskService) Delete(id int) error {
    if err := s.repo.Delete(id); err != nil {
        s.logError("delete", fmt.Sprintf("Failed to delete task: %v", err), map[string]interface{}{"task_id": id, "error": err.Error()})
        return fmt.Errorf("failed to delete task: %w", err)
    }

    // Invalidate all task-related caches if enabled
    if s.config.EnableCache {
        if err := s.cache.InvalidateByTags(s.config.CacheKeys.TaskListRef, fmt.Sprintf(s.config.CacheKeys.TaskReference, id)); err != nil {
            s.logError("delete", 
                fmt.Sprintf("Failed to invalidate cache for task %d delete: %v", id, err), 
                map[string]interface{}{"task_id": id, "error": err.Error()})
        }
    }

    return nil
}

// ListByPage retrieves a paginated list of tasks
func (s *TaskService) ListByPage(page, limit int, order string) ([]*models.Task, int64, error) {
    var cachedData struct {
        Tasks      []*models.Task `json:"tasks"`
        TotalCount int64         `json:"totalCount"`
    }
    
    // Try to get from cache first if enabled
    if s.config.EnableCache {
        cacheKey := fmt.Sprintf(s.config.CacheKeys.TaskPageKey, page, limit, order)
        if err := s.cache.Get(cacheKey, &cachedData); err == nil {
            // If we have cached data, verify it's still valid
            totalCount, err := s.repo.GetTotalCount()
            if err != nil {
                s.logError("list-by-page", fmt.Sprintf("Failed to get total count: %v", err), map[string]interface{}{"error": err.Error()})
                return nil, 0, fmt.Errorf("failed to get total count: %w", err)
            }

            // If total count matches, return cached data
            if totalCount == cachedData.TotalCount {
                return cachedData.Tasks, cachedData.TotalCount, nil
            }

            // If total count doesn't match, invalidate cache and continue
            if err := s.cache.InvalidateByTags("tasks:list"); err != nil {
                s.logError("list-by-page", 
                    fmt.Sprintf("Failed to invalidate cache for tasks list: %v", err), 
                    map[string]interface{}{"error": err.Error()})
            }
        }
    }

    // Get fresh data from repository
    tasks, totalCount, err := s.repo.ListByPage(page, limit, order)
    if err != nil {
        s.logError("list-by-page", fmt.Sprintf("Failed to list tasks by page: %v", err), map[string]interface{}{"error": err.Error()})
        return nil, 0, fmt.Errorf("failed to list tasks by page: %w", err)
    }

    // Cache the results with tags if enabled
    if s.config.EnableCache {
        cacheKey := fmt.Sprintf(s.config.CacheKeys.TaskPageKey, page, limit, order)
        if err := s.cache.SetWithTags(cacheKey, struct {
            Tasks      []*models.Task `json:"tasks"`
            TotalCount int64         `json:"totalCount"`
        }{
            Tasks:      tasks,
            TotalCount: totalCount,
        }, s.config.CacheTTL, "tasks:list", fmt.Sprintf("tasks:page:%d", page)); err != nil {
            s.logError("list-by-page", 
                fmt.Sprintf("Failed to cache tasks: %v", err), 
                map[string]interface{}{"error": err.Error()})
        }
    }

    return tasks, totalCount, nil
}

// Cache invalidation methods moved from handler
func (s *TaskService) InvalidateTaskCache(id int) error {
    if !s.config.EnableCache {
        return nil
    }
    
    cacheKey := fmt.Sprintf(s.config.CacheKeys.TaskKey, id)
    if err := s.cache.Delete(cacheKey); err != nil {
        s.logError("invalidate-task-cache", 
            fmt.Sprintf("Failed to delete task cache for id %d: %v", id, err), 
            map[string]interface{}{"task_id": id, "error": err.Error()})
        return err
    }
    return nil
}

func (s *TaskService) InvalidateListCache() error {
    if !s.config.EnableCache {
        return nil
    }
    
    if err := s.cache.InvalidateByTags(s.config.CacheKeys.TaskListRef); err != nil {
        s.logError("invalidate-list-cache", 
            fmt.Sprintf("Failed to invalidate list cache: %v", err), 
            map[string]interface{}{"error": err.Error()})
        return err
    }
    return nil
}

func (s *TaskService) InvalidatePageCache() error {
    if !s.config.EnableCache {
        return nil
    }
    
    if err := s.cache.Delete(s.config.CacheKeys.TaskPageCache); err != nil {
        s.logError("invalidate-page-cache", 
            fmt.Sprintf("Failed to invalidate page cache: %v", err), 
            map[string]interface{}{"error": err.Error()})
        return err
    }
    return nil
}

// logError is a helper method that handles error logging based on configuration
func (s *TaskService) logError(operation, errorMsg string, metadata map[string]interface{}) {
    if !s.config.EnableLogging {
        return
    }
    
    if s.config.AsyncLogging {
        // For async logging, we don't wait for the result
        go func() {
            s.errorLog.LogError(context.Background(), "task-service", operation, errorMsg, metadata)
        }()
    } else {
        // For sync logging, we wait for the result
        s.errorLog.LogError(context.Background(), "task-service", operation, errorMsg, metadata)
    }
}