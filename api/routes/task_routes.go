package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/v1/task"
)

const (
    basePath    = "/tasks/task"
    taskIDPath  = basePath + "/:id"
    taskDonePath = taskIDPath + "/done"
)

func SetupTaskRoutes(r *gin.Engine, handler *task.Handler) {
    taskGroup := r.Group(basePath)
    {
        taskGroup.GET("/list", handler.List)
        taskGroup.GET("/:id", handler.ListById)
        taskGroup.POST("", handler.Create)
        taskGroup.PATCH("/:id", handler.Update)
        taskGroup.PATCH("/:id/done", handler.Done)
        taskGroup.DELETE("/:id", handler.Delete)
    }

}
