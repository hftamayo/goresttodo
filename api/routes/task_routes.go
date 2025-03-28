package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/v1/task"
)

func SetupTaskRoutes(app *gin.Engine, handler *task.Handler) {
	const todoIDPath = "/gotodo/task/:id"

	app.GET("/gotodo/task/list", handler.List)
	app.GET(todoIDPath, handler.ListById)
	app.POST("/gotodo/task/new", handler.Create)
	app.PATCH(todoIDPath, handler.Update)
	app.PATCH(todoIDPath+"/done", handler.Done)
	app.DELETE(todoIDPath, handler.Delete)

}
