package main

import (
	"binance-pooler/pkg/core"
	"context"
	"log"
)

// go run cmd/exec/main.go
func main() {
	app, err := core.NewApp(context.Background())
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer app.Exit(context.Background())
}
