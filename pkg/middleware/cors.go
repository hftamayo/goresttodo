package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/pkg/config"
)

func CORSMiddleware(config *config.EnvVars) gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")

        // Check if the origin is allowed
        for _, allowedOrigin := range config.FeOrigins {
            if origin == allowedOrigin {
                c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
                break
            }
        }
        
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

        // Handle preflight requests
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}