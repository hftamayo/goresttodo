package http

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hftamayo/gotodo/api/v1/models"
)

// ETagGenerator provides utilities for generating ETags
type ETagGenerator struct{}

// NewETagGenerator creates a new ETag generator
func NewETagGenerator() *ETagGenerator {
    return &ETagGenerator{}
}

// Generate creates an ETag from any data by converting to JSON and hashing
func (e *ETagGenerator) Generate(data interface{}) string {
    jsonData, err := json.Marshal(data)
    if err != nil {
        // Fallback to a timestamp-based tag if marshaling fails
        return fmt.Sprintf("\"%x\"", sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano()))))
    }
    
    hash := sha256.Sum256(jsonData)
    return fmt.Sprintf("\"%x\"", hash)
}

// GenerateFromTasks creates an ETag specifically for task collections
func (e *ETagGenerator) GenerateFromTasks(tasks []*models.Task) string {
    hash := sha256.New()
    for _, task := range tasks {
        hash.Write([]byte(fmt.Sprintf("%d-%s-%t-%d", 
            task.ID, task.Title, task.Done, task.UpdatedAt.UnixNano())))
    }
    return fmt.Sprintf("\"%x\"", hash.Sum(nil))
}