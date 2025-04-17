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

type HealthResponse struct {
    App AppHealthDetails `json:"app"`
    Db  DbHealthDetails `json:"db"`
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
    return &HealthHandler{
        db:        db,
        startTime: time.Now(),
    }
}

func (h *HealthHandler) Check(c *gin.Context) {
    now := time.Now()
    var mem runtime.MemStats
    runtime.ReadMemStats(&mem)

    health := HealthResponse{
        App: AppHealthDetails{
            Timestamp: now.UTC().Format(time.RFC3339),
            Uptime:    now.Sub(h.startTime).Seconds(),
            MemoryUsage: MemoryUsage{
                Total: mem.TotalAlloc,
                Free:  mem.Frees,
            },
            StartTime: h.startTime.Unix(),
        },
        Db: DbHealthDetails{
            Timestamp: now.UTC().Format(time.RFC3339),
        },
    }

    // Check database connection
    start := time.Now()
    if sqlDB, err := h.db.DB(); err != nil {
        health.Db.DatabaseStatus = "error"
        health.Db.Error = "Database connection error: " + err.Error()
        c.JSON(http.StatusServiceUnavailable, health)
        return
    } else if err := sqlDB.Ping(); err != nil {
        health.Db.DatabaseStatus = "error"
        health.Db.Error = "Database ping failed: " + err.Error()
        c.JSON(http.StatusServiceUnavailable, health)
        return
    }

    health.Db.ConnectionTime = time.Since(start).Seconds()
    health.Db.DatabaseStatus = "healthy"

    c.JSON(http.StatusOK, health)
}