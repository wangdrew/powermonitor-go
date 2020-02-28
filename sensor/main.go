package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/wangdrew/powermonitor-go/models"
	"log"
	"os"
	"time"
)

func main() {
	ctx := context.Background()
	log.Println("starting power monitor")
	h, err := os.Hostname()
	fmt.Println(err)
	fmt.Println(h)

	var influxURL, influxToken, influxBucket, influxOrg, serialPath string
	var mqttUrl, mqttTopic, mqttUser, mqttPass string
	// fixme: env vars can override these
	flag.StringVar(&influxURL, "influxUrl", "https://us-west-2-1.aws.cloud2.influxdata.com",
		"influx cloud instance URL")
	flag.StringVar(&influxToken, "influxToken", "", "influx cloud authorization token")
	flag.StringVar(&influxBucket, "influxBucket", "", "influx cloud bucket name")
	flag.StringVar(&influxOrg, "influxOrg", "", "influx cloud organization name")
	flag.StringVar(&serialPath, "serialPath", "/dev/ttyUSB0",
		"serial device system filepath that the ECM1240 is connected to")
	flag.StringVar(&mqttUrl, "mqttUrl", "", "MQTT URL hostname")
	flag.StringVar(&mqttTopic, "mqttTopic", "", "MQTT broker topic")
	flag.StringVar(&mqttUser, "mqttUser", "", "MQTT username")
	flag.StringVar(&mqttPass, "mqttPass", "", "MQTT username")
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
	//mqttUrl := "tcp://192.168.69.95:1883"
	//mqttTopic := "test-topic"
	//mqttUser := ""
	//mqttPass := ""
	mqttClientID := "foobar1"
	mqttOutput := NewMqtt(mqttUrl, mqttClientID, mqttTopic, mqttUser, mqttPass)
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
