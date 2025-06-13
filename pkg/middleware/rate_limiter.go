package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/pkg/utils"
)

// RateLimiter is a middleware that limits the number of requests
func RateLimiter(limiter *utils.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := c.ClientIP()
		if clientID == "" {
			clientID = "unknown"
		}

		// Determine operation type based on HTTP method
		var op utils.OperationType
		switch c.Request.Method {
		case "GET":
			op = utils.OperationRead
		case "POST", "PUT", "PATCH", "DELETE":
			op = utils.OperationWrite
		default:
			op = utils.OperationRead
		}

		allowed, limit, retryTime, err := limiter.Allow(clientID, op)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Rate limit error",
			})
			c.Abort()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.FormatInt(limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(limit-1, 10))
		if !retryTime.IsZero() {
			c.Header("X-RateLimit-Reset", strconv.FormatInt(retryTime.Unix(), 10))
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": retryTime.Unix(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}