package todo

import (
	"github.com/gofiber/fiber/v2"
)

func CreateTodo(c *fiber.Ctx) error {
	repo := &TodoRepositoryImpl{}
	service := NewTodoService(repo)
	todo := &Todo{}
	err := service.CreateTodo(todo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error: failed to create a new task": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "task created successfully", "data": todo})
}

func UpdateTodo(c *fiber.Ctx) error {
	repo := &TodoRepositoryImpl{}
	service := NewTodoService(repo)

	// Parse the ID from the URL parameter.
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	// Parse the updated todo from the request body.
	todo := &Todo{}
	if err := c.BodyParser(todo); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// Set the ID of the todo to the ID from the URL parameter.
	todo.Id = id

	err = service.UpdateTodo(todo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update task", "details": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Task updated successfully", "data": todo})
}

func UpdateTodoDone(c *fiber.Ctx) error {
}

func GetAllTodos(c *fiber.Ctx) error {
}

func GetTodoById(c *fiber.Ctx) error {
}

func DeleteTodoById(c *fiber.Ctx) error {
}
