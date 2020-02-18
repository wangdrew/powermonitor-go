package main

import (
	"context"
	"github.com/influxdata/influxdb-client-go"
	protocol "github.com/influxdata/line-protocol"
	"github.com/stretchr/testify/assert"
	"github.com/wangdrew/powermonitor-go/models"
	"testing"
	"time"
)

func TestInfluxOutput(t *testing.T) {
	client := &MockInflux{}
	output := InfluxOutput{
		Client: client,
	}
	n := 10
	metrics := make(chan models.PowerMetrics, n)
	stop := make(chan struct{})
	go func() { output.Start(metrics, stop) }()

	for i := 0; i < n; i++ {
		metrics <- models.PowerMetrics{{}}
	}
	time.Sleep(100 * time.Millisecond) // give output goroutine time to write to its client
	assert.Equal(t, 10, client.NumWriteCalled)
	assert.Equal(t, 10, client.NumMetricsReceived)

	close(stop)
	time.Sleep(100 * time.Millisecond) // give output goroutine time to return
	assert.Equal(t, 1, client.NumCloseCalled)
}

func TestMapMetrics(t *testing.T) {
	ts := time.Date(2000, 01, 01, 00, 00, 00, 00, time.UTC)
	actual := mapMetrics(models.PowerMetrics{{
		VoltageV:   120,
		PowerW:     100,
		EnergyWs:   200,
		SensorName: "foo",
		Ts:         ts,
	}})
	expected := &influxdb.RowMetric{
		NameStr: "power-metrics",
		// influxdb library seems to alphabetize these fields. Order matters!
		Fields: []*protocol.Field{
			{
				Key:   "energyWs",
				Value: float64(200),
			},
			{
				Key:   "powerW",
				Value: float64(100),
			},
			{
				Key:   "voltageV",
				Value: float64(120),
			},
		},
		Tags: []*protocol.Tag{
			{
				Key:   "sensorName",
				Value: "foo",
			},
		},
		TS: ts,
	}
	assert.Len(t, actual, 1)
	assert.Equal(t, expected, actual[0].(*influxdb.RowMetric))
}

type MockInflux struct {
	NumWriteCalled     int
	NumCloseCalled     int
	NumMetricsReceived int
}

func (me *MockInflux) Write(ctx context.Context, bucket, org string, m ...influxdb.Metric) (n int, err error) {
	me.NumWriteCalled += 1
	me.NumMetricsReceived += len(m)
	return len(m), nil
}

func (me *MockInflux) Close() error {
	me.NumCloseCalled += 1
	return nil
}
