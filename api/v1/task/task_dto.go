package task

import (
	"crypto/sha256"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/utils"
)

const (
    DefaultOrder = "desc"
    DefaultLimit = 10
    MaxLimit    = 100
)

type CreateTaskRequest struct {
    Title       string `json:"title" binding:"required"`
    Description string `json:"description"`
    Owner       uint   `json:"owner" binding:"required"`
}

type UpdateTaskRequest struct {
    Title       string `json:"title" binding:"required"`
    Description string `json:"description" binding:"required"`
}

type CursorPaginationQuery struct {
    Cursor string `form:"cursor"`
    Limit  int    `form:"limit" binding:"omitempty,gt=0"`
    Order  string `form:"order" binding:"omitempty,oneof=asc desc"`
}

type PagePaginationQuery struct {
    Page  int    `form:"page" binding:"omitempty,gt=0"`
    Limit int    `form:"limit" binding:"omitempty,gt=0"`
    Order string `form:"order" binding:"omitempty,oneof=asc desc"`
}

type PaginationMeta struct {
    NextCursor  string `json:"nextCursor"`
    PrevCursor  string `json:"prevCursor,omitempty"`
    Limit       int    `json:"limit"`
    TotalCount  int64  `json:"totalCount"`
    HasMore     bool   `json:"hasMore"`
    CurrentPage int    `json:"currentPage"`
    TotalPages  int    `json:"totalPages"`
    Order       string `json:"order"`
    HasPrev     bool   `json:"hasPrev"`              // Add hasPrev flag
    IsFirstPage bool   `json:"isFirstPage"`          // Explicit first page flag
    IsLastPage  bool   `json:"isLastPage"`           // Explicit last page flag    
}

type TaskListResponse struct {
    Tasks      []*TaskResponse `json:"tasks"`
    Pagination PaginationMeta  `json:"pagination"`
    ETag       string          `json:"etag,omitempty"`    // Add ETag for client caching
    LastModified string        `json:"lastModified,omitempty"` // Last-Modified header value    
}

type TaskResponse struct {
    ID          uint      `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Done        bool     `json:"done"`
    Owner       uint     `json:"owner"`
    CreatedAt   time.Time `json:"createdAt" binding:"required"`
    UpdatedAt   time.Time `json:"updatedAt" binding:"required"`    
}

type TaskOperationResponse struct {
    Code          int           `json:"code" binding:"required"`
    ResultMessage string        `json:"resultMessage" binding:"required"`
    Data          interface{}   `json:"data,omitempty"`
    Timestamp     int64         `json:"timestamp"`      // Add timestamp for cache invalidation
    CacheTTL      int           `json:"cacheTTL,omitempty"` // Time in seconds the response can be cached    
}

type ErrorResponse struct {
    Code          int    `json:"code"`
    ResultMessage string `json:"resultMessage"`
    Error         string `json:"error,omitempty"`
}

type TaskOperationStatus struct {
    IsLoading       bool   `json:"isLoading,omitempty"`
    LastUpdated     int64  `json:"lastUpdated,omitempty"`
    RefreshRequired bool   `json:"refreshRequired,omitempty"`
    Message         string `json:"message,omitempty"`
}

func ToTaskResponse(task *models.Task) *TaskResponse {
    return &TaskResponse{
        ID:          task.ID,
        Title:       task.Title,
        Description: task.Description,
        Done:        task.Done,
        Owner:       task.Owner,
        CreatedAt:   task.CreatedAt,
        UpdatedAt:   task.UpdatedAt,
    }
}

func TasksToResponse(tasks []*models.Task) []*TaskResponse {
    taskResponses := make([]*TaskResponse, len(tasks))
    for i, task := range tasks {
        taskResponses[i] = ToTaskResponse(task)
    }
    return taskResponses
}

// NewTaskOperationResponse creates a new TaskOperationResponse with success status
func NewTaskOperationResponse(data interface{}) TaskOperationResponse {
    return TaskOperationResponse{
        Code:          http.StatusOK,
        ResultMessage: utils.OperationSuccess,
        Data:         data,
        Timestamp:     time.Now().Unix(),
        CacheTTL:      60, // Cache for 60 seconds on client        
    }
}

// buildListResponse creates a paginated list response
func buildListResponse(tasks []*models.Task, query CursorPaginationQuery, nextCursor, prevCursor string, totalCount int64) TaskOperationResponse {
    // For cursor-based pagination, we don't need to calculate current page
    // as it's not relevant to the user
    currentPage := 1
    totalPages := int(math.Ceil(float64(totalCount) / float64(query.Limit)))

    isFirstPage := prevCursor == ""
    isLastPage := nextCursor == ""
    hasPrev := prevCursor != ""    

    // Calculate if there are more records
    hasMore := nextCursor != ""

    listResponse := TaskListResponse{
        Tasks: TasksToResponse(tasks),
        Pagination: PaginationMeta{
            NextCursor:  nextCursor,
            PrevCursor:  prevCursor,
            Limit:       query.Limit,
            TotalCount:  totalCount,
            HasMore:     hasMore,
            CurrentPage: currentPage,
            TotalPages:  totalPages,
            Order:       query.Order,
            HasPrev:     hasPrev,              // Add hasPrev flag
            IsFirstPage: isFirstPage,
            IsLastPage:  isLastPage,            
        },
        ETag:          generateETag(tasks),  // Generate unique identifier based on content
        LastModified:  time.Now().UTC().Format(http.TimeFormat),        
    }

    return NewTaskOperationResponse(listResponse)
}

func NewErrorResponse(code int, resultMessage string, err string) *ErrorResponse {
    return &ErrorResponse{
        Code:          code,
        ResultMessage: resultMessage,
        Error:         err,
    }
}

func generateETag(tasks []*models.Task) string {
    // Simple implementation - in production you might want something more sophisticated
    hash := sha256.New()
    for _, task := range tasks {
        hash.Write([]byte(fmt.Sprintf("%d-%s-%t-%d", 
            task.ID, task.Title, task.Done, task.UpdatedAt.UnixNano())))
    }
    return fmt.Sprintf("\"%x\"", hash.Sum(nil))
}