package core

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewInMemoryClient(t *testing.T) {
	client := NewInMemoryClient()
	assert.NotNil(t, client)
	assert.NotNil(t, client.data)
	assert.Equal(t, 0, len(client.data), "New client should have empty data map")
}

func TestInMemoryClient_Set(t *testing.T) {
	client := NewInMemoryClient()

	tests := []struct {
		name       string
		key        string
		value      interface{}
		expiration time.Duration
		expectErr  bool
		description string
	}{
		{
			name:       "set string value",
			key:        "test_key",
			value:      "test_value",
			expiration: 0,
			expectErr:  false,
			description: "Should successfully set a string value",
		},
		{
			name:       "set integer value",
			key:        "int_key",
			value:      42,
			expiration: 0,
			expectErr:  false,
			description: "Should successfully set an integer value",
		},
		{
			name:       "set struct value",
			key:        "struct_key",
			value:      map[string]interface{}{"name": "test", "age": 25},
			expiration: 0,
			expectErr:  false,
			description: "Should successfully set a struct value",
		},
		{
			name:       "set with expiration",
			key:        "expiring_key",
			value:      "expiring_value",
			expiration: time.Second * 5,
			expectErr:  false,
			description: "Should successfully set a value with expiration",
		},
		{
			name:       "set empty key",
			key:        "",
			value:      "value",
			expiration: 0,
			expectErr:  false,
			description: "Should allow empty key",
		},
		{
			name:       "set nil value",
			key:        "nil_key",
			value:      nil,
			expiration: 0,
			expectErr:  false,
			description: "Should allow nil value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.Set(tt.key, tt.value, tt.expiration)
			
			if tt.expectErr {
				assert.Error(t, err, "Test: %s", tt.description)
			} else {
				assert.NoError(t, err, "Test: %s", tt.description)
				
				// Verify the value was stored
				assert.True(t, client.Exists(tt.key), "Key should exist after setting")
			}
		})
	}
}

func TestInMemoryClient_Get(t *testing.T) {
	client := NewInMemoryClient()

	tests := []struct {
		name        string
		setup       func()
		key         string
		dest        interface{}
		expectErr   bool
		expectedVal interface{}
		description string
	}{
		{
			name: "get existing string value",
			setup: func() {
				client.Set("string_key", "test_value", 0)
			},
			key:         "string_key",
			dest:        new(string),
			expectErr:   false,
			expectedVal: "test_value",
			description: "Should retrieve existing string value",
		},
		{
			name: "get existing integer value",
			setup: func() {
				client.Set("int_key", 42, 0)
			},
			key:         "int_key",
			dest:        new(int),
			expectErr:   false,
			expectedVal: 42,
			description: "Should retrieve existing integer value",
		},
		{
			name: "get existing struct value",
			setup: func() {
				client.Set("struct_key", map[string]interface{}{"name": "test", "age": 25}, 0)
			},
			key:         "struct_key",
			dest:        new(map[string]interface{}),
			expectErr:   false,
			expectedVal: map[string]interface{}{"name": "test", "age": 25},
			description: "Should retrieve existing struct value",
		},
		{
			name: "get non-existent key",
			setup: func() {
				// No setup needed
			},
			key:         "non_existent",
			dest:        new(string),
			expectErr:   true,
			expectedVal: nil,
			description: "Should return error for non-existent key",
		},
		{
			name: "get expired key",
			setup: func() {
				client.Set("expired_key", "value", time.Millisecond*1)
				time.Sleep(time.Millisecond * 10) // Wait for expiration
			},
			key:         "expired_key",
			dest:        new(string),
			expectErr:   true,
			expectedVal: nil,
			description: "Should return error for expired key",
		},
		{
			name: "get with nil destination",
			setup: func() {
				client.Set("nil_dest_key", "value", 0)
			},
			key:         "nil_dest_key",
			dest:        nil,
			expectErr:   true,
			expectedVal: nil,
			description: "Should return error for nil destination",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			
			err := client.Get(tt.key, tt.dest)
			
			if tt.expectErr {
				assert.Error(t, err, "Test: %s", tt.description)
			} else {
				assert.NoError(t, err, "Test: %s", tt.description)
				assert.Equal(t, tt.expectedVal, getValueFromPointer(tt.dest), "Test: %s", tt.description)
			}
		})
	}
}

