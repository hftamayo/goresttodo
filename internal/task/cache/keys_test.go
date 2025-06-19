package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTaskKeyGenerator(t *testing.T) {
	generator := NewTaskKeyGenerator()
	assert.NotNil(t, generator)
	assert.NotNil(t, generator.generator)
	assert.Equal(t, "tasks", generator.generator.Prefix)
}

func TestTaskKeyGenerator_ForTask(t *testing.T) {
	generator := NewTaskKeyGenerator()

	tests := []struct {
		name     string
		id       int
		expected string
		description string
	}{
		{
			name:     "positive ID",
			id:       123,
			expected: "tasks_id_123",
			description: "Should generate key for positive task ID",
		},
		{
			name:     "zero ID",
			id:       0,
			expected: "tasks_id_0",
			description: "Should generate key for zero task ID",
		},
		{
			name:     "negative ID",
			id:       -1,
			expected: "tasks_id_-1",
			description: "Should generate key for negative task ID",
		},
		{
			name:     "large ID",
			id:       999999999,
			expected: "tasks_id_999999999",
			description: "Should generate key for large task ID",
		},
		{
			name:     "single digit ID",
			id:       5,
			expected: "tasks_id_5",
			description: "Should generate key for single digit task ID",
		},
		{
			name:     "double digit ID",
			id:       42,
			expected: "tasks_id_42",
			description: "Should generate key for double digit task ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.ForTask(tt.id)
			assert.Equal(t, tt.expected, result, "Test: %s", tt.description)
		})
	}
}

func TestTaskKeyGenerator_ForCursorList(t *testing.T) {
	generator := NewTaskKeyGenerator()

	tests := []struct {
		name     string
		cursor   string
		limit    int
		order    string
		expected string
		description string
	}{
		{
			name:     "valid cursor with asc order",
			cursor:   "cursor123",
			limit:    10,
			order:    "asc",
			expected: "tasks_cursor_cursor123_limit_10_order_asc",
			description: "Should generate key for cursor list with asc order",
		},
		{
			name:     "valid cursor with desc order",
			cursor:   "cursor456",
			limit:    20,
			order:    "desc",
			expected: "tasks_cursor_cursor456_limit_20_order_desc",
			description: "Should generate key for cursor list with desc order",
		},
		{
			name:     "empty cursor",
			cursor:   "",
			limit:    15,
			order:    "asc",
			expected: "tasks_cursor__limit_15_order_asc",
			description: "Should generate key for empty cursor",
		},
		{
			name:     "zero limit",
			cursor:   "cursor789",
			limit:    0,
			order:    "desc",
			expected: "tasks_cursor_cursor789_limit_0_order_desc",
			description: "Should generate key for zero limit",
		},
		{
			name:     "negative limit",
			cursor:   "cursor101",
			limit:    -5,
			order:    "asc",
			expected: "tasks_cursor_cursor101_limit_-5_order_asc",
			description: "Should generate key for negative limit",
		},
		{
			name:     "large limit",
			cursor:   "cursor202",
			limit:    1000,
			order:    "desc",
			expected: "tasks_cursor_cursor202_limit_1000_order_desc",
			description: "Should generate key for large limit",
		},
		{
			name:     "cursor with special characters",
			cursor:   "cursor@123!",
			limit:    25,
			order:    "asc",
			expected: "tasks_cursor_cursor@123!_limit_25_order_asc",
			description: "Should generate key for cursor with special characters",
		},
		{
			name:     "cursor with spaces",
			cursor:   "cursor with spaces",
			limit:    30,
			order:    "desc",
			expected: "tasks_cursor_cursor with spaces_limit_30_order_desc",
			description: "Should generate key for cursor with spaces",
		},
		{
			name:     "order with special characters",
			cursor:   "cursor303",
			limit:    35,
			order:    "order@123!",
			expected: "tasks_cursor_cursor303_limit_35_order_order@123!",
			description: "Should generate key for order with special characters",
		},
		{
			name:     "all parameters with edge values",
			cursor:   "",
			limit:    0,
			order:    "",
			expected: "tasks_cursor__limit_0_order_",
			description: "Should generate key for all empty/zero parameters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.ForCursorList(tt.cursor, tt.limit, tt.order)
			assert.Equal(t, tt.expected, result, "Test: %s", tt.description)
		})
	}
}

