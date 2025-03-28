package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/pkg/utils"
)

func RateLimitMiddleware(rateLimiter *utils.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		allowed, err := rateLimiter.Allow(clientIP)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":          http.StatusInternalServerError,
				"resultMessage": "INTERNAL_SERVER_ERROR",
			})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":          http.StatusTooManyRequests,
				"resultMessage": "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
