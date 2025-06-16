package task

import (
	"testing"
	"time"

	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTaskRepository is a mock implementation of TaskRepository
type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) List(limit int, cursor string, order string) ([]*models.Task, string, string, error) {
	args := m.Called(limit, cursor, order)
	return args.Get(0).([]*models.Task), args.String(1), args.String(2), args.Error(3)
}

func (m *MockTaskRepository) ListById(id int) (*models.Task, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskRepository) SearchByTitle(title string) (*models.Task, error) {
	args := m.Called(title)
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskRepository) Create(task *models.Task) (*models.Task, error) {
	args := m.Called(task)
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskRepository) Update(id int, task *models.Task) (*models.Task, error) {
	args := m.Called(id, task)
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskRepository) MarkAsDone(id int) (*models.Task, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTaskRepository) GetTotalCount() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTaskRepository) ListByPage(page int, limit int, order string) ([]*models.Task, int64, error) {
	args := m.Called(page, limit, order)
	return args.Get(0).([]*models.Task), args.Get(1).(int64), args.Error(2)
}

// MockCache is a mock implementation of utils.Cache
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(key string, value interface{}) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *MockCache) Set(key string, value interface{}, expiration time.Duration) error {
	args := m.Called(key, value, expiration)
	return args.Error(0)
}

func (m *MockCache) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockCache) SetWithTags(key string, value interface{}, expiration time.Duration, tags ...string) error {
	args := m.Called(key, value, expiration, tags)
	return args.Error(0)
}

func (m *MockCache) InvalidateByTags(tags ...string) error {
	args := m.Called(tags)
	return args.Error(0)
}

func TestNewTaskService(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	mockCache := new(MockCache)

	service := NewTaskService(mockRepo, mockCache)
	assert.NotNil(t, service)
}

func TestTaskService_List(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	mockCache := new(MockCache)
	service := NewTaskService(mockRepo, mockCache)

	now := time.Now()
	tasks := []*models.Task{
		{
			ID:          1,
			Title:       "Task 1",
			Description: "Description 1",
			Done:        false,
			Owner:       1,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          2,
			Title:       "Task 2",
			Description: "Description 2",
			Done:        true,
			Owner:       1,
			CreatedAt:   now.Add(-time.Hour),
			UpdatedAt:   now.Add(-time.Hour),
		},
	}

	tests := []struct {
		name       string
		cursor     string
		limit      int
		setupCache func()
		setupRepo  func()
		wantTasks  []*models.Task
		wantCursor string
		wantCount  int64
		wantErr    bool
	}{
		{
			name:   "successful list from cache",
			cursor: "",
			limit:  10,
			setupCache: func() {
				mockCache.On("Get", "tasks_cursor__limit_10", mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(1).(*struct {
						Tasks      []*models.Task
						Pagination CursorPaginationMeta
						TotalCount int64
					})
					arg.Tasks = tasks
					arg.Pagination = CursorPaginationMeta{NextCursor: "", HasMore: false, Count: 2}
					arg.TotalCount = 2
				}).Return(nil)
			},
			setupRepo:  func() {},
			wantTasks:  tasks,
			wantCursor: "",
			wantCount:  2,
			wantErr:    false,
		},
		{
			name:   "successful list from repository",
			cursor: "",
			limit:  10,
			setupCache: func() {
				mockCache.On("Get", "tasks_cursor__limit_10", mock.Anything).Return(assert.AnError)
				mockCache.On("Set", "tasks_cursor__limit_10", mock.Anything, defaultCacheTime).Return(nil)
			},
			setupRepo: func() {
				mockRepo.On("List", 10, "", "").Return(tasks, "", "", nil)
				mockRepo.On("GetTotalCount").Return(int64(2), nil)
			},
			wantTasks:  tasks,
			wantCursor: "",
			wantCount:  2,
			wantErr:    false,
		},
		{
			name:   "repository error",
			cursor: "",
			limit:  10,
			setupCache: func() {
				mockCache.On("Get", "tasks_cursor__limit_10", mock.Anything).Return(assert.AnError)
			},
			setupRepo: func() {
				mockRepo.On("List", 10, "", "").Return(nil, "", "", assert.AnError)
			},
			wantTasks:  nil,
			wantCursor: "",
			wantCount:  0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupCache()
			tt.setupRepo()
			gotTasks, gotCursor, gotCount, err := service.List(tt.cursor, tt.limit, "")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTasks, gotTasks)
				assert.Equal(t, tt.wantCursor, gotCursor)
				assert.Equal(t, tt.wantCount, gotCount)
			}
		})
	}
}

