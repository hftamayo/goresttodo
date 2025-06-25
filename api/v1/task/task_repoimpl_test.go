package task

import (
	"errors"
	"testing"
	"time"

	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/cursor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockDB mocks the gorm.DB
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0).(*gorm.DB)
}

func (m *MockDB) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	mockArgs := m.Called(dest, conds)
	return mockArgs.Get(0).(*gorm.DB)
}

func (m *MockDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	mockArgs := m.Called(dest, conds)
	return mockArgs.Get(0).(*gorm.DB)
}

func (m *MockDB) Create(value interface{}) *gorm.DB {
	mockArgs := m.Called(value)
	return mockArgs.Get(0).(*gorm.DB)
}

func (m *MockDB) Save(value interface{}) *gorm.DB {
	mockArgs := m.Called(value)
	return mockArgs.Get(0).(*gorm.DB)
}

func (m *MockDB) Delete(value interface{}, conds ...interface{}) *gorm.DB {
	mockArgs := m.Called(value, conds)
	return mockArgs.Get(0).(*gorm.DB)
}

func (m *MockDB) Count(count *int64) *gorm.DB {
	mockArgs := m.Called(count)
	return mockArgs.Get(0).(*gorm.DB)
}

func (m *MockDB) Offset(offset int) *gorm.DB {
	mockArgs := m.Called(offset)
	return mockArgs.Get(0).(*gorm.DB)
}

func (m *MockDB) Limit(limit int) *gorm.DB {
	mockArgs := m.Called(limit)
	return mockArgs.Get(0).(*gorm.DB)
}

func (m *MockDB) Order(value interface{}) *gorm.DB {
	mockArgs := m.Called(value)
	return mockArgs.Get(0).(*gorm.DB)
}

func (m *MockDB) Error() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDB) RowsAffected() int64 {
	args := m.Called()
	return args.Get(0).(int64)
}

func TestNewTaskRepositoryImpl(t *testing.T) {
	tests := []struct {
		name    string
		db      *gorm.DB
		wantErr bool
	}{
		{
			name:    "nil database",
			db:      nil,
			wantErr: true,
		},
		{
			name:    "valid database",
			db:      &gorm.DB{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewTaskRepositoryImpl(tt.db)
			if tt.wantErr {
				assert.Nil(t, repo)
			} else {
				assert.NotNil(t, repo)
			}
		})
	}
}

