package config

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrorLogger interface for different implementations
type ErrorLogger interface {
	LogError(operation string, err error) error
	LogInfo(operation string, message string) error
	Close() error
}

// LogEntry represents a log entry
type LogEntry struct {
	Operation string
	Message   string
	Error     error
	Timestamp time.Time
	Level     string // "error", "info", "warn"
}

// RedisErrorLogger implements ErrorLogger using Redis
type RedisErrorLogger struct {
	client *redis.Client
}

// MemoryErrorLogger implements ErrorLogger using in-memory storage (for testing)
type MemoryErrorLogger struct {
	logs []LogEntry
}

// Factory function to create appropriate logger based on config
func NewErrorLogger(loggerType string) (ErrorLogger, error) {
	switch loggerType {
	case "redis":
		return NewRedisErrorLogger()
	case "memory":
		return NewMemoryErrorLogger()
	default:
		return NewMemoryErrorLogger() // Default for testing
	}
}

// NonBlockingErrorLogger wraps any ErrorLogger to make logging non-blocking
type NonBlockingErrorLogger struct {
	logger ErrorLogger
}

func NewNonBlockingErrorLogger(logger ErrorLogger) *NonBlockingErrorLogger {
	return &NonBlockingErrorLogger{logger: logger}
}

func (n *NonBlockingErrorLogger) LogError(operation string, err error) error {
	go func() {
		if err := n.logger.LogError(operation, err); err != nil {
			// Only log if logging itself fails (rare)
			log.Printf("Failed to log error: %v", err)
		}
	}()
	return nil // Always return nil since it's non-blocking
}

func (n *NonBlockingErrorLogger) LogInfo(operation string, message string) error {
	go func() {
		if err := n.logger.LogInfo(operation, message); err != nil {
			log.Printf("Failed to log info: %v", err)
		}
	}()
	return nil
}

func (n *NonBlockingErrorLogger) Close() error {
	return n.logger.Close()
}

// Legacy function for backward compatibility
func ErrorLogConnect() (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	return redisClient, nil
}

func NewRedisErrorLogger() (*RedisErrorLogger, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	
	return &RedisErrorLogger{client: redisClient}, nil
}

func (r *RedisErrorLogger) LogError(operation string, err error) error {
	ctx := context.Background()
	logEntry := map[string]interface{}{
		"operation": operation,
		"error":     err.Error(),
		"timestamp": time.Now().Format(time.RFC3339),
		"level":     "error",
	}
	return r.client.HSet(ctx, "logs", time.Now().UnixNano(), logEntry).Err()
}

func (r *RedisErrorLogger) LogInfo(operation string, message string) error {
	ctx := context.Background()
	logEntry := map[string]interface{}{
		"operation": operation,
		"message":   message,
		"timestamp": time.Now().Format(time.RFC3339),
		"level":     "info",
	}
	return r.client.HSet(ctx, "logs", time.Now().UnixNano(), logEntry).Err()
}

func (r *RedisErrorLogger) Close() error {
	return r.client.Close()
}

func NewMemoryErrorLogger() (*MemoryErrorLogger, error) {
	return &MemoryErrorLogger{
		logs: make([]LogEntry, 0),
	}, nil
}

func (m *MemoryErrorLogger) LogError(operation string, err error) error {
	entry := LogEntry{
		Operation: operation,
		Error:     err,
		Timestamp: time.Now(),
		Level:     "error",
	}
	m.logs = append(m.logs, entry)
	return nil
}

func (m *MemoryErrorLogger) LogInfo(operation string, message string) error {
	entry := LogEntry{
		Operation: operation,
		Message:   message,
		Timestamp: time.Now(),
		Level:     "info",
	}
	m.logs = append(m.logs, entry)
	return nil
}

func (m *MemoryErrorLogger) Close() error {
	// Clear logs for memory cleanup
	m.logs = nil
	return nil
}

// GetLogs returns all stored logs (useful for testing)
func (m *MemoryErrorLogger) GetLogs() []LogEntry {
	return m.logs
}
