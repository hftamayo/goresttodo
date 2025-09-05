package health

import (
	"testing"
)

func TestHealthHandler_AppStatus(t *testing.T) {
	// Skip this test as it requires complex integration setup with real gorm.DB
	t.Skip("Skipping AppStatus test - requires complex setup with real gorm.DB")
}

func TestHealthHandler_DbStatus(t *testing.T) {
	// Skip this test as it requires complex integration setup with real gorm.DB
	t.Skip("Skipping DbStatus test - requires complex setup with real gorm.DB")
} 