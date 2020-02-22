package main

import (
	"context"
	"github.com/influxdata/influxdb-client-go"
	"github.com/wangdrew/powermonitor-go/models"
	"log"
)

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
				log.Printf("error closing output client: %+v", err)
			}
			return
		case m := <-metrics:
			_, err := me.Client.Write(context.Background(), me.bucket, me.org, me.mapMetrics(m)...)
			if err != nil {
				log.Printf("error writting metric to influx: %+v", err)
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

func (me *InfluxOutput) mapMetrics(metrics models.PowerMetrics) []influxdb.Metric {
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
