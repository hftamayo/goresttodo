package cursor

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Cursor represents a generic pagination cursor that can work with any domain
type Cursor[T any] struct {
    ID        T         `json:"id"`
    Timestamp time.Time `json:"timestamp"`
    Extra     string    `json:"extra,omitempty"` // Optional field for additional sorting criteria
}

// Options provides configuration for cursor encoding/decoding
type Options struct {
    Field     string // Field to use for timestamp (e.g., "created_at", "updated_at")
    Direction string // Sort direction ("ASC" or "DESC")
}

// ValidateOptions checks if the provided options are valid
func ValidateOptions(opts Options) error {
    if opts.Field == "" {
        return fmt.Errorf("field option is required")
    }
    if opts.Direction != "ASC" && opts.Direction != "DESC" {
        return fmt.Errorf("direction must be either ASC or DESC")
    }
    return nil
}

// Encode converts a cursor to a base64 string
func Encode[T any](c Cursor[T], opts Options) (string, error) {
    if err := ValidateOptions(opts); err != nil {
        return "", err
    }    
    // Convert ID to string based on type
    var idStr string
    switch v := any(c.ID).(type) {
    case uint:
        idStr = strconv.FormatUint(uint64(v), 10)
    case int:
        idStr = strconv.Itoa(v)
    case string:
        idStr = v
    default:
        return "", fmt.Errorf("unsupported ID type: %T", c.ID)
    }

    str := fmt.Sprintf("%s:%d:%s", idStr, c.Timestamp.Unix(), c.Extra)
    return base64.StdEncoding.EncodeToString([]byte(str)), nil
}

// Decode converts a base64 string back to a cursor
func Decode[T any](str string) (Cursor[T], error) {
    if str == "" {
        return Cursor[T]{}, nil
    }    
    bytes, err := base64.StdEncoding.DecodeString(str)
    if err != nil {
        return Cursor[T]{}, fmt.Errorf("failed to decode cursor: %w", err)
    }

    parts := strings.Split(string(bytes), ":")
    if len(parts) < 2 {
        return Cursor[T]{}, fmt.Errorf("invalid cursor format")
    }

    // Parse timestamp
    timestamp, err := strconv.ParseInt(parts[1], 10, 64)
    if err != nil {
        return Cursor[T]{}, fmt.Errorf("invalid cursor timestamp: %w", err)
    }

    // Parse ID based on type
    var id T
    switch any(id).(type) {
    case uint:
        parsed, err := strconv.ParseUint(parts[0], 10, 64)
        if err != nil {
            return Cursor[T]{}, fmt.Errorf("invalid cursor ID: %w", err)
        }
        id = any(uint(parsed)).(T)
    case int:
        parsed, err := strconv.Atoi(parts[0])
        if err != nil {
            return Cursor[T]{}, fmt.Errorf("invalid cursor ID: %w", err)
        }
        id = any(parsed).(T)
    case string:
        id = any(parts[0]).(T)
    default:
        return Cursor[T]{}, fmt.Errorf("unsupported ID type: %T", id)
    }

    cursor := Cursor[T]{
        ID:        id,
        Timestamp: time.Unix(timestamp, 0),
    }

    // Add extra data if provided
    if len(parts) > 2 {
        cursor.Extra = parts[2]
    }

    return cursor, nil
}

// NewCursor creates a new cursor instance with validation
func NewCursor[T any](id T, timestamp time.Time, extra string) Cursor[T] {
    return Cursor[T]{
        ID:        id,
        Timestamp: timestamp,
        Extra:     extra,
    }
}

// IsEmpty checks if a cursor is empty/initial
func (c Cursor[T]) IsEmpty() bool {
    return c.Timestamp.IsZero()
}