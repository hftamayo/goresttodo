package todo

type TodoRepository interface {
	GetById(id int) (*Todo, error)
	GetAll() ([]*Todo, error)
	Create(todo *Todo) error
	Update(todo *Todo) error
	Delete(id int) error
}
