package task

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/v1/models"
	"gorm.io/gorm"
)

type Handler struct {
	Db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{Db: db}
}

func NewTaskRepositoryImpl(db *gorm.DB) *TaskRepositoryImpl {
	return &TaskRepositoryImpl{Db: db}
}

func (h *Handler) List(c *gin.Context) {
	db := h.Db
	repo := NewTaskRepositoryImpl(db)
	service := NewTaskService(repo)
	tasks, err := service.List()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to fetch tasks", " details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tasks fetched successfully", "data": tasks})
}

func (h *Handler) ListById(c *gin.Context) {
	db := h.Db
	repo := NewTaskRepositoryImpl(db)
	service := NewTaskService(repo)

	// Parse the ID from the URL parameter.
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	task, err := service.ListById(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task fetched successfully", "data": task})
}

func (h *Handler) Create(c *gin.Context) {
	db := h.Db
	repo := NewTaskRepositoryImpl(db)
	service := NewTaskService(repo)
	task := &models.Task{}

	if err := c.ShouldBindJSON(task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot parse JSON"})
		return
	}
	err := service.Create(task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create a new task", "details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "task created successfully", "data": task})
}

func (h *Handler) Update(c *gin.Context) {
	db := h.Db
	repo := NewTaskRepositoryImpl(db)
	service := NewTaskService(repo)

	// Parse the ID from the URL parameter.
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Parse the updated todo from the request body.
	task := &models.Task{}
	if err := c.ShouldBindJSON(task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot parse JSON"})
		return
	}

	// Set the ID of the todo to the ID from the URL parameter.
	task.ID = uint(id)

	err = service.Update(task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully", "data": task})
}

func (h *Handler) Done(c *gin.Context) {
	db := h.Db
	repo := NewTaskRepositoryImpl(db)
	service := NewTaskService(repo)

	// Parse the ID from the URL parameter.
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var body map[string]bool
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot parse JSON"})
		return
	}
	done, ok := body["done"]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'done' field in request body"})
	}
	task := &models.Task{
		Model: gorm.Model{ID: uint(id)},
		Done:  done,
	}

	task, err = service.Done(int(task.ID), done) // Pass the ID of the todo instead of the todo itself.
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully", "data": task})
}

func (h *Handler) Delete(c *gin.Context) {
	db := h.Db
	repo := NewTaskRepositoryImpl(db)
	service := NewTaskService(repo)

	// Parse the ID from the URL parameter.
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	err = service.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}
