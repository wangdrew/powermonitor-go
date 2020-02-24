package models

import "time"

type PowerMetric struct {
	SensorName string    `json:"sensorName"`
	Ts         time.Time `json:"ts"`
	PowerW     float64   `json:"powerW"`
	EnergyWs   float64   `json:"energyWs"`
	VoltageV   float64   `json:"voltageV"`
}

type PowerMetrics []PowerMetric
