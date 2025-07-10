package task

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/config"
	"github.com/hftamayo/gotodo/pkg/utils"
)

type TaskService struct {
	repo   TaskRepository
	cache  *utils.Cache
	config *TaskServiceConfig
}

var _ TaskServiceInterface = (*TaskService)(nil)

var cachedData struct {
	Tasks      []*models.Task `json:"tasks"`
	Pagination PaginationMeta `json:"pagination"`
	TotalCount int64          `json:"totalCount"`
}

func NewTaskService(repo TaskRepository, cache *utils.Cache, config *TaskServiceConfig) TaskServiceInterface {
	if config == nil {
		config = DefaultTaskServiceConfig()
	}
	return &TaskService{repo: repo, cache: cache, config: config}
}

// Legacy constructor for backward compatibility
func NewTaskServiceLegacy(repo TaskRepository, cache *utils.Cache) TaskServiceInterface {
	return NewTaskService(repo, cache, DefaultTaskServiceConfig())
}

func (s *TaskService) List(cursor string, limit int, order string) ([]*models.Task, string, string, int64, error) {
	// Use the validation helper
	query := validatePaginationQuery(CursorPaginationQuery{
		Cursor: cursor,
		Limit:  limit,
		Order:  order,
	})

	// Try to get from cache first if caching is enabled
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("tasks_cursor_%s_limit_%d_order_%s",
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
		s.logError("TaskService_List", err)
		return nil, "", "", 0, fmt.Errorf("failed to list tasks: %w", err)
	}

	// Get total count
	totalCount, err := s.repo.GetTotalCount()
	if err != nil {
		s.logError("TaskService_List_TotalCount", err)
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
		Tasks      []*models.Task `json:"tasks"`
		Pagination PaginationMeta `json:"pagination"`
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

	// Cache the result if caching is enabled
	if s.config.EnableCache {
		if cacheBytes, err := json.Marshal(cacheData); err == nil {
			s.cache.Set(fmt.Sprintf("tasks_cursor_%s_limit_%d_order_%s",
				query.Cursor, query.Limit, query.Order), string(cacheBytes), time.Duration(s.config.CacheTTL)*time.Second)
		}
	}

	return tasks, nextCursor, prevCursor, totalCount, nil
}

// validateTaskExistence checks if a task exists and hasn't been modified
func (s *TaskService) validateTaskExistence(id int, cachedTask *models.Task) (*models.Task, error) {
	existingTask, err := s.repo.ListById(id)
	if err != nil {
		s.logError("TaskService_ValidateTaskExistence", err)
		return nil, fmt.Errorf("failed to verify task existence: %w", err)
	}

	// If task doesn't exist anymore, invalidate cache and return error
	if existingTask == nil {
		s.InvalidateTaskCache(id)
		return nil, fmt.Errorf("task with id %d not found", id)
	}

	// If task exists and hasn't been modified, return cached version
	if cachedTask != nil && cachedTask.UpdatedAt.Equal(existingTask.UpdatedAt) {
		return cachedTask, nil
	}

	// If task has been modified, invalidate cache and continue
	s.InvalidateTaskCache(id)

	return existingTask, nil
}

// ListById retrieves a task by its ID
func (s *TaskService) ListById(id int) (*models.Task, error) {
	cacheKey := fmt.Sprintf("task_%d", id)
	var task *models.Task

	// Try to get from cache first if caching is enabled
	if s.config.EnableCache {
		if err := s.cache.Get(cacheKey, &task); err == nil {
			return s.validateTaskExistence(id, task)
		}
	}

	// Get fresh data from repository
	task, err := s.repo.ListById(id)
	if err != nil {
		s.logError("TaskService_ListById", err)
		return nil, fmt.Errorf("failed to get task by id: %w", err)
	}

	if task == nil {
		return nil, fmt.Errorf("task with id %d not found", id)
	}

	// Cache the result if caching is enabled
	if s.config.EnableCache {
		if err := s.cache.Set(cacheKey, task, time.Duration(s.config.CacheTTL)*time.Second); err != nil {
			s.logError("TaskService_ListById_Cache", err)
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
		s.logError("TaskService_Create_Search", err)
		return nil, fmt.Errorf("failed to check for duplicate title: %w", err)
	}

	if existingTask != nil {
		return nil, fmt.Errorf("task with title %s already exists", task.Title)
	}

	createdTask, err := s.repo.Create(task)
	if err != nil {
		s.logError("TaskService_Create", err)
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Invalidate caches if caching is enabled
	if s.config.EnableCache {
		s.InvalidateListCache()
		s.InvalidatePageCache()
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
		s.logError("TaskService_Update_Verify", err)
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
		s.logError("TaskService_Update", err)
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	// Invalidate caches if caching is enabled
	if s.config.EnableCache {
		s.InvalidateTaskCache(id)
		s.InvalidateListCache()
		s.InvalidatePageCache()
	}

	return updatedTask, nil
}

func (s *TaskService) MarkAsDone(id int) (*models.Task, error) {
	existingTask, err := s.repo.ListById(id)
	if err != nil {
		s.logError("TaskService_MarkAsDone_Get", err)
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if existingTask == nil {
		return nil, fmt.Errorf(errTaskNotFoundFmt, id)
	}

	updatedTask, err := s.repo.MarkAsDone(id)
	if err != nil {
		s.logError("TaskService_MarkAsDone", err)
		return nil, fmt.Errorf("failed to mark task as done: %w", err)
	}

	// Invalidate caches if caching is enabled
	if s.config.EnableCache {
		s.InvalidateTaskCache(id)
		s.InvalidateListCache()
		s.InvalidatePageCache()
	}

	return updatedTask, nil
}

func (s *TaskService) Delete(id int) error {
	if err := s.repo.Delete(id); err != nil {
		s.logError("TaskService_Delete", err)
		return fmt.Errorf("failed to delete task: %w", err)
	}

	// Invalidate caches if caching is enabled
	if s.config.EnableCache {
		s.InvalidateTaskCache(id)
		s.InvalidateListCache()
		s.InvalidatePageCache()
	}

	return nil
}

// ListByPage retrieves a paginated list of tasks
func (s *TaskService) ListByPage(page, limit int, order string) ([]*models.Task, int64, error) {
	cacheKey := fmt.Sprintf("tasks_page_%d_%d_%s", page, limit, order)
	var cachedData struct {
		Tasks      []*models.Task `json:"tasks"`
		TotalCount int64          `json:"totalCount"`
	}

	// Try to get from cache first if caching is enabled
	if s.config.EnableCache {
		if err := s.cache.Get(cacheKey, &cachedData); err == nil {
			// If we have cached data, verify it's still valid
			totalCount, err := s.repo.GetTotalCount()
			if err != nil {
				s.logError("TaskService_ListByPage_TotalCount", err)
				return nil, 0, fmt.Errorf("failed to get total count: %w", err)
			}

			// If total count matches, return cached data
			if totalCount == cachedData.TotalCount {
				return cachedData.Tasks, cachedData.TotalCount, nil
			}

			// If total count doesn't match, invalidate cache and continue
			s.InvalidatePageCache()
		}
	}

	// Get fresh data from repository
	tasks, totalCount, err := s.repo.ListByPage(page, limit, order)
	if err != nil {
		s.logError("TaskService_ListByPage", err)
		return nil, 0, fmt.Errorf("failed to list tasks by page: %w", err)
	}

	// Cache the results if caching is enabled
	if s.config.EnableCache {
		if err := s.cache.Set(cacheKey, struct {
			Tasks      []*models.Task `json:"tasks"`
			TotalCount int64          `json:"totalCount"`
		}{
			Tasks:      tasks,
			TotalCount: totalCount,
		}, time.Duration(s.config.CacheTTL)*time.Second); err != nil {
			s.logError("TaskService_ListByPage_Cache", err)
		}
	}

	return tasks, totalCount, nil
}

// Cache operations (moved from handler)
func (s *TaskService) InvalidateTaskCache(id int) error {
	if !s.config.EnableCache {
		return nil
	}
	
	// Invalidate specific task cache
	cacheKey := fmt.Sprintf("task_%d", id)
	if err := s.cache.Delete(cacheKey); err != nil {
		s.logError("TaskService_InvalidateTaskCache", err)
		return err
	}
	return nil
}

func (s *TaskService) InvalidateListCache() error {
	if !s.config.EnableCache {
		return nil
	}
	
	// Invalidate all list-related caches
	if err := s.cache.Delete("tasks_cursor_*"); err != nil {
		s.logError("TaskService_InvalidateListCache", err)
		return err
	}
	return nil
}

func (s *TaskService) InvalidatePageCache() error {
	if !s.config.EnableCache {
		return nil
	}
	
	// Invalidate all page-related caches
	if err := s.cache.Delete("tasks_page_*"); err != nil {
		s.logError("TaskService_InvalidatePageCache", err)
		return err
	}
	return nil
}

// Helper method for error logging
func (s *TaskService) logError(operation string, err error) {
	if s.config.EnableLogging && s.config.ErrorLogger != nil {
		if s.config.AsyncLogging {
			go func() {
				if logErr := s.config.ErrorLogger.LogError(operation, err); logErr != nil {
					// Only log if logging itself fails (rare)
					fmt.Printf("Failed to log error: %v\n", logErr)
				}
			}()
		} else {
			if logErr := s.config.ErrorLogger.LogError(operation, err); logErr != nil {
				fmt.Printf("Failed to log error: %v\n", logErr)
			}
		}
	}
}