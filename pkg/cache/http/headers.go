package http

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// Headers manages HTTP cache-related headers
type Headers struct{}

// NewHeaders creates a new HTTP headers utility
func NewHeaders() *Headers {
    return &Headers{}
}

// SetETag sets the ETag header if value is not empty
func (h *Headers) SetETag(c *gin.Context, etag string) {
    if etag != "" {
        c.Header("ETag", etag)
    }
}

// AddCacheControl sets appropriate cache-control headers
func (h *Headers) AddCacheControl(c *gin.Context, isModifying bool, maxAge int) {
    if isModifying {
        // For POST, PUT, DELETE - tell browsers not to cache
        c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
        c.Header("Pragma", "no-cache")
        c.Header("Expires", "0")
    } else {
        // For GET - allow controlled caching
        c.Header("Cache-Control", "private, max-age=" + strconv.Itoa(maxAge))
        c.Header("Vary", "Authorization")
    }
}

// IsNotModified checks if content hasn't changed based on If-None-Match header
func (h *Headers) IsNotModified(c *gin.Context, etag string) bool {
    ifNoneMatch := c.GetHeader("If-None-Match")
    if ifNoneMatch == "" {
        return false
    }
    
    // Split by comma and check each ETag
    etags := strings.Split(ifNoneMatch, ",")
    for _, candidate := range etags {
        candidate = strings.TrimSpace(candidate)
        if candidate == etag || candidate == "W/"+etag {
            return true
        }
    }
    
    return false
}