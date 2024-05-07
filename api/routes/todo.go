package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hftamayo/gotodo/api/v1/todo"
	"github.com/jinzhu/gorm"
)

const todoIDPath = "/gotodo/todo/:id"

func SetupRoutes(app *fiber.App, db *gorm.DB) {
	app.Post("/gotodo/todo", func(c *fiber.Ctx) error {
		return todo.CreateTodo(c, db)
	})

	app.Patch(todoIDPath+"/done", func(c *fiber.Ctx) error {
		return todo.UpdateTodoDone(c, db)
	})

	app.Patch(todoIDPath, func(c *fiber.Ctx) error {
		return todo.UpdateTodo(c, db)
	})

	app.Get("/gotodo/todos", func(c *fiber.Ctx) error {
		return todo.GetAllTodos(c, db)
	})

	app.Get(todoIDPath, func(c *fiber.Ctx) error {
		return todo.GetTodoById(c, db)
	})

	app.Delete(todoIDPath, func(c *fiber.Ctx) error {
		return todo.DeleteTodoById(c, db)
	})

}
