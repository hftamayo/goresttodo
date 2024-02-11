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
    body    string  `json:"body"

func main() {
    fmt.Print("This is a test")
    app := fiber.New()

    app.Get("/gotodo/healthcheck", func(c *fiber.Ctx) error {
        return c.SendString("GoToDo RestAPI is up and running")
    })

    log.Fatal(app.Listen(":8001"))
}
