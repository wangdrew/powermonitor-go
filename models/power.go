package models

import "time"

type PowerMetric struct {
	SensorName string
	Ts         time.Time
	PowerW     float32
	EnergyWs   float32
	VoltageV   float32
}

type PowerMetrics []PowerMetric
