package main
import (
    "fmt"
    "log"
    "github.com/gofiber/fiber/v2"
)

func main() {
    fmt.Print("This is a test")
    app := fiber.New()

    app.Get("/gotodo/healthcheck", func(c *fiber.Ctx) error {
        return c.SendString("GoToDo RestAPI is up and running")
    })

    log.Fatal(app.Listen(":8001"))
}
