package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/v1/todo"
)

func SetupRoutes(app *gin.Engine, handler *todo.Handler) {
	const todoIDPath = "/gotodo/todo/:id"

	app.POST("/gotodo/todo", handler.CreateTodo)

	app.PATCH(todoIDPath+"/done", handler.UpdateTodoDone)

	app.PATCH(todoIDPath, handler.UpdateTodo)

	app.GET("/gotodo/todos", handler.GetAllTodos)

	app.GET(todoIDPath, handler.GetTodoById)

	app.DELETE(todoIDPath, handler.DeleteTodoById)

}
