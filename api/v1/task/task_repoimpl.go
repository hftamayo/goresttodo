package task

import (
	"errors"
	"fmt"

	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/cursor"
	"gorm.io/gorm"
)

type TaskRepositoryImpl struct {
	db *gorm.DB
}

const (
    defaultLimit = 10
    maxLimit    = 100
)

func NewTaskRepositoryImpl(db *gorm.DB) TaskRepository {
	if db == nil {
		fmt.Errorf("database connection is required")
		return nil
	}
	return &TaskRepositoryImpl{db: db}
}

func (r *TaskRepositoryImpl) GetTotalCount() (int64, error) {
    var count int64
    result := r.db.Model(&models.Task{}).Count(&count)
    if result.Error != nil {
        return 0, fmt.Errorf("failed to get total count: %w", result.Error)
    }
    
    return count, nil
}

func (r *TaskRepositoryImpl) List(limit int, cursorStr string) ([]*models.Task, string, error) {
    if limit <= 0 {
        limit = defaultLimit
    }
    if limit > maxLimit {
        limit = maxLimit
    }

    query := r.db.Model(&models.Task{}).
    Order("created_at DESC, id DESC").
    Select("id, title, description, done, owner, created_at, updated_at")
    // If cursor is provided, decode and apply conditions
    if cursorStr != "" {
        c, err := cursor.Decode[uint](cursorStr)
        if err != nil {
            return nil, "", fmt.Errorf("invalid cursor: %w", err)
        }
        
        query = query.Where("(created_at < ?) OR (created_at = ? AND id < ?)", 
            c.Timestamp, 
            c.Timestamp, 
            c.ID)
    }

    var tasks []*models.Task
    if err := query.Limit(limit + 1).Find(&tasks).Error; err != nil {
        return nil, "", fmt.Errorf("failed to fetch tasks: %w", err)
    }

    var nextCursor string
    hasMore := len(tasks) > limit

    if hasMore {
        lastTask := tasks[len(tasks)-1]
        c := cursor.Cursor[uint]{
            ID:        lastTask.ID,
            Timestamp: lastTask.CreatedAt,
        }
        nextCursor, _ = cursor.Encode(c, cursor.Options{
            Field:     "created_at",
            Direction: "DESC",
        })
        tasks = tasks[:limit]
    }

    return tasks, nextCursor, nil
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

func (r *TaskRepositoryImpl) Update(id int, task *models.Task) (*models.Task, error) {
    if task == nil {
        return nil, errors.New("task cannot be nil")
    }
	
     var existingTask models.Task
    if err := r.db.First(&existingTask, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, fmt.Errorf("task not found: %d", id)
        }
        return nil, fmt.Errorf("failed to verify task existence: %w", err)
    }

    task.Owner = existingTask.Owner

    tx := r.db.Begin()
    if err := tx.Save(task).Error; err != nil {
        tx.Rollback()
        return nil, fmt.Errorf("failed to update task: %w", err)
    }

    var updatedTask models.Task
    if err := tx.First(&updatedTask, id).Error; err != nil {
        tx.Rollback()
        return nil, fmt.Errorf("failed to fetch updated task: %w", err)
    }


    if err := tx.Commit().Error; err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    return &updatedTask, nil
}

func (r *TaskRepositoryImpl) MarkAsDone(id int) (*models.Task, error) {
    var task models.Task
    result := r.db.Model(&models.Task{}).Where("id = ?", id).Update("done", true)
    if result.Error != nil {
        return nil, fmt.Errorf("failed to mark task as done: %w", result.Error)
    }
    if result.RowsAffected == 0 {
        return fmt.Errorf("task not found: %d", task.ID)
    }
    
    if err := r.db.First(&task, id).Error; err != nil {
        return nil, fmt.Errorf("failed to fetch updated task: %w", err)
    }
    
    return &task, nil
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
