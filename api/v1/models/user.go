package models

import (
	"errors"
	"strings"

	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Id       int    `gorm:"primary_key" json:"id"`
	Name     string `gorm:"type:varchar(50)" json:"name"`
	Email    string `gorm:"type:varchar(50)" json:"email"`
	Password string `gorm:"type:varchar(50)" json:"password"`
}

func (u *User) Validate() error {
	if u.Name == "" {
		return errors.New("name is required")
	}
	if u.Email == "" {
		return errors.New("email is required")
	}
	if u.Password == "" {
		return errors.New("password is required")
	}
	if len(u.Password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}
	if !strings.Contains(u.Email, "@") {
		return errors.New("email is invalid")
	}
	return nil
}
