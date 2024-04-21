package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/hftamayo/gotodo/api/v1/todo"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var db *gorm.DB

func main() {
	var err error
	db, err = gorm.Open("postgres", "host=localhost port=5432 user=postgres dbname=gotodo password=postgres sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3002",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	todos := []Todo{}

	app.Get("/gotodo/healthcheck", func(c *fiber.Ctx) error {
		return c.SendString("GoToDo RestAPI is up and running")
	})

	app.Post("/gotodo/todo", todo.CreateTodo)
	app.Patch("/gotodo/todo/:id/done", todo.UpdateTodoDone)
	app.Patch("/gotodo/todo/:id", todo.UpdateTodo)
	app.Get("/gotodo/todos", todo.GetAllTodos)
	app.Get("/gotodo/todo/:id", todo.GetTodoById)
	app.Delete("/gotodo/todo/:id", todo.DeleteTodoById)

	log.Fatal(app.Listen(":8001"))
}
