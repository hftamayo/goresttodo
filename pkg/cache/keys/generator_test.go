package keys

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGenerator(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		description string
	}{
		{
			name:   "empty prefix",
			prefix: "",
			description: "Should create generator with empty prefix",
		},
		{
			name:   "simple prefix",
			prefix: "app",
			description: "Should create generator with simple prefix",
		},
		{
			name:   "complex prefix",
			prefix: "my-app-v1",
			description: "Should create generator with complex prefix",
		},
		{
			name:   "prefix with special characters",
			prefix: "app_name@123",
			description: "Should create generator with special characters in prefix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewGenerator(tt.prefix)
			assert.NotNil(t, generator)
			assert.Equal(t, tt.prefix, generator.Prefix, "Test: %s", tt.description)
		})
	}
}

func TestGenerator_Build(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		parts    []string
		expected string
		description string
	}{
		{
			name:     "no prefix, no parts",
			prefix:   "",
			parts:    []string{},
			expected: "",
			description: "Should return empty string for no prefix and no parts",
		},
		{
			name:     "no prefix, single part",
			prefix:   "",
			parts:    []string{"user"},
			expected: "user",
			description: "Should return single part for no prefix",
		},
		{
			name:     "no prefix, multiple parts",
			prefix:   "",
			parts:    []string{"user", "123", "profile"},
			expected: "user_123_profile",
			description: "Should join multiple parts with underscore for no prefix",
		},
		{
			name:     "with prefix, no parts",
			prefix:   "app",
			parts:    []string{},
			expected: "app_",
			description: "Should return prefix with trailing underscore for no parts",
		},
		{
			name:     "with prefix, single part",
			prefix:   "app",
			parts:    []string{"user"},
			expected: "app_user",
			description: "Should return prefix_part for single part",
		},
		{
			name:     "with prefix, multiple parts",
			prefix:   "app",
			parts:    []string{"user", "123", "profile"},
			expected: "app_user_123_profile",
			description: "Should return prefix_part1_part2_part3 for multiple parts",
		},
		{
			name:     "complex prefix, multiple parts",
			prefix:   "my-app-v1",
			parts:    []string{"user", "123", "profile"},
			expected: "my-app-v1_user_123_profile",
			description: "Should handle complex prefix with hyphens",
		},
		{
			name:     "prefix with underscore, multiple parts",
			prefix:   "app_name",
			parts:    []string{"user", "123", "profile"},
			expected: "app_name_user_123_profile",
			description: "Should handle prefix with existing underscores",
		},
		{
			name:     "empty parts with spaces",
			prefix:   "app",
			parts:    []string{"", "user", ""},
			expected: "app__user_",
			description: "Should handle empty parts correctly",
		},
		{
			name:     "parts with special characters",
			prefix:   "app",
			parts:    []string{"user@123", "profile!test"},
			expected: "app_user@123_profile!test",
			description: "Should handle special characters in parts",
		},
		{
			name:     "parts with spaces",
			prefix:   "app",
			parts:    []string{"user name", "profile data"},
			expected: "app_user name_profile data",
			description: "Should preserve spaces in parts",
		},
		{
			name:     "parts with numbers",
			prefix:   "app",
			parts:    []string{"user", "123", "456"},
			expected: "app_user_123_456",
			description: "Should handle numeric parts correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewGenerator(tt.prefix)
			result := generator.Build(tt.parts...)
			assert.Equal(t, tt.expected, result, "Test: %s", tt.description)
		})
	}
}

func TestGenerator_ForList(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		params   map[string]string
		expected string
		description string
	}{
		{
			name:     "no prefix, empty params",
			prefix:   "",
			params:   map[string]string{},
			expected: "list",
			description: "Should return 'list' for no prefix and empty params",
		},
		{
			name:     "no prefix, single param",
			prefix:   "",
			params:   map[string]string{"user_id": "123"},
			expected: "list_user_id_123",
			description: "Should return 'list_key_value' for single param",
		},
		{
			name:     "no prefix, multiple params",
			prefix:   "",
			params:   map[string]string{"user_id": "123", "status": "active"},
			expected: "list_status_active_user_id_123",
			description: "Should return sorted params for multiple params",
		},
		{
			name:     "with prefix, empty params",
			prefix:   "app",
			params:   map[string]string{},
			expected: "app_list",
			description: "Should return 'prefix_list' for empty params",
		},
		{
			name:     "with prefix, single param",
			prefix:   "app",
			params:   map[string]string{"user_id": "123"},
			expected: "app_list_user_id_123",
			description: "Should return 'prefix_list_key_value' for single param",
		},
		{
			name:     "with prefix, multiple params",
			prefix:   "app",
			params:   map[string]string{"user_id": "123", "status": "active"},
			expected: "app_list_status_active_user_id_123",
			description: "Should return sorted params with prefix for multiple params",
		},
		{
			name:     "complex prefix, multiple params",
			prefix:   "my-app-v1",
			params:   map[string]string{"user_id": "123", "status": "active"},
			expected: "my-app-v1_list_status_active_user_id_123",
			description: "Should handle complex prefix with multiple params",
		},
		{
			name:     "params with special characters",
			prefix:   "app",
			params:   map[string]string{"user@id": "123", "status!": "active"},
			expected: "app_list_status!_active_user@id_123",
			description: "Should handle special characters in param keys and values",
		},
		{
			name:     "params with spaces",
			prefix:   "app",
			params:   map[string]string{"user id": "123", "user status": "active"},
			expected: "app_list_user id_123_user status_active",
			description: "Should preserve spaces in param keys and values",
		},
		{
			name:     "params with numbers",
			prefix:   "app",
			params:   map[string]string{"id": "123", "limit": "50", "offset": "0"},
			expected: "app_list_id_123_limit_50_offset_0",
			description: "Should handle numeric param values correctly",
		},
		{
			name:     "empty param values",
			prefix:   "app",
			params:   map[string]string{"user_id": "", "status": "active"},
			expected: "app_list_status_active_user_id_",
			description: "Should handle empty param values correctly",
		},
		{
			name:     "many params for sorting test",
			prefix:   "app",
			params: map[string]string{
				"z_key": "z_value",
				"a_key": "a_value",
				"m_key": "m_value",
				"1_key": "1_value",
			},
			expected: "app_list_1_key_1_value_a_key_a_value_m_key_m_value_z_key_z_value",
			description: "Should sort params alphabetically by key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewGenerator(tt.prefix)
			result := generator.ForList(tt.params)
			assert.Equal(t, tt.expected, result, "Test: %s", tt.description)
		})
	}
}

