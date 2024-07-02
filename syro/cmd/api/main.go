package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"syro/internal/api"
	"syro/pkg/app"

	"github.com/go-chi/chi/v5"
)

// go run cmd/pooler/main.go
func main() {
	app, err := app.New(context.Background())
	if err != nil {
		msg := fmt.Sprintf("failed to create app in go pooler: %v", err.Error())
		log.Fatalf(msg)
	}
	defer app.Exit(context.Background())

	r := chi.NewRouter()
	api.New(app, r).Routes()

	addr := fmt.Sprintf("%v:%v", app.Conf().Api.Host, app.Conf().Api.Port)

	log.Println("Starting HTTP server on " + addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("failed to start HTTP server: %v", err)
	}

	// on exit, kill the app running on port 8080
	// kill -9 $(lsof -t -i:8080)
}
