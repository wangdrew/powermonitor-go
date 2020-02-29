package main

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/kelseyhightower/envconfig"
	"github.com/wangdrew/powermonitor-go/models"
	"log"
	"net/http"
	"os"
	"sync"
)

type Config struct {
	Port      int    `default:"8081"`
	MqttURL   string `default:"tcp://mqtt:1883"`
	MqttTopic string `default:"power"`
	MqttUser  string `default:""`
	MqttPass  string `default:""`
}

func main() {
	var c = Config{}
	if err := envconfig.Process("", &c); err != nil {
		log.Fatal(err)
	}
	log.Printf("configuration: %+v\n", c)

	opts := mqtt.NewClientOptions().AddBroker(c.MqttURL).SetClientID("power-subscriber") // todo
	cl := mqtt.NewClient(opts)
	if token := cl.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error()) // fixme
	}

	msgHandler := func(cl mqtt.Client, msg mqtt.Message) {
		var metric models.PowerMetric
		if err := json.Unmarshal(msg.Payload(), &metric); err != nil {
			log.Printf("error deserializing MQTT message: %v", err)
			return
		}
		if metric.SensorName == "" {
			log.Printf("metric missing sensorname, payload: %s", string(msg.Payload()))
			return
		}
		m.UpdatePower(metric.SensorName, metric)
	}

	if token := cl.Subscribe(c.MqttTopic, 0, msgHandler); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1) //fixme
	}

	http.HandleFunc("/metrics", getMetrics)
	fmt.Printf("starting power server on :%d\n", c.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", c.Port), nil))
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
