package todo

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
	return s.repo.Update(todo)
}