func TestGenerator_ForList_SortingConsistency(t *testing.T) {
	// Test that parameter sorting is consistent regardless of map insertion order
	generator := NewGenerator("app")
	
	params1 := map[string]string{
		"user_id": "123",
		"status":  "active",
		"limit":   "50",
	}
	
	params2 := map[string]string{
		"status":  "active",
		"limit":   "50",
		"user_id": "123",
	}
	
	params3 := map[string]string{
		"limit":   "50",
		"user_id": "123",
		"status":  "active",
	}
	
	result1 := generator.ForList(params1)
	result2 := generator.ForList(params2)
	result3 := generator.ForList(params3)
	
	expected := "app_list_limit_50_status_active_user_id_123"
	
	assert.Equal(t, expected, result1, "Should produce consistent result regardless of insertion order 1")
	assert.Equal(t, expected, result2, "Should produce consistent result regardless of insertion order 2")
	assert.Equal(t, expected, result3, "Should produce consistent result regardless of insertion order 3")
	assert.Equal(t, result1, result2, "Results should be identical for same params in different order")
	assert.Equal(t, result2, result3, "Results should be identical for same params in different order")
}

func TestGenerator_Integration(t *testing.T) {
	// Test complete workflow with both methods
	generator := NewGenerator("my-app")
	
	// Test Build method
	userKey := generator.Build("user", "123", "profile")
	assert.Equal(t, "my-app_user_123_profile", userKey)
	
	// Test ForList method
	listParams := map[string]string{
		"user_id": "123",
		"status":  "active",
	}
	listKey := generator.ForList(listParams)
	assert.Equal(t, "my-app_list_status_active_user_id_123", listKey)
	
	// Test that keys are different
	assert.NotEqual(t, userKey, listKey, "Different operations should produce different keys")
}

func TestGenerator_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		parts    []string
		params   map[string]string
		description string
	}{
		{
			name:     "very long prefix",
			prefix:   "very-long-prefix-that-exceeds-normal-length-and-might-cause-issues",
			parts:    []string{"test"},
			params:   map[string]string{"test": "value"},
			description: "Should handle very long prefix",
		},
		{
			name:     "very long parts",
			prefix:   "app",
			parts:    []string{"very-long-part-that-exceeds-normal-length-and-might-cause-issues"},
			params:   map[string]string{"test": "value"},
			description: "Should handle very long parts",
		},
		{
			name:     "very long param values",
			prefix:   "app",
			parts:    []string{"test"},
			params:   map[string]string{"key": "very-long-value-that-exceeds-normal-length-and-might-cause-issues"},
			description: "Should handle very long param values",
		},
		{
			name:     "unicode characters",
			prefix:   "app",
			parts:    []string{"用户", "123"},
			params:   map[string]string{"用户": "123", "status": "活跃"},
			description: "Should handle unicode characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewGenerator(tt.prefix)
			
			// Test Build method
			buildResult := generator.Build(tt.parts...)
			assert.NotEmpty(t, buildResult, "Build should not return empty for: %s", tt.description)
			
			// Test ForList method
			forListResult := generator.ForList(tt.params)
			assert.NotEmpty(t, forListResult, "ForList should not return empty for: %s", tt.description)
		})
	}
}

func BenchmarkGenerator_Build(b *testing.B) {
	generator := NewGenerator("app")
	parts := []string{"user", "123", "profile", "data"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator.Build(parts...)
	}
}

func BenchmarkGenerator_ForList(b *testing.B) {
	generator := NewGenerator("app")
	params := map[string]string{
		"user_id": "123",
		"status":  "active",
		"limit":   "50",
		"offset":  "0",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator.ForList(params)
	}
}

func BenchmarkGenerator_ForList_LargeParams(b *testing.B) {
	generator := NewGenerator("app")
	params := map[string]string{
		"user_id": "123",
		"status":  "active",
		"limit":   "50",
		"offset":  "0",
		"sort":    "created_at",
		"order":   "desc",
		"filter":  "completed",
		"search":  "test",
		"category": "work",
		"priority": "high",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator.ForList(params)
	}
} 