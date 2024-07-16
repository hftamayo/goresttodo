package todotest

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	repoimpl "github.com/hftamayo/gotodo/api/v1/todo"
)

func setupDBMock() (*gorm.DB, sqlmock.Sqlmock, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("Error reading context data: %w", err)
	}

	db, mock, err := sqlmock.New() // Create instance of SQL mock.
	if err != nil {
		return nil, nil, err
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"))

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
		Conn:                 db,
	}), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	return gormDB, mock, nil
}

func TestGetById(t *testing.T) {
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

func TestGetByIdNotFound(t *testing.T) {
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

func TestGetByIdError(t *testing.T) {
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

func TestGetAll(t *testing.T) {
	gormDB, mock, err := setupDBMock()
	assert.NoError(t, err)

	repo := repoimpl.TodoRepositoryImpl{Db: gormDB}

	// Mocking the SQL query response.
	rows := sqlmock.NewRows([]string{"id", "title", "completed"}).
		AddRow(1, "Test Todo 1", false).
		AddRow(2, "Test Todo 2", true)
	mock.ExpectQuery("^SELECT \\* FROM `todos`").
		WillReturnRows(rows)

	// Call the method
	todos, err := repo.GetAll(1, 10)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, todos)
	assert.Len(t, todos, 2)
	assert.Equal(t, "Test Todo 1", todos[0].Title)
	assert.Equal(t, "Test Todo 2", todos[1].Title)
}

func TestGetAllError(t *testing.T) {
	gormDB, mock, err := setupDBMock()
	assert.NoError(t, err)

	repo := repoimpl.TodoRepositoryImpl{Db: gormDB}

	// Mocking the SQL query response for an error.
	mock.ExpectQuery("^SELECT \\* FROM `todos`").
		WillReturnError(errors.New("unexpected error"))

	// Call the method
	todos, err := repo.GetAll(1, 10)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, todos)
}
