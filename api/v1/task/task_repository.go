package task

import (
	"github.com/hftamayo/gotodo/api/v1/models"
)

type TaskRepository interface {
	List(limit int, cursor string) ([]*models.Task, string, error)
	ListById(id int) (*models.Task, error)
	Create(task *models.Task) error
	Update(id int, task *models.Task) error
	MarkAsDone(id int) (*models.Task, error)
	Delete(id int) error
	GetTotalCount() (int64, error)
}

// Ensure TaskRepositoryImpl implements TaskRepository at compile time
var _ TaskRepository = (*TaskRepositoryImpl)(nil)