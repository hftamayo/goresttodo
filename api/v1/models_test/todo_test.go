package models_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/config"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// Test Create Todo
func TestCreateTodo(t *testing.T) {
	gormDB, mock, err := config.SetupDBMock()
	assert.NoError(t, err)

	// Mocking the SQL query response
	mock.ExpectBegin()
	mock.ExpectExec("^INSERT INTO `todos`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	todo := &models.Todo{
		Title: "Test Todo",
		Done:  false,
		Body:  "Test Body",
	}

	err = gormDB.Create(todo).Error
	assert.NoError(t, err)
	assert.Equal(t, uint(1), todo.ID)
}

// Test Read Todo
func TestGetTodoByID(t *testing.T) {
	gormDB, mock, err := config.SetupDBMock()
	assert.NoError(t, err)

	// Mocking the SQL query response
	rows := sqlmock.NewRows([]string{"id", "title", "done", "body"}).
		AddRow(1, "Test Todo", false, "Test Body")
	mock.ExpectQuery("^SELECT (.+) FROM `todos` WHERE `todos`.`id` = ?").
		WithArgs(1).
		WillReturnRows(rows)

	var todo models.Todo
	err = gormDB.First(&todo, 1).Error
	assert.NoError(t, err)
	assert.Equal(t, "Test Todo", todo.Title)
	assert.Equal(t, false, todo.Done)
	assert.Equal(t, "Test Body", todo.Body)
}

// Test Update Todo
func TestUpdateTodo(t *testing.T) {
	gormDB, mock, err := config.SetupDBMock()
	assert.NoError(t, err)

	// Mocking the SQL query response
	mock.ExpectBegin()
	mock.ExpectExec("^UPDATE `todos` SET").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	todo := &models.Todo{
		Model: gorm.Model{ID: 1},
		Title: "Updated Todo",
		Done:  true,
		Body:  "Updated Body",
	}

	err = gormDB.Save(todo).Error
	assert.NoError(t, err)
}

// Test Delete Todo
func TestDeleteTodo(t *testing.T) {
	gormDB, mock, err := config.SetupDBMock()
	assert.NoError(t, err)

	// Mocking the SQL query response
	mock.ExpectBegin()
	mock.ExpectExec("^DELETE FROM `todos` WHERE `todos`.`id` = ?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	todo := &models.Todo{
		Model: gorm.Model{ID: 1},
	}

	err = gormDB.Delete(todo).Error
	assert.NoError(t, err)
}
