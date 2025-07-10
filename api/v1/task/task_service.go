package task

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/config"
	"github.com/hftamayo/gotodo/pkg/utils"
)

type TaskService struct {
	repo       TaskRepository
	cache      *utils.Cache
	errorLog   config.ErrorLogger
}

var _ TaskServiceInterface = (*TaskService)(nil)

var cachedData struct {
    Tasks      []*models.Task        `json:"tasks"`
    Pagination PaginationMeta  `json:"pagination"`
    TotalCount int64          `json:"totalCount"`
}

func NewTaskService(repo TaskRepository, cache *utils.Cache) TaskServiceInterface {
	// Create a memory-based error logger for testing/development
	// In production, this would be injected or configured
	errorLogger := config.NewMemoryErrorLogger()
	nonBlockingLogger := config.NewNonBlockingErrorLogger(errorLogger)
	
	return &TaskService{
		repo:     repo, 
		cache:    cache,
		errorLog: nonBlockingLogger,
	}
}

// NewTaskServiceWithErrorLogger creates a task service with a custom error logger
func NewTaskServiceWithErrorLogger(repo TaskRepository, cache *utils.Cache, errorLog config.ErrorLogger) TaskServiceInterface {
	return &TaskService{
		repo:     repo,
		cache:    cache,
		errorLog: errorLog,
	}
}

