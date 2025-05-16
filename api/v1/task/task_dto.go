package task

import (
	"time"

	"github.com/hftamayo/gotodo/api/v1/models"
)

type CreateTaskRequest struct {
    Title       string `json:"title" binding:"required"`
    Owner       uint   `json:"owner" binding:"required"`
}

type UpdateTaskRequest struct {
    Title       string `json:"title" binding:"required"`
    Description string `json:"description" binding:"required"`
}

type CursorPaginationQuery struct {
    Cursor string `form:"cursor"`
    Limit  int    `form:"limit"`
    Order  string `form:"order" binding:"omitempty,oneof=asc desc"` // Add ordering option
}

type PaginationMeta struct {
    NextCursor  string `json:"nextCursor"`
    PrevCursor  string `json:"prevCursor,omitempty"` // Add previous cursor
    Limit       int    `json:"limit"`
    TotalCount  int64  `json:"totalCount"`
    HasMore     bool   `json:"hasMore"`
    CurrentPage int    `json:"currentPage"`
}

type TaskListResponse struct {
    Tasks      []*TaskResponse `json:"tasks"`
    Pagination struct {
        NextCursor string `json:"nextCursor"`
        Limit     int    `json:"limit"`
        TotalCount int64  `json:"totalCount"`
        HasMore    bool   `json:"hasMore"`
    } `json:"pagination"`
}

type TaskResponse struct {
    ID          uint   `json:"id"`
    Title       string `json:"title"`
    Description string `json:"description"`
    Done        bool   `json:"done"`
    Owner       uint   `json:"owner"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`    
}

type TaskOperationResponse struct {
    Code          int           `json:"code"`
    ResultMessage string        `json:"resultMessage"`
    Data          interface{} `json:"data"`
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

func NewErrorResponse(code int, resultMessage string, err string) *ErrorResponse {
    return &ErrorResponse{
        Code:          code,
        ResultMessage: resultMessage,
        Error:         err,
    }
}