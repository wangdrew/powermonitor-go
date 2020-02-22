package main

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
	"github.com/wangdrew/powermonitor-go/models"
	"testing"
	"time"
)

func TestMqttOutput(t *testing.T) {
	n := 5
	cl := MockMqtt{}
	output := MqttOutput{Client: &cl}
	metrics := make(chan models.PowerMetrics, n)
	stop := make(chan struct{})

	go func() { output.Start(metrics, stop) }()
	for i := 0; i < n; i++ {
		metrics <- models.PowerMetrics{{}, {}} // publish two measurements per metric
	}
	time.Sleep(100 * time.Millisecond) // give output goroutine time to write to its client
	assert.Equal(t, 10, cl.NumPublishCalled)
	assert.Len(t, cl.PublishCalledWith, 10)
	close(stop)
	time.Sleep(100 * time.Millisecond) // give output goroutine time to stop
	assert.Equal(t, 1, cl.NumDisconnectCalled)
}

type MockMqtt struct {
	NumConnectCalled    int
	NumDisconnectCalled int
	NumPublishCalled    int
	PublishCalledWith   []interface{}
}

func (me *MockMqtt) Connect() mqtt.Token {
	me.NumConnectCalled += 1
	return &mqtt.DummyToken{}
}

func (me *MockMqtt) Disconnect(quiesce uint) {
	me.NumDisconnectCalled += 1
}

func (me *MockMqtt) Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	me.NumPublishCalled += 1
	me.PublishCalledWith = append(me.PublishCalledWith, payload)
	return &mqtt.DummyToken{}
}
