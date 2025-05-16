package task

import "github.com/hftamayo/gotodo/api/v1/models"

type TaskServiceInterface interface {
    List(cursor string, limit int, order string) ([]*models.Task, string, string, int64, error)
    ListById(id int) (*models.Task, error)
    Create(task *models.Task) (*models.Task, error)
    Update(id int, task *models.Task) (*models.Task, error)
    Delete(id int) error
    MarkAsDone(id int) (*models.Task, error)
}