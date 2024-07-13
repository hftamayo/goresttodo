package models

import "gorm.io/gorm"

// the default gorm struct is this:
// type Model struct {
//     ID        uint `gorm:"primary_key"`
//     CreatedAt time.Time
//     UpdatedAt time.Time
//     DeletedAt *time.Time `sql:"index"`
// }

type Todo struct {
	gorm.Model
	Title string `gorm:"type:varchar(100)" json:"title"`
	Done  bool   `gorm:"default:false" json:"done"`
	Body  string `gorm:"type:text" json:"body"`
}
