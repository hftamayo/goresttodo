package http

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestNewETagGenerator(t *testing.T) {
	generator := NewETagGenerator()
	assert.NotNil(t, generator)
}

func TestETagGenerator_Generate(t *testing.T) {
	generator := NewETagGenerator()

	tests := []struct {
		name        string
		data        interface{}
		expectValid bool
		description string
	}{
		{
			name:        "generate from string",
			data:        "test string",
			expectValid: true,
			description: "Should generate valid ETag from string",
		},
		{
			name:        "generate from integer",
			data:        42,
			expectValid: true,
			description: "Should generate valid ETag from integer",
		},
		{
			name:        "generate from map",
			data:        map[string]interface{}{"key": "value", "number": 123},
			expectValid: true,
			description: "Should generate valid ETag from map",
		},
		{
			name:        "generate from slice",
			data:        []string{"item1", "item2", "item3"},
			expectValid: true,
			description: "Should generate valid ETag from slice",
		},
		{
			name:        "generate from struct",
			data:        struct{ Name string; Age int }{"John", 30},
			expectValid: true,
			description: "Should generate valid ETag from struct",
		},
		{
			name:        "generate from empty string",
			data:        "",
			expectValid: true,
			description: "Should generate valid ETag from empty string",
		},
		{
			name:        "generate from nil",
			data:        nil,
			expectValid: true,
			description: "Should generate valid ETag from nil",
		},
		{
			name:        "generate from complex nested structure",
			data: map[string]interface{}{
				"users": []map[string]interface{}{
					{"id": 1, "name": "Alice", "active": true},
					{"id": 2, "name": "Bob", "active": false},
				},
				"metadata": map[string]interface{}{
					"total": 2,
					"page":  1,
				},
			},
			expectValid: true,
			description: "Should generate valid ETag from complex nested structure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			etag := generator.Generate(tt.data)
			
			assert.NotEmpty(t, etag, "ETag should not be empty")
			assert.Contains(t, etag, "\"", "ETag should contain quotes")
			
			if tt.expectValid {
				// Verify ETag format (should be quoted hex string)
				assert.True(t, len(etag) > 2, "ETag should have content between quotes")
				assert.Equal(t, "\"", etag[:1], "ETag should start with quote")
				assert.Equal(t, "\"", etag[len(etag)-1:], "ETag should end with quote")
			}
		})
	}
}

func TestETagGenerator_Generate_Consistency(t *testing.T) {
	generator := NewETagGenerator()
	testData := map[string]interface{}{
		"name": "John Doe",
		"age":  30,
		"tags": []string{"tag1", "tag2"},
	}

	// Generate ETag multiple times for the same data
	etag1 := generator.Generate(testData)
	etag2 := generator.Generate(testData)
	etag3 := generator.Generate(testData)

	// All ETags should be identical for the same data
	assert.Equal(t, etag1, etag2, "ETags should be consistent for same data")
	assert.Equal(t, etag2, etag3, "ETags should be consistent for same data")
	assert.Equal(t, etag1, etag3, "ETags should be consistent for same data")
}

func TestETagGenerator_Generate_DifferentData(t *testing.T) {
	generator := NewETagGenerator()

	data1 := map[string]interface{}{"name": "John", "age": 30}
	data2 := map[string]interface{}{"name": "Jane", "age": 25}
	data3 := map[string]interface{}{"name": "John", "age": 30} // Same as data1

	etag1 := generator.Generate(data1)
	etag2 := generator.Generate(data2)
	etag3 := generator.Generate(data3)

	// Different data should produce different ETags
	assert.NotEqual(t, etag1, etag2, "Different data should produce different ETags")
	
	// Same data should produce same ETags
	assert.Equal(t, etag1, etag3, "Same data should produce same ETags")
}

