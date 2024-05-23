package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/hftamayo/gotodo/api/routes"
	"github.com/hftamayo/gotodo/api/v1/todo"
	"github.com/hftamayo/gotodo/pkg/config"
	"github.com/jinzhu/gorm"
)

var db *gorm.DB

func main() {
	var err error
	fmt.Printf("Starting ToDo RestAPI\n")

	app := fiber.New()
	fmt.Printf("reading environment...\n")
	envVars, err := config.LoadEnvVars()
	if err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	} else {
		fmt.Printf("verify data layer availability...\n")
		_, err := config.CheckDataLayerAvailability(envVars)
		if err != nil {
			log.Fatalf("Error: Data layer is not available: %v", err)
		} else {
			fmt.Printf("connecting to the database...\n")
			db, err := config.DataLayerConnect(envVars)
			if err != nil {
				log.Fatalf("Failed to connect to the database: %v", err)
			} else {
				fmt.Printf("connected to the database, loading last stage: \n")
				app.Use(cors.New(cors.Config{
					AllowOrigins:     "http://localhost:5173",
					AllowHeaders:     "Origin, Content-Type, Accept",
					AllowCredentials: true,
				}))

				handler := todo.NewHandler(db)
				routes.SetupRoutes(app, handler)
				fmt.Printf("API is up and running")

				app.Get("/gotodo/healthcheck", func(c *fiber.Ctx) error {
					return c.SendString("GoToDo RestAPI is up and running")
				})

				log.Fatal(app.Listen(":8001"))

			}
		}

	}

}
