package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/hftamayo/gotodo/api/routes"
	"github.com/hftamayo/gotodo/pkg/config"
)

func main() {
	var err error

	app := fiber.New()
	envVars, err := config.LoadEnvVars()
	if err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	} else {
		if !config.CheckDataLayerAvailability(envVars) {
			log.Fatalf("Error: Data layer is not available: %v", err)
		} else {
			db, err := config.DataLayerConnect(envVars)
			if err != nil {
				log.Fatalf("Failed to connect to the database: %v", err)
			} else {
				app.Use(cors.New(cors.Config{
					AllowOrigins:     "http://localhost:5173",
					AllowHeaders:     "Origin, Content-Type, Accept",
					AllowCredentials: true,
				}))

				app.Use(func(c *fiber.Ctx) error {
					c.Locals("db", db)
					return c.Next()
				})

				routes.SetupRoutes(app, db)

				app.Get("/gotodo/healthcheck", func(c *fiber.Ctx) error {
					return c.SendString("GoToDo RestAPI is up and running")
				})

				log.Fatal(app.Listen(":8001"))

			}
		}

	}

}
