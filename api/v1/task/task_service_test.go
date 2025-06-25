package task

import (
	"errors"
	"testing"
	"time"

	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/utils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestTaskService_ListByPage(t *testing.T) {
	tests := []struct {
		name          string
		page          int
		limit         int
		order         string
		setupMocks    func(*MockTaskRepository)
		expectedTasks []*models.Task
		expectedCount int64
		expectedError error
	}{
		{
			name:  "successful list",
			page:  1,
			limit: 10,
			order: "desc",
			setupMocks: func(mockRepo *MockTaskRepository) {
				tasks := []*models.Task{
					{
						Model: gorm.Model{
							ID:        1,
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Title:       "Test Task 1",
						Description: "Test Description 1",
						Done:        false,
						Owner:       1,
					},
				}
				mockRepo.GetTotalCountFunc = func() (int64, error) {
					return int64(1), nil
				}
				// Note: ListByPage would need to be implemented in the repository
				// For now, we'll mock the individual calls that would be made
			},
			expectedTasks: []*models.Task{
				{
					Model: gorm.Model{
						ID:        1,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Title:       "Test Task 1",
					Description: "Test Description 1",
					Done:        false,
					Owner:       1,
				},
			},
			expectedCount: 1,
			expectedError: nil,
		},
		{
			name:  "repository error",
			page:  1,
			limit: 10,
			order: "desc",
			setupMocks: func(mockRepo *MockTaskRepository) {
				mockRepo.GetTotalCountFunc = func() (int64, error) {
					return int64(0), errors.New("database error")
				}
			},
			expectedTasks: nil,
			expectedCount: 0,
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockTaskRepository{}
			mockCache := &utils.Cache{} // Use actual cache type
			tt.setupMocks(mockRepo)

			service := NewTaskService(mockRepo, mockCache)

			tasks, count, err := service.ListByPage(tt.page, tt.limit, tt.order)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
				// Note: We can't easily compare tasks due to time fields
				assert.NotNil(t, tasks)
			}
		})
	}
}

func TestTaskService_ListById(t *testing.T) {
	tests := []struct {
		name          string
		taskID        int
		setupMocks    func(*MockTaskRepository, *utils.Cache)
		expectedTask  *models.Task
		expectedError error
	}{
		{
			name:   "successful get by id",
			taskID: 1,
			setupMocks: func(mockRepo *MockTaskRepository, mockCache *utils.Cache) {
				task := &models.Task{
					Model: gorm.Model{
						ID:        1,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Title:       "Test Task",
					Description: "Test Description",
					Done:        false,
					Owner:       1,
				}
				mockRepo.ListByIdFunc = func(id int) (*models.Task, error) {
					return task, nil
				}
			},
			expectedTask: &models.Task{
				Model: gorm.Model{
					ID:        1,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Title:       "Test Task",
				Description: "Test Description",
				Done:        false,
				Owner:       1,
			},
			expectedError: nil,
		},
		{
			name:   "task not found",
			taskID: 999,
			setupMocks: func(mockRepo *MockTaskRepository, mockCache *utils.Cache) {
				mockRepo.ListByIdFunc = func(id int) (*models.Task, error) {
					return nil, errors.New("task not found")
				}
			},
			expectedTask:  nil,
			expectedError: errors.New("task not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockTaskRepository{}
			mockCache := &utils.Cache{} // Use actual cache type
			tt.setupMocks(mockRepo, mockCache)

			service := NewTaskService(mockRepo, mockCache)

			task, err := service.ListById(tt.taskID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTask.ID, task.ID)
				assert.Equal(t, tt.expectedTask.Title, task.Title)
			}
		})
	}
}

func TestTaskService_Create(t *testing.T) {
	tests := []struct {
		name          string
		task          *models.Task
		setupMocks    func(*MockTaskRepository, *utils.Cache)
		expectedTask  *models.Task
		expectedError error
	}{
		{
			name: "successful create",
			task: &models.Task{
				Title:       "New Task",
				Description: "New Description",
				Done:        false,
				Owner:       1,
			},
			setupMocks: func(mockRepo *MockTaskRepository, mockCache *utils.Cache) {
				createdTask := &models.Task{
					Model: gorm.Model{
						ID:        1,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Title:       "New Task",
					Description: "New Description",
					Done:        false,
					Owner:       1,
				}
				mockRepo.CreateFunc = func(task *models.Task) (*models.Task, error) {
					return createdTask, nil
				}
			},
			expectedTask: &models.Task{
				Model: gorm.Model{
					ID:        1,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Title:       "New Task",
				Description: "New Description",
				Done:        false,
				Owner:       1,
			},
			expectedError: nil,
		},
		{
			name: "repository error",
			task: &models.Task{
				Title:       "New Task",
				Description: "New Description",
				Done:        false,
				Owner:       1,
			},
			setupMocks: func(mockRepo *MockTaskRepository, mockCache *utils.Cache) {
				mockRepo.CreateFunc = func(task *models.Task) (*models.Task, error) {
					return nil, errors.New("database error")
				}
			},
			expectedTask:  nil,
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockTaskRepository{}
			mockCache := &utils.Cache{} // Use actual cache type
			tt.setupMocks(mockRepo, mockCache)

			service := NewTaskService(mockRepo, mockCache)

			task, err := service.Create(tt.task)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTask.ID, task.ID)
				assert.Equal(t, tt.expectedTask.Title, task.Title)
			}
		})
	}
}

func TestTaskService_Update(t *testing.T) {
	tests := []struct {
		name          string
		taskID        int
		task          *models.Task
		setupMocks    func(*MockTaskRepository, *utils.Cache)
		expectedTask  *models.Task
		expectedError error
	}{
		{
			name:   "successful update",
			taskID: 1,
			task: &models.Task{
				Title:       "Updated Task",
				Description: "Updated Description",
				Done:        true,
				Owner:       1,
			},
			setupMocks: func(mockRepo *MockTaskRepository, mockCache *utils.Cache) {
				updatedTask := &models.Task{
					Model: gorm.Model{
						ID:        1,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Title:       "Updated Task",
					Description: "Updated Description",
					Done:        true,
					Owner:       1,
				}
				mockRepo.UpdateFunc = func(id int, task *models.Task) (*models.Task, error) {
					return updatedTask, nil
				}
			},
			expectedTask: &models.Task{
				Model: gorm.Model{
					ID:        1,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Title:       "Updated Task",
				Description: "Updated Description",
				Done:        true,
				Owner:       1,
			},
			expectedError: nil,
		},
		{
			name:   "repository error",
			taskID: 1,
			task: &models.Task{
				Title:       "Updated Task",
				Description: "Updated Description",
				Done:        true,
				Owner:       1,
			},
			setupMocks: func(mockRepo *MockTaskRepository, mockCache *utils.Cache) {
				mockRepo.UpdateFunc = func(id int, task *models.Task) (*models.Task, error) {
					return nil, errors.New("database error")
				}
			},
			expectedTask:  nil,
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockTaskRepository{}
			mockCache := &utils.Cache{} // Use actual cache type
			tt.setupMocks(mockRepo, mockCache)

			service := NewTaskService(mockRepo, mockCache)

			task, err := service.Update(tt.taskID, tt.task)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTask.ID, task.ID)
				assert.Equal(t, tt.expectedTask.Title, task.Title)
			}
		})
	}
}

