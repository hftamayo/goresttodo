package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/routes"
	"github.com/hftamayo/gotodo/pkg/config"
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
    if err := config.CheckDataLayerAvailability(envVars); err != nil {
        log.Fatalf("Error: Data layer is not available, exiting...: %v", err)
    }

    fmt.Printf("connecting to the database...\n")
    db, err := config.DataLayerConnect(envVars)
    if err != nil {
        log.Fatalf("Failed to connect to the database: %v", err)
    }

    fmt.Printf("connected to the database, loading last stage: \n")
    router.SetupRouter(r, db)

    fmt.Printf("GoToDo API is up and running\n")
    log.Fatal(r.Run(":8001"))
}	
