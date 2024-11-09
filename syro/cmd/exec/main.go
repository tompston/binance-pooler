package main

import (
	"context"
	"log"
	"syro/internal/pooler/services/binance_service"
	"syro/pkg/app"
)

// go run cmd/exec/main.go
func main() {
	app, err := app.New(context.Background())
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer app.Exit(context.Background())

	s := binance_service.New(app, 1).WithDebugMode()
	s.Tmp(true)

	// t1 := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	// t2 := time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC)
	// interval := timeset.MilisToDuration(900000)

	// timeset.ChunkTimeRange(t1, t2, interval, 900, 10)
	// for _, chunk := range gapChunks {
	// 	log.Printf("chunk: %s - %s", chunk.From, chunk.To)
	// }
}