func TestInMemoryClient_Delete(t *testing.T) {
	client := NewInMemoryClient()

	// Setup: add some test data
	client.Set("key1", "value1", 0)
	client.Set("key2", "value2", 0)
	client.Set("key3", "value3", 0)

	tests := []struct {
		name        string
		key         string
		expectErr   bool
		shouldExist bool
		description string
	}{
		{
			name:        "delete existing key",
			key:         "key1",
			expectErr:   false,
			shouldExist: false,
			description: "Should successfully delete existing key",
		},
		{
			name:        "delete non-existent key",
			key:         "non_existent",
			expectErr:   false,
			shouldExist: false,
			description: "Should not error when deleting non-existent key",
		},
		{
			name:        "delete empty key",
			key:         "",
			expectErr:   false,
			shouldExist: false,
			description: "Should handle empty key deletion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if key exists before deletion
			existedBefore := client.Exists(tt.key)
			
			err := client.Delete(tt.key)
			
			assert.NoError(t, err, "Test: %s", tt.description)
			assert.False(t, client.Exists(tt.key), "Key should not exist after deletion")
			
			if existedBefore {
				assert.False(t, client.Exists(tt.key), "Previously existing key should be deleted")
			}
		})
	}
}

func TestInMemoryClient_DeletePattern(t *testing.T) {
	client := NewInMemoryClient()

	// Setup: add test data with patterns
	client.Set("user_1", "value1", 0)
	client.Set("user_2", "value2", 0)
	client.Set("user_3", "value3", 0)
	client.Set("task_1", "value4", 0)
	client.Set("task_2", "value5", 0)
	client.Set("other", "value6", 0)

	tests := []struct {
		name           string
		pattern        string
		expectErr      bool
		remainingKeys  []string
		description    string
	}{
		{
			name:          "delete user pattern",
			pattern:       "user_*",
			expectErr:     false,
			remainingKeys: []string{"task_1", "task_2", "other"},
			description:   "Should delete all keys matching user_* pattern",
		},
		{
			name:          "delete task pattern",
			pattern:       "task_*",
			expectErr:     false,
			remainingKeys: []string{"user_1", "user_2", "user_3", "other"},
			description:   "Should delete all keys matching task_* pattern",
		},
		{
			name:          "delete specific key",
			pattern:       "other",
			expectErr:     false,
			remainingKeys: []string{"user_1", "user_2", "user_3", "task_1", "task_2"},
			description:   "Should delete specific key",
		},
		{
			name:          "delete non-matching pattern",
			pattern:       "nonexistent_*",
			expectErr:     false,
			remainingKeys: []string{"user_1", "user_2", "user_3", "task_1", "task_2", "other"},
			description:   "Should not delete anything for non-matching pattern",
		},
		{
			name:          "delete all with wildcard",
			pattern:       "*",
			expectErr:     false,
			remainingKeys: []string{},
			description:   "Should delete all keys with wildcard pattern",
		},
		{
			name:          "invalid pattern",
			pattern:       "[invalid",
			expectErr:     true,
			remainingKeys: []string{"user_1", "user_2", "user_3", "task_1", "task_2", "other"},
			description:   "Should return error for invalid pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset cache for each test
			client.Clear()
			client.Set("user_1", "value1", 0)
			client.Set("user_2", "value2", 0)
			client.Set("user_3", "value3", 0)
			client.Set("task_1", "value4", 0)
			client.Set("task_2", "value5", 0)
			client.Set("other", "value6", 0)

			err := client.DeletePattern(tt.pattern)
			
			if tt.expectErr {
				assert.Error(t, err, "Test: %s", tt.description)
			} else {
				assert.NoError(t, err, "Test: %s", tt.description)
			}

			// Verify remaining keys
			for _, key := range tt.remainingKeys {
				assert.True(t, client.Exists(key), "Key %s should remain after pattern deletion", key)
			}

			// Verify deleted keys
			allKeys := []string{"user_1", "user_2", "user_3", "task_1", "task_2", "other"}
			for _, key := range allKeys {
				if !contains(tt.remainingKeys, key) {
					assert.False(t, client.Exists(key), "Key %s should be deleted", key)
				}
			}
		})
	}
}

