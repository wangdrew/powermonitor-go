package main

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/wangdrew/powermonitor-go/models"
)

type MqttOutput struct {
	Client MqttClient
	topic string
	connectionToken mqtt.Token
	retained bool
	qos int
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
			for _, dp := range m {
				me.Client.Publish(me.topic, byte(me.qos), me.retained, dp)
			}
		}
	}
}

func NewMqtt(url, clientID, topic, username, password string) (*MqttOutput, error) {
	cl := mqtt.NewClient(
		mqtt.NewClientOptions().
			AddBroker(url).
			SetClientID(clientID).
			SetUsername(username).
			SetPassword(password))

	token := cl.Connect()
	if token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return &MqttOutput{
		Client: cl,
		topic: topic,
		connectionToken: token,
		retained: false, // todo: maybe this should be configurable?
		qos: 0,
	}, nil
}
