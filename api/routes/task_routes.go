package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/v1/todo"
)

func SetupTodoRoutes(app *gin.Engine, handler *todo.Handler) {
	const todoIDPath = "/gotodo/task/:id"

	app.POST("/gotodo/task/new", handler.CreateTodo)
	app.PATCH(todoIDPath+"/done", handler.UpdateTodoDone)
	app.PATCH(todoIDPath, handler.UpdateTodo)
	app.GET("/gotodo/task/list", handler.GetAllTodos)
	app.GET(todoIDPath, handler.GetTodoById)
	app.DELETE(todoIDPath, handler.DeleteTodoById)

}
