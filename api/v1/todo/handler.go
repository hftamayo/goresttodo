package todo

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/jinzhu/gorm"
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

func (h *Handler) CreateTodo(c *fiber.Ctx) error {
	db := h.db
	repo := NewTodoRepositoryImpl(db)
	service := NewTodoService(repo)
	todo := &models.Todo{}

	if err := c.BodyParser(todo); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}
	err := service.CreateTodo(todo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error: failed to create a new task": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "task created successfully", "data": todo})
}

func (h *Handler) UpdateTodo(c *fiber.Ctx) error {
	db := h.db
	repo := NewTodoRepositoryImpl(db)
	service := NewTodoService(repo)

	// Parse the ID from the URL parameter.
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	// Parse the updated todo from the request body.
	todo := &models.Todo{}
	if err := c.BodyParser(todo); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// Set the ID of the todo to the ID from the URL parameter.
	todo.ID = uint(id)

	err = service.UpdateTodo(todo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update task", "details": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Task updated successfully", "data": todo})
}

func (h *Handler) UpdateTodoDone(c *fiber.Ctx) error {
	db := h.db
	repo := NewTodoRepositoryImpl(db)
	service := NewTodoService(repo)

	// Parse the ID from the URL parameter.
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var body map[string]bool
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}
	done, ok := body["done"]
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing 'done' field in request body"})
	}
	todo := &models.Todo{
		Model: gorm.Model{ID: uint(id)},
		Done:  done,
	}

	err = service.MarkTodoAsDone(int(todo.ID)) // Pass the ID of the todo instead of the todo itself.
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update task", "details": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Task updated successfully", "data": todo})
}

func (h *Handler) GetAllTodos(c *fiber.Ctx) error {
	db := h.db
	repo := NewTodoRepositoryImpl(db)
	service := NewTodoService(repo)
	todos, err := service.GetAllTodos()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch tasks", "details": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Tasks fetched successfully", "data": todos})
}

func (h *Handler) GetTodoById(c *fiber.Ctx) error {
	db := h.db
	repo := NewTodoRepositoryImpl(db)
	service := NewTodoService(repo)

	// Parse the ID from the URL parameter.
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	todo, err := service.GetTodoById(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch task", "details": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Task fetched successfully", "data": todo})
}

func (h *Handler) DeleteTodoById(c *fiber.Ctx) error {
	db := h.db
	repo := NewTodoRepositoryImpl(db)
	service := NewTodoService(repo)

	// Parse the ID from the URL parameter.
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	err = service.DeleteTodoById(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete task", "details": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Task deleted successfully"})
}
