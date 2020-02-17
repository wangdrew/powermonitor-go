package main

import (
	"github.com/wangdrew/powermonitor-go/models"
	"log"
	"time"
)

type Runner struct {
	Source Source
	Timer  Timing
	Output chan<- models.PowerMetrics
}

func NewRunner(Source Source, Timer Timing, Output chan<- models.PowerMetrics) *Runner {
	return &Runner{
		Source: Source,
		Timer:  Timer,
		Output: Output,
	}
}

// Run uses the current go-routine to read PowerMetrics from a Source and send
// it on the Runner's Output channel
func (me *Runner) Run() {
	for {
		select {
		case <-me.Timer.Stop():
			return
		case <-me.Timer.Trigger():
			metrics, err := me.Source.Read()
			if err != nil {
				log.Printf("error reading metric from source: %+v", err)
			}
			if len(metrics) == 0 {
				continue // metrics are empty
			}
			select {
			case me.Output <- metrics:
			default:
				log.Println("output channel is full, skipping metric")
			}
		}
	}
}

type Timing interface {
	Stop() chan struct{}
	Trigger() <-chan time.Time
}

type TickerTimer struct {
	Ticker   *time.Ticker
	StopChan chan struct{}
}

func (t *TickerTimer) Stop() chan struct{}       { return t.StopChan }
func (t *TickerTimer) Trigger() <-chan time.Time { return t.Ticker.C }

func NewTickerTimer(interval time.Duration) Timing {
	return &TickerTimer{
		Ticker:   time.NewTicker(interval),
		StopChan: make(chan struct{}),
	}
}

type ExternalTimer struct {
	Trig     chan time.Time
	StopChan chan struct{}
}

func (e *ExternalTimer) Stop() chan struct{}       { return e.StopChan }
func (e *ExternalTimer) Trigger() <-chan time.Time { return e.Trig }

func NewExternalTimer() *ExternalTimer {
	return &ExternalTimer{
		Trig:     make(chan time.Time),
		StopChan: make(chan struct{}),
	}
}
