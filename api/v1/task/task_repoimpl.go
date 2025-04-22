package task

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hftamayo/gotodo/api/v1/models"
	"gorm.io/gorm"
)

type TaskRepositoryImpl struct {
	db *gorm.DB
}

func NewTaskRepositoryImpl(db *gorm.DB) TaskRepository {
	if db == nil {
		fmt.Errorf("database connection is required")
		return nil
	}
	return &TaskRepositoryImpl{db: db}
}

const (
    defaultLimit = 10
    maxLimit    = 100
)

type cursor struct {
    ID uint `json:"id"`
    CreatedAt time.Time `json:"created_at"`
}


func (r *TaskRepositoryImpl) List(limit int, cursorStr string) ([]*models.Task, string, error) {
    if limit <= 0 {
        limit = defaultLimit
    }
    if limit > maxLimit {
        limit = maxLimit
    }

    query := r.db.Order("created_at DESC, id DESC")

    // If cursor is provided, decode and apply conditions
    if cursorStr != "" {
        c, err := decodeCursor(cursorStr)
        if err != nil {
            return nil, "", fmt.Errorf("invalid cursor: %w", err)
        }
        
        query = query.Where("(created_at, id) < (?, ?)", c.CreatedAt, c.ID)
    }

    var tasks []*models.Task
    if err := query.Limit(limit + 1).Find(&tasks).Error; err != nil {
        return nil, "", fmt.Errorf("failed to fetch tasks: %w", err)
    }

    var nextCursor string
    hasMore := len(tasks) > limit

    // If we have more items than limit, create next cursor
    if hasMore {
        lastTask := tasks[len(tasks)-1]
        nextCursor = encodeCursor(cursor{
            ID:        lastTask.ID,
            CreatedAt: lastTask.CreatedAt,
        })
        tasks = tasks[:limit] // Remove the extra item
    }

    return tasks, nextCursor, nil
}

// Helper function to encode cursor
func encodeCursor(c cursor) string {
    str := fmt.Sprintf("%d:%d", c.ID, c.CreatedAt.Unix())
    return base64.StdEncoding.EncodeToString([]byte(str))
}

// Helper function to decode cursor
func decodeCursor(str string) (cursor, error) {
    bytes, err := base64.StdEncoding.DecodeString(str)
    if err != nil {
        return cursor{}, fmt.Errorf("failed to decode cursor: %w", err)
    }

    parts := strings.Split(string(bytes), ":")
    if len(parts) != 2 {
        return cursor{}, fmt.Errorf("invalid cursor format")
    }

    id, err := strconv.ParseUint(parts[0], 10, 32)
    if err != nil {
        return cursor{}, fmt.Errorf("invalid cursor ID: %w", err)
    }

    timestamp, err := strconv.ParseInt(parts[1], 10, 64)
    if err != nil {
        return cursor{}, fmt.Errorf("invalid cursor timestamp: %w", err)
    }

    return cursor{
        ID:        uint(id),
        CreatedAt: time.Unix(timestamp, 0),
    }, nil
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