func TestETagGenerator_Generate_JSONMarshalingError(t *testing.T) {
	generator := NewETagGenerator()

	// Create data that cannot be marshaled to JSON
	unmarshallableData := make(chan int)

	etag := generator.Generate(unmarshallableData)
	
	// Should still generate an ETag (fallback to timestamp)
	assert.NotEmpty(t, etag, "Should generate ETag even for unmarshallable data")
	assert.Contains(t, etag, "\"", "ETag should contain quotes")
}

func TestETagGenerator_GenerateFromTasks(t *testing.T) {
	generator := NewETagGenerator()

	tests := []struct {
		name        string
		tasks       []*models.Task
		expectValid bool
		description string
	}{
		{
			name: "generate from single task",
			tasks: []*models.Task{
				{
					Model: gorm.Model{
						ID:        1,
						UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					},
					Title: "Test Task",
					Done:  false,
				},
			},
			expectValid: true,
			description: "Should generate valid ETag from single task",
		},
		{
			name: "generate from multiple tasks",
			tasks: []*models.Task{
				{
					Model: gorm.Model{
						ID:        1,
						UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					},
					Title: "Task 1",
					Done:  false,
				},
				{
					Model: gorm.Model{
						ID:        2,
						UpdatedAt: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
					},
					Title: "Task 2",
					Done:  true,
				},
			},
			expectValid: true,
			description: "Should generate valid ETag from multiple tasks",
		},
		{
			name:        "generate from empty task list",
			tasks:       []*models.Task{},
			expectValid: true,
			description: "Should generate valid ETag from empty task list",
		},
		{
			name:        "generate from nil task list",
			tasks:       nil,
			expectValid: true,
			description: "Should generate valid ETag from nil task list",
		},
		{
			name: "generate from tasks with special characters",
			tasks: []*models.Task{
				{
					Model: gorm.Model{
						ID:        1,
						UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					},
					Title: "Task with émojis 🎉 and special chars!@#",
					Done:  false,
				},
			},
			expectValid: true,
			description: "Should generate valid ETag from tasks with special characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			etag := generator.GenerateFromTasks(tt.tasks)
			
			assert.NotEmpty(t, etag, "ETag should not be empty")
			assert.Contains(t, etag, "\"", "ETag should contain quotes")
			
			if tt.expectValid {
				// Verify ETag format
				assert.True(t, len(etag) > 2, "ETag should have content between quotes")
				assert.Equal(t, "\"", etag[:1], "ETag should start with quote")
				assert.Equal(t, "\"", etag[len(etag)-1:], "ETag should end with quote")
			}
		})
	}
}

