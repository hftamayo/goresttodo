package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/routes"
	"github.com/hftamayo/gotodo/pkg/config"
	"github.com/hftamayo/gotodo/pkg/middleware"
	"github.com/hftamayo/gotodo/pkg/utils"
)

func main() {
	fmt.Printf("Starting GoToDo API\n")

	r := gin.Default()
	fmt.Printf("reading environment...\n")
	envVars, err := config.LoadEnvVars()
	if err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

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

	fmt.Printf("loading error logger...\n")
	redisClient, err := config.ErrorLogConnect()
	if err != nil {
		log.Fatalf("Failed to connect to the error logger: %v", err)
	}

	//setting up the cache
	fmt.Printf("setting up the cache...\n")
	cache := utils.NewCache(redisClient)

	//setting up the rate limiter
	fmt.Printf("setting up the rate limiter...\n")
	rateLimiter := utils.NewRateLimiter(redisClient, 100, time.Minute)
	r.Use(middleware.RateLimitMiddleware(rateLimiter))

	fmt.Printf("connected to the database, loading last stage: \n")
	routes.SetupRouter(r, db, redisClient, cache)

	fmt.Printf("GoToDo API is up and running\n")
	log.Fatal(r.Run(":8001"))
}
