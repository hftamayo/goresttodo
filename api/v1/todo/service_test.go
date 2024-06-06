package todo

import (
	"testing"

	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/api/v1/todo/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTodoRepository struct {
	mock.Mock
}

func (m *MockTodoRepository) Create(todo *models.Todo) error {
	args := m.Called(todo)
	return args.Error(0)
}

func (m *MockTodoRepository) GetById(id int) (*models.Todo, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Todo), args.Error(1)
}

func (m *MockTodoRepository) GetAll(page, pageSize int) ([]*models.Todo, error) {
	args := m.Called(page, pageSize)
	return args.Get(0).([]*models.Todo), args.Error(1)
}

func (m *MockTodoRepository) Update(todo *models.Todo) error {
	args := m.Called(todo)
	return args.Error(0)
}

func (m *MockTodoRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestTodoService_Create(t *testing.T) {
	repo := new(MockTodoRepository)
	todo := &models.Todo{Title: "Test Todo"}
	repo.On("Create", todo).Return(nil)
	service := service.NewTodoService(repo)
	err := service.CreateTodo(todo)
	repo.AssertExpectations(t)
	assert.NoError(t, err)
}
