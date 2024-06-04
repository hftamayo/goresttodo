package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hftamayo/gotodo/api/v1/todo"
)

func SetupRoutes(app *fiber.App, handler *todo.Handler) {
	const todoIDPath = "/gotodo/todo/:id"

	app.Post("/gotodo/todo", handler.CreateTodo)

	app.Patch(todoIDPath+"/done", handler.UpdateTodoDone)

	app.Patch(todoIDPath, handler.UpdateTodo)

	app.Get("/gotodo/todos", handler.GetAllTodos)

	app.Get(todoIDPath, handler.GetTodoById)

	app.Delete(todoIDPath, handler.DeleteTodoById)

}
