package main

import (
	"context"
	"fmt"
	"log"
	"syro/internal/pooler/services/binance_service"
	"syro/pkg/app"
	"syro/pkg/sy"
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
	storage := app.CronStorage()
	scheduler := sy.NewCronScheduler(cron, "go-pooler").WithStorage(storage)

	if err := binance_service.New(app, 3).
		WithDebugMode().
		AddJobs(scheduler); err != nil {
		log.Fatalf("failed to add binance jobs: %v", err)
	}

	// fmt.Printf("scheduler.Jobs: %#v\n", scheduler.Jobs)
	// fmt.Printf("format string")
	scheduler.Start()
	select {} // run forever

	// for _, job := range scheduler.Jobs {
	// 	fmt.Println(job.Freq, job.Name, job.Func)
	// }
}
