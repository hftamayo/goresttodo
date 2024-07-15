package todo

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/v1/models"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func NewTodoRepositoryImpl(db *gorm.DB) *TodoRepositoryImpl {
	return &TodoRepositoryImpl{db: db}
}

func (h *Handler) CreateTodo(c *gin.Context) {
	db := h.db
	repo := NewTodoRepositoryImpl(db)
	service := NewTodoService(repo)
	todo := &models.Todo{}

	if err := c.ShouldBindJSON(todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot parse JSON"})
		return
	}
	err := service.CreateTodo(todo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create a new task", "details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "task created successfully", "data": todo})
}

func (h *Handler) UpdateTodo(c *gin.Context) {
	db := h.db
	repo := NewTodoRepositoryImpl(db)
	service := NewTodoService(repo)

	// Parse the ID from the URL parameter.
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Parse the updated todo from the request body.
	todo := &models.Todo{}
	if err := c.ShouldBindJSON(todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot parse JSON"})
		return
	}

	// Set the ID of the todo to the ID from the URL parameter.
	todo.ID = uint(id)

	err = service.UpdateTodo(todo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully", "data": todo})
}

func (h *Handler) UpdateTodoDone(c *gin.Context) {
	db := h.db
	repo := NewTodoRepositoryImpl(db)
	service := NewTodoService(repo)

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
	todo := &models.Todo{
		Model: gorm.Model{ID: uint(id)},
		Done:  done,
	}

	todo, err = service.MarkTodoAsDone(int(todo.ID), done) // Pass the ID of the todo instead of the todo itself.
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully", "data": todo})
}

func (h *Handler) GetAllTodos(c *gin.Context) {
	db := h.db
	repo := NewTodoRepositoryImpl(db)
	service := NewTodoService(repo)
	todos, err := service.GetAllTodos()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to fetch tasks", " details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tasks fetched successfully", "data": todos})
}

func (h *Handler) GetTodoById(c *gin.Context) {
	db := h.db
	repo := NewTodoRepositoryImpl(db)
	service := NewTodoService(repo)

	// Parse the ID from the URL parameter.
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	todo, err := service.GetTodoById(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task fetched successfully", "data": todo})
}

func (h *Handler) DeleteTodoById(c *gin.Context) {
	db := h.db
	repo := NewTodoRepositoryImpl(db)
	service := NewTodoService(repo)

	// Parse the ID from the URL parameter.
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	err = service.DeleteTodoById(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}
