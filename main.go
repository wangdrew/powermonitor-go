package main

import (
	"context"
	"fmt"
	"github.com/wangdrew/powermonitor-go/models"
	"log"
	"os"
	"time"
)

func main() {
	ctx := context.Background()
	serialPath := os.Getenv("SERIAL_PATH")
	if serialPath == "" {
		serialPath = "/dev/ttyUSB0"
	}

	timer := NewTickerTimer(2000 * time.Millisecond)
	output := make(chan models.PowerMetrics, 100)
	source := NewECM1240Source("power", serialPath)
	if err := source.Init(); err != nil {
		log.Fatal(err)
	}
	go func() {
		<-ctx.Done()
		close(timer.Stop())
	}()

	go func() {
		for {
			fmt.Println(<-output)
		}
	}()

	NewRunner(source, timer, output).Run()
}
