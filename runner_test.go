package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/wangdrew/powermonitor-go/models"
	"testing"
	"time"
)

type MockSource struct {
	NumReadCalls int
}

func (s *MockSource) Init() error { return nil }

func (s *MockSource) Read() (models.PowerMetrics, error) {
	s.NumReadCalls += 1
	return models.PowerMetrics{{}}, nil
}

func TestRunner_Run(t *testing.T) {
	n := 10
	source := &MockSource{}
	timer := NewExternalTimer()
	output := make(chan models.PowerMetrics, n)
	r := Runner{
		Source: source,
		Timer:  timer,
		Output: output,
	}
	go func() { r.Run() }()
	for i := 0; i < n; i++ {
		timer.Trig <- time.Now()
	}
	time.Sleep(100 * time.Millisecond) // allow Run to output before closing the channel
	close(output)

	assert.Equal(t, 10, source.NumReadCalls)
	for len(output) > 0 {
		assert.Equal(t, <-output, models.PowerMetrics{{}})
	}
	close(timer.Stop())
}
