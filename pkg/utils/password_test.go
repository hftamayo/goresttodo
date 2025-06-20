package utils

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: TestValidPassword,
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "password with special characters",
			password: TestStrongPassword,
			wantErr:  false,
		},
		{
			name:     "very long password",
			password: strings.Repeat("a", 1000),
			wantErr:  false,
		},
		{
			name:     "password with unicode characters",
			password: TestUnicodePassword,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("HashPassword() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("HashPassword() unexpected error = %v", err)
				return
			}
			
			if hash == "" {
				t.Errorf("HashPassword() returned empty hash")
			}
			
			// Verify the hash is valid bcrypt format
			if !strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$2b$") && !strings.HasPrefix(hash, "$2y$") {
				t.Errorf("HashPassword() returned invalid bcrypt hash format: %s", hash)
			}
		})
	}
}

func TestHashPassword_Consistency(t *testing.T) {
	password := TestValidPassword
	
	// Hash the same password multiple times
	hash1, err1 := HashPassword(password)
	hash2, err2 := HashPassword(password)
	hash3, err3 := HashPassword(password)
	
	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatalf("HashPassword() failed: %v, %v, %v", err1, err2, err3)
	}
	
	// Hashes should be different due to salt
	if hash1 == hash2 || hash1 == hash3 || hash2 == hash3 {
		t.Errorf("HashPassword() returned identical hashes for same password, expected different hashes due to salt")
	}
	
	// All hashes should be valid
	if err := bcrypt.CompareHashAndPassword([]byte(hash1), []byte(password)); err != nil {
		t.Errorf("HashPassword() generated invalid hash: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash2), []byte(password)); err != nil {
		t.Errorf("HashPassword() generated invalid hash: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash3), []byte(password)); err != nil {
		t.Errorf("HashPassword() generated invalid hash: %v", err)
	}
}

func TestHashPassword_Verification(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "simple password",
			password: TestValidPassword,
		},
		{
			name:     "complex password",
			password: TestStrongPassword,
		},
		{
			name:     "password with spaces",
			password: TestPasswordWithSpace,
		},
		{
			name:     "numeric password",
			password: TestNumericPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if err != nil {
				t.Fatalf("HashPassword() failed: %v", err)
			}
			
			// Verify the hash matches the original password
			err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(tt.password))
			if err != nil {
				t.Errorf("Hash verification failed: %v", err)
			}
			
			// Verify the hash doesn't match wrong password
			wrongPassword := tt.password + "wrong"
			err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(wrongPassword))
			if err == nil {
				t.Errorf("Hash verification should have failed for wrong password")
			}
		})
	}
}

func TestHashPassword_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		password string
		desc     string
	}{
		{
			name:     "single character",
			password: "a",
			desc:     "very short password",
		},
		{
			name:     "whitespace only",
			password: "   ",
			desc:     "password with only whitespace",
		},
		{
			name:     "null bytes",
			password: string([]byte{0, 1, 2, 3}),
			desc:     "password with null bytes",
		},
		{
			name:     "emoji password",
			password: TestEmojiPassword,
			desc:     "password with emojis",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if err != nil {
				t.Errorf("HashPassword() failed for %s: %v", tt.desc, err)
				return
			}
			
			// Verify the hash is valid
			err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(tt.password))
			if err != nil {
				t.Errorf("Hash verification failed for %s: %v", tt.desc, err)
			}
		})
	}
}

func TestHashPassword_Performance(t *testing.T) {
	password := TestValidPassword
	
	// Test that hashing doesn't take too long
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}
	
	if hash == "" {
		t.Errorf("HashPassword() returned empty hash")
	}
	
	// Verify the hash is valid
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		t.Errorf("Hash verification failed: %v", err)
	}
}

func BenchmarkHashPassword(b *testing.B) {
	password := TestValidPassword
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := HashPassword(password)
		if err != nil {
			b.Fatalf("HashPassword() failed: %v", err)
		}
	}
}

func BenchmarkHashPassword_Empty(b *testing.B) {
	password := ""
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := HashPassword(password)
		if err != nil {
			b.Fatalf("HashPassword() failed: %v", err)
		}
	}
}

func BenchmarkHashPassword_Long(b *testing.B) {
	password := strings.Repeat("a", 100)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := HashPassword(password)
		if err != nil {
			b.Fatalf("HashPassword() failed: %v", err)
		}
	}
} 