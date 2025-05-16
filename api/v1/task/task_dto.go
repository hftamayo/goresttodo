package task

import (
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

type PaginationMeta struct {
    NextCursor  string `json:"nextCursor"`
    PrevCursor  string `json:"prevCursor,omitempty"`
    Limit       int    `json:"limit"`
    TotalCount  int64  `json:"totalCount"`
    HasMore     bool   `json:"hasMore"`
    CurrentPage int    `json:"currentPage"`
    TotalPages  int    `json:"totalPages"`
    Order       string `json:"order"`
}

type TaskListResponse struct {
    Tasks      []*TaskResponse `json:"tasks"`
    Pagination PaginationMeta  `json:"pagination"`
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
}

type ErrorResponse struct {
    Code          int    `json:"code"`
    ResultMessage string `json:"resultMessage"`
    Error         string `json:"error,omitempty"`
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
    }
}

// buildListResponse creates a paginated list response
func buildListResponse(tasks []*models.Task, query CursorPaginationQuery, nextCursor, prevCursor string, totalCount int64) TaskOperationResponse {
    currentPage := 1
    if query.Cursor != "" {
        currentPage = int(totalCount/int64(query.Limit)) + 1
    }
    totalPages := int(math.Ceil(float64(totalCount) / float64(query.Limit)))

    listResponse := TaskListResponse{
        Tasks: TasksToResponse(tasks),
        Pagination: PaginationMeta{
            NextCursor:  nextCursor,
            PrevCursor:  prevCursor,
            Limit:       query.Limit,
            TotalCount:  totalCount,
            HasMore:     nextCursor != "",
            CurrentPage: currentPage,
            TotalPages:  totalPages,
            Order:       query.Order,
        },
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