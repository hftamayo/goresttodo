package task

import (
	"time"

	"github.com/hftamayo/gotodo/api/v1/models"
)

type CreateTaskRequest struct {
    Title       string `json:"title" binding:"required"`
    Description string `json:"description" binding:"required"`
    Owner       uint   `json:"owner" binding:"required"`
}

type UpdateTaskRequest struct {
    Title       string `json:"title" binding:"required"`
    Description string `json:"description" binding:"required"`
}

type CursorPaginationQuery struct {
    Limit  int    `form:"limit" binding:"required,min=1,max=100"`
    Cursor string `form:"cursor"`
}

type CursorPaginationMeta struct {
    NextCursor string `json:"nextCursor,omitempty"`
    HasMore    bool   `json:"hasMore"`
    Count      int    `json:"count"`
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
    Task          *TaskResponse `json:"task"`
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