package mocks

import (
	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/stretchr/testify/mock"
)

type MockTaskService struct {
    mock.Mock
}

func (m *MockTaskService) List() ([]*models.Task, error) {
    args := m.Called()
    return args.Get(0).([]*models.Task), args.Error(1)
}

func (m *MockTaskService) ListById(id int) (*models.Task, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskService) Create(task *models.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskService) Update(task *models.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskService) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}