func TestInMemoryClient_Exists(t *testing.T) {
	client := NewInMemoryClient()

	// Setup: add test data
	client.Set("existing_key", "value", 0)
	client.Set("expiring_key", "value", time.Millisecond*1)

	tests := []struct {
		name        string
		key         string
		wait        time.Duration
		expected    bool
		description string
	}{
		{
			name:        "existing key",
			key:         "existing_key",
			wait:        0,
			expected:    true,
			description: "Should return true for existing key",
		},
		{
			name:        "non-existing key",
			key:         "non_existing",
			wait:        0,
			expected:    false,
			description: "Should return false for non-existing key",
		},
		{
			name:        "expired key",
			key:         "expiring_key",
			wait:        time.Millisecond * 10,
			expected:    false,
			description: "Should return false for expired key",
		},
		{
			name:        "empty key",
			key:         "",
			wait:        0,
			expected:    false,
			description: "Should return false for empty key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wait > 0 {
				time.Sleep(tt.wait)
			}
			
			exists := client.Exists(tt.key)
			assert.Equal(t, tt.expected, exists, "Test: %s", tt.description)
		})
	}
}

func TestInMemoryClient_Clear(t *testing.T) {
	client := NewInMemoryClient()

	// Setup: add some test data
	client.Set("key1", "value1", 0)
	client.Set("key2", "value2", 0)
	client.Set("key3", "value3", 0)

	// Verify data exists
	assert.True(t, client.Exists("key1"))
	assert.True(t, client.Exists("key2"))
	assert.True(t, client.Exists("key3"))

	// Clear the cache
	err := client.Clear()
	assert.NoError(t, err)

	// Verify all data is gone
	assert.False(t, client.Exists("key1"))
	assert.False(t, client.Exists("key2"))
	assert.False(t, client.Exists("key3"))
}

func TestInMemoryClient_Concurrency(t *testing.T) {
	client := NewInMemoryClient()
	var wg sync.WaitGroup
	numGoroutines := 100
	numOperations := 10

	// Test concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)
				err := client.Set(key, value, 0)
				assert.NoError(t, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify all values were written
	for i := 0; i < numGoroutines; i++ {
		for j := 0; j < numOperations; j++ {
			key := fmt.Sprintf("key_%d_%d", i, j)
			expectedValue := fmt.Sprintf("value_%d_%d", i, j)
			
			var actualValue string
			err := client.Get(key, &actualValue)
			assert.NoError(t, err)
			assert.Equal(t, expectedValue, actualValue)
		}
	}
}

func TestInMemoryClient_Expiration(t *testing.T) {
	client := NewInMemoryClient()

	// Test with short expiration
	client.Set("short_exp", "value", time.Millisecond*10)
	
	// Should exist immediately
	assert.True(t, client.Exists("short_exp"))
	
	// Wait for expiration
	time.Sleep(time.Millisecond * 20)
	
	// Should not exist after expiration
	assert.False(t, client.Exists("short_exp"))

	// Test with no expiration
	client.Set("no_exp", "value", 0)
	time.Sleep(time.Millisecond * 10)
	assert.True(t, client.Exists("no_exp"), "Key with no expiration should persist")
}

func TestInMemoryClient_JSONHandling(t *testing.T) {
	client := NewInMemoryClient()

	// Test complex struct
	type TestStruct struct {
		Name    string            `json:"name"`
		Age     int               `json:"age"`
		Tags    []string          `json:"tags"`
		Details map[string]string `json:"details"`
	}

	testData := TestStruct{
		Name: "John Doe",
		Age:  30,
		Tags: []string{"tag1", "tag2", "tag3"},
		Details: map[string]string{
			"city":    "New York",
			"country": "USA",
		},
	}

	// Set the struct
	err := client.Set("struct_key", testData, 0)
	assert.NoError(t, err)

	// Get the struct back
	var retrieved TestStruct
	err = client.Get("struct_key", &retrieved)
	assert.NoError(t, err)
	assert.Equal(t, testData, retrieved)
}

// Helper functions
func getValueFromPointer(ptr interface{}) interface{} {
	if ptr == nil {
		return nil
	}
	
	switch v := ptr.(type) {
	case *string:
		return *v
	case *int:
		return *v
	case *map[string]interface{}:
		return *v
	default:
		return v
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func BenchmarkInMemoryClient_Set(b *testing.B) {
	client := NewInMemoryClient()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i)
		client.Set(key, "value", 0)
	}
}

func BenchmarkInMemoryClient_Get(b *testing.B) {
	client := NewInMemoryClient()
	client.Set("test_key", "test_value", 0)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var value string
		client.Get("test_key", &value)
	}
}

func BenchmarkInMemoryClient_Exists(b *testing.B) {
	client := NewInMemoryClient()
	client.Set("test_key", "test_value", 0)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Exists("test_key")
	}
} 