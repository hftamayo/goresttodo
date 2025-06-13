package cache

import (
	"fmt"

	"github.com/hftamayo/gotodo/pkg/cache/keys"
)

// TaskKeyGenerator provides task-specific cache key generation
type TaskKeyGenerator struct {
    generator *keys.Generator
}

// NewTaskKeyGenerator creates a new task key generator
func NewTaskKeyGenerator() *TaskKeyGenerator {
    return &TaskKeyGenerator{
        generator: keys.NewGenerator("tasks"),
    }
}

// ForTask generates a key for a single task
func (t *TaskKeyGenerator) ForTask(id int) string {
    return t.generator.Build("id", fmt.Sprintf("%d", id))
}

// ForCursorList generates a key for cursor-based pagination
func (t *TaskKeyGenerator) ForCursorList(cursor string, limit int, order string) string {
    return t.generator.Build("cursor", cursor, "limit", fmt.Sprintf("%d", limit), "order", order)
}

// ForPageList generates a key for page-based pagination
func (t *TaskKeyGenerator) ForPageList(page int, limit int, order string) string {
    return t.generator.Build("page", fmt.Sprintf("%d", page), "limit", fmt.Sprintf("%d", limit), "order", order)
}