package errorlog

type ErrorLogRepository interface {
	LogError(operation string, err error) error
}
