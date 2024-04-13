package todo

type TodoRepository interface {
	GetById(id int) (*TodoModel, error)
	GetAll() ([]*TodoModel, error)
	Create(todo *TodoModel) error
	Update(todo *TodoModel) error
	Delete(id int) error
}
