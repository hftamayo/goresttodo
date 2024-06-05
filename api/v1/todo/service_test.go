package todo

import (
	"testing"

	"github.com/hftamayo/gotodo/api/v1/models"
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

func TestTodoService_Create(t *testing.T) {
	repo := new(MockTodoRepository)
	todo := &Todo{Title: "Test Todo"}
	repo.On("Create", todo).Return(nil)
	service := NewTodoService(repo)
	err := service.Create(todo)
	repo.AssertExpectations(t)
	assert.NoError(t, err)
}
