package task

import "github.com/hftamayo/gotodo/api/v1/models"

type TaskServiceInterface interface {
    List(limit int, cursorStr string) ([]*models.Task, *CursorPaginationMeta, error)
    ListById(id int) (*models.Task, error)
    Create(task *models.Task) error
    Update(task *models.Task) error
    Delete(id int) error
    Done(id int, done bool) (*models.Task, error)
}