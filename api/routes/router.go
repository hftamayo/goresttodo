package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/v1/errorlog"
	"github.com/hftamayo/gotodo/api/v1/health"
	"github.com/hftamayo/gotodo/api/v1/task"
	"github.com/hftamayo/gotodo/pkg/utils"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func SetupRouter(r *gin.Engine, db *gorm.DB, redisClient *redis.Client, cache *utils.Cache) {
	setupCORS(r)

	logRepo := errorlog.NewErrorLogRepositoryImpl(redisClient)
	taskRepo := task.NewTaskRepositoryImpl(db)

	errorLogService := errorlog.NewErrorLogService(logRepo)
	taskService := task.NewTaskService(taskRepo, cache)

	taskHandler := task.NewHandler(taskService, errorLogService, cache)
	healthHandler := health.NewHealthHandler(db)

	SetupTaskRoutes(r, taskHandler)
	SetupHealthCheckRoutes(r, healthHandler)
}

func setupCORS(r *gin.Engine) {
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:5173"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
        AllowHeaders:     []string{"Origin, Content-Type, Accept"},
        AllowCredentials: true,
    }))
}