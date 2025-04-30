package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/pkg/config"
)

func CORSMiddleware(config *config.EnvVars) gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")
		allowed := false
		/*

        fmt.Printf("\n=== CORS Debug Information ===\n")
        fmt.Printf("Server running on: http://localhost:%d\n", config.AppPort)
        
        // Debug request details
        fmt.Printf("\n--- Request Details ---\n")
        fmt.Printf("Origin received: '%s' (length: %d)\n", origin, len(origin))
        fmt.Printf("Method: %s\n", c.Request.Method)
        fmt.Printf("Path: %s\n", c.Request.URL.Path)

        // Debug configured origins
        fmt.Printf("\n--- Configured Origins ---\n")
        fmt.Printf("Number of configured origins: %d\n", len(config.FeOrigins))
        for i, origin := range config.FeOrigins {
            fmt.Printf("Origin[%d]: '%s' (length: %d)\n", i, origin, len(origin))
        }

        // Origin matching with detailed logging
        fmt.Printf("\n--- Origin Matching Process ---\n")
		*/
        for _, allowedOrigin := range config.FeOrigins {
            //fmt.Printf("Comparing received '%s' with allowed '%s'\n", origin, allowedOrigin)			
            if origin == allowedOrigin {
                allowed = true
                c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
                break
            }
        }		

        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		
        // Handle preflight requests first
        if c.Request.Method == "OPTIONS" {
            if allowed {
                c.AbortWithStatus(204)
            } else {
                c.AbortWithStatus(403)
            }
            return
        }

        // If not allowed, return 403
        if !allowed {
            fmt.Printf("\n--- Access Denied ---\n")
            c.AbortWithStatus(403)
            return
        }

        fmt.Printf("\n--- Request Authorized ---\n")
        c.Next()
    }
}