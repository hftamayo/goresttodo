package task

import (
	"github.com/hftamayo/gotodo/api/v1/models"
)

type TaskRepository interface {
    List(limit int, cursor string, order string) ([]*models.Task, string, string, error)
	ListById(id int) (*models.Task, error)
	Create(task *models.Task) (*models.Task, error)
	Update(id int, task *models.Task)(*models.Task, error)
	MarkAsDone(id int) (*models.Task, error)
	Delete(id int) error
	GetTotalCount() (int64, error)
	ListByPage(page int, limit int, order string) ([]*models.Task, int64, error)
}

// Ensure TaskRepositoryImpl implements TaskRepository at compile time
var _ TaskRepository = (*TaskRepositoryImpl)(nil)