package main

import (
	"binance-pooler/pkg/app"
	"binance-pooler/pkg/syro"
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	app, err := app.New(context.Background())
	if err != nil {
		msg := fmt.Sprintf("failed to create app in go pooler: %v", err.Error())
		log.Fatalf(msg)
	}
	defer app.Exit(context.Background())

	// Create a new Fiber instance
	api := fiber.New(fiber.Config{
		DisableStartupMessage: true, // Disable Fiber's startup message
	})

	api.Use(
		cors.New(cors.Config{
			AllowCredentials: false,
			AllowOrigins:     "*", // NOTE: change this to a list of urls from which fetch requests are allowed
		}),
	)

	// Define routes
	api.Get("/logs", func(c *fiber.Ctx) error {
		data, err := syro.RequestLogs(app.Logger(), c.OriginalURL())
		if err != nil {
			return c.Status(400).SendString(err.Error())
		}

		return c.JSON(data)
	})

	// Start the server
	addr := fmt.Sprintf("%v:%v", app.Conf().Api.Host, app.Conf().Api.Port)
	log.Println("Starting HTTP server on " + addr)

	if err := api.Listen(addr); err != nil {
		log.Fatalf("failed to start HTTP server: %v", err)
	}
}

// on exit, kill the app running on port 8080
// kill -9 $(lsof -t -i:4444)
