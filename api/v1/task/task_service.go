package task

import (
	"errors"
	"strconv"
	"time"

	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/utils"
)

var _ TaskServiceInterface = (*TaskService)(nil)

type TaskService struct {
	repo  TaskRepository
	cache *utils.Cache
}

func NewTaskService(repo TaskRepository, cache *utils.Cache) *TaskService {
	return &TaskService{repo: repo, cache: cache}
}

func (s *TaskService) List() ([]*models.Task, error) {
	var tasks []*models.Task
	cacheKey := "tasks_list"

	// Try to get tasks from cache
	err := s.cache.Get(cacheKey, &tasks)
	if err == nil {
		return tasks, nil
	}

	// If cache miss, get tasks from repository
	tasks, err = s.repo.List(1, 10)
	if err != nil {
		return nil, err
	}

	// Cache the tasks
	s.cache.Set(cacheKey, tasks, 10*time.Minute)

	return tasks, nil
}

func (s *TaskService) ListById(id int) (*models.Task, error) {
	var task *models.Task
	cacheKey := "task_" + strconv.Itoa(id)

	// Try to get task from cache
	err := s.cache.Get(cacheKey, &task)
	if err == nil {
		return task, nil
	}

	// If cache miss, get task from repository
	task, err = s.repo.ListById(id)
	if err != nil {
		return nil, err
	}

	// Cache the task
	s.cache.Set(cacheKey, task, 10*time.Minute)

	return task, nil
}

func (s *TaskService) Create(task *models.Task) error {
	return s.repo.Create(task)
}

func (s *TaskService) Update(task *models.Task) error {
	existingTask, err := s.repo.ListById(int(task.ID))
	if err != nil {
		return err
	}

	if existingTask == nil {
		return errors.New("Todo not found")
	}

	return s.repo.Update(task)
}

func (s *TaskService) Done(id int, done bool) (*models.Task, error) {
	existingTask, err := s.repo.ListById(id)
	if err != nil {
		return nil, err
	}

	if existingTask == nil {
		return nil, errors.New("Todo not found")
	}

	existingTask.Done = done

	// Save the updated in the database.
	err = s.repo.Update(existingTask)
	if err != nil {
		return nil, err
	}
	return existingTask, nil
}

func (s *TaskService) Delete(id int) error {
	return s.repo.Delete(id)
}