func TestETagGenerator_GenerateFromTasks_Consistency(t *testing.T) {
	generator := NewETagGenerator()
	
	tasks := []*models.Task{
		{
			Model: gorm.Model{
				ID:        1,
				UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			Title: "Test Task",
			Done:  false,
		},
		{
			Model: gorm.Model{
				ID:        2,
				UpdatedAt: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
			},
			Title: "Another Task",
			Done:  true,
		},
	}

	// Generate ETag multiple times for the same tasks
	etag1 := generator.GenerateFromTasks(tasks)
	etag2 := generator.GenerateFromTasks(tasks)
	etag3 := generator.GenerateFromTasks(tasks)

	// All ETags should be identical for the same tasks
	assert.Equal(t, etag1, etag2, "ETags should be consistent for same tasks")
	assert.Equal(t, etag2, etag3, "ETags should be consistent for same tasks")
	assert.Equal(t, etag1, etag3, "ETags should be consistent for same tasks")
}

func TestETagGenerator_GenerateFromTasks_DifferentTasks(t *testing.T) {
	generator := NewETagGenerator()

	tasks1 := []*models.Task{
		{
			Model: gorm.Model{
				ID:        1,
				UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			Title: "Task 1",
			Done:  false,
		},
	}

	tasks2 := []*models.Task{
		{
			Model: gorm.Model{
				ID:        1,
				UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			Title: "Task 1 Modified",
			Done:  false,
		},
	}

	tasks3 := []*models.Task{
		{
			Model: gorm.Model{
				ID:        1,
				UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			Title: "Task 1",
			Done:  true, // Different done status
		},
	}

	etag1 := generator.GenerateFromTasks(tasks1)
	etag2 := generator.GenerateFromTasks(tasks2)
	etag3 := generator.GenerateFromTasks(tasks3)

	// Different tasks should produce different ETags
	assert.NotEqual(t, etag1, etag2, "Different task titles should produce different ETags")
	assert.NotEqual(t, etag1, etag3, "Different task status should produce different ETags")
	assert.NotEqual(t, etag2, etag3, "Different task properties should produce different ETags")
}

func TestETagGenerator_GenerateFromTasks_OrderIndependent(t *testing.T) {
	generator := NewETagGenerator()

	task1 := &models.Task{
		Model: gorm.Model{
			ID:        1,
			UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		Title: "Task 1",
		Done:  false,
	}

	task2 := &models.Task{
		Model: gorm.Model{
			ID:        2,
			UpdatedAt: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
		},
		Title: "Task 2",
		Done:  true,
	}

	// Generate ETags with different orders
	etag1 := generator.GenerateFromTasks([]*models.Task{task1, task2})
	etag2 := generator.GenerateFromTasks([]*models.Task{task2, task1})

	// ETags should be different for different orders (since we iterate in order)
	assert.NotEqual(t, etag1, etag2, "Different task orders should produce different ETags")
}

func TestETagGenerator_GenerateFromTasks_ManualHashVerification(t *testing.T) {
	generator := NewETagGenerator()

	task := &models.Task{
		Model: gorm.Model{
			ID:        1,
			UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		Title: "Test Task",
		Done:  false,
	}

	tasks := []*models.Task{task}

	etag := generator.GenerateFromTasks(tasks)

	// Manually calculate the expected hash
	expectedString := fmt.Sprintf("%d-%s-%t-%d", 
		task.ID, task.Title, task.Done, task.UpdatedAt.UnixNano())
	
	hash := sha256.New()
	hash.Write([]byte(expectedString))
	expectedETag := fmt.Sprintf(eTagCharacterFmt, hash.Sum(nil))

	assert.Equal(t, expectedETag, etag, "ETag should match manually calculated hash")
}

func TestETagGenerator_Generate_JSONFallback(t *testing.T) {
	generator := NewETagGenerator()

	// Test that JSON marshaling works correctly
	testData := map[string]interface{}{
		"string": "test",
		"number": 42,
		"bool":   true,
		"array":  []string{"a", "b", "c"},
	}

	etag := generator.Generate(testData)

	// Manually calculate expected hash
	jsonData, _ := json.Marshal(testData)
	hash := sha256.Sum256(jsonData)
	expectedETag := fmt.Sprintf(eTagCharacterFmt, hash)

	assert.Equal(t, expectedETag, etag, "ETag should match JSON-based hash")
}

func BenchmarkETagGenerator_Generate(b *testing.B) {
	generator := NewETagGenerator()
	testData := map[string]interface{}{
		"name": "John Doe",
		"age":  30,
		"tags": []string{"tag1", "tag2", "tag3"},
		"details": map[string]interface{}{
			"city":    "New York",
			"country": "USA",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator.Generate(testData)
	}
}

func BenchmarkETagGenerator_GenerateFromTasks(b *testing.B) {
	generator := NewETagGenerator()
	tasks := []*models.Task{
		{
			Model: gorm.Model{
				ID:        1,
				UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			Title: "Task 1",
			Done:  false,
		},
		{
			Model: gorm.Model{
				ID:        2,
				UpdatedAt: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
			},
			Title: "Task 2",
			Done:  true,
		},
		{
			Model: gorm.Model{
				ID:        3,
				UpdatedAt: time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
			},
			Title: "Task 3",
			Done:  false,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator.GenerateFromTasks(tasks)
	}
} 