func TestTaskService_ListById(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	mockCache := new(MockCache)
	service := NewTaskService(mockRepo, mockCache)

	validTask := &models.Task{
		ID:          1,
		Title:       "Test Task",
		Description: "Test Description",
		Done:        false,
		Owner:       1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		name       string
		id         int
		setupCache func()
		setupRepo  func()
		want       *models.Task
		wantErr    bool
	}{
		{
			name: "successful get from cache",
			id:   1,
			setupCache: func() {
				mockCache.On("Get", "task_1", mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(1).(**models.Task)
					*arg = validTask
				}).Return(nil)
			},
			setupRepo: func() {},
			want:      validTask,
			wantErr:   false,
		},
		{
			name: "successful get from repository",
			id:   1,
			setupCache: func() {
				mockCache.On("Get", "task_1", mock.Anything).Return(assert.AnError)
				mockCache.On("Set", "task_1", validTask, 10*time.Minute).Return(nil)
			},
			setupRepo: func() {
				mockRepo.On("ListById", 1).Return(validTask, nil)
			},
			want:    validTask,
			wantErr: false,
		},
		{
			name: "task not found",
			id:   999,
			setupCache: func() {
				mockCache.On("Get", "task_999", mock.Anything).Return(assert.AnError)
			},
			setupRepo: func() {
				mockRepo.On("ListById", 999).Return(nil, nil)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupCache()
			tt.setupRepo()
			got, err := service.ListById(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestTaskService_Create(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	mockCache := new(MockCache)
	service := NewTaskService(mockRepo, mockCache)

	validTask := &models.Task{
		Title:       "New Task",
		Description: "New Description",
		Done:        false,
		Owner:       1,
	}

	tests := []struct {
		name       string
		task       *models.Task
		setupCache func()
		setupRepo  func()
		want       *models.Task
		wantErr    bool
	}{
		{
			name: "successful create",
			task: validTask,
			setupCache: func() {
				mockCache.On("Delete", "tasks_list*").Return(nil)
			},
			setupRepo: func() {
				mockRepo.On("Create", validTask).Return(validTask, nil)
			},
			want:    validTask,
			wantErr: false,
		},
		{
			name:       "nil task",
			task:       nil,
			setupCache: func() {},
			setupRepo:  func() {},
			want:       nil,
			wantErr:    true,
		},
		{
			name: "repository error",
			task: validTask,
			setupCache: func() {},
			setupRepo: func() {
				mockRepo.On("Create", validTask).Return(nil, assert.AnError)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupCache()
			tt.setupRepo()
			got, err := service.Create(tt.task)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestTaskService_Update(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	mockCache := new(MockCache)
	service := NewTaskService(mockRepo, mockCache)

	existingTask := &models.Task{
		ID:          1,
		Title:       "Existing Task",
		Description: "Existing Description",
		Done:        false,
		Owner:       1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	updatedTask := &models.Task{
		ID:          1,
		Title:       "Updated Task",
		Description: "Updated Description",
		Done:        true,
		Owner:       1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		name       string
		id         int
		task       *models.Task
		setupCache func()
		setupRepo  func()
		want       *models.Task
		wantErr    bool
	}{
		{
			name: "successful update",
			id:   1,
			task: updatedTask,
			setupCache: func() {
				mockCache.On("Delete", "task_1").Return(nil)
				mockCache.On("Delete", "tasks_list*").Return(nil)
			},
			setupRepo: func() {
				mockRepo.On("ListById", 1).Return(existingTask, nil)
				mockRepo.On("Update", 1, updatedTask).Return(updatedTask, nil)
			},
			want:    updatedTask,
			wantErr: false,
		},
		{
			name:       "nil task",
			id:         1,
			task:       nil,
			setupCache: func() {},
			setupRepo:  func() {},
			want:       nil,
			wantErr:    true,
		},
		{
			name: "inconsistent task ID",
			id:   2,
			task: updatedTask,
			setupCache: func() {},
			setupRepo:  func() {},
			want:       nil,
			wantErr:    true,
		},
		{
			name: "task not found",
			id:   999,
			task: updatedTask,
			setupCache: func() {},
			setupRepo: func() {
				mockRepo.On("ListById", 999).Return(nil, nil)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupCache()
			tt.setupRepo()
			got, err := service.Update(tt.id, tt.task)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestTaskService_MarkAsDone(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	mockCache := new(MockCache)
	service := NewTaskService(mockRepo, mockCache)

	existingTask := &models.Task{
		ID:          1,
		Title:       "Task",
		Description: "Description",
		Done:        false,
		Owner:       1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	updatedTask := &models.Task{
		ID:          1,
		Title:       "Task",
		Description: "Description",
		Done:        true,
		Owner:       1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		name       string
		id         int
		setupCache func()
		setupRepo  func()
		want       *models.Task
		wantErr    bool
	}{
		{
			name: "successful mark as done",
			id:   1,
			setupCache: func() {
				mockCache.On("Delete", "task_1").Return(nil)
				mockCache.On("Delete", "tasks_list*").Return(nil)
			},
			setupRepo: func() {
				mockRepo.On("ListById", 1).Return(existingTask, nil)
				mockRepo.On("MarkAsDone", 1).Return(updatedTask, nil)
			},
			want:    updatedTask,
			wantErr: false,
		},
		{
			name: "task not found",
			id:   999,
			setupCache: func() {},
			setupRepo: func() {
				mockRepo.On("ListById", 999).Return(nil, nil)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "repository error",
			id:   1,
			setupCache: func() {},
			setupRepo: func() {
				mockRepo.On("ListById", 1).Return(existingTask, nil)
				mockRepo.On("MarkAsDone", 1).Return(nil, assert.AnError)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupCache()
			tt.setupRepo()
			got, err := service.MarkAsDone(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestTaskService_Delete(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	mockCache := new(MockCache)
	service := NewTaskService(mockRepo, mockCache)

	tests := []struct {
		name       string
		id         int
		setupCache func()
		setupRepo  func()
		wantErr    bool
	}{
		{
			name: "successful delete",
			id:   1,
			setupCache: func() {
				mockCache.On("Delete", "task_1").Return(nil)
				mockCache.On("Delete", "tasks_list*").Return(nil)
			},
			setupRepo: func() {
				mockRepo.On("Delete", 1).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repository error",
			id:   1,
			setupCache: func() {},
			setupRepo: func() {
				mockRepo.On("Delete", 1).Return(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupCache()
			tt.setupRepo()
			err := service.Delete(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskService_ListByPage(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	mockCache := new(MockCache)
	service := NewTaskService(mockRepo, mockCache)

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
		setupCache func()
		setupRepo  func()
		wantTasks  []*models.Task
		wantCount  int64
		wantErr    bool
	}{
		{
			name:  "successful list with default values",
			page:  1,
			limit: 0,
			order: "",
			setupCache: func() {
				mockCache.On("Get", "tasks_page_1_limit_10_order_desc", mock.Anything).Return(assert.AnError)
				mockCache.On("Set", "tasks_page_1_limit_10_order_desc", mock.Anything, utils.DefaultCacheTime).Return(nil)
			},
			setupRepo: func() {
				mockRepo.On("ListByPage", 1, utils.DefaultLimit, utils.DefaultOrder).Return(tasks, int64(2), nil)
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
			setupCache: func() {
				mockCache.On("Get", "tasks_page_2_limit_1_order_asc", mock.Anything).Return(assert.AnError)
				mockCache.On("Set", "tasks_page_2_limit_1_order_asc", mock.Anything, utils.DefaultCacheTime).Return(nil)
			},
			setupRepo: func() {
				mockRepo.On("ListByPage", 2, 1, "asc").Return(tasks[1:2], int64(2), nil)
			},
			wantTasks: tasks[1:2],
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:  "repository error",
			page:  1,
			limit: 10,
			order: "desc",
			setupCache: func() {
				mockCache.On("Get", "tasks_page_1_limit_10_order_desc", mock.Anything).Return(assert.AnError)
			},
			setupRepo: func() {
				mockRepo.On("ListByPage", 1, 10, "desc").Return(nil, int64(0), assert.AnError)
			},
			wantTasks: nil,
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupCache()
			tt.setupRepo()
			gotTasks, gotCount, err := service.ListByPage(tt.page, tt.limit, tt.order)
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