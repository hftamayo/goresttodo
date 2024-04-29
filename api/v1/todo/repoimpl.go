package todo

import (
	"errors"

	"github.com/jinzhu/gorm"
)

type TodoRepositoryImpl struct {
	// Fields for database connection go here.
}

func (r *TodoRepositoryImpl) GetById(id int) (*Todo, error) {
	var todo Todo
	if result := r.db.First(&todo, id); result.Error != nil {
		// If the record is not found, GORM returns a "record not found" error.
		// You might want to return nil, nil in this case instead of nil, error.
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &todo, nil
}

func (r *TodoRepositoryImpl) GetAll() ([]*Todo, error) {
	var todos []*Todo
	if result := r.db.Find(&todos); result.Error != nil {
		return nil, result.Error
	}
	return todos, nil
}

func (r *TodoRepositoryImpl) Create(todo *Todo) error {
	if result := r.db.Create(todo); result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *TodoRepositoryImpl) Update(todo *Todo) error {
	if result := r.db.Save(todo); result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *TodoRepositoryImpl) Delete(id int) error {
	// Implement this method.
	// Delete the todo with the given id from the database.
	// Return nil if successful, or an error if something goes wrong.
}
