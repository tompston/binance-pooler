package main

import (
	"binance-pooler/pkg/app"
	"binance-pooler/pkg/syro"
	"context"
	"fmt"
	"log"
	"strconv"

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
	appServer := fiber.New(fiber.Config{
		DisableStartupMessage: true, // Disable Fiber's startup message
	})

	appServer.Use(
		cors.New(cors.Config{
			AllowCredentials: false,
			AllowOrigins:     "*", // NOTE: change this to a list of urls from which fetch requests are allowed
		}),
	)

	// Define routes
	appServer.Get("/logs", func(c *fiber.Ctx) error {
		filter := syro.LogFilter{
			Limit: 500,
			// Level: ,
		}
		_ = filter

		// Call the syro HTMX handler and pass the Fiber context
		return c.Type("text/html").SendString("")
	})

	// Start the server
	addr := fmt.Sprintf("%v:%v", app.Conf().Api.Host, app.Conf().Api.Port)
	log.Println("Starting HTTP server on " + addr)

	if err := appServer.Listen(addr); err != nil {
		log.Fatalf("failed to start HTTP server: %v", err)
	}
}

func stringToFloatPointer(s string) *float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &f
}

// on exit, kill the app running on port 8080
// kill -9 $(lsof -t -i:4444)
