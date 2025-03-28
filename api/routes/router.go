package routes

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/v1/errorlog"
	"github.com/hftamayo/gotodo/api/v1/task"
	"github.com/hftamayo/gotodo/pkg/utils"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func SetupRouter(r *gin.Engine, db *gorm.DB, redisClient *redis.Client, cache *utils.Cache) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin, Content-Type, Accept"},
		AllowCredentials: true,
	}))

	logRepo := errorlog.NewErrorLogRepositoryImpl(redisClient)
	taskRepo := task.NewTaskRepositoryImpl(db)

	taskService := task.NewTaskService(taskRepo, cache)
	errorLogService := errorlog.NewErrorLogService(logRepo)

	taskHandler := task.NewHandler(db, taskService, errorLogService)

	SetupTaskRoutes(r, taskHandler)

	r.GET("/gotodo/healthcheck", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H {
			"code":          http.StatusOK,
			"resultMessage": "GoToDo RestAPI is up and running",
		})
	})
}
