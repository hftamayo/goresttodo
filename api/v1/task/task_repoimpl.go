package task

import (
	"errors"

	"github.com/hftamayo/gotodo/api/v1/models"
	"gorm.io/gorm"
)

type TaskRepositoryImpl struct {
	Db *gorm.DB
}

func (r *TaskRepositoryImpl) List(page, pageSize int) ([]*models.Task, error) {
	var tasks []*models.Task
	offset := (page - 1) * pageSize

    result := r.Db.
        Order("created_at DESC").
        Offset(offset).
        Limit(pageSize).
        Find(&tasks)

    if result.Error != nil {
        return nil, result.Error
    }

    return tasks, nil
}

func (r *TaskRepositoryImpl) ListById(id int) (*models.Task, error) {
	var task models.Task
	if result := r.Db.First(&task, id); result.Error != nil {
		// If the record is not found, GORM returns a "record not found" error.
		// You might want to return nil, nil in this case instead of nil, error.
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &task, nil
}

func (r *TaskRepositoryImpl) Create(task *models.Task) error {
	if result := r.Db.Create(task); result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *TaskRepositoryImpl) Update(task *models.Task) error {
     var existingTask models.Task
    if err := r.Db.First(&existingTask, task.ID).Error; err != nil {
        return err
    }

    task.Owner = existingTask.Owner

    return r.Db.Save(task).Error
}

func (r *TaskRepositoryImpl) Delete(id int) error {
	task := &models.Task{}
	if result := r.Db.First(task, id); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil
		}
		return result.Error
	}

	if result := r.Db.Delete(task); result.Error != nil {
		return result.Error
	}
	return nil
}
