package main

import (
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"syro/pkg/lib/logger"
	"syro/pkg/lib/timeset"

	"time"
)

/*
NOTE: need to install this -> https://github.com/brendangregg/FlameGraph
  - mac: brew install flamegraph
  - linux: ?

// build the binary and run it to generate the cpu profile
go build -o out cmd/scripts/flamegraph/main.go
./out

// create a file which contains the profile in text format (%age breakdown)
go tool pprof -text -nodecount=1000 ./out cpu.prof > out.txt

// create and open the flamegraph svg
go tool pprof -raw -output=cpu.txt ./out cpu.prof
stackcollapse-go.pl cpu.txt | flamegraph.pl > flame.svg && open flame.svg
*/
func main() {
	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

	now := time.Now()
	fmt.Println("Hello, World!")

	// app, err := app.New(context.Background())
	// if err != nil {
	// 	msg := fmt.Sprintf("failed to create app in go pooler: %v", err.Error())
	// 	log.Fatalf(msg)
	// }
	// defer app.Exit(context.Background())

	// for i := 0; i < 1_000; i++ {
	// 	app.Logger().SetEvent("main").Info("Hello, World!")
	// }

	logger := logger.NewConsoleLogger(nil)
	for i := 0; i < 1_000; i++ {
		logger.SetEvent("main").Info("Hello, World!")
	}

	fmt.Printf("execution time: %v\n", timeset.SecSince(now))
}
