package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/v1/health"
	"github.com/hftamayo/gotodo/api/v1/task"
	"github.com/hftamayo/gotodo/pkg/config"
	"gorm.io/gorm"
)

func SetupRouter(r *gin.Engine, db *gorm.DB, cache config.CacheInterface, errorLogger config.ErrorLogger) {
	
	taskRepo := task.NewTaskRepositoryImpl(db)

	// Create task service with custom configuration
	taskServiceConfig := task.DefaultTaskServiceConfig()
	taskServiceConfig.ErrorLogger = errorLogger
	taskService := task.NewTaskServiceWithConfig(taskRepo, cache, taskServiceConfig)

	taskHandler := task.NewHandler(taskService)
	healthHandler := health.NewHealthHandler(db)

	SetupTaskRoutes(r, taskHandler)
	SetupHealthCheckRoutes(r, healthHandler)
}
