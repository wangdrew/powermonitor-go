package main

import (
	"context"
	"fmt"
	"github.com/influxdata/influxdb-client-go"
	"github.com/wangdrew/powermonitor-go/models"
)

type Output interface {
	Start(metrics chan models.PowerMetrics, stop chan struct{})
}

type InfluxOutput struct {
	Client InfluxClient
	bucket string
	org    string
}

type InfluxClient interface {
	Write(ctx context.Context, bucket, org string, m ...influxdb.Metric) (n int, err error)
	Close() error
}

func (me *InfluxOutput) Start(metrics chan models.PowerMetrics, stop chan struct{}) {
	for {
		select {
		case <-stop:
			if err := me.Client.Close(); err != nil {
				fmt.Errorf("%+v", err)
			}
			return
		case m := <-metrics:
			_, err := me.Client.Write(context.Background(), me.bucket, me.org, mapMetrics(m)...)
			if err != nil {
				fmt.Errorf("%+v", err) // continue on error
			}
		}
	}
}

func NewInflux(url, token, org, bucket string) (Output, error) {
	influx, err := influxdb.New(url, token)
	if err != nil {
		return nil, err
	}
	return &InfluxOutput{
		Client: influx,
		bucket: bucket,
		org:    org,
	}, nil
}

func mapMetrics(metrics models.PowerMetrics) []influxdb.Metric {
	ret := make([]influxdb.Metric, 0)
	for _, m := range metrics {
		ret = append(ret,
			influxdb.NewRowMetric(
				map[string]interface{}{
					"voltageV": m.VoltageV,
					"energyWs": m.EnergyWs,
					"powerW":   m.PowerW,
				},
				"power-metrics",
				map[string]string{
					"sensorName": m.SensorName,
				},
				m.Ts),
		)
	}
	return ret
}
