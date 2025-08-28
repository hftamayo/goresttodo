package task

import (
	"time"

	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/config"
)

// CacheInterface defines the contract for cache operations
type CacheInterface interface {
	Get(key string, dest interface{}) error
	Set(key string, value interface{}, ttl time.Duration) error
	SetWithTags(key string, value interface{}, ttl time.Duration, tags ...string) error
	Delete(key string) error
	InvalidateByTags(tags ...string) error
}

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
	// Cache configuration
	EnableCache     bool
	CacheTTL        time.Duration // in time.Duration
	CacheKeys       CacheKeyConfig
	
	// Logging configuration
	EnableLogging   bool
	AsyncLogging    bool
	ErrorLogger     config.ErrorLogger
	
	// Validation configuration
	ValidationConfig ValidationConfig
}

// CacheKeyConfig holds all cache key patterns
type CacheKeyConfig struct {
	TaskKey          string // "task_%d"
	TaskPageKey      string // "tasks_page_%d_%d_%s"
	TaskCursorKey    string // "tasks_cursor_%s_limit_%d_order_%s"
	TaskListRef      string // "tasks:list"
	TaskReference    string // "task:%d"
	TaskPageCache    string // "task_page_*"
}

// ValidationConfig holds validation-related configuration
type ValidationConfig struct {
	ErrTaskNotFoundFmt string // "task with id %d not found"
	ErrInvalidateCacheFmt string // "Failed to invalidate cache: %v\n"
}

// DefaultTaskServiceConfig returns default service configuration
func DefaultTaskServiceConfig() *TaskServiceConfig {
	return &TaskServiceConfig{
		EnableCache:   true,
		CacheTTL:      5 * time.Minute, // 5 minutes
		CacheKeys: CacheKeyConfig{
			TaskKey:       "task_%d",
			TaskPageKey:   "tasks_page_%d_%d_%s",
			TaskCursorKey: "tasks_cursor_%s_limit_%d_order_%s",
			TaskListRef:   "tasks:list",
			TaskReference: "task:%d",
			TaskPageCache: "task_page_*",
		},
		EnableLogging: true,
		AsyncLogging:  true,
		ValidationConfig: ValidationConfig{
			ErrTaskNotFoundFmt:     "task with id %d not found",
			ErrInvalidateCacheFmt:  "Failed to invalidate cache: %v\n",
		},
	}
}