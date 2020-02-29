package main

import (
	"context"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/wangdrew/powermonitor-go/models"
	"log"
	"os"
	"time"
)

type Config struct {
	SerialPath   string `default:"/dev/ttyUSB0"`
	EnableInflux bool   `default:"true"`
	EnableMqtt   bool   `default:"true"`
	InfluxURL    string `default:"https://us-west-2-1.aws.cloud2.influxdata.com"`
	InfluxToken  string
	InfluxBucket string
	InfluxOrg    string
	MqttURL      string `default:"tcp://mqtt:1883"`
	MqttTopic    string `default:"power"`
	MqttUser     string `default:""`
	MqttPass     string `default:""`
}

func main() {
	log.Println("starting power monitor")
	var (
		ctx = context.Background()
		c   = Config{}
	)
	if err := envconfig.Process("", &c); err != nil {
		log.Fatal(err)
	}
	log.Printf("configuration: %+v\n", c)

	timer := NewTickerTimer(2000 * time.Millisecond) // poll ECM-1240 every 2 seconds
	metrics := make(chan models.PowerMetrics, 100)   // metrics buffer decoupling source from output sink
	stopOutput := make(chan struct{})
	metricStreams := Clone(metrics, 2)

	// stop outputs and source on ctx.Done signal
	go func() {
		<-ctx.Done()
		close(timer.Stop())
		close(stopOutput)
	}()

	if c.EnableInflux {
		influxOutput, err := NewInflux(c.InfluxURL, c.InfluxToken, c.InfluxOrg, c.InfluxBucket)
		if err != nil {
			log.Fatal(err)
		}
		go func() {
			influxOutput.Start(metricStreams[0], stopOutput)
		}()
	}

	if c.EnableMqtt {
		h, _ := os.Hostname()
		mqttClientID := fmt.Sprintf("powersensor-%s", h)
		mqttOutput := NewMqtt(c.MqttURL, mqttClientID, c.MqttTopic, c.MqttUser, c.MqttPass)
		if err != nil {
			log.Fatal(err)
		}
		go func() {
			mqttOutput.Start(metricStreams[1], stopOutput)
		}()
	}

	source := NewECM1240Source("power", c.SerialPath)
	if err := source.Init(); err != nil {
		log.Fatal(err)
	}
	NewRunner(source, timer, metrics).Run()
}
