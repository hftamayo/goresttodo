package task

import (
	"errors"
	"fmt"

	"github.com/hftamayo/gotodo/api/v1/models"
	"gorm.io/gorm"
)

type TaskRepositoryImpl struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	if db == nil {
		fmt.Errorf("database connection is required")
		return nil
	}
	return &TaskRepositoryImpl{db: db}
}


func (r *TaskRepositoryImpl) List(page, pageSize int) ([]*models.Task, error) {
	if page < 1 {
        return nil, fmt.Errorf("page must be greater than 0")
    }
    if pageSize < 1 || pageSize > 100 {
        return nil, fmt.Errorf("pageSize must be between 1 and 100")
    }

	var tasks []*models.Task
	offset := (page - 1) * pageSize

    result := r.db.
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
    if id < 1 {
        return nil, fmt.Errorf("invalid task id: %d", id)
    }

	var task models.Task
	if result := r.db.First(&task, id); result.Error != nil {
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
    if task == nil {
        return errors.New("task cannot be nil")
    }
	
	if result := r.db.Create(task); result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *TaskRepositoryImpl) Update(task *models.Task) error {
    if task == nil {
        return errors.New("task cannot be nil")
    }
	
     var existingTask models.Task
    if err := r.db.First(&existingTask, task.ID).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return fmt.Errorf("task not found: %d", task.ID)
        }
        return fmt.Errorf("failed to verify task existence: %w", err)
    }

    task.Owner = existingTask.Owner

    if err := r.db.Save(task).Error; err != nil {
        return fmt.Errorf("failed to update task: %w", err)
    }

    return nil
}

func (r *TaskRepositoryImpl) Delete(id int) error {
    if id < 1 {
        return fmt.Errorf("invalid task id: %d", id)
    }

    result := r.db.Delete(&models.Task{}, id)

    if result.Error != nil {
        return fmt.Errorf("failed to delete task: %w", result.Error)
    }

    if result.RowsAffected == 0 {
        return fmt.Errorf("task not found: %d", id)
    }

    return nil
}
