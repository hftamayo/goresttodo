package todo

import (
	"errors"

	"github.com/hftamayo/gotodo/api/v1/models"
	"gorm.io/gorm"
)

type TodoRepositoryImpl struct {
	db *gorm.DB
}

func (r *TodoRepositoryImpl) GetById(id int) (*models.Todo, error) {
	var todo models.Todo
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

func (r *TodoRepositoryImpl) GetAll(page, pageSize int) ([]*models.Todo, error) {
	var todos []*models.Todo
	offset := (page - 1) * pageSize
	if result := r.db.Offset(offset).Limit(pageSize).Find(&todos); result.Error != nil {
		return nil, result.Error
	}
	return todos, nil
}

func (r *TodoRepositoryImpl) Create(todo *models.Todo) error {
	if result := r.db.Create(todo); result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *TodoRepositoryImpl) Update(todo *models.Todo) error {
	if result := r.db.Save(todo); result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *TodoRepositoryImpl) Delete(id int) error {
	todo := &models.Todo{}
	if result := r.db.First(todo, id); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil
		}
		return result.Error
	}

	if result := r.db.Delete(todo); result.Error != nil {
		return result.Error
	}
	return nil
}
