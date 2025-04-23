package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/v1/health"
)

func SetupHealthCheckRoutes(app *gin.Engine, handler *health.HealthHandler) {
	const hcPath = "/tasks/healthcheck"

	app.GET(hcPath+"/app", handler.AppStatus)
	app.GET(hcPath+"/db", handler.DbStatus)

}
