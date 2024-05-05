package todo

import "errors"

type TodoService struct {
	repo TodoRepository
}

func NewTodoService(repo TodoRepository) *TodoService {
	return &TodoService{repo}
}

func (s *TodoService) CreateTodo(todo *Todo) error {
	return s.repo.Create(todo)
}

func (s *TodoService) UpdateTodo(todo *Todo) error {
	existingTodo, err := s.repo.GetById(todo.Id)
	if err != nil {
		return err
	}

	if existingTodo == nil {
		return errors.New("Todo not found")
	}

	return s.repo.Update(todo)
}
