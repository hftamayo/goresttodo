package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/v1/task"
)

func SetupTaskRoutes(app *gin.Engine, handler *task.Handler) {
	const todoIDPath = "/tasks/task/:id"

	app.GET("/tasks/task/list", handler.List)
	app.GET(todoIDPath, handler.ListById)
	app.POST("/tasks/task", handler.Create)
	app.PATCH(todoIDPath, handler.Update)
	app.PATCH(todoIDPath+"/done", handler.Done)
	app.DELETE(todoIDPath, handler.Delete)

}
