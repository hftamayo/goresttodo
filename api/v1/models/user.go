package models

import (
	"gorm.io/gorm"
)

// the default gorm struct is this:
// type Model struct {
//     ID        uint `gorm:"primary_key"`
//     CreatedAt time.Time
//     UpdatedAt time.Time
//     DeletedAt *time.Time `sql:"index"`
// }

type User struct {
	gorm.Model
	Name     string `gorm:"type:varchar(50)" json:"name"`
	Email    string `gorm:"type:varchar(50)" json:"email"`
	Password string `gorm:"type:varchar(50)" json:"password"`
	status	 bool   `gorm:"default:true" json:"status"`
	Todos    []Todo `gorm:"foreignKey:UserID" json:"todos"`
}
