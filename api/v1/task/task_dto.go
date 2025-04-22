package task

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