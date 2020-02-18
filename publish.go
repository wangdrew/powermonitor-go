package main

import "github.com/wangdrew/powermonitor-go/models"

type Publisher interface {
	Publish(<-chan models.PowerMetric)
}
