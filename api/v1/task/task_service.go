package task

import (
	"fmt"
	"time"

	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/utils"
)

var _ TaskServiceInterface = (*TaskService)(nil)

type TaskService struct {
	repo  TaskRepository
	cache *utils.Cache
}

func NewTaskService(repo TaskRepository, cache *utils.Cache) TaskServiceInterface {
	return &TaskService{repo: repo, cache: cache}
}

func (s *TaskService) List(page, pageSize int) ([]*models.Task, error) {
	var tasks []*models.Task
    cacheKey := fmt.Sprintf("tasks_list_page_%d_size_%d", page, pageSize)

	// Try to get tasks from cache
    if err := s.cache.Get(cacheKey, &tasks); err == nil {
        return tasks, nil
    }

    tasks, err := s.repo.List(page, pageSize)
    if err != nil {
        return nil, fmt.Errorf("failed to list tasks: %w", err)
    }

    s.cache.Set(cacheKey, tasks, 10*time.Minute)
    return tasks, nil
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

func (s *TaskService) Create(task *models.Task) error {
    if task == nil {
        return fmt.Errorf("invalid task data")
    }

    if err := s.repo.Create(task); err != nil {
        return fmt.Errorf("failed to create task: %w", err)
    }

    // Invalidate list cache
    s.cache.Delete("tasks_list*")
    return nil
}

func (s *TaskService) Update(task *models.Task) error {
    if task == nil {
        return ErrInvalidTask
    }

    existingTask, err := s.repo.ListById(int(task.ID))
    if err != nil {
        return fmt.Errorf("failed to verify task existence: %w", err)
    }

    if existingTask == nil {
        return fmt.Errorf("task with id %d not found", task.ID)
    }

    if err := s.repo.Update(task); err != nil {
        return fmt.Errorf("failed to update task: %w", err)
    }

    // Invalidate caches
    s.cache.Delete(fmt.Sprintf("task_%d", task.ID))
    s.cache.Delete("tasks_list*")
    return nil
}

func (s *TaskService) Done(id int, done bool) (*models.Task, error) {
    existingTask, err := s.repo.ListById(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get task: %w", err)
    }

    if existingTask == nil {
        return nil, fmt.Errorf("task with id %d not found", id)
    }

    existingTask.Done = done
    if err := s.repo.Update(existingTask); err != nil {
        return nil, fmt.Errorf("failed to update task status: %w", err)
    }

    // Invalidate caches
    s.cache.Delete(fmt.Sprintf("task_%d", id))
    s.cache.Delete("tasks_list*")
    return existingTask, nil
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
