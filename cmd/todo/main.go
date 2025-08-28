package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/routes"
	"github.com/hftamayo/gotodo/pkg/config"
	"github.com/hftamayo/gotodo/pkg/middleware"
)

func main() {
	fmt.Printf("Starting GoToDo API\n")

	r := gin.Default()
	fmt.Printf("reading environment...\n")
	envVars, err := config.LoadEnvVars()
	if err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	fmt.Printf("setting up CORS...\n")
	r.Use(middleware.CORSMiddleware(envVars))

	fmt.Printf("verify data layer availability...\n")
	db, err := config.CheckDataLayerAvailability(envVars)
	if err != nil {
		log.Fatalf("Error: Data layer is not available, exiting...: %v", err)
	}

	fmt.Printf("connecting to the database...\n")
	db, err = config.DataLayerConnect(envVars)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Setting up cache with new architecture
	fmt.Printf("Setting up cache...\n")
	cache, err := config.NewCacheWithDefaults()
	if err != nil {
		log.Printf("Warning: Failed to setup Redis cache, falling back to memory cache: %v", err)
		cache = config.NewMemoryCache()
	}

	// Setting up error logger with new architecture
	fmt.Printf("Setting up error logger...\n")
	errorLogger, err := config.NewErrorLogger("redis")
	if err != nil {
		log.Printf("Warning: Failed to setup Redis error logger, falling back to memory logger: %v", err)
		errorLogger = config.NewErrorLoggerWithDefaults()
	}

	// Setting up rate limiter (still using Redis for now)
	fmt.Printf("setting up the rate limiter...\n")
	redisClient, err := config.ErrorLogConnect()
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis for rate limiter, rate limiting will be disabled: %v", err)
		// TODO: Implement in-memory rate limiter fallback
	} else {
		rateLimiter := config.SetupRateLimiter(redisClient, 100, time.Minute)
		r.Use(middleware.RateLimiter(rateLimiter))
	}

	fmt.Printf("Setting up routes... \n")
	routes.SetupRouter(r, db, cache, errorLogger)

    // Server configuration
    server := &http.Server{
        Addr:         fmt.Sprintf(":%d", envVars.AppPort),
        Handler:      r,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
    }	

    // Graceful shutdown
    go func() {
        fmt.Printf("GoToDo API is running on port %d\n", envVars.AppPort)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    fmt.Println("Shutting down server...")
    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }

    fmt.Println("Server exiting")
}