package task

import (
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

type DoneTaskRequest struct {
    Done bool `json:"done" binding:"required"`
}

type PaginationQuery struct {
    Page     int `form:"page" binding:"required,min=1"`
    PageSize int `form:"limit" binding:"required,min=1,max=100"`
}

type PaginationMeta struct {
    Page     int `json:"page"`
    PageSize int `json:"pageSize"`
}

type TaskListResponse struct {
    Tasks      []*TaskResponse `json:"tasks"`
    Pagination PaginationMeta  `json:"pagination"`
}

type TaskResponse struct {
    ID          uint   `json:"id"`
    Title       string `json:"title"`
    Description string `json:"description"`
    Done        bool   `json:"done"`
    Owner       uint   `json:"owner"`
}

func ToTaskResponse(task *models.Task) *TaskResponse {
    return &TaskResponse{
        ID:          task.ID,
        Title:       task.Title,
        Description: task.Description,
        Done:        task.Done,
        Owner:       task.Owner,
    }
}

func TasksToResponse(tasks []*models.Task) []*TaskResponse {
    taskResponses := make([]*TaskResponse, len(tasks))
    for i, task := range tasks {
        taskResponses[i] = ToTaskResponse(task)
    }
    return taskResponses
}