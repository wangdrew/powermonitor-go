package main

import "github.com/wangdrew/powermonitor-go/models"

type Source interface {
	Init() error
	Read() (models.PowerMetrics, error)
}
