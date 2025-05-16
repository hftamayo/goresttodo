package task

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/utils"
)

const (
    defaultCacheTime        = 10 * time.Minute
    serviceDefaultLimit     = 10
    serviceMaxLimit         = 100
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
    // Try to get from cache first
    cacheKey := fmt.Sprintf("tasks_cursor_%s_limit_%d_order_%s", cursor, limit, order)
    if err := s.cache.Get(cacheKey, &cachedData); err == nil {
        return cachedData.Tasks, cachedData.Pagination.NextCursor, cachedData.Pagination.PrevCursor, cachedData.TotalCount, nil
    }

    // Get from repository
    tasks, nextCursor, prevCursor, err := s.repo.List(limit, cursor, order)
    if err != nil {
        return nil, "", "", 0, fmt.Errorf("failed to list tasks: %w", err)
    }

    // Get total count
    totalCount, err := s.repo.GetTotalCount()
    if err != nil {
        return nil, "", "", 0, fmt.Errorf("failed to get total count: %w", err)
    }

    // Calculate current page
    currentPage := 1
    if cursor != "" {
        currentPage = int(totalCount/int64(limit)) + 1
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
            HasMore:     nextCursor != "",
            Limit:       limit,
            TotalCount:  totalCount,
            CurrentPage: currentPage,
        },
        TotalCount: totalCount,
    }
    
    if cacheBytes, err := json.Marshal(cacheData); err == nil {
        s.cache.Set(cacheKey, string(cacheBytes), defaultCacheTime)
    }

    return tasks, nextCursor, prevCursor, totalCount, nil
}

func (s *TaskService) ListById(id int) (*models.Task, error) {
	var task *models.Task
    cacheKey := fmt.Sprintf("task_%d", id)

    if err := s.cache.Get(cacheKey, &task); err == nil {
        return task, nil
    }

    task, err := s.repo.ListById(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get task by id: %w", err)
    }

    if task == nil {
        return nil, fmt.Errorf("task with id %d not found", id)
    }

    s.cache.Set(cacheKey, task, 10*time.Minute)
    return task, nil
}

func (s *TaskService) Create(task *models.Task) (*models.Task, error) {
    if task == nil {
        return nil, fmt.Errorf("invalid task data")
    }

    createdTask, err := s.repo.Create(task)
    if err != nil {
        return nil, fmt.Errorf("failed to create task: %w", err)
    }

    // Invalidate list cache
    s.cache.Delete("tasks_list*")
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
        return nil, fmt.Errorf("task with id %d not found", task.ID)
    }

    // Prevent modification of immutable fields
    task.ID = uint(id)
    task.CreatedAt = existingTask.CreatedAt    

    updatedTask, err := s.repo.Update(id, task)
     if err != nil {
        return nil, fmt.Errorf("failed to update task: %w", err)
    }

    // Invalidate caches
    s.cache.Delete(fmt.Sprintf("task_%d", task.ID))
    s.cache.Delete("tasks_list*")
    return updatedTask, nil
}

func (s *TaskService) MarkAsDone(id int) (*models.Task, error) {
    existingTask, err := s.repo.ListById(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get task: %w", err)
    }

    if existingTask == nil {
        return nil, fmt.Errorf("task with id %d not found", id)
    }

    updatedTask, err := s.repo.MarkAsDone(id)
    if err != nil {
        return nil, fmt.Errorf("failed to mark task as done: %w", err)
    }

    // Invalidate caches
    s.cache.Delete(fmt.Sprintf("task_%d", id))
    s.cache.Delete("tasks_list*")
    return updatedTask, nil
}

func (s *TaskService) Delete(id int) error {
    if err := s.repo.Delete(id); err != nil {
        return fmt.Errorf("failed to delete task: %w", err)
    }

    // Invalidate caches
    s.cache.Delete(fmt.Sprintf("task_%d", id))
    s.cache.Delete("tasks_list*")
    return nil
}
