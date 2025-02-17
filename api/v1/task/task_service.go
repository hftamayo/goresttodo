package task

import (
	"errors"

	"github.com/hftamayo/gotodo/api/v1/models"
)

type TaskService struct {
	repo TaskRepository
}

func NewTodoService(repo TaskRepository) *TaskService {
	return &TaskService{repo}
}

func (s *TaskService) List() ([]*models.Task, error) {
	return s.repo.List(1, 10)
}

func (s *TaskService) ListById(id int) (*models.Task, error) {
	return s.repo.ListById(id)
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
