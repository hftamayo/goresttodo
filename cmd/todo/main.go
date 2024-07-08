package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hftamayo/gotodo/api/routes"
	"github.com/hftamayo/gotodo/api/v1/todo"
	"github.com/hftamayo/gotodo/pkg/config"
	"github.com/jinzhu/gorm"
)

var db *gorm.DB

func main() {
	var err error
	fmt.Printf("Starting ToDo GraphQLAPI\n")

	r := gin.Default()
	fmt.Printf("reading environment...\n")
	envVars, err := config.LoadEnvVars()
	if err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	} else {
		fmt.Printf("verify data layer availability...\n")
		_, err := config.CheckDataLayerAvailability(envVars)
		if err != nil {
			log.Fatalf("Error: Data layer is not available, exiting...: %v", err)
		} else {
			fmt.Printf("connecting to the database...\n")
			db, err := config.DataLayerConnect(envVars)
			if err != nil {
				log.Fatalf("Failed to connect to the database: %v", err)
			} else {
				fmt.Printf("connected to the database, loading last stage: \n")
				r.Use(cors.New(cors.Config{
					AllowOrigins:     []string{"http://localhost:5173"},
					AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
					AllowHeaders:     []string{"Origin, Content-Type, Accept"},
					AllowCredentials: true,
				}))

				handler := todo.NewHandler(db)
				routes.SetupRoutes(r, handler)
				fmt.Printf("API is up and running")

				r.GET("/gotodo/healthcheck", func(c *gin.Context) {
					c.String(http.StatusOK, "GoToDo RestAPI is up and running")
				})

				log.Fatal(r.Run(":8001"))

			}
		}

	}

}
