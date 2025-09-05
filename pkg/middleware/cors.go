package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/pkg/config"
)

func CORSMiddleware(config *config.EnvVars) gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")
		allowed := false
		
        
        // Check if origin is allowed (only if origin is present)
        if origin != "" {
            for _, allowedOrigin := range config.FeOrigins {
                if origin == allowedOrigin {
                    allowed = true
                    break
                }
            }
        } else {
            // If no origin header, it's not allowed
            allowed = false
        }
		
        // Handle preflight requests first
        if c.Request.Method == "OPTIONS" {
            if allowed {
                c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
                c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
                c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
                c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
                c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
                c.AbortWithStatus(204)
            } else {
                c.AbortWithStatus(403)
            }
            return
        }

        // If not allowed, return 403
        if !allowed {
            c.AbortWithStatus(403)
            return
        }

        // Set CORS headers for allowed requests
        c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        
        c.Next()
    }
}