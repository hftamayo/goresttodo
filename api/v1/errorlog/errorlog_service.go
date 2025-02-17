package errorlog

type ErrorLogService struct {
	repo ErrorLogRepository
}

func NewErrorLogService(repo ErrorLogRepository) *ErrorLogService {
	return &ErrorLogService{repo}
}

func (s *ErrorLogService) LogError(operation string, err error) error {
	return s.repo.LogError(operation, err)
}
