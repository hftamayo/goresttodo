package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/pkg/config"
)

func CORSMiddleware(config *config.EnvVars) gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")
		allowed := false

        // Check if the origin is allowed
        for _, allowedOrigin := range config.FeOrigins {
            if origin == allowedOrigin {
                c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				allowed = true
                break
            }
        }
        
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        // Handle preflight requests
        if c.Request.Method == "OPTIONS" {
            if allowed {
                c.AbortWithStatus(204)
            } else {
                c.AbortWithStatus(403)
            }
            return
        }

        // If origin is not allowed, abort with 403
        if !allowed {
            c.AbortWithStatus(403)
            return
        }

        c.Next()
    }
}