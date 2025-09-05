package utils

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
    // Truncate password to 72 bytes (bcrypt limit)
    passwordBytes := []byte(password)
    if len(passwordBytes) > 72 {
        passwordBytes = passwordBytes[:72]
    }
    
    bytes, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
    return string(bytes), err
}