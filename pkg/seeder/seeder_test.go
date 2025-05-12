package seeder

import (
	"os"
	"testing"

	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockDB is a mock implementation of gorm.DB
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Begin() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Create(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Commit() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Rollback() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func TestSeedData(t *testing.T) {
	// Set up test environment variables
	os.Setenv("ADMINISTRADOR_PASSWORD", "admin123")
	os.Setenv("SUPERVISOR_PASSWORD", "super123")
	os.Setenv("USER01_PASSWORD", "user123")
	os.Setenv("USER02_PASSWORD", "user456")
	defer func() {
		os.Unsetenv("ADMINISTRADOR_PASSWORD")
		os.Unsetenv("SUPERVISOR_PASSWORD")
		os.Unsetenv("USER01_PASSWORD")
		os.Unsetenv("USER02_PASSWORD")
	}()

	tests := []struct {
		name          string
		setupMock     func(*MockDB)
		expectedError bool
	}{
		{
			name: "successful seeding",
			setupMock: func(m *MockDB) {
				// Mock successful transaction
				tx := &gorm.DB{}
				m.On("Begin").Return(tx)
				m.On("Create", mock.AnythingOfType("*models.User")).Return(tx)
				m.On("Create", mock.AnythingOfType("*models.Task")).Return(tx)
				m.On("Commit").Return(tx)
			},
			expectedError: false,
		},
		{
			name: "failed transaction begin",
			setupMock: func(m *MockDB) {
				// Mock failed transaction begin
				tx := &gorm.DB{}
				tx.Error = assert.AnError
				m.On("Begin").Return(tx)
			},
			expectedError: true,
		},
		{
			name: "failed user creation",
			setupMock: func(m *MockDB) {
				// Mock failed user creation
				tx := &gorm.DB{}
				m.On("Begin").Return(tx)
				txWithError := &gorm.DB{}
				txWithError.Error = assert.AnError
				m.On("Create", mock.AnythingOfType("*models.User")).Return(txWithError)
				m.On("Rollback").Return(tx)
			},
			expectedError: true,
		},
		{
			name: "failed task creation",
			setupMock: func(m *MockDB) {
				// Mock failed task creation
				tx := &gorm.DB{}
				m.On("Begin").Return(tx)
				m.On("Create", mock.AnythingOfType("*models.User")).Return(tx)
				txWithError := &gorm.DB{}
				txWithError.Error = assert.AnError
				m.On("Create", mock.AnythingOfType("*models.Task")).Return(txWithError)
				m.On("Rollback").Return(tx)
			},
			expectedError: true,
		},
		{
			name: "failed commit",
			setupMock: func(m *MockDB) {
				// Mock failed commit
				tx := &gorm.DB{}
				m.On("Begin").Return(tx)
				m.On("Create", mock.AnythingOfType("*models.User")).Return(tx)
				m.On("Create", mock.AnythingOfType("*models.Task")).Return(tx)
				txWithError := &gorm.DB{}
				txWithError.Error = assert.AnError
				m.On("Commit").Return(txWithError)
				m.On("Rollback").Return(tx)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DB
			mockDB := new(MockDB)
			tt.setupMock(mockDB)

			// Execute test
			err := SeedData(mockDB)

			// Assert results
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			mockDB.AssertExpectations(t)
		})
	}
}

func TestSeedDataEnvironmentVariables(t *testing.T) {
	// Test with missing environment variables
	mockDB := new(MockDB)
	tx := &gorm.DB{}
	mockDB.On("Begin").Return(tx)
	mockDB.On("Create", mock.AnythingOfType("*models.User")).Return(tx)
	mockDB.On("Create", mock.AnythingOfType("*models.Task")).Return(tx)
	mockDB.On("Commit").Return(tx)

	// Clear environment variables
	os.Unsetenv("ADMINISTRADOR_PASSWORD")
	os.Unsetenv("SUPERVISOR_PASSWORD")
	os.Unsetenv("USER01_PASSWORD")
	os.Unsetenv("USER02_PASSWORD")

	// Execute test
	err := SeedData(mockDB)

	// Assert that the function handles missing environment variables gracefully
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
} 