func TestTaskKeyGenerator_ForPageList(t *testing.T) {
	generator := NewTaskKeyGenerator()

	tests := []struct {
		name     string
		page     int
		limit    int
		order    string
		expected string
		description string
	}{
		{
			name:     "first page with asc order",
			page:     1,
			limit:    10,
			order:    "asc",
			expected: "tasks_page_1_limit_10_order_asc",
			description: "Should generate key for first page with asc order",
		},
		{
			name:     "second page with desc order",
			page:     2,
			limit:    20,
			order:    "desc",
			expected: "tasks_page_2_limit_20_order_desc",
			description: "Should generate key for second page with desc order",
		},
		{
			name:     "zero page",
			page:     0,
			limit:    15,
			order:    "asc",
			expected: "tasks_page_0_limit_15_order_asc",
			description: "Should generate key for zero page",
		},
		{
			name:     "negative page",
			page:     -1,
			limit:    25,
			order:    "desc",
			expected: "tasks_page_-1_limit_25_order_desc",
			description: "Should generate key for negative page",
		},
		{
			name:     "large page number",
			page:     999999,
			limit:    30,
			order:    "asc",
			expected: "tasks_page_999999_limit_30_order_asc",
			description: "Should generate key for large page number",
		},
		{
			name:     "zero limit",
			page:     5,
			limit:    0,
			order:    "desc",
			expected: "tasks_page_5_limit_0_order_desc",
			description: "Should generate key for zero limit",
		},
		{
			name:     "negative limit",
			page:     10,
			limit:    -5,
			order:    "asc",
			expected: "tasks_page_10_limit_-5_order_asc",
			description: "Should generate key for negative limit",
		},
		{
			name:     "large limit",
			page:     15,
			limit:    1000,
			order:    "desc",
			expected: "tasks_page_15_limit_1000_order_desc",
			description: "Should generate key for large limit",
		},
		{
			name:     "order with special characters",
			page:     20,
			limit:    35,
			order:    "order@123!",
			expected: "tasks_page_20_limit_35_order_order@123!",
			description: "Should generate key for order with special characters",
		},
		{
			name:     "order with spaces",
			page:     25,
			limit:    40,
			order:    "order with spaces",
			expected: "tasks_page_25_limit_40_order_order with spaces",
			description: "Should generate key for order with spaces",
		},
		{
			name:     "all parameters with edge values",
			page:     0,
			limit:    0,
			order:    "",
			expected: "tasks_page_0_limit_0_order_",
			description: "Should generate key for all zero/empty parameters",
		},
		{
			name:     "single digit values",
			page:     1,
			limit:    5,
			order:    "asc",
			expected: "tasks_page_1_limit_5_order_asc",
			description: "Should generate key for single digit values",
		},
		{
			name:     "double digit values",
			page:     42,
			limit:    50,
			order:    "desc",
			expected: "tasks_page_42_limit_50_order_desc",
			description: "Should generate key for double digit values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.ForPageList(tt.page, tt.limit, tt.order)
			assert.Equal(t, tt.expected, result, "Test: %s", tt.description)
		})
	}
}

func TestTaskKeyGenerator_Integration(t *testing.T) {
	// Test complete workflow with all methods
	generator := NewTaskKeyGenerator()
	
	// Test ForTask method
	taskKey := generator.ForTask(123)
	assert.Equal(t, "tasks_id_123", taskKey)
	
	// Test ForCursorList method
	cursorKey := generator.ForCursorList("cursor123", 10, "asc")
	assert.Equal(t, "tasks_cursor_cursor123_limit_10_order_asc", cursorKey)
	
	// Test ForPageList method
	pageKey := generator.ForPageList(1, 10, "desc")
	assert.Equal(t, "tasks_page_1_limit_10_order_desc", pageKey)
	
	// Test that all keys are different
	assert.NotEqual(t, taskKey, cursorKey, "Task key should be different from cursor key")
	assert.NotEqual(t, cursorKey, pageKey, "Cursor key should be different from page key")
	assert.NotEqual(t, taskKey, pageKey, "Task key should be different from page key")
	
	// Test that all keys have the correct prefix
	assert.Contains(t, taskKey, "tasks_", "Task key should contain tasks prefix")
	assert.Contains(t, cursorKey, "tasks_", "Cursor key should contain tasks prefix")
	assert.Contains(t, pageKey, "tasks_", "Page key should contain tasks prefix")
}

