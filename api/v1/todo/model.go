package todo

type TodoModel struct {
	Id    int    `gorm:"primary_key"`
	Title string `gorm:"type:varchar(100)"`
	Done  bool   `gorm:"default:false"`
	Body  string `gorm:"type:text"`
}
