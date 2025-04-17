package health

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
    db        *gorm.DB
    startTime time.Time
}

type AppHealthDetails struct {
    Timestamp   string       `json:"timestamp"`
    Uptime      float64     `json:"uptime"`
    MemoryUsage MemoryUsage `json:"memoryUsage"`
    StartTime   int64       `json:"startTime"`
}

type MemoryUsage struct {
    Total uint64 `json:"total"`
    Free  uint64 `json:"free"`
}

type DbHealthDetails struct {
    Timestamp      string  `json:"timestamp"`
    ConnectionTime float64 `json:"connectionTime,omitempty"`
    DatabaseStatus string  `json:"databaseStatus,omitempty"`
    Error         string  `json:"error,omitempty"`
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
    return &HealthHandler{
        db:        db,
        startTime: time.Now(),
    }
}

func (h *HealthHandler) AppStatus(c *gin.Context) {
    now := time.Now()
    var mem runtime.MemStats
    runtime.ReadMemStats(&mem)

    health := AppHealthDetails{
        Timestamp: now.UTC().Format(time.RFC3339),
        Uptime:    now.Sub(h.startTime).Seconds(),
        MemoryUsage: MemoryUsage{
            Total: mem.TotalAlloc,
            Free:  mem.Frees,
        },
        StartTime: h.startTime.Unix(),
    }

    c.JSON(http.StatusOK, health)
}

func (h *HealthHandler) DbStatus(c *gin.Context) {
    now := time.Now()
    health := DbHealthDetails{
        Timestamp: now.UTC().Format(time.RFC3339),
    }

    start := time.Now()
    if sqlDB, err := h.db.DB(); err != nil {
        health.DatabaseStatus = "error"
        health.Error = "Database connection error: " + err.Error()
        c.JSON(http.StatusServiceUnavailable, health)
        return
    } else if err := sqlDB.Ping(); err != nil {
        health.DatabaseStatus = "error"
        health.Error = "Database ping failed: " + err.Error()
        c.JSON(http.StatusServiceUnavailable, health)
        return
    }

    health.ConnectionTime = time.Since(start).Seconds()
    health.DatabaseStatus = "healthy"

    c.JSON(http.StatusOK, health)
}