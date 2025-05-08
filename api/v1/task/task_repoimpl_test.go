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

func TestTaskRepositoryImpl_List(t *testing.T) {
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
		name      string
		limit     int
		cursorStr string
		setup     func()
		want      []*models.Task
		wantNext  string
		wantErr   bool
	}{
		{
			name:  "successful list with default limit",
			limit: 0,
			setup: func() {
				mockDB.On("Model", &models.Task{}).Return(&gorm.DB{})
				mockDB.On("Order", "created_at DESC, id DESC").Return(&gorm.DB{})
				mockDB.On("Select", "id, title, description, done, owner, created_at, updated_at").Return(&gorm.DB{})
				mockDB.On("Limit", defaultLimit+1).Return(&gorm.DB{})
				mockDB.On("Find", mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*[]*models.Task)
					*arg = tasks[:1]
				}).Return(&gorm.DB{Error: nil})
			},
			want:     tasks[:1],
			wantNext: "",
			wantErr:  false,
		},
		{
			name:  "successful list with custom limit",
			limit: 2,
			setup: func() {
				mockDB.On("Model", &models.Task{}).Return(&gorm.DB{})
				mockDB.On("Order", "created_at DESC, id DESC").Return(&gorm.DB{})
				mockDB.On("Select", "id, title, description, done, owner, created_at, updated_at").Return(&gorm.DB{})
				mockDB.On("Limit", 3).Return(&gorm.DB{})
				mockDB.On("Find", mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*[]*models.Task)
					*arg = tasks
				}).Return(&gorm.DB{Error: nil})
			},
			want:     tasks,
			wantNext: "",
			wantErr:  false,
		},
		{
			name:      "successful list with cursor",
			limit:     1,
			cursorStr: "cursor123",
			setup: func() {
				mockDB.On("Model", &models.Task{}).Return(&gorm.DB{})
				mockDB.On("Order", "created_at DESC, id DESC").Return(&gorm.DB{})
				mockDB.On("Select", "id, title, description, done, owner, created_at, updated_at").Return(&gorm.DB{})
				mockDB.On("Where", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&gorm.DB{})
				mockDB.On("Limit", 2).Return(&gorm.DB{})
				mockDB.On("Find", mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*[]*models.Task)
					*arg = tasks[:1]
				}).Return(&gorm.DB{Error: nil})
			},
			want:     tasks[:1],
			wantNext: "",
			wantErr:  false,
		},
		{
			name:  "database error",
			limit: 1,
			setup: func() {
				mockDB.On("Model", &models.Task{}).Return(&gorm.DB{})
				mockDB.On("Order", "created_at DESC, id DESC").Return(&gorm.DB{})
				mockDB.On("Select", "id, title, description, done, owner, created_at, updated_at").Return(&gorm.DB{})
				mockDB.On("Limit", 2).Return(&gorm.DB{})
				mockDB.On("Find", mock.Anything).Return(&gorm.DB{Error: assert.AnError})
			},
			want:     nil,
			wantNext: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got, next, err := repo.List(tt.limit, tt.cursorStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
				assert.Equal(t, tt.wantNext, next)
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