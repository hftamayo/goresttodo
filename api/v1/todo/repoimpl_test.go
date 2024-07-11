package todo

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"yourapp/repoimpl and model"
)

func setupDBMock() (*gorm.DB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New() // Create instance of SQL mock.
	if err != nil {
		return nil, nil, err
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	return gormDB, mock, nil
}

func TestGetById_Found(t *testing.T) {
	gormDB, mock, err := setupDBMock()
	assert.NoError(t, err)

	repo := repoimpl.TodoRepositoryImpl{Db: gormDB}

	// Mocking the SQL query response.
	rows := sqlmock.NewRows([]string{"id", "title", "completed"}).
		AddRow(1, "Test Todo", false)
	mock.ExpectQuery("^SELECT \\* FROM `todos` WHERE `todos`.`id` = \\?").
		WithArgs(1).
		WillReturnRows(rows)

	// Call the method
	todo, err := repo.GetById(1)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, todo)
	assert.Equal(t, "Test Todo", todo.Title)
}

func TestGetById_NotFound(t *testing.T) {
	gormDB, mock, err := setupDBMock()
	assert.NoError(t, err)

	repo := repoimpl.TodoRepositoryImpl{Db: gormDB}

	// Mocking the SQL query response for not found.
	mock.ExpectQuery("^SELECT \\* FROM `todos` WHERE `todos`.`id` = \\?").
		WithArgs(1).
		WillReturnError(gorm.ErrRecordNotFound)

	// Call the method
	todo, err := repo.GetById(1)

	// Assertions
	assert.NoError(t, err)
	assert.Nil(t, todo)
}

func TestGetById_Error(t *testing.T) {
	gormDB, mock, err := setupDBMock()
	assert.NoError(t, err)

	repo := repoimpl.TodoRepositoryImpl{Db: gormDB}

	// Mocking the SQL query response for an error.
	mock.ExpectQuery("^SELECT \\* FROM `todos` WHERE `todos`.`id` = \\?").
		WithArgs(1).
		WillReturnError(errors.New("unexpected error"))

	// Call the method
	todo, err := repo.GetById(1)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, todo)
}
