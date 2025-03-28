package task

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/v1/errorlog"
	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/utils"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Handler struct {
	Db              *gorm.DB
	Service         *TaskService
	ErrorLogService *errorlog.ErrorLogService
	cache 		 	*utils.Cache
	redisClient 	*redis.Client
}

func NewHandler(db *gorm.DB, service *TaskService, errorLogService *errorlog.ErrorLogService) *Handler {
	redisClient := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379", // Redis server address
        Password: "",               // no password set
        DB:       0,                // use default DB
    })	

	return &Handler{
		Db:              db,
		Service:         service,
		ErrorLogService: errorLogService,
		cache:           utils.NewCache(redisClient),
		redisClient:     redisClient,
	}
}

func NewTaskRepositoryImpl(db *gorm.DB) *TaskRepositoryImpl {
	return &TaskRepositoryImpl{Db: db}
}

func (h *Handler) List(c *gin.Context) {
	db := h.Db
	repo := NewTaskRepositoryImpl(db)

	service := NewTaskService(repo, h.cache)
	tasks, err := service.List()
	if err != nil {
		h.ErrorLogService.LogError("Task_list", err)

		c.JSON(http.StatusBadRequest, gin.H{
			"code":          http.StatusBadRequest,
			"resultMessage": utils.OperationFailed,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":          http.StatusOK,
		"resultMessage": utils.OperationSuccess,
		"tasks":         tasks,
	})
}

func (h *Handler) ListById(c *gin.Context) {
	db := h.Db
	repo := NewTaskRepositoryImpl(db)
	service := NewTaskService(repo, h.cache)

	// Parse the ID from the URL parameter.
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.ErrorLogService.LogError("Task_list_by_id", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":          http.StatusBadRequest,
			"resultMessage": utils.OperationFailed,
		})
		return
	}
	task, err := service.ListById(id)
	if err != nil {
		h.ErrorLogService.LogError("Task_list_by_id", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":          http.StatusInternalServerError,
			"resultMessage": utils.OperationFailed,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":          http.StatusOK,
		"resultMessage": utils.OperationSuccess,
		"task":          task,
	})
}

func (h *Handler) Create(c *gin.Context) {
	db := h.Db
	repo := NewTaskRepositoryImpl(db)
	service := NewTaskService(repo, h.cache)
	task := &models.Task{}

	if err := c.ShouldBindJSON(task); err != nil {
		h.ErrorLogService.LogError("Task_create", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":          http.StatusBadRequest,
			"resultMessage": utils.OperationFailed,
		})
		return
	}
	err := service.Create(task)
	if err != nil {
		h.ErrorLogService.LogError("Task_create", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":          http.StatusInternalServerError,
			"resultMessage": utils.OperationFailed,
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"code":          http.StatusCreated,
		"resultMessage": utils.OperationSuccess,
		"task":          task,
	})
}

func (h *Handler) Update(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "code":          http.StatusBadRequest,
            "resultMessage": utils.OperationFailed,
            "error":        "Invalid ID",
        })
        return
    }

    existingTask, err := h.Service.ListById(id)
    if err != nil {
        h.ErrorLogService.LogError("Task_update_fetch", err)
        c.JSON(http.StatusNotFound, gin.H{
            "code":          http.StatusNotFound,
            "resultMessage": utils.OperationFailed,
            "error":        "Task not found",
        })
        return
    }

    updatedTask := &models.Task{}
    if err := c.ShouldBindJSON(updatedTask); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "code":          http.StatusBadRequest,
            "resultMessage": utils.OperationFailed,
            "error":        "Invalid request body",
        })
        return
    }

    updatedTask.ID = uint(id)
    updatedTask.Owner = existingTask.Owner

    err = h.Service.Update(updatedTask)
    if err != nil {
        h.ErrorLogService.LogError("Task_update", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "code":          http.StatusInternalServerError,
            "resultMessage": utils.OperationFailed,
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "code":          http.StatusOK,
        "resultMessage": utils.OperationSuccess,
        "task":         updatedTask,
    })
}

func (h *Handler) Done(c *gin.Context) {
	db := h.Db
	repo := NewTaskRepositoryImpl(db)
	service := NewTaskService(repo, h.cache)

	// Parse the ID from the URL parameter.
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.ErrorLogService.LogError("Task_done", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":          http.StatusBadRequest,
			"resultMessage": utils.OperationFailed,
		})
		return
	}

	var body map[string]bool
	if err := c.ShouldBindJSON(&body); err != nil {
		h.ErrorLogService.LogError("Task_done", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":          http.StatusBadRequest,
			"resultMessage": utils.OperationFailed,
		})
		return
	}
	done, ok := body["done"]
	if !ok {
		h.ErrorLogService.LogError("Task_done", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":          http.StatusBadRequest,
			"resultMessage": utils.OperationFailed,
		})
	}
	task := &models.Task{
		Model: gorm.Model{ID: uint(id)},
		Done:  done,
	}

	task, err = service.Done(int(task.ID), done) // Pass the ID of the todo instead of the todo itself.
	if err != nil {
		h.ErrorLogService.LogError("Task_done", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":          http.StatusInternalServerError,
			"resultMessage": utils.OperationFailed,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":          http.StatusOK,
		"resultMessage": utils.OperationSuccess,
		"task":          task,
	})
}

func (h *Handler) Delete(c *gin.Context) {
	db := h.Db
	repo := NewTaskRepositoryImpl(db)
	service := NewTaskService(repo, h.cache)

	// Parse the ID from the URL parameter.
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.ErrorLogService.LogError("Task_delete", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":          http.StatusBadRequest,
			"resultMessage": utils.OperationFailed,
		})
		return
	}

	err = service.Delete(id)
	if err != nil {
		h.ErrorLogService.LogError("Task_delete", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":          http.StatusInternalServerError,
			"resultMessage": utils.OperationFailed,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":          http.StatusOK,
		"resultMessage": utils.OperationSuccess,
	})
}
