package task

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/utils"
)

type TaskService struct {
	repo  TaskRepository
	cache *utils.Cache
}

var _ TaskServiceInterface = (*TaskService)(nil)

var cachedData struct {
    Tasks      []*models.Task        `json:"tasks"`
    Pagination PaginationMeta  `json:"pagination"`
    TotalCount int64          `json:"totalCount"`
}

func NewTaskService(repo TaskRepository, cache *utils.Cache) TaskServiceInterface {
	return &TaskService{repo: repo, cache: cache}
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

func (s *TaskService) ListById(id int) (*models.Task, error) {
    cacheKey := fmt.Sprintf("task_%d", id)
    var task *models.Task

    // Try to get from cache first
    if err := s.cache.Get(cacheKey, &task); err == nil {
        // Verify the task still exists and hasn't been modified
        existingTask, err := s.repo.ListById(id)
        if err != nil {
            return nil, fmt.Errorf("failed to verify task existence: %w", err)
        }

        // If task doesn't exist anymore, invalidate cache and return error
        if existingTask == nil {
            if err := s.cache.InvalidateByTags(fmt.Sprintf(errTaskReference, id)); err != nil {
                fmt.Printf(errInvalidateCacheFmt, err)
            }
            return nil, fmt.Errorf(errTaskNotFoundFmt, id)
        }

        // If task exists and hasn't been modified, return cached version
        if task != nil && task.UpdatedAt.Equal(existingTask.UpdatedAt) {
            return task, nil
        }

        // If task has been modified, invalidate cache and continue
        if err := s.cache.InvalidateByTags(fmt.Sprintf(errTaskReference, id)); err != nil {
            fmt.Printf(errInvalidateCacheFmt, err)
        }
    }

    // Get fresh data from repository
    task, err := s.repo.ListById(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get task by id: %w", err)
    }

    if task == nil {
        return nil, fmt.Errorf(errTaskNotFoundFmt, id)
    }

    // Cache the result with tags
    if err := s.cache.SetWithTags(cacheKey, task, utils.DefaultCacheTime, 
        fmt.Sprintf(errTaskReference, id)); err != nil {
        fmt.Printf("Failed to cache task: %v\n", err)
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
        fmt.Printf(errInvalidateCacheFmt, err)
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
        fmt.Printf(errInvalidateCacheFmt, err)
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
        fmt.Printf(errInvalidateCacheFmt, err)
    }

    return updatedTask, nil
}

func (s *TaskService) Delete(id int) error {
    if err := s.repo.Delete(id); err != nil {
        return fmt.Errorf("failed to delete task: %w", err)
    }

    // Invalidate all task-related caches
    if err := s.cache.InvalidateByTags(taskServiceListRef, fmt.Sprintf(errTaskReference, id)); err != nil {
        fmt.Printf(errInvalidateCacheFmt, err)
    }

    return nil
}

func (s *TaskService) ListByPage(page int, limit int, order string) ([]*models.Task, int64, error) {
    // Try to get from cache first
    cacheKey := fmt.Sprintf("tasks_page_%d_limit_%d_order_%s", page, limit, order)
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
        if err := s.cache.InvalidateByTags(taskServiceListRef); err != nil {
            fmt.Printf(errInvalidateCacheFmt, err)
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
    }, utils.DefaultCacheTime, taskServiceListRef, fmt.Sprintf("tasks:page:%d", page)); err != nil {
        fmt.Printf("Failed to cache tasks: %v\n", err)
    }

    return tasks, totalCount, nil
}
