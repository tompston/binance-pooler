package main

import (
	"binance-pooler/pkg/app"
	"context"
	"log"
)

// go run cmd/exec/main.go
func main() {
	app, err := app.New(context.Background())
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer app.Exit(context.Background())
}
