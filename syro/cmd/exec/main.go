package main

import (
	"context"
	"log"
	"syro/internal/pooler/services/binance"
	"syro/pkg/app"
)

// go run cmd/exec/main.go
func main() {
	app, err := app.New(context.Background())
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer app.Exit(context.Background())

	s := binance.New(app, 2)
	s.Tmp()
}
