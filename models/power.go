package models

import "time"

type PowerMetric struct {
	SensorName string
	Ts         time.Time
	PowerW     float64
	EnergyWs   float64
	VoltageV   float64
}

type PowerMetrics []PowerMetric
