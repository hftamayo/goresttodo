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

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3002",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	db, err := config.DataLayerConnect()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	routes.SetupRoutes(app, db)

	app.Get("/gotodo/healthcheck", func(c *fiber.Ctx) error {
		return c.SendString("GoToDo RestAPI is up and running")
	})

	log.Fatal(app.Listen(":8001"))
}
