package task

import "github.com/hftamayo/gotodo/api/v1/models"

type TaskServiceInterface interface {
    List(cursor string, limit int) ([]*models.Task, string, int64, error)
    ListById(id int) (*models.Task, error)
    Create(task *models.Task) error
    Update(id int, task *models.Task) error
    Delete(id int) error
    Done(id int) (*models.Task, error)
}