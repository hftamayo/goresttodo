package models

import "github.com/jinzhu/gorm"

type Todo struct {
	gorm.Model
	Id    int    `gorm:"primary_key" json:"id"`
	Title string `gorm:"type:varchar(100)" json:"title"`
	Done  bool   `gorm:"default:false" json:"done"`
	Body  string `gorm:"type:text" json:"body"`
}
