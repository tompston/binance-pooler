package main

import (
	"context"
	"fmt"
	"log"
	"syro/internal/pooler/services/binance"
	"syro/pkg/app"
	"syro/pkg/lib/scheduler"
	"time"

	"github.com/robfig/cron/v3"
)

// go run cmd/pooler/main.go
func main() {
	loc, err := time.LoadLocation("Europe/Riga")
	if err != nil {
		log.Fatalf(err.Error())
	}

	app, err := app.New(context.Background())
	if err != nil {
		msg := fmt.Sprintf("failed to create app in go pooler: %v", err.Error())
		log.Fatalf(msg)
	}
	defer app.Exit(context.Background())

	cron := cron.New(cron.WithLocation(loc))
	scheduler := scheduler.NewScheduler(cron, app.CronStorage())

	binance.New(app, 3).Run(scheduler)

	scheduler.Start()
	select {} // run forever
}
