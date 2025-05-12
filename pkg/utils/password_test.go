package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name          string
		password      string
		expectedError bool
	}{
		{
			name:          "valid password",
			password:      "validPassword123!",
			expectedError: false,
		},
		{
			name:          "empty password",
			password:      "",
			expectedError: false,
		},
		{
			name:          "very long password",
			password:      strings.Repeat("a", 1000),
			expectedError: false,
		},
		{
			name:          "password with special characters",
			password:      "!@#$%^&*()_+-=[]{}|;:,.<>?",
			expectedError: false,
		},
		{
			name:          "password with unicode characters",
			password:      "パスワード123",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Hash the password
			hashedPassword, err := HashPassword(tt.password)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Empty(t, hashedPassword)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hashedPassword)
				assert.NotEqual(t, tt.password, hashedPassword)

				// Verify the hash can be compared with bcrypt
				err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(tt.password))
				assert.NoError(t, err)

				// Verify the hash is different for the same password (due to salt)
				secondHash, err := HashPassword(tt.password)
				assert.NoError(t, err)
				assert.NotEqual(t, hashedPassword, secondHash)

				// Verify the hash can still be compared with bcrypt
				err = bcrypt.CompareHashAndPassword([]byte(secondHash), []byte(tt.password))
				assert.NoError(t, err)
			}
		})
	}
}

func TestHashPassword_InvalidInput(t *testing.T) {
	// Test with nil password
	hashedPassword, err := HashPassword("")
	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)

	// Test with extremely long password
	longPassword := strings.Repeat("a", 10000)
	hashedPassword, err = HashPassword(longPassword)
	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)

	// Verify the hash can be compared with bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(longPassword))
	assert.NoError(t, err)
}

func TestHashPassword_Consistency(t *testing.T) {
	password := "testPassword123!"
	
	// Hash the same password multiple times
	hashes := make([]string, 5)
	for i := 0; i < 5; i++ {
		hash, err := HashPassword(password)
		assert.NoError(t, err)
		hashes[i] = hash
	}

	// Verify all hashes are different (due to salt)
	for i := 0; i < len(hashes); i++ {
		for j := i + 1; j < len(hashes); j++ {
			assert.NotEqual(t, hashes[i], hashes[j], "Hashes should be different due to salt")
		}
	}

	// Verify all hashes can be compared with the original password
	for _, hash := range hashes {
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		assert.NoError(t, err, "Hash should be comparable with original password")
	}
} 