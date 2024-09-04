package models_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupDBMock() (*gorm.DB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	dsn := "host=localhost user=youruser dbname=yourdb sslmode=disable password=yourpassword"
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

func TestCreateUser(t *testing.T) {
	gormDB, mock, err := setupDBMock()
	assert.NoError(t, err)

	// Mocking the SQL query response
	mock.ExpectBegin()
	mock.ExpectExec("^INSERT INTO `users`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	user := &models.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password",
	}

	err = gormDB.Create(user).Error
	assert.NoError(t, err)
	assert.Equal(t, uint(1), user.ID)
}

func TestGetUserByID(t *testing.T) {
	gormDB, mock, err := setupDBMock()
	assert.NoError(t, err)

	// Mocking the SQL query response
	rows := sqlmock.NewRows([]string{"id", "name", "email", "password"}).
		AddRow(1, "Test User", "test@example.com", "password")
	mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE `users`.`id` = ?").
		WithArgs(1).
		WillReturnRows(rows)

	var user models.User
	err = gormDB.First(&user, 1).Error
	assert.NoError(t, err)
	assert.Equal(t, "Test User", user.Name)
	assert.Equal(t, "test@example.com", user.Email)
}

func TestUpdateUser(t *testing.T) {
	gormDB, mock, err := setupDBMock()
	assert.NoError(t, err)

	// Mocking the SQL query response
	mock.ExpectBegin()
	mock.ExpectExec("^UPDATE `users` SET").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	user := &models.User{
		Model:    gorm.Model{ID: 1},
		Name:     "Updated User",
		Email:    "updated@example.com",
		Password: "newpassword",
	}

	err = gormDB.Save(user).Error
	assert.NoError(t, err)
}

func TestDeleteUser(t *testing.T) {
	gormDB, mock, err := setupDBMock()
	assert.NoError(t, err)

	// Mocking the SQL query response
	mock.ExpectBegin()
	mock.ExpectExec("^DELETE FROM `users` WHERE `users`.`id` = ?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	user := &models.User{
		Model: gorm.Model{ID: 1},
	}

	err = gormDB.Delete(user).Error
	assert.NoError(t, err)
}