func TestTaskService_MarkAsDone(t *testing.T) {
	tests := []struct {
		name          string
		taskID        int
		setupMocks    func(*MockTaskRepository, *utils.Cache)
		expectedTask  *models.Task
		expectedError error
	}{
		{
			name:   "successful mark as done",
			taskID: 1,
			setupMocks: func(mockRepo *MockTaskRepository, mockCache *utils.Cache) {
				updatedTask := &models.Task{
					Model: gorm.Model{
						ID:        1,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Title:       "Test Task",
					Description: "Test Description",
					Done:        true,
					Owner:       1,
				}
				mockRepo.MarkAsDoneFunc = func(id int) (*models.Task, error) {
					return updatedTask, nil
				}
			},
			expectedTask: &models.Task{
				Model: gorm.Model{
					ID:        1,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Title:       "Test Task",
				Description: "Test Description",
				Done:        true,
				Owner:       1,
			},
			expectedError: nil,
		},
		{
			name:   "repository error",
			taskID: 1,
			setupMocks: func(mockRepo *MockTaskRepository, mockCache *utils.Cache) {
				mockRepo.MarkAsDoneFunc = func(id int) (*models.Task, error) {
					return nil, errors.New("database error")
				}
			},
			expectedTask:  nil,
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockTaskRepository{}
			mockCache := &utils.Cache{} // Use actual cache type
			tt.setupMocks(mockRepo, mockCache)

			service := NewTaskService(mockRepo, mockCache)

			task, err := service.MarkAsDone(tt.taskID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTask.ID, task.ID)
				assert.True(t, task.Done)
			}
		})
	}
}

func TestTaskService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		taskID        int
		setupMocks    func(*MockTaskRepository, *utils.Cache)
		expectedError error
	}{
		{
			name:   "successful delete",
			taskID: 1,
			setupMocks: func(mockRepo *MockTaskRepository, mockCache *utils.Cache) {
				mockRepo.DeleteFunc = func(id int) error {
					return nil
				}
			},
			expectedError: nil,
		},
		{
			name:   "repository error",
			taskID: 1,
			setupMocks: func(mockRepo *MockTaskRepository, mockCache *utils.Cache) {
				mockRepo.DeleteFunc = func(id int) error {
					return errors.New("database error")
				}
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockTaskRepository{}
			mockCache := &utils.Cache{} // Use actual cache type
			tt.setupMocks(mockRepo, mockCache)

			service := NewTaskService(mockRepo, mockCache)

			err := service.Delete(tt.taskID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
} 