package main

import (
	"context"
	"log"
	"syro/pkg/app"
	"syro/pkg/lib/logbook"
)

// go run cmd/exec/main.go
func main() {
	app, err := app.New(context.Background())
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer app.Exit(context.Background())

	// s := binance.New(app, 2)
	// s.Tmp()

	logger := logbook.NewConsoleLogger(nil).
		SetSource("main").
		SetEvent("sub-main")

	numLogs := 10
	for i := 0; i < numLogs; i++ {
		logger.Debug("this is a test")
	}
}
