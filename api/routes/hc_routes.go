package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/v1/hc"
)

func SetupHealthCheckRoutes(app *gin.Engine, handler *hc.Handler) {
	const hcPath = "/tasks/healthcheck"

	app.GET(hcPath, handler.AppStatus)
	app.GET(hcPath, handler.DbStatus)

}
