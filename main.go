package main

import (
	"context"
	"github.com/wangdrew/powermonitor-go/models"
	"log"
	"os"
	"time"
)

func main() {
	ctx := context.Background()
	influxURL := ""
	influxToken := ""
	influxBucket := ""
	influxOrg := ""

	serialPath := os.Getenv("SERIAL_PATH")
	if serialPath == "" {
		serialPath = "/dev/ttyUSB0"
	}

	timer := NewTickerTimer(2000 * time.Millisecond)
	metrics := make(chan models.PowerMetrics, 100)
	stopOutput := make(chan struct{})
	output, err := NewInflux(influxURL, influxToken, influxOrg, influxBucket)
	if err != nil {
		log.Fatal(err)
	}
	source := NewECM1240Source("power", serialPath)
	if err := source.Init(); err != nil {
		log.Fatal(err)
	}
	go func() {
		<-ctx.Done()
		close(timer.Stop())
		close(stopOutput)
	}()

	go func() {
		output.Start(metrics, stopOutput)
	}()

	NewRunner(source, timer, metrics).Run()
}
