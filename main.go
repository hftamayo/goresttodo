package main
import (
    "fmt"
    "log"
    "github.com/gofiber/fiber/v2"
)

type Todo struct {
    id      int     `json:"id"`
    title   string  `json:"title"`
    done    bool    `json:"done"`
    body    string  `json:"body"`
}

func main() {
    fmt.Print("This is a test")
    app := fiber.New()

    app.Use(cors.New(cors.Config{
        AllowOrigins: "http://localhost:3000",
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

        todo.ID = len(todos) + 1

        todos = append(todos, *todo)

        return c.JSON(todos)
    })

    app.Patch("/gotodo/todo/:id/done", func(c *fiber.Ctx) error {
        id, err := c.ParamsInt("id")

        if err != nil {
            return c.Status(401).SendString("Invalid id")
        }

        for i, t := range todos {
            if t.ID == id {
                todos[i].Done = true
                break
            }
        }
        return c.JSON(todos)
    })

    app.Get("/gotodo/todos", func(c *fiber.Ctx) error {
        return c.JSON(todos)
    })

    log.Fatal(app.Listen(":8001"))
}
