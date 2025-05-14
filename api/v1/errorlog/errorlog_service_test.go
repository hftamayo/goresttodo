package errorlog

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockErrorLogRepository is a mock implementation of ErrorLogRepository
type MockErrorLogRepository struct {
	mock.Mock
}

func (m *MockErrorLogRepository) LogError(operation string, err error) error {
	args := m.Called(operation, err)
	return args.Error(0)
}

func TestErrorLogService_LogError(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		err       error
		setupMock func(*MockErrorLogRepository)
		expectErr bool
	}{
		{
			name:      "success",
			operation: "test_op",
			err:       errors.New("test error"),
			setupMock: func(m *MockErrorLogRepository) {
				m.On("LogError", "test_op", errors.New("test error")).Return(nil)
			},
			expectErr: false,
		},
		{
			name:      "repository error",
			operation: "test_op",
			err:       errors.New("test error"),
			setupMock: func(m *MockErrorLogRepository) {
				m.On("LogError", "test_op", errors.New("test error")).Return(errors.New("repository error"))
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockErrorLogRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}
			service := NewErrorLogService(mockRepo)
			err := service.LogError(tt.operation, tt.err)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
} 