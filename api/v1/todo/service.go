package todo

import (
	"errors"

	"github.com/hftamayo/gotodo/api/v1/models"
)

type TodoService struct {
	repo TodoRepository
}

func NewTodoService(repo TodoRepository) *TodoService {
	return &TodoService{repo}
}

func (s *TodoService) CreateTodo(todo *models.Todo) error {
	return s.repo.Create(todo)
}

func (s *TodoService) UpdateTodo(todo *models.Todo) error {
	existingTodo, err := s.repo.GetById(int(todo.ID))
	if err != nil {
		return err
	}

	if existingTodo == nil {
		return errors.New("Todo not found")
	}

	return s.repo.Update(todo)
}

func (s *TodoService) MarkTodoAsDone(id int, done bool) (*models.Todo, error) {
	// Fetch the existing todo from the database.
	existingTodo, err := s.repo.GetById(id)
	if err != nil {
		return nil, err
	}

	// If the todo does not exist, return an error.
	if existingTodo == nil {
		return nil, errors.New("Todo not found")
	}

	// Mark the todo as done.
	existingTodo.Done = done

	// Save the updated todo in the database.
	err = s.repo.Update(existingTodo)
	if err != nil {
		return nil, err
	}
	return existingTodo, nil
}

func (s *TodoService) GetAllTodos() ([]*models.Todo, error) {
	return s.repo.GetAll(1, 10)
}

func (s *TodoService) GetTodoById(id int) (*models.Todo, error) {
	return s.repo.GetById(id)
}

func (s *TodoService) DeleteTodoById(id int) error {
	return s.repo.Delete(id)
}
