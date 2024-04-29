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
}

func UpdateTodoDone(c *fiber.Ctx) error {
}

func GetAllTodos(c *fiber.Ctx) error {
}

func GetTodoById(c *fiber.Ctx) error {
}

func DeleteTodoById(c *fiber.Ctx) error {
}
