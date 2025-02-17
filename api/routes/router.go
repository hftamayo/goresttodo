package routes

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/hftamayo/gotodo/api/v1/errorlog"
	"github.com/hftamayo/gotodo/api/v1/task"
	"gorm.io/gorm"
)

func SetupRouter(r *gin.Engine, db *gorm.DB, redisClient *redis.Client) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin, Content-Type, Accept"},
		AllowCredentials: true,
	}))

	logRepo := errorlog.NewErrorLogRepositoryImpl(redisClient)
	taskRepo := task.NewTaskRepositoryImpl(db)

	taskService := task.NewTaskService(taskRepo)
	errorLogService := errorlog.NewErrorLogService(logRepo)

	taskHandler := task.NewHandler(db, taskService, errorLogService)

	SetupTaskRoutes(r, taskHandler)

	r.GET("/gotodo/healthcheck", func(c *gin.Context) {
		c.String(http.StatusOK, "GoToDo RestAPI is up and running")
	})
}
