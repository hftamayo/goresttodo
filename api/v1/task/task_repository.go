package task

import (
	"github.com/hftamayo/gotodo/api/v1/models"
)

type TaskRepository interface {
	List(page, pageSize int) ([]*models.Task, error)
	ListById(id int) (*models.Task, error)
	Create(task *models.Task) error
	Update(task *models.Task) error
	Delete(id int) error
}

// Ensure TaskRepositoryImpl implements TaskRepository at compile time
var _ TaskRepository = (*TaskRepositoryImpl)(nil)