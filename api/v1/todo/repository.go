package todo

import (
	"github.com/hftamayo/gotodo/api/v1/models"
)

type TodoRepository interface {
	GetById(id int) (*models.Todo, error)
	GetAll(page, pageSize int) ([]*models.Todo, error)
	Create(todo *models.Todo) error
	Update(todo *models.Todo) error
	Delete(id int) error
}
