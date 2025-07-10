package task

import (
	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/config"
)

// TaskServiceInterface defines the contract for task operations
type TaskServiceInterface interface {
	// Core CRUD operations
	List(cursor string, limit int, order string) ([]*models.Task, string, string, int64, error)
	ListById(id int) (*models.Task, error)
	Create(task *models.Task) (*models.Task, error)
	Update(id int, task *models.Task) (*models.Task, error)
	Delete(id int) error
	MarkAsDone(id int) (*models.Task, error)
	ListByPage(page int, limit int, order string) ([]*models.Task, int64, error)
	
	// Cache operations (moved from handler)
	InvalidateTaskCache(id int) error
	InvalidateListCache() error
	InvalidatePageCache() error
}

// TaskServiceConfig holds configuration for the task service
type TaskServiceConfig struct {
	EnableCache     bool
	EnableLogging   bool
	AsyncLogging    bool
	CacheTTL        int // in seconds
	ErrorLogger     config.ErrorLogger
}

// DefaultTaskServiceConfig returns default service configuration
func DefaultTaskServiceConfig() *TaskServiceConfig {
	return &TaskServiceConfig{
		EnableCache:   true,
		EnableLogging: true,
		AsyncLogging:  true,
		CacheTTL:      300, // 5 minutes
	}
}