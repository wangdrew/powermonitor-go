package main

import "github.com/wangdrew/powermonitor-go/models"

type Output interface {
	Start(metrics chan models.PowerMetrics, stop chan struct{})
}
