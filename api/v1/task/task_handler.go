package task

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/v1/errorlog"
	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/utils"
	"gorm.io/gorm"
)

const (
    ErrInvalidID = "Invalid ID parameter"
    ErrTaskNotFound = "Task not found"
    ErrInvalidRequest = "Invalid request body"
    ErrInvalidPaginationParams = "Invalid pagination parameters"
)

type Handler struct {
	service         TaskServiceInterface
	errorLogService *errorlog.ErrorLogService
	cache 		 	*utils.Cache
}

func NewHandler(service TaskServiceInterface, errorLogService *errorlog.ErrorLogService, cache *utils.Cache) *Handler {
    if service == nil {
        panic("task service is required")
    }
    if errorLogService == nil {
        panic("error log service is required")
    }
    if cache == nil {
        panic("cache is required")
    }

	return &Handler{
        service:         service,
        errorLogService: errorLogService,
        cache:           cache,
	}
}

func (h *Handler) List(c *gin.Context) {
    var query CursorPaginationQuery
    if err := c.ShouldBindQuery(&query); err != nil {
        h.errorLogService.LogError("Task_list_validation", err)
        c.JSON(http.StatusBadRequest, gin.H{
            "code": http.StatusBadRequest,
            "resultMessage": utils.OperationFailed,
            "error": ErrInvalidPaginationParams,
        })
        return
    }

    if query.Limit <= 0 {
        query.Limit = defaultLimit
    }
    if query.Limit > maxLimit {
        query.Limit = maxLimit
    }    

    tasks, nextCursor, totalCount, err := h.service.List(query.Cursor, query.Limit)
    if err != nil {
        h.errorLogService.LogError("Task_list", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "code": http.StatusInternalServerError,
            "resultMessage": utils.OperationFailed,
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "code": http.StatusOK,
        "resultMessage": utils.OperationSuccess,
        "data": gin.H{
            "tasks": TasksToResponse(tasks),
            "pagination": gin.H{
                "nextCursor": nextCursor,
                "limit": query.Limit,
                "totalCount": totalCount,
                "hasMore": nextCursor != "",
            },
        },
    })
}


func (h *Handler) ListById(c *gin.Context) {
	// Parse the ID from the URL parameter.
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorLogService.LogError("Task_list_by_id", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"resultMessage": utils.OperationFailed,
			"error":        ErrInvalidID,
		})
		return
	}
	task, err := h.service.ListById(id)
	if err != nil {
		h.errorLogService.LogError("Task_list_by_id", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"resultMessage": utils.OperationFailed,
		})
		return
	}

    if task == nil {
        c.JSON(http.StatusNotFound, gin.H{
            "code": http.StatusNotFound,
            "resultMessage": utils.OperationFailed,
            "error":        ErrTaskNotFound,
        })
        return
    }

	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"resultMessage": utils.OperationSuccess,
		"data":         ToTaskResponse(task),
	})
}

func (h *Handler) Create(c *gin.Context) {
	var createRequest CreateTaskRequest
	if err := c.ShouldBindJSON(&createRequest); err != nil {
		h.errorLogService.LogError("Task_create", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"resultMessage": utils.OperationFailed,
			"error":        ErrInvalidRequest,
		})
		return
	}
    task := &models.Task{
        Title:       createRequest.Title,
        Description: createRequest.Description,
        Owner:       createRequest.Owner,
    }

    if err := h.service.Create(task); err != nil {
        h.errorLogService.LogError("Task_create", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "code": http.StatusInternalServerError,
            "resultMessage": utils.OperationFailed,
        })
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "code": http.StatusCreated,
        "resultMessage": utils.OperationSuccess,
        "data":         ToTaskResponse(task),
    })
}

func (h *Handler) Update(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "code": http.StatusBadRequest,
            "resultMessage": utils.OperationFailed,
            "error":        ErrInvalidID,
        })
        return
    }

    var updateRequest UpdateTaskRequest
    if err := c.ShouldBindJSON(&updateRequest); err != nil {
		h.errorLogService.LogError("Task_update_binding", err)
        c.JSON(http.StatusBadRequest, gin.H{
            "code": http.StatusBadRequest,
            "resultMessage": utils.OperationFailed,
            "error":        "Invalid request body",
        })
        return
    }

    task := &models.Task{
        Model:       gorm.Model{ID: uint(id)},
        Title:       updateRequest.Title,
        Description: updateRequest.Description,
    }	

    if err := h.service.Update(task); err != nil {
        h.errorLogService.LogError("Task_update", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "code": http.StatusInternalServerError,
            "resultMessage": utils.OperationFailed,
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "code": http.StatusOK,
        "resultMessage": utils.OperationSuccess,
        "data":         ToTaskResponse(task),
    })
}

func (h *Handler) Done(c *gin.Context) {
	// Parse the ID from the URL parameter.
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorLogService.LogError("Task_done", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"resultMessage": utils.OperationFailed,
			"error":        ErrInvalidID,
		})
		return
	}

    var doneRequest DoneTaskRequest
    if err := c.ShouldBindJSON(&doneRequest); err != nil {
        h.errorLogService.LogError("Task_done_binding", err)
        c.JSON(http.StatusBadRequest, gin.H{
            "code": http.StatusBadRequest,
            "resultMessage": utils.OperationFailed,
            "error":        ErrInvalidRequest,
        })
        return
    }

    task, err := h.service.Done(id, doneRequest.Done)
    if err != nil {
        h.errorLogService.LogError("Task_done", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "code": http.StatusInternalServerError,
            "resultMessage": utils.OperationFailed,
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "code": http.StatusOK,
        "resultMessage": utils.OperationSuccess,
        "data":         ToTaskResponse(task),
    })
}

func (h *Handler) Delete(c *gin.Context) {
	// Parse the ID from the URL parameter.
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorLogService.LogError("Task_delete", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"resultMessage": utils.OperationFailed,
			"error":        ErrInvalidID,
		})
		return
	}

    if err := h.service.Delete(id); err != nil {
        h.errorLogService.LogError("Task_delete", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "code": http.StatusInternalServerError,
            "resultMessage": utils.OperationFailed,
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "code": http.StatusOK,
        "resultMessage": utils.OperationSuccess,
    })
}
