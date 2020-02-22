package main

import (
	"context"
	"flag"
	"github.com/wangdrew/powermonitor-go/models"
	"log"
	"time"
)

func main() {
	ctx := context.Background()
	log.Println("starting power monitor")

	var influxURL, influxToken, influxBucket, influxOrg, serialPath string
	flag.StringVar(&influxURL, "influxUrl", "https://us-west-2-1.aws.cloud2.influxdata.com",
		"influx cloud instance URL")
	flag.StringVar(&influxToken, "influxToken", "", "influx cloud authorization token")
	flag.StringVar(&influxBucket, "influxBucket", "", "influx cloud bucket name")
	flag.StringVar(&influxOrg, "influxOrg", "", "influx cloud organization name")
	flag.StringVar(&serialPath, "serialPath", "/dev/ttyUSB0",
		"serial device system filepath that the ECM1240 is connected to")
	flag.Parse()

	timer := NewTickerTimer(2000 * time.Millisecond) // poll ECM-1240 every 2 seconds
	metrics := make(chan models.PowerMetrics, 100)   // metrics buffer decoupling source from output sink
	stopOutput := make(chan struct{})
	metricStreams := Clone(metrics, 2)

	// influx
	influxOutput, err := NewInflux(influxURL, influxToken, influxOrg, influxBucket)
	if err != nil {
		log.Fatal(err)
	}

	// mqtt
	// fixme: parameterize main with these
	mqttUrl := "tcp://192.168.69.95:1883"
	mqttClientID := "foo"
	mqttTopic := "test-topic"
	mqttUser := ""
	mqttPass := ""
	mqttOutput, err := NewMqtt(mqttUrl, mqttClientID, mqttTopic, mqttUser, mqttPass)
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
		influxOutput.Start(metricStreams[0], stopOutput)
	}()
	go func() {
		mqttOutput.Start(metricStreams[1], stopOutput)
	}()

	NewRunner(source, timer, metrics).Run()
}
