package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/hftamayo/gotodo/pkg/utils"
)

// ErrorLogger defines the interface for error logging operations
type ErrorLogger interface {
	LogError(ctx context.Context, service, operation, errorMsg string, metadata map[string]interface{}) error
	Close() error
}

// RedisErrorLogger implements ErrorLogger using Redis
type RedisErrorLogger struct {
	client utils.RedisClientInterface
}

// NewRedisErrorLogger creates a new Redis-based error logger
func NewRedisErrorLogger() (*RedisErrorLogger, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisErrorLogger{client: redisClient}, nil
}

// LogError logs an error to Redis
func (r *RedisErrorLogger) LogError(ctx context.Context, service, operation, errorMsg string, metadata map[string]interface{}) error {
	errorLog := map[string]interface{}{
		"service":   service,
		"operation": operation,
		"error":     errorMsg,
		"timestamp": time.Now().Unix(),
		"metadata":  metadata,
	}

	key := fmt.Sprintf("errorlog:%s:%d", service, time.Now().Unix())
	return r.client.HMSet(ctx, key, errorLog).Err()
}

// Close closes the Redis connection
func (r *RedisErrorLogger) Close() error {
	return r.client.Close()
}

// MemoryErrorLogger implements ErrorLogger using in-memory storage
type MemoryErrorLogger struct {
	errors []map[string]interface{}
	mu     sync.RWMutex
}

// NewMemoryErrorLogger creates a new memory-based error logger
func NewMemoryErrorLogger() *MemoryErrorLogger {
	return &MemoryErrorLogger{
		errors: make([]map[string]interface{}, 0),
	}
}

// LogError logs an error to memory
func (m *MemoryErrorLogger) LogError(ctx context.Context, service, operation, errorMsg string, metadata map[string]interface{}) error {
	errorLog := map[string]interface{}{
		"service":   service,
		"operation": operation,
		"error":     errorMsg,
		"timestamp": time.Now().Unix(),
		"metadata":  metadata,
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors = append(m.errors, errorLog)
	return nil
}

// GetErrors returns all logged errors (for testing/debugging)
func (m *MemoryErrorLogger) GetErrors() []map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	errors := make([]map[string]interface{}, len(m.errors))
	copy(errors, m.errors)
	return errors
}

// Close closes the memory logger (no-op)
func (m *MemoryErrorLogger) Close() error {
	return nil
}

// NonBlockingErrorLogger wraps an ErrorLogger to make operations non-blocking
type NonBlockingErrorLogger struct {
	logger ErrorLogger
}

// NewNonBlockingErrorLogger creates a new non-blocking error logger
func NewNonBlockingErrorLogger(logger ErrorLogger) *NonBlockingErrorLogger {
	return &NonBlockingErrorLogger{logger: logger}
}

// LogError logs an error without blocking the main application flow
func (n *NonBlockingErrorLogger) LogError(ctx context.Context, service, operation, errorMsg string, metadata map[string]interface{}) error {
	go func() {
		if err := n.logger.LogError(ctx, service, operation, errorMsg, metadata); err != nil {
			// Log the logging error to stderr to avoid infinite loops
			log.Printf("Failed to log error: %v", err)
		}
	}()
	return nil
}

// Close closes the underlying logger
func (n *NonBlockingErrorLogger) Close() error {
	return n.logger.Close()
}

// NewErrorLogger creates an error logger based on the specified type
// This is a factory function to simplify logger creation
func NewErrorLogger(loggerType string) (ErrorLogger, error) {
	switch loggerType {
	case "redis":
		return NewRedisErrorLogger()
	case "memory":
		return NewMemoryErrorLogger(), nil
	case "nonblocking":
		memoryLogger := NewMemoryErrorLogger()
		return NewNonBlockingErrorLogger(memoryLogger), nil
	default:
		// Default to memory logger
		return NewMemoryErrorLogger(), nil
	}
}

// NewErrorLoggerWithDefaults creates a non-blocking memory logger by default
// This is useful for testing and development
func NewErrorLoggerWithDefaults() ErrorLogger {
	memoryLogger := NewMemoryErrorLogger()
	return NewNonBlockingErrorLogger(memoryLogger)
}

// ErrorLogConnect creates a Redis client (legacy function for backward compatibility)
func ErrorLogConnect() (utils.RedisClientInterface, error) {
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
