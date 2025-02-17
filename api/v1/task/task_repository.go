package task

import (
	"github.com/hftamayo/gotodo/api/v1/models"
)

type TaskRepository interface {
	List(page, pageSize int) ([]*models.Task, error)
	ListById(id int) (*models.Task, error)
	Create(todo *models.Task) error
	Update(todo *models.Task) error
	Delete(id int) error
}
