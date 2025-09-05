package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/hftamayo/gotodo/api/v1/models"
)

func ValidateUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := validateUser(user); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func validateUser(user models.User) error {
	// Check if full name is empty or only whitespace
	if strings.TrimSpace(user.FullName) == "" {
		return errors.New("name is required")
	}
	
	// Check if email is empty or only whitespace
	if strings.TrimSpace(user.Email) == "" {
		return errors.New("email is required")
	}
	
	// Check if password is empty
	if user.Password == "" {
		return errors.New("password is required")
	}
	
	// Check password length
	if len(user.Password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}
	
	// More robust email validation
	email := strings.TrimSpace(user.Email)
	if !isValidEmail(email) {
		return errors.New("email is invalid")
	}
	
	return nil
}

// isValidEmail performs basic email validation
func isValidEmail(email string) bool {
	// Check if email contains @
	if !strings.Contains(email, "@") {
		return false
	}
	
	// Split by @ and check both parts
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	
	// Check that both parts are not empty
	if strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return false
	}
	
	return true
}
