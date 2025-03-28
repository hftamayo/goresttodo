package models

import "gorm.io/gorm"

// the default gorm struct is this:
// type Model struct {
//     ID        uint `gorm:"primary_key"`
//     CreatedAt time.Time
//     UpdatedAt time.Time
//     DeletedAt *time.Time `sql:"index"`
// }

type Task struct {
	gorm.Model
	Title string `gorm:"type:varchar(100)" json:"title"`
	Description  string `gorm:"type:text" json:"body"`
	Done  bool   `gorm:"default:false" json:"done"`
	Owner  uint   `json:"owner"` 
    User    User   `gorm:"foreignKey:Owner" json:"user"` 
}