func TestTaskKeyGenerator_Consistency(t *testing.T) {
	// Test that the same inputs always produce the same outputs
	generator := NewTaskKeyGenerator()
	
	// Test ForTask consistency
	taskKey1 := generator.ForTask(123)
	taskKey2 := generator.ForTask(123)
	assert.Equal(t, taskKey1, taskKey2, "ForTask should produce consistent results")
	
	// Test ForCursorList consistency
	cursorKey1 := generator.ForCursorList("cursor123", 10, "asc")
	cursorKey2 := generator.ForCursorList("cursor123", 10, "asc")
	assert.Equal(t, cursorKey1, cursorKey2, "ForCursorList should produce consistent results")
	
	// Test ForPageList consistency
	pageKey1 := generator.ForPageList(1, 10, "desc")
	pageKey2 := generator.ForPageList(1, 10, "desc")
	assert.Equal(t, pageKey1, pageKey2, "ForPageList should produce consistent results")
}

func TestTaskKeyGenerator_EdgeCases(t *testing.T) {
	generator := NewTaskKeyGenerator()

	tests := []struct {
		name     string
		description string
		testFunc func() string
	}{
		{
			name:     "very large task ID",
			description: "Should handle very large task ID",
			testFunc: func() string {
				return generator.ForTask(999999999999)
			},
		},
		{
			name:     "very large page number",
			description: "Should handle very large page number",
			testFunc: func() string {
				return generator.ForPageList(999999999999, 10, "asc")
			},
		},
		{
			name:     "very large limit",
			description: "Should handle very large limit",
			testFunc: func() string {
				return generator.ForPageList(1, 999999999999, "desc")
			},
		},
		{
			name:     "very long cursor",
			description: "Should handle very long cursor string",
			testFunc: func() string {
				longCursor := "very-long-cursor-string-that-exceeds-normal-length-and-might-cause-issues"
				return generator.ForCursorList(longCursor, 10, "asc")
			},
		},
		{
			name:     "very long order",
			description: "Should handle very long order string",
			testFunc: func() string {
				longOrder := "very-long-order-string-that-exceeds-normal-length-and-might-cause-issues"
				return generator.ForCursorList("cursor123", 10, longOrder)
			},
		},
		{
			name:     "unicode characters in cursor",
			description: "Should handle unicode characters in cursor",
			testFunc: func() string {
				return generator.ForCursorList("cursor用户123", 10, "asc")
			},
		},
		{
			name:     "unicode characters in order",
			description: "Should handle unicode characters in order",
			testFunc: func() string {
				return generator.ForCursorList("cursor123", 10, "order用户")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.testFunc()
			assert.NotEmpty(t, result, "Test: %s", tt.description)
			assert.Contains(t, result, "tasks_", "Test: %s", tt.description)
		})
	}
}

func TestTaskKeyGenerator_Performance(t *testing.T) {
	generator := NewTaskKeyGenerator()
	
	// Test that key generation doesn't take too long
	taskKey := generator.ForTask(123)
	cursorKey := generator.ForCursorList("cursor123", 10, "asc")
	pageKey := generator.ForPageList(1, 10, "desc")
	
	assert.NotEmpty(t, taskKey)
	assert.NotEmpty(t, cursorKey)
	assert.NotEmpty(t, pageKey)
}

func BenchmarkTaskKeyGenerator_ForTask(b *testing.B) {
	generator := NewTaskKeyGenerator()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator.ForTask(123)
	}
}

func BenchmarkTaskKeyGenerator_ForCursorList(b *testing.B) {
	generator := NewTaskKeyGenerator()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator.ForCursorList("cursor123", 10, "asc")
	}
}

func BenchmarkTaskKeyGenerator_ForPageList(b *testing.B) {
	generator := NewTaskKeyGenerator()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator.ForPageList(1, 10, "desc")
	}
}

func BenchmarkTaskKeyGenerator_AllMethods(b *testing.B) {
	generator := NewTaskKeyGenerator()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator.ForTask(123)
		generator.ForCursorList("cursor123", 10, "asc")
		generator.ForPageList(1, 10, "desc")
	}
} 