package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	app := fiber.New()

	app.Use(logger.New())
	app.Use(recover.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("./index.html")

	})
	app.Listen("127.0.0.1:8080")
}
