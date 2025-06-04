package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/pkg/utils"
)

func RateLimitMiddleware(rateLimiter *utils.RateLimiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        clientIP := c.ClientIP()
        
        // Determine operation type based on HTTP method and headers
        operation := utils.OperationRead
        if c.Request.Method != "GET" {
            operation = utils.OperationWrite
        }
        
        // Special case: prefetch requests
        if c.Request.Method == "GET" && (c.GetHeader("X-Purpose") == "prefetch" || c.Query("prefetch") != "") {
            operation = utils.OperationPrefetch
        }

        // Check if the request is allowed
        allowed, limit, retryTime, err := rateLimiter.AllowOperation(clientIP, operation)
        
        // Set rate limit headers
        c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
        
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "code":          http.StatusInternalServerError,
                "resultMessage": "INTERNAL_SERVER_ERROR",
                "error":         "Rate limiter error",
            })
            c.Abort()
            return
        }

        if !allowed {
            // Calculate retry after in seconds
            retryAfter := int(time.Until(retryTime).Seconds())
            if retryAfter < 1 {
                retryAfter = 1
            }
            
            // Set Retry-After header
            c.Header("Retry-After", strconv.Itoa(retryAfter))
            c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", retryTime.Unix()))
            
            c.JSON(http.StatusTooManyRequests, gin.H{
                "code":          http.StatusTooManyRequests,
                "resultMessage": "RATE_LIMIT_EXCEEDED",
                "retry_after":   retryAfter,
                "message":       fmt.Sprintf("Too many requests. Try again in %d seconds.", retryAfter),
            })
            c.Abort()
            return
        }

        c.Next()
    }
}