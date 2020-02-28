package main

import (
	"encoding/json"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/wangdrew/powermonitor-go/models"
	"log"
)

type MqttOutput struct {
	Client          MqttClient
	topic           string
	connectionToken mqtt.Token
	retained        bool
	qos             int
}

type MqttClient interface {
	Connect() mqtt.Token
	Disconnect(quiesce uint)
	Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token
}

func (me *MqttOutput) Start(metrics chan models.PowerMetrics, stop chan struct{}) {
	for {
		select {
		case <-stop:
			me.Client.Disconnect(250) // wait 250ms for existing work to be completed
			return
		case m := <-metrics:
			if me.connectionToken == nil {
				if err := me.Connect(); err != nil {
					log.Printf("error connecting to mqtt: %v", err)
				}
			}
			for _, dp := range m {
				msg, err := me.mapMetrics(dp)
				if err != nil {
					log.Printf("error serializing powerMetric to JSON: %+v", err)
					break
				}
				tok := me.Client.Publish(me.topic, byte(me.qos), me.retained, msg)
				if err := tok.Error(); tok.Wait() && err != nil {
					log.Printf("error writing metric to mqtt: %+v", err)
				}
			}
		}
	}
}

func (me *MqttOutput) Connect() error {
	token := me.Client.Connect()
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	me.connectionToken = token
	return nil
}

func NewMqtt(url, clientID, topic, username, password string) *MqttOutput {
	return &MqttOutput{
		Client: mqtt.NewClient(
			mqtt.NewClientOptions().
				AddBroker(url).
				SetClientID(clientID).
				SetUsername(username).
				SetPassword(password)),
		topic:    topic,
		retained: false, // todo: maybe this should be configurable?
		qos:      0,
	}
}

// mapMetrics serializes into JSON
func (me *MqttOutput) mapMetrics(metrics models.PowerMetric) ([]byte, error) {
	return json.Marshal(metrics)
}
