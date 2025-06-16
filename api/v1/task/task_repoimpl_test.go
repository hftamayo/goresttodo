package task

import (
	"testing"
	"time"

	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/cursor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockDB is a mock implementation of *gorm.DB
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Model(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(dest, conds)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Create(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Save(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Delete(value interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(value, conds)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Begin() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Count(count *int64) *gorm.DB {
	args := m.Called(count)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Order(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Select(query interface{}, args ...interface{}) *gorm.DB {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0).(*gorm.DB)
}

func (m *MockDB) Offset(offset int) *gorm.DB {
	args := m.Called(offset)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Limit(limit int) *gorm.DB {
	args := m.Called(limit)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(dest, conds)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0).(*gorm.DB)
}

func (m *MockDB) Updates(values interface{}) *gorm.DB {
	args := m.Called(values)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Commit() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Rollback() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
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
	mockDB := new(MockDB)
	repo := &TaskRepositoryImpl{db: mockDB}

	tests := []struct {
		name    string
		setup   func()
		want    int64
		wantErr bool
	}{
		{
			name: "successful count",
			setup: func() {
				mockDB.On("Model", &models.Task{}).Return(&gorm.DB{})
				mockDB.On("Count", mock.Anything).Return(&gorm.DB{Error: nil})
			},
			want:    10,
			wantErr: false,
		},
		{
			name: "database error",
			setup: func() {
				mockDB.On("Model", &models.Task{}).Return(&gorm.DB{})
				mockDB.On("Count", mock.Anything).Return(&gorm.DB{Error: assert.AnError})
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got, err := repo.GetTotalCount()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestTaskRepositoryImpl_ListById(t *testing.T) {
	mockDB := new(MockDB)
	repo := &TaskRepositoryImpl{db: mockDB}

	validTask := &models.Task{
		ID:          1,
		Title:       "Test Task",
		Description: "Test Description",
		Done:        false,
		Owner:       "test@example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		name    string
		id      int
		setup   func()
		want    *models.Task
		wantErr bool
	}{
		{
			name: "valid task",
			id:   1,
			setup: func() {
				mockDB.On("First", mock.Anything, 1).Return(&gorm.DB{Error: nil})
			},
			want:    validTask,
			wantErr: false,
		},
		{
			name: "task not found",
			id:   999,
			setup: func() {
				mockDB.On("First", mock.Anything, 999).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "database error",
			id:   1,
			setup: func() {
				mockDB.On("First", mock.Anything, 1).Return(&gorm.DB{Error: assert.AnError})
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got, err := repo.ListById(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
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
	mockDB := new(MockDB)
	repo := &TaskRepositoryImpl{db: mockDB}

	validTask := &models.Task{
		Title:       "New Task",
		Description: "New Description",
		Done:        false,
		Owner:       "test@example.com",
	}

	tests := []struct {
		name    string
		task    *models.Task
		setup   func()
		want    *models.Task
		wantErr bool
	}{
		{
			name: "successful create",
			task: validTask,
			setup: func() {
				mockDB.On("Create", validTask).Return(&gorm.DB{Error: nil})
			},
			want:    validTask,
			wantErr: false,
		},
		{
			name:    "nil task",
			task:    nil,
			setup:   func() {},
			want:    nil,
			wantErr: true,
		},
		{
			name: "database error",
			task: validTask,
			setup: func() {
				mockDB.On("Create", validTask).Return(&gorm.DB{Error: assert.AnError})
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got, err := repo.Create(tt.task)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestTaskRepositoryImpl_Update(t *testing.T) {
	mockDB := new(MockDB)
	repo := &TaskRepositoryImpl{db: mockDB}

	existingTask := &models.Task{
		ID:          1,
		Title:       "Existing Task",
		Description: "Existing Description",
		Done:        false,
		Owner:       "test@example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	updatedTask := &models.Task{
		ID:          1,
		Title:       "Updated Task",
		Description: "Updated Description",
		Done:        true,
		Owner:       "test@example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		name    string
		id      int
		task    *models.Task
		setup   func()
		want    *models.Task
		wantErr bool
	}{
		{
			name: "successful update",
			id:   1,
			task: updatedTask,
			setup: func() {
				mockDB.On("First", mock.Anything, 1).Return(&gorm.DB{Error: nil})
				mockDB.On("Begin").Return(&gorm.DB{})
				mockDB.On("Save", updatedTask).Return(&gorm.DB{Error: nil})
				mockDB.On("First", mock.Anything, 1).Return(&gorm.DB{Error: nil})
				mockDB.On("Commit").Return(&gorm.DB{Error: nil})
			},
			want:    updatedTask,
			wantErr: false,
		},
		{
			name:    "nil task",
			id:      1,
			task:    nil,
			setup:   func() {},
			want:    nil,
			wantErr: true,
		},
		{
			name: "task not found",
			id:   999,
			task: updatedTask,
			setup: func() {
				mockDB.On("First", mock.Anything, 999).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "database error during update",
			id:   1,
			task: updatedTask,
			setup: func() {
				mockDB.On("First", mock.Anything, 1).Return(&gorm.DB{Error: nil})
				mockDB.On("Begin").Return(&gorm.DB{})
				mockDB.On("Save", updatedTask).Return(&gorm.DB{Error: assert.AnError})
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got, err := repo.Update(tt.id, tt.task)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestTaskRepositoryImpl_MarkAsDone(t *testing.T) {
	mockDB := new(MockDB)
	repo := &TaskRepositoryImpl{db: mockDB}

	updatedTask := &models.Task{
		ID:          1,
		Title:       "Task",
		Description: "Description",
		Done:        true,
		Owner:       "test@example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		name    string
		id      int
		setup   func()
		want    *models.Task
		wantErr bool
	}{
		{
			name: "successful mark as done",
			id:   1,
			setup: func() {
				mockDB.On("Model", &models.Task{}).Return(&gorm.DB{})
				mockDB.On("Where", "id = ?", 1).Return(&gorm.DB{})
				mockDB.On("Updates", mock.Anything).Return(&gorm.DB{RowsAffected: 1, Error: nil})
				mockDB.On("First", mock.Anything, 1).Return(&gorm.DB{Error: nil})
			},
			want:    updatedTask,
			wantErr: false,
		},
		{
			name: "task not found",
			id:   999,
			setup: func() {
				mockDB.On("Model", &models.Task{}).Return(&gorm.DB{})
				mockDB.On("Where", "id = ?", 999).Return(&gorm.DB{})
				mockDB.On("Updates", mock.Anything).Return(&gorm.DB{RowsAffected: 0, Error: nil})
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "database error",
			id:   1,
			setup: func() {
				mockDB.On("Model", &models.Task{}).Return(&gorm.DB{})
				mockDB.On("Where", "id = ?", 1).Return(&gorm.DB{})
				mockDB.On("Updates", mock.Anything).Return(&gorm.DB{Error: assert.AnError})
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got, err := repo.MarkAsDone(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestTaskRepositoryImpl_Delete(t *testing.T) {
	mockDB := new(MockDB)
	repo := &TaskRepositoryImpl{db: mockDB}

	tests := []struct {
		name    string
		id      int
		setup   func()
		wantErr bool
	}{
		{
			name: "successful delete",
			id:   1,
			setup: func() {
				mockDB.On("Delete", &models.Task{}, 1).Return(&gorm.DB{RowsAffected: 1, Error: nil})
			},
			wantErr: false,
		},
		{
			name: "invalid id",
			id:   0,
			setup: func() {},
			wantErr: true,
		},
		{
			name: "task not found",
			id:   999,
			setup: func() {
				mockDB.On("Delete", &models.Task{}, 999).Return(&gorm.DB{RowsAffected: 0, Error: nil})
			},
			wantErr: true,
		},
		{
			name: "database error",
			id:   1,
			setup: func() {
				mockDB.On("Delete", &models.Task{}, 1).Return(&gorm.DB{Error: assert.AnError})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := repo.Delete(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
} 