func TestTaskRepositoryImpl_GetTotalCount(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*MockDB)
		expectedCount int64
		expectedError error
	}{
		{
			name: "successful count",
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Count", mock.AnythingOfType("*int64")).Return(mockDB).Once()
				mockDB.On("Error").Return(nil).Once()
			},
			expectedCount: 10,
			expectedError: nil,
		},
		{
			name: "database error",
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Count", mock.AnythingOfType("*int64")).Return(mockDB).Once()
				mockDB.On("Error").Return(errors.New("database error")).Once()
			},
			expectedCount: 0,
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockDB{}
			tt.setupMock(mockDB)

			repo := &TaskRepositoryImpl{
				db: mockDB,
			}

			count, err := repo.GetTotalCount()

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestTaskRepositoryImpl_ListById(t *testing.T) {
	tests := []struct {
		name          string
		taskID        int
		setupMock     func(*MockDB)
		expectedTask  *models.Task
		expectedError error
	}{
		{
			name:   "successful get by id",
			taskID: 1,
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Where", "id = ?", 1).Return(mockDB).Once()
				mockDB.On("First", mock.AnythingOfType("*models.Task")).Return(mockDB).Once()
				mockDB.On("Error").Return(nil).Once()
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
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Where", "id = ?", 999).Return(mockDB).Once()
				mockDB.On("First", mock.AnythingOfType("*models.Task")).Return(mockDB).Once()
				mockDB.On("Error").Return(gorm.ErrRecordNotFound).Once()
			},
			expectedTask:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:   "database error",
			taskID: 1,
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Where", "id = ?", 1).Return(mockDB).Once()
				mockDB.On("First", mock.AnythingOfType("*models.Task")).Return(mockDB).Once()
				mockDB.On("Error").Return(errors.New("database error")).Once()
			},
			expectedTask:  nil,
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockDB{}
			tt.setupMock(mockDB)

			repo := &TaskRepositoryImpl{
				db: mockDB,
			}

			task, err := repo.ListById(tt.taskID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTask, task)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestTaskRepositoryImpl_ListByPage(t *testing.T) {
	mockDB := new(MockDB)
	repo := &TaskRepositoryImpl{db: mockDB}

	now := time.Now()
	tasks := []*models.Task{
		{
			ID:          1,
			Title:       "Task 1",
			Description: "Description 1",
			Done:        false,
			Owner:       "test@example.com",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          2,
			Title:       "Task 2",
			Description: "Description 2",
			Done:        true,
			Owner:       "test@example.com",
			CreatedAt:   now.Add(-time.Hour),
			UpdatedAt:   now.Add(-time.Hour),
		},
	}

	tests := []struct {
		name       string
		page       int
		limit      int
		order      string
		setup      func()
		wantTasks  []*models.Task
		wantCount  int64
		wantErr    bool
	}{
		{
			name:  "successful list with default values",
			page:  1,
			limit: 0,
			order: "",
			setup: func() {
				mockDB.On("Model", &models.Task{}).Return(mockDB)
				mockDB.On("Count", mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*int64)
					*arg = 2
				}).Return(&gorm.DB{Error: nil})
				mockDB.On("Order", "created_at desc, id desc").Return(mockDB)
				mockDB.On("Select", "id, title, description, done, owner, created_at, updated_at").Return(mockDB)
				mockDB.On("Offset", 0).Return(mockDB)
				mockDB.On("Limit", utils.DefaultLimit).Return(mockDB)
				mockDB.On("Find", mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*[]*models.Task)
					*arg = tasks
				}).Return(&gorm.DB{Error: nil})
			},
			wantTasks: tasks,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:  "successful list with custom values",
			page:  2,
			limit: 1,
			order: "asc",
			setup: func() {
				mockDB.On("Model", &models.Task{}).Return(mockDB)
				mockDB.On("Count", mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*int64)
					*arg = 2
				}).Return(&gorm.DB{Error: nil})
				mockDB.On("Order", "created_at asc, id asc").Return(mockDB)
				mockDB.On("Select", "id, title, description, done, owner, created_at, updated_at").Return(mockDB)
				mockDB.On("Offset", 1).Return(mockDB)
				mockDB.On("Limit", 1).Return(mockDB)
				mockDB.On("Find", mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*[]*models.Task)
					*arg = tasks[1:2]
				}).Return(&gorm.DB{Error: nil})
			},
			wantTasks: tasks[1:2],
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:  "database error on count",
			page:  1,
			limit: 10,
			order: "desc",
			setup: func() {
				mockDB.On("Model", &models.Task{}).Return(mockDB)
				mockDB.On("Count", mock.Anything).Return(&gorm.DB{Error: assert.AnError})
			},
			wantTasks: nil,
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:  "database error on find",
			page:  1,
			limit: 10,
			order: "desc",
			setup: func() {
				mockDB.On("Model", &models.Task{}).Return(mockDB)
				mockDB.On("Count", mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*int64)
					*arg = 2
				}).Return(&gorm.DB{Error: nil})
				mockDB.On("Order", "created_at desc, id desc").Return(mockDB)
				mockDB.On("Select", "id, title, description, done, owner, created_at, updated_at").Return(mockDB)
				mockDB.On("Offset", 0).Return(mockDB)
				mockDB.On("Limit", 10).Return(mockDB)
				mockDB.On("Find", mock.Anything).Return(&gorm.DB{Error: assert.AnError})
			},
			wantTasks: nil,
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			gotTasks, gotCount, err := repo.ListByPage(tt.page, tt.limit, tt.order)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTasks, gotTasks)
				assert.Equal(t, tt.wantCount, gotCount)
			}
		})
	}
}

