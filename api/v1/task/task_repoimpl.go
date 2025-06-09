package task

import (
	"errors"
	"fmt"

	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/cursor"
	"github.com/hftamayo/gotodo/pkg/utils"
	"gorm.io/gorm"
)

type TaskRepositoryImpl struct {
	db *gorm.DB
}

func NewTaskRepositoryImpl(db *gorm.DB) TaskRepository {
    if db == nil {
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

func (r *TaskRepositoryImpl) List(limit int, cursorStr string, order string) ([]*models.Task, string, string, error) {
    if limit <= 0 {
        limit = DefaultLimit
    }
    if limit > MaxLimit {
        limit = MaxLimit
    }

    fmt.Printf("DEBUG: Starting List with limit=%d, cursor=%s\n", limit, cursorStr)

    // Base query with deleted_at IS NULL condition
    query := r.db.Model(&models.Task{}).
        Where("deleted_at IS NULL").
        Order("id DESC").  // Always order by id DESC to start from most recent
        Select("id, title, description, done, owner, created_at, updated_at")

    if cursorStr != "" {
        c, err := cursor.Decode[uint](cursorStr)
        if err != nil {
            return nil, "", "", fmt.Errorf("invalid cursor: %w", err)
        }
        fmt.Printf("DEBUG: Decoded cursor - ID: %d\n", c.ID)

        // Get all IDs less than the cursor ID
        var ids []uint
        if err := r.db.Model(&models.Task{}).
            Where("deleted_at IS NULL AND id < ?", c.ID).
            Pluck("id", &ids).Error; err != nil {
            return nil, "", "", fmt.Errorf("failed to get IDs: %w", err)
        }

        // If we have no IDs, return empty result
        if len(ids) == 0 {
            return []*models.Task{}, "", "", nil
        }

        // Use the IDs in the query
        query = query.Where("id IN ?", ids)
    }

    // Log the SQL query
    sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
        return tx.Session(&gorm.Session{DryRun: true})
    })
    fmt.Printf("DEBUG: SQL Query: %s\n", sql)

    var tasks []*models.Task
    if err := query.Limit(limit + 1).Find(&tasks).Error; err != nil {
        return nil, "", "", fmt.Errorf("failed to fetch tasks: %w", err)
    }

    fmt.Printf("DEBUG: Found %d tasks\n", len(tasks))
    for _, task := range tasks {
        fmt.Printf("DEBUG: Task ID: %d\n", task.ID)
    }

    var nextCursor, prevCursor string
    hasMore := len(tasks) > limit

    if hasMore {
        // Get the last item before removing it
        lastTask := tasks[len(tasks)-1]
        tasks = tasks[:limit] // Remove the extra item
        fmt.Printf("DEBUG: Has more pages, last task ID: %d\n", lastTask.ID)

        // Create cursor for the next page
        c := cursor.Cursor[uint]{
            ID: lastTask.ID,
        }

        var err error
        nextCursor, err = cursor.Encode(c, cursor.Options{
            Field:     "id",
            Direction: "DESC",
        })
        if err != nil {
            return nil, "", "", fmt.Errorf("failed to encode next cursor: %w", err)
        }
        fmt.Printf("DEBUG: Next cursor: %s\n", nextCursor)
    }

    // Generate previous cursor from first item if we have tasks
    if len(tasks) > 0 {
        firstTask := tasks[0]
        fmt.Printf("DEBUG: First task ID: %d\n", firstTask.ID)
        c := cursor.Cursor[uint]{
            ID: firstTask.ID,
        }

        var err error
        prevCursor, err = cursor.Encode(c, cursor.Options{
            Field:     "id",
            Direction: "DESC",
        })
        if err != nil {
            return nil, "", "", fmt.Errorf("failed to encode previous cursor: %w", err)
        }
        fmt.Printf("DEBUG: Previous cursor: %s\n", prevCursor)
    }

    return tasks, nextCursor, prevCursor, nil
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

func (r *TaskRepositoryImpl) SearchByTitle(title string) (*models.Task, error) {
    var task models.Task
    
    result := r.db.Where("title = ? AND deleted_at IS NULL", title).First(&task)
    if result.Error != nil {
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
            return nil, nil // No task found with this title
        }
        return nil, fmt.Errorf("error searching task by title: %w", result.Error)
    }
    
    return &task, nil
}

func (r *TaskRepositoryImpl) Create(task *models.Task) (*models.Task, error) {
    if task == nil {
        return nil, errors.New("task cannot be nil")
    }
	
	if result := r.db.Create(task); result.Error != nil {
		return nil, result.Error
	}
	return task, nil
}

func (r *TaskRepositoryImpl) Update(id int, task *models.Task) (*models.Task, error) {
    if task == nil {
        return nil, errors.New("task cannot be nil")
    }
	
     var existingTask models.Task
    if err := r.db.First(&existingTask, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, fmt.Errorf(errTaskNotFoundFmt, id)
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
    result := r.db.Model(&models.Task{}).Where("id = ?", id).Updates(map[string]interface{}{
        "done":       true,
    })
    
    if result.Error != nil {
        return nil, fmt.Errorf("failed to mark task as done: %w", result.Error)
    }
    if result.RowsAffected == 0 {
        return nil, fmt.Errorf(errTaskNotFoundFmt, id)
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
        return fmt.Errorf(errTaskNotFoundFmt, id)
    }

    return nil
}

func (r *TaskRepositoryImpl) ListByPage(page int, limit int, order string) ([]*models.Task, int64, error) {
    if page <= 0 {
        page = 1
    }
    if limit <= 0 {
        limit = utils.DefaultLimit
    }
    if limit > utils.MaxLimit {
        limit = utils.MaxLimit
    }

    // Default order if not provided
    if order == "" {
        order = utils.DefaultOrder
    }

    // Calculate offset
    offset := (page - 1) * limit

    // Get total count
    var totalCount int64
    if err := r.db.Model(&models.Task{}).Count(&totalCount).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to get total count: %w", err)
    }

    // Get paginated data
    var tasks []*models.Task
    query := r.db.Model(&models.Task{}).
        Order(fmt.Sprintf("created_at %s, id %s", order, order)).
        Select("id, title, description, done, owner, created_at, updated_at").
        Offset(offset).
        Limit(limit)

    if err := query.Find(&tasks).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to fetch tasks: %w", err)
    }

    return tasks, totalCount, nil
}
