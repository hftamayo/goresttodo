package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadEnvVars(t *testing.T) {
	// Save original environment variables
	originalEnv := make(map[string]string)
	envKeys := []string{
		"POSTGRES_HOST",
		"POSTGRES_PORT",
		"POSTGRES_USER",
		"POSTGRES_PASSWORD",
		"POSTGRES_DB",
		"GOAPP_PORT",
		"GOAPP_MODE",
		"SEED_DEVELOPMENT",
		"SEED_PRODUCTION",
		"FRONTEND_ORIGINS",
	}

	for _, key := range envKeys {
		if value, exists := os.LookupEnv(key); exists {
			originalEnv[key] = value
		}
	}

	// Cleanup function to restore environment variables
	defer func() {
		for key, value := range originalEnv {
			os.Setenv(key, value)
		}
		for _, key := range envKeys {
			if _, exists := originalEnv[key]; !exists {
				os.Unsetenv(key)
			}
		}
	}()

	t.Run("successful load with default values", func(t *testing.T) {
		// Clear environment variables
		for _, key := range envKeys {
			os.Unsetenv(key)
		}

		// Set required password
		os.Setenv("POSTGRES_PASSWORD", "testpassword")

		// Load environment variables
		envVars, err := LoadEnvVars()
		assert.NoError(t, err, "Should load environment variables successfully")
		assert.NotNil(t, envVars, "Environment variables should not be nil")

		// Verify default values
		assert.Equal(t, "localhost", envVars.Host, "Default host should be localhost")
		assert.Equal(t, 5432, envVars.Port, "Default port should be 5432")
		assert.Equal(t, "postgres", envVars.User, "Default user should be postgres")
		assert.Equal(t, "gotodo_dev", envVars.Dbname, "Default database should be gotodo_dev")
		assert.Equal(t, 8001, envVars.AppPort, "Default app port should be 8001")
		assert.Equal(t, "development", envVars.Mode, "Default mode should be development")
		assert.Equal(t, 30, envVars.timeOut, "Default timeout should be 30")
		assert.False(t, envVars.seedDev, "Default seed development should be false")
		assert.False(t, envVars.seedProd, "Default seed production should be false")
		assert.Equal(t, []string{"http://localhost:5173"}, envVars.FeOrigins, "Default frontend origins should be correct")
	})

	t.Run("successful load with custom values", func(t *testing.T) {
		// Set custom environment variables
		os.Setenv("POSTGRES_HOST", "custom-host")
		os.Setenv("POSTGRES_PORT", "5433")
		os.Setenv("POSTGRES_USER", "custom-user")
		os.Setenv("POSTGRES_PASSWORD", "custom-password")
		os.Setenv("POSTGRES_DB", "custom-db")
		os.Setenv("GOAPP_PORT", "8002")
		os.Setenv("GOAPP_MODE", "production")
		os.Setenv("SEED_DEVELOPMENT", "true")
		os.Setenv("SEED_PRODUCTION", "true")
		os.Setenv("FRONTEND_ORIGINS", "http://localhost:3000,http://localhost:3001")

		// Load environment variables
		envVars, err := LoadEnvVars()
		assert.NoError(t, err, "Should load environment variables successfully")
		assert.NotNil(t, envVars, "Environment variables should not be nil")

		// Verify custom values
		assert.Equal(t, "custom-host", envVars.Host, "Host should match custom value")
		assert.Equal(t, 5433, envVars.Port, "Port should match custom value")
		assert.Equal(t, "custom-user", envVars.User, "User should match custom value")
		assert.Equal(t, "custom-password", envVars.Password, "Password should match custom value")
		assert.Equal(t, "custom-db", envVars.Dbname, "Database should match custom value")
		assert.Equal(t, 8002, envVars.AppPort, "App port should match custom value")
		assert.Equal(t, "production", envVars.Mode, "Mode should match custom value")
		assert.True(t, envVars.seedDev, "Seed development should be true")
		assert.True(t, envVars.seedProd, "Seed production should be true")
		assert.Equal(t, []string{"http://localhost:3000", "http://localhost:3001"}, envVars.FeOrigins, "Frontend origins should match custom values")
	})

	t.Run("missing required password", func(t *testing.T) {
		// Clear environment variables
		for _, key := range envKeys {
			os.Unsetenv(key)
		}

		// Load environment variables
		envVars, err := LoadEnvVars()
		assert.Error(t, err, "Should return error for missing password")
		assert.Nil(t, envVars, "Environment variables should be nil")
		assert.Contains(t, err.Error(), "POSTGRES_PASSWORD is required", "Error message should mention missing password")
	})

	t.Run("invalid port values", func(t *testing.T) {
		// Set required password
		os.Setenv("POSTGRES_PASSWORD", "testpassword")

		// Test invalid postgres port
		os.Setenv("POSTGRES_PORT", "invalid")
		envVars, err := LoadEnvVars()
		assert.Error(t, err, "Should return error for invalid postgres port")
		assert.Nil(t, envVars, "Environment variables should be nil")
		assert.Contains(t, err.Error(), "invalid postgres port", "Error message should mention invalid postgres port")

		// Test invalid app port
		os.Setenv("POSTGRES_PORT", "5432")
		os.Setenv("GOAPP_PORT", "invalid")
		envVars, err = LoadEnvVars()
		assert.Error(t, err, "Should return error for invalid app port")
		assert.Nil(t, envVars, "Environment variables should be nil")
		assert.Contains(t, err.Error(), "invalid app port", "Error message should mention invalid app port")
	})

	t.Run("frontend origins parsing", func(t *testing.T) {
		// Set required password
		os.Setenv("POSTGRES_PASSWORD", "testpassword")

		// Test single origin
		os.Setenv("FRONTEND_ORIGINS", "http://localhost:3000")
		envVars, err := LoadEnvVars()
		assert.NoError(t, err, "Should load environment variables successfully")
		assert.Equal(t, []string{"http://localhost:3000"}, envVars.FeOrigins, "Should parse single origin correctly")

		// Test multiple origins with spaces
		os.Setenv("FRONTEND_ORIGINS", "http://localhost:3000, http://localhost:3001,  http://localhost:3002")
		envVars, err = LoadEnvVars()
		assert.NoError(t, err, "Should load environment variables successfully")
		assert.Equal(t, []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:3002",
		}, envVars.FeOrigins, "Should parse multiple origins with spaces correctly")
	})
} 