func TestTaskRepositoryImpl_Create(t *testing.T) {
	tests := []struct {
		name          string
		task          *models.Task
		setupMock     func(*MockDB)
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
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Create", mock.AnythingOfType("*models.Task")).Return(mockDB).Once()
				mockDB.On("Error").Return(nil).Once()
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
			name: "database error",
			task: &models.Task{
				Title:       "New Task",
				Description: "New Description",
				Done:        false,
				Owner:       1,
			},
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Create", mock.AnythingOfType("*models.Task")).Return(mockDB).Once()
				mockDB.On("Error").Return(errors.New("database error")).Once()
			},
			expectedTask:  nil,
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockDB{}
			tt.setupMock(mockDB)

			repo := &TaskRepositoryImpl{
				db: mockDB,
			}

			task, err := repo.Create(tt.task)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTask, task)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestTaskRepositoryImpl_Update(t *testing.T) {
	tests := []struct {
		name          string
		taskID        int
		task          *models.Task
		setupMock     func(*MockDB)
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
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Where", "id = ?", 1).Return(mockDB).Once()
				mockDB.On("Save", mock.AnythingOfType("*models.Task")).Return(mockDB).Once()
				mockDB.On("Error").Return(nil).Once()
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
			name:   "database error",
			taskID: 1,
			task: &models.Task{
				Title:       "Updated Task",
				Description: "Updated Description",
				Done:        true,
				Owner:       1,
			},
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Where", "id = ?", 1).Return(mockDB).Once()
				mockDB.On("Save", mock.AnythingOfType("*models.Task")).Return(mockDB).Once()
				mockDB.On("Error").Return(errors.New("database error")).Once()
			},
			expectedTask:  nil,
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockDB{}
			tt.setupMock(mockDB)

			repo := &TaskRepositoryImpl{
				db: mockDB,
			}

			task, err := repo.Update(tt.taskID, tt.task)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTask, task)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestTaskRepositoryImpl_MarkAsDone(t *testing.T) {
	tests := []struct {
		name          string
		taskID        int
		setupMock     func(*MockDB)
		expectedTask  *models.Task
		expectedError error
	}{
		{
			name:   "successful mark as done",
			taskID: 1,
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Where", "id = ?", 1).Return(mockDB).Once()
				mockDB.On("Save", mock.AnythingOfType("*models.Task")).Return(mockDB).Once()
				mockDB.On("Error").Return(nil).Once()
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
			name:   "database error",
			taskID: 1,
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Where", "id = ?", 1).Return(mockDB).Once()
				mockDB.On("Save", mock.AnythingOfType("*models.Task")).Return(mockDB).Once()
				mockDB.On("Error").Return(errors.New("database error")).Once()
			},
			expectedTask:  nil,
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockDB{}
			tt.setupMock(mockDB)

			repo := &TaskRepositoryImpl{
				db: mockDB,
			}

			task, err := repo.MarkAsDone(tt.taskID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTask, task)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestTaskRepositoryImpl_Delete(t *testing.T) {
	tests := []struct {
		name          string
		taskID        int
		setupMock     func(*MockDB)
		expectedError error
	}{
		{
			name:   "successful delete",
			taskID: 1,
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Where", "id = ?", 1).Return(mockDB).Once()
				mockDB.On("Delete", mock.AnythingOfType("*models.Task")).Return(mockDB).Once()
				mockDB.On("Error").Return(nil).Once()
				mockDB.On("RowsAffected").Return(int64(1)).Once()
			},
			expectedError: nil,
		},
		{
			name:   "task not found",
			taskID: 999,
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Where", "id = ?", 999).Return(mockDB).Once()
				mockDB.On("Delete", mock.AnythingOfType("*models.Task")).Return(mockDB).Once()
				mockDB.On("Error").Return(nil).Once()
				mockDB.On("RowsAffected").Return(int64(0)).Once()
			},
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:   "database error",
			taskID: 1,
			setupMock: func(mockDB *MockDB) {
				mockDB.On("Where", "id = ?", 1).Return(mockDB).Once()
				mockDB.On("Delete", mock.AnythingOfType("*models.Task")).Return(mockDB).Once()
				mockDB.On("Error").Return(errors.New("database error")).Once()
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockDB{}
			tt.setupMock(mockDB)

			repo := &TaskRepositoryImpl{
				db: mockDB,
			}

			err := repo.Delete(tt.taskID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
} 