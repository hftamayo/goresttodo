package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockGormDB is a mock implementation of gorm.DB
type MockGormDB struct {
	mock.Mock
}

func (m *MockGormDB) DB() (*gorm.DB, error) {
	args := m.Called()
	return args.Get(0).(*gorm.DB), args.Error(1)
}

func (m *MockGormDB) AutoMigrate(values ...interface{}) error {
	args := m.Called(values)
	return args.Error(0)
}

func TestBuildConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		envVars  *EnvVars
		expected string
	}{
		{
			name: "valid connection string",
			envVars: &EnvVars{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				Dbname:   "testdb",
				timeOut:  30,
			},
			expected: "host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=disable connect_timeout=30",
		},
		{
			name: "connection string with special characters",
			envVars: &EnvVars{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres@user",
				Password: "pass@word",
				Dbname:   "test-db",
				timeOut:  30,
			},
			expected: "host=localhost port=5432 user=postgres@user password=pass@word dbname=test-db sslmode=disable connect_timeout=30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildConnectionString(tt.envVars)
			assert.Equal(t, tt.expected, result, "Connection string should match expected format")
		})
	}
}

func TestIsTestEnviro(t *testing.T) {
	tests := []struct {
		name     string
		envVars  *EnvVars
		expected bool
	}{
		{
			name: "development mode",
			envVars: &EnvVars{
				Mode: "development",
			},
			expected: false,
		},
		{
			name: "testing mode",
			envVars: &EnvVars{
				Mode: "testing",
			},
			expected: true,
		},
		{
			name: "production mode",
			envVars: &EnvVars{
				Mode: "production",
			},
			expected: true,
		},
		{
			name: "invalid mode",
			envVars: &EnvVars{
				Mode: "invalid",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTestEnviro(tt.envVars)
			assert.Equal(t, tt.expected, result, "Test environment check should match expected result")
		})
	}
}

func TestMaskConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "mask password in connection string",
			input:    "host=localhost port=5432 user=postgres password=secret dbname=testdb",
			expected: "host=localhost port=5432 user=postgres password=***** dbname=testdb",
		},
		{
			name:     "connection string without password",
			input:    "host=localhost port=5432 user=postgres dbname=testdb",
			expected: "host=localhost port=5432 user=postgres dbname=testdb",
		},
		{
			name:     "empty connection string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskConnectionString(tt.input)
			assert.Equal(t, tt.expected, result, "Masked connection string should match expected format")
		})
	}
}

func TestCheckDataLayerAvailability(t *testing.T) {
	// Skip this test if PostgreSQL is not available
	t.Skip("Skipping PostgreSQL connection test - requires PostgreSQL server")

	tests := []struct {
		name        string
		envVars     *EnvVars
		mockDB      *MockGormDB
		expectError bool
	}{
		{
			name: "successful connection",
			envVars: &EnvVars{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				Dbname:   "testdb",
				timeOut:  30,
			},
			mockDB:      new(MockGormDB),
			expectError: false,
		},
		{
			name: "connection failure",
			envVars: &EnvVars{
				Host:     "invalid-host",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				Dbname:   "testdb",
				timeOut:  30,
			},
			mockDB:      new(MockGormDB),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock DB operations
			if !tt.expectError {
				tt.mockDB.On("DB").Return(new(gorm.DB), nil)
			}

			// Test connection
			db, err := CheckDataLayerAvailability(tt.envVars)
			if tt.expectError {
				assert.Error(t, err, "Should return error for failed connection")
				assert.Nil(t, db, "Database connection should be nil")
			} else {
				assert.NoError(t, err, "Should connect successfully")
				assert.NotNil(t, db, "Database connection should not be nil")
			}

			// Verify mock expectations
			tt.mockDB.AssertExpectations(t)
		})
	}
}

func TestDataLayerConnect(t *testing.T) {
	// Skip this test if PostgreSQL is not available
	t.Skip("Skipping PostgreSQL connection test - requires PostgreSQL server")

	tests := []struct {
		name        string
		envVars     *EnvVars
		mockDB      *MockGormDB
		expectError bool
	}{
		{
			name: "successful connection in development mode",
			envVars: &EnvVars{
				Mode:     "development",
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				Dbname:   "testdb",
				timeOut:  30,
				seedDev:  false,
				seedProd: false,
			},
			mockDB:      new(MockGormDB),
			expectError: false,
		},
		{
			name: "connection in testing mode",
			envVars: &EnvVars{
				Mode: "testing",
			},
			mockDB:      new(MockGormDB),
			expectError: true,
		},
		{
			name: "connection with seeding required",
			envVars: &EnvVars{
				Mode:     "development",
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				Dbname:   "testdb",
				timeOut:  30,
				seedDev:  true,
				seedProd: false,
			},
			mockDB:      new(MockGormDB),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock DB operations
			if !tt.expectError {
				tt.mockDB.On("AutoMigrate", mock.Anything).Return(nil)
			}

			// Test connection
			db, err := DataLayerConnect(tt.envVars)
			if tt.expectError {
				assert.Error(t, err, "Should return error for failed connection")
				assert.Nil(t, db, "Database connection should be nil")
			} else {
				assert.NoError(t, err, "Should connect successfully")
				assert.NotNil(t, db, "Database connection should not be nil")
			}

			// Verify mock expectations
			tt.mockDB.AssertExpectations(t)
		})
	}
} 