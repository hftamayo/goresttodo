package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type Todo struct {
	id    int    `json:"id"`
	title string `json:"title"`
	done  bool   `json:"done"`
	body  string `json:"body"`
}

func main() {
	fmt.Print("This is a test")
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3002",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	todos := []Todo{}

	app.Get("/gotodo/healthcheck", func(c *fiber.Ctx) error {
		return c.SendString("GoToDo RestAPI is up and running")
	})

	app.Post("/gotodo/todos", func(c *fiber.Ctx) error {
		todo := &Todo{}

		if err := c.BodyParser(todo); err != nil {
			return err
		}

		todo.id = len(todos) + 1

		todos = append(todos, *todo)

		return c.JSON(todos)
	})

	app.Patch("/gotodo/todo/:id/done", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")

		if err != nil {
			return c.Status(401).SendString("Invalid id")
		}

		for i, t := range todos {
			if t.id == id {
				todos[i].done = true
				break
			}
		}
		return c.JSON(todos)
	})

	app.Patch("/gotodo/todo/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(401).SendString("Invalid id")
		}

		// Find the todo with the given ID
		var foundTodo *Todo
		for i, t := range todos {
			if t.id == id {
				foundTodo = &todos[i]
				break
			}
		}

		if foundTodo == nil {
			return c.Status(404).SendString("Todo not found")
		}

		// Parse request body to get updated properties
		var updatedTodo Todo
		if err := c.BodyParser(&updatedTodo); err != nil {
			return c.Status(400).SendString("Invalid request body")
		}

		// Update the properties
		foundTodo.title = updatedTodo.title
		foundTodo.done = updatedTodo.done
		foundTodo.body = updatedTodo.body

		return c.JSON(foundTodo)
	})

	app.Get("/gotodo/todos", func(c *fiber.Ctx) error {
		return c.JSON(todos)
	})

	app.Get("/gotodo/todo/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).SendString("Invalid ID")
		}

		// Find the todo with the given ID
		var foundTodo Todo
		for _, t := range todos {
			if t.id == id {
				foundTodo = t
				break
			}
		}

		if foundTodo.id == 0 {
			return c.Status(404).SendString("Todo not found")
		}

		return c.JSON(foundTodo)
	})

	app.Delete("/gotodo/todo/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(401).SendString("Invalid id")
		}

		// Find the index of the todo with the given ID
		var index int
		for i, t := range todos {
			if t.id == id {
				index = i
				break
			}
		}

		// Remove the todo from the slice
		if index >= 0 && index < len(todos) {
			todos = append(todos[:index], todos[index+1:]...)
		} else {
			return c.Status(404).SendString("Todo not found")
		}

		return c.JSON(todos)

	})

	log.Fatal(app.Listen(":8001"))
}
