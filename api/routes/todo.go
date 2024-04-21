package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hftamayo/gotodo/api/v1/todo"
)

const todoIDPath = "/gotodo/todo/:id"

func SetupRoutes(app *fiber.App) {
	app.Post("/gotodo/todo", todo.CreateTodo)
	app.Patch(todoIDPath+"/done", todo.UpdateTodoDone)
	app.Patch(todoIDPath, todo.UpdateTodo)
	app.Get("/gotodo/todos", todo.GetAllTodos)
	app.Get(todoIDPath, todo.GetTodoById)
	app.Delete(todoIDPath, todo.DeleteTodoById)
}
