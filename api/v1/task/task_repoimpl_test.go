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
	// This test would require a real database connection
	// For unit testing, we would typically use a test database or mock the gorm.DB
	t.Skip("Skipping test that requires database connection")
}

func TestTaskRepositoryImpl_ListById(t *testing.T) {
	// This test would require a real database connection
	// For unit testing, we would typically use a test database or mock the gorm.DB
	t.Skip("Skipping test that requires database connection")
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
	// This test would require a real database connection
	// For unit testing, we would typically use a test database or mock the gorm.DB
	t.Skip("Skipping test that requires database connection")
}

func TestTaskRepositoryImpl_Update(t *testing.T) {
	// This test would require a real database connection
	// For unit testing, we would typically use a test database or mock the gorm.DB
	t.Skip("Skipping test that requires database connection")
}

func TestTaskRepositoryImpl_MarkAsDone(t *testing.T) {
	// This test would require a real database connection
	// For unit testing, we would typically use a test database or mock the gorm.DB
	t.Skip("Skipping test that requires database connection")
}

func TestTaskRepositoryImpl_Delete(t *testing.T) {
	// This test would require a real database connection
	// For unit testing, we would typically use a test database or mock the gorm.DB
	t.Skip("Skipping test that requires database connection")
}

// Test helper functions that can be tested without database
func TestValidateListParams(t *testing.T) {
	tests := []struct {
		name      string
		limit     int
		order     string
		expectErr bool
	}{
		{
			name:      "valid parameters",
			limit:     10,
			order:     "desc",
			expectErr: false,
		},
		{
			name:      "invalid limit - zero",
			limit:     0,
			order:     "desc",
			expectErr: true,
		},
		{
			name:      "invalid limit - negative",
			limit:     -1,
			order:     "desc",
			expectErr: true,
		},
		{
			name:      "invalid order - empty",
			limit:     10,
			order:     "",
			expectErr: true,
		},
		{
			name:      "invalid order - invalid value",
			limit:     10,
			order:     "invalid",
			expectErr: true,
		},
		{
			name:      "valid order - asc",
			limit:     10,
			order:     "asc",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &TaskRepositoryImpl{}
			err := repo.validateListParams(tt.limit, tt.order)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseCursor(t *testing.T) {
	tests := []struct {
		name        string
		cursorStr   string
		expectTime  bool
		expectError bool
	}{
		{
			name:        "empty cursor",
			cursorStr:   "",
			expectTime:  false,
			expectError: false,
		},
		{
			name:        "valid cursor",
			cursorStr:   "2023-01-01T00:00:00Z",
			expectTime:  true,
			expectError: false,
		},
		{
			name:        "invalid cursor format",
			cursorStr:   "invalid-date",
			expectTime:  false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &TaskRepositoryImpl{}
			time, err := repo.parseCursor(tt.cursorStr)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expectTime {
					assert.False(t, time.IsZero())
				} else {
					assert.True(t, time.IsZero())
				}
			}
		})
	}
}

// Test data structures
func TestTaskModelStructure(t *testing.T) {
	// Test that the Task model has the expected structure
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

	assert.Equal(t, uint(1), task.ID)
	assert.Equal(t, "Test Task", task.Title)
	assert.Equal(t, "Test Description", task.Description)
	assert.False(t, task.Done)
	assert.Equal(t, uint(1), task.Owner)
	assert.False(t, task.CreatedAt.IsZero())
	assert.False(t, task.UpdatedAt.IsZero())
}

func TestTaskRepositoryInterface(t *testing.T) {
	// Test that TaskRepositoryImpl implements TaskRepository interface
	var _ TaskRepository = (*TaskRepositoryImpl)(nil)
}

// Mock for testing purposes (if needed in the future)
type MockTaskRepository struct {
	GetTotalCountFunc func() (int64, error)
	ListByIdFunc      func(id int) (*models.Task, error)
	CreateFunc        func(task *models.Task) (*models.Task, error)
	UpdateFunc        func(id int, task *models.Task) (*models.Task, error)
	MarkAsDoneFunc    func(id int) (*models.Task, error)
	DeleteFunc        func(id int) error
}

func (m *MockTaskRepository) GetTotalCount() (int64, error) {
	if m.GetTotalCountFunc != nil {
		return m.GetTotalCountFunc()
	}
	return 0, errors.New("not implemented")
}

func (m *MockTaskRepository) ListById(id int) (*models.Task, error) {
	if m.ListByIdFunc != nil {
		return m.ListByIdFunc(id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockTaskRepository) Create(task *models.Task) (*models.Task, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(task)
	}
	return nil, errors.New("not implemented")
}

func (m *MockTaskRepository) Update(id int, task *models.Task) (*models.Task, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(id, task)
	}
	return nil, errors.New("not implemented")
}

func (m *MockTaskRepository) MarkAsDone(id int) (*models.Task, error) {
	if m.MarkAsDoneFunc != nil {
		return m.MarkAsDoneFunc(id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockTaskRepository) Delete(id int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return errors.New("not implemented")
}

func TestMockTaskRepository(t *testing.T) {
	mockRepo := &MockTaskRepository{
		GetTotalCountFunc: func() (int64, error) {
			return 10, nil
		},
		ListByIdFunc: func(id int) (*models.Task, error) {
			return &models.Task{
				Model: gorm.Model{ID: uint(id)},
				Title: "Test Task",
			}, nil
		},
	}

	count, err := mockRepo.GetTotalCount()
	assert.NoError(t, err)
	assert.Equal(t, int64(10), count)

	task, err := mockRepo.ListById(1)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), task.ID)
	assert.Equal(t, "Test Task", task.Title)
} 