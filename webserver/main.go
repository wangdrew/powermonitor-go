package main

import (
	"encoding/json"
	"flag"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/wangdrew/powermonitor-go/models"
	"log"
	"net/http"
	"os"
	"sync"
)

func main() {
	var port int
	var mqttURL, mqttTopic, mqttPass, mqttUser string
	flag.IntVar(&port, "port", 8081, "port to bind the webserver to")
	flag.StringVar(&mqttURL, "mqttURL", "tcp://localhost:1883", "mqtt broker url")
	flag.StringVar(&mqttTopic, "mqttTopic", "test-topic", "mqtt topic name")
	flag.StringVar(&mqttUser, "mqttUsername", "", "mqtt username")
	flag.StringVar(&mqttPass, "mqttPassword", "", "mqtt password")
	flag.Parse()

	opts := mqtt.NewClientOptions().AddBroker(mqttURL).SetClientID("power-subscriber") // todo
	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error()) // fixme
	}

	msgHandler := func(cl mqtt.Client, msg mqtt.Message) {
		var metric models.PowerMetric
		if err := json.Unmarshal(msg.Payload(), &m); err != nil {
			log.Printf("error deserializing MQTT message: %v", err)
			return
		}
		if metric.SensorName == "" {
			log.Printf("metric missing sensorname, payload: %s", string(msg.Payload()))
		}
		m.UpdatePower(metric.SensorName, metric)
	}

	if token := c.Subscribe(mqttTopic, 0, msgHandler); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1) //fixme
	}

	http.HandleFunc("/metrics", getMetrics)
	fmt.Printf("starting server on :%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

// In-memory metrics store
var m = &Metrics{
	Power: make(map[string]models.PowerMetric),
}

type Metrics struct {
	sync.RWMutex
	Power map[string]models.PowerMetric
}

func (m *Metrics) UpdatePower(key string, value models.PowerMetric) {
	m.Lock()
	defer m.Unlock()
	m.Power[key] = value
}

func (m *Metrics) ToJson() ([]byte, error) {
	m.RLock()
	defer m.RUnlock()
	return json.Marshal(m)
}

func getMetrics(w http.ResponseWriter, r *http.Request) {
	resp, err := m.ToJson()
	if err != nil {
		log.Printf("error serializing metrics to JSON: %v", err)
	}
	fmt.Fprintf(w, string(resp))
}