func (s *TaskService) List(cursor string, limit int, order string) ([]*models.Task, string, string, int64, error) {
    // Use the validation helper
    query := validatePaginationQuery(CursorPaginationQuery{
        Cursor: cursor,
        Limit:  limit,
        Order:  order,
    })

    // Try to get from cache first
    cacheKey := fmt.Sprintf("tasks_cursor_%s_limit_%d_order_%s", 
        query.Cursor, query.Limit, query.Order)
    if err := s.cache.Get(cacheKey, &cachedData); err == nil {
        return cachedData.Tasks, cachedData.Pagination.NextCursor, 
            cachedData.Pagination.PrevCursor, cachedData.TotalCount, nil
    }

    // Get from repository with extra record for hasMore check
    tasks, nextCursor, prevCursor, err := s.repo.List(query.Limit+1, 
        query.Cursor, query.Order)
    if err != nil {
        return nil, "", "", 0, fmt.Errorf("failed to list tasks: %w", err)
    }

    // Get total count
    totalCount, err := s.repo.GetTotalCount()
    if err != nil {
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
    
    

    if cacheBytes, err := json.Marshal(cacheData); err == nil {
        s.cache.Set(cacheKey, string(cacheBytes), utils.DefaultCacheTime)
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
        if err := s.cache.InvalidateByTags(fmt.Sprintf(errTaskReference, id)); err != nil {
            s.errorLog.LogError(context.Background(), "task-service", "validate-task-existence", 
                fmt.Sprintf("Failed to invalidate cache for task %d: %v", id, err), 
                map[string]interface{}{"task_id": id, "error": err.Error()})
        }
        return nil, fmt.Errorf("task with id %d not found", id)
    }

    // If task exists and hasn't been modified, return cached version
    if cachedTask != nil && cachedTask.UpdatedAt.Equal(existingTask.UpdatedAt) {
        return cachedTask, nil
    }

    // If task has been modified, invalidate cache and continue
    if err := s.cache.InvalidateByTags(fmt.Sprintf(errTaskReference, id)); err != nil {
        s.errorLog.LogError(context.Background(), "task-service", "validate-task-existence", 
            fmt.Sprintf("Failed to invalidate cache for task %d: %v", id, err), 
            map[string]interface{}{"task_id": id, "error": err.Error()})
    }

    return existingTask, nil
}

// ListById retrieves a task by its ID
func (s *TaskService) ListById(id int) (*models.Task, error) {
    cacheKey := fmt.Sprintf("task_%d", id)
    var task *models.Task

    // Try to get from cache first
    if err := s.cache.Get(cacheKey, &task); err == nil {
        return s.validateTaskExistence(id, task)
    }

    // Get fresh data from repository
    task, err := s.repo.ListById(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get task by id: %w", err)
    }

    if task == nil {
        return nil, fmt.Errorf("task with id %d not found", id)
    }

    // Cache the result with tags
    if err := s.cache.SetWithTags(cacheKey, task, utils.DefaultCacheTime, 
        fmt.Sprintf(errTaskReference, id)); err != nil {
        s.errorLog.LogError(context.Background(), "task-service", "list-by-id", 
            fmt.Sprintf("Failed to cache task %d: %v", id, err), 
            map[string]interface{}{"task_id": id, "error": err.Error()})
    }

    return task, nil
}

func (s *TaskService) Create(task *models.Task) (*models.Task, error) {
    if task == nil {
        return nil, fmt.Errorf("invalid task data")
    }

    existingTask, err := s.repo.SearchByTitle(task.Title)
    if err != nil {
        return nil, fmt.Errorf("failed to check for duplicate title: %w", err)
    }

    if existingTask != nil {
        return nil, fmt.Errorf("task with title %s already exists", task.Title)
    }

    createdTask, err := s.repo.Create(task)
    if err != nil {
        return nil, fmt.Errorf("failed to create task: %w", err)
    }

    // Invalidate all task-related caches
    if err := s.cache.InvalidateByTags(taskServiceListRef); err != nil {
        s.errorLog.LogError(context.Background(), "task-service", "create", 
            fmt.Sprintf("Failed to invalidate cache for task list: %v", err), 
            map[string]interface{}{"error": err.Error()})
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
        return nil, fmt.Errorf("failed to verify task existence: %w", err)
    }

    if existingTask == nil {
        return nil, fmt.Errorf(errTaskNotFoundFmt, task.ID)
    }

    // Prevent modification of immutable fields
    task.ID = uint(id)
    task.CreatedAt = existingTask.CreatedAt    

    updatedTask, err := s.repo.Update(id, task)
    if err != nil {
        return nil, fmt.Errorf("failed to update task: %w", err)
    }

    // Invalidate all task-related caches
    if err := s.cache.InvalidateByTags(taskServiceListRef, fmt.Sprintf(errTaskReference, id)); err != nil {
        s.errorLog.LogError(context.Background(), "task-service", "update", 
            fmt.Sprintf("Failed to invalidate cache for task %d update: %v", id, err), 
            map[string]interface{}{"task_id": id, "error": err.Error()})
    }

    return updatedTask, nil
}

func (s *TaskService) MarkAsDone(id int) (*models.Task, error) {
    existingTask, err := s.repo.ListById(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get task: %w", err)
    }

    if existingTask == nil {
        return nil, fmt.Errorf(errTaskNotFoundFmt, id)
    }

    updatedTask, err := s.repo.MarkAsDone(id)
    if err != nil {
        return nil, fmt.Errorf("failed to mark task as done: %w", err)
    }

    // Invalidate all task-related caches
    if err := s.cache.InvalidateByTags(taskServiceListRef, fmt.Sprintf(errTaskReference, id)); err != nil {
        s.errorLog.LogError(context.Background(), "task-service", "mark-as-done", 
            fmt.Sprintf("Failed to invalidate cache for task %d mark as done: %v", id, err), 
            map[string]interface{}{"task_id": id, "error": err.Error()})
    }

    return updatedTask, nil
}

func (s *TaskService) Delete(id int) error {
    if err := s.repo.Delete(id); err != nil {
        return fmt.Errorf("failed to delete task: %w", err)
    }

    // Invalidate all task-related caches
    if err := s.cache.InvalidateByTags(taskServiceListRef, fmt.Sprintf(errTaskReference, id)); err != nil {
        s.errorLog.LogError(context.Background(), "task-service", "delete", 
            fmt.Sprintf("Failed to invalidate cache for task %d delete: %v", id, err), 
            map[string]interface{}{"task_id": id, "error": err.Error()})
    }

    return nil
}

// ListByPage retrieves a paginated list of tasks
func (s *TaskService) ListByPage(page, limit int, order string) ([]*models.Task, int64, error) {
    cacheKey := fmt.Sprintf("tasks_page_%d_%d_%s", page, limit, order)
    var cachedData struct {
        Tasks      []*models.Task `json:"tasks"`
        TotalCount int64         `json:"totalCount"`
    }
    
    if err := s.cache.Get(cacheKey, &cachedData); err == nil {
        // If we have cached data, verify it's still valid
        totalCount, err := s.repo.GetTotalCount()
        if err != nil {
            return nil, 0, fmt.Errorf("failed to get total count: %w", err)
        }

        // If total count matches, return cached data
        if totalCount == cachedData.TotalCount {
            return cachedData.Tasks, cachedData.TotalCount, nil
        }

        // If total count doesn't match, invalidate cache and continue
        if err := s.cache.InvalidateByTags("tasks:list"); err != nil {
            s.errorLog.LogError(context.Background(), "task-service", "list-by-page", 
                fmt.Sprintf("Failed to invalidate cache for tasks list: %v", err), 
                map[string]interface{}{"error": err.Error()})
        }
    }

    // Get fresh data from repository
    tasks, totalCount, err := s.repo.ListByPage(page, limit, order)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to list tasks by page: %w", err)
    }

    // Cache the results with tags
    if err := s.cache.SetWithTags(cacheKey, struct {
        Tasks      []*models.Task `json:"tasks"`
        TotalCount int64         `json:"totalCount"`
    }{
        Tasks:      tasks,
        TotalCount: totalCount,
    }, utils.DefaultCacheTime, "tasks:list", fmt.Sprintf("tasks:page:%d", page)); err != nil {
        s.errorLog.LogError(context.Background(), "task-service", "list-by-page", 
            fmt.Sprintf("Failed to cache tasks: %v", err), 
            map[string]interface{}{"error": err.Error()})
    }

    return tasks, totalCount, nil
}