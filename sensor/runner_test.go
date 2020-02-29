package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/wangdrew/powermonitor-go/models"
	"testing"
	"time"
)

type MockSource struct {
	NumReadCalls int
	ReturnArgs   []interface{}
}

func (s *MockSource) Init() error { return nil }

func (s *MockSource) Read() (models.PowerMetrics, error) {
	s.NumReadCalls += 1
	if s.ReturnArgs[1] == nil {
		return s.ReturnArgs[0].(models.PowerMetrics), nil
	}
	return s.ReturnArgs[0].(models.PowerMetrics), s.ReturnArgs[1].(error)
}

func BeforeEach(numMetrics int) (*MockSource, *ExternalTimer, chan models.PowerMetrics, Runner) {
	source := &MockSource{
		ReturnArgs: []interface{}{models.PowerMetrics{{}}, nil},
	}
	timer := NewExternalTimer()
	output := make(chan models.PowerMetrics, numMetrics)
	r := Runner{
		Source: source,
		Timer:  timer,
		Output: output,
	}
	return source, timer, output, r
}

func TestRunner_Run(t *testing.T) {
	n := 10
	source, timer, output, r := BeforeEach(n)
	go func() { r.Run() }()
	for i := 0; i < n; i++ {
		timer.Trig <- time.Now()
	}
	time.Sleep(50 * time.Millisecond) // allow Runner time to run before closing the channel
	close(output)

	assert.Equal(t, 10, source.NumReadCalls)
	for len(output) > 0 {
		assert.Equal(t, <-output, models.PowerMetrics{{}})
	}
	close(timer.Stop())
}

func TestRunner_ChannelFull(t *testing.T) {
	_, timer, output, r := BeforeEach(1)
	go func() { r.Run() }()

	timer.Trig <- time.Now()
	time.Sleep(50 * time.Millisecond) // allow Runner time to run
	assert.Len(t, output, 1)          // output channel should still be full
	<-output
	assert.Len(t, output, 0) // output channel should be clear
	timer.Trig <- time.Now()
	time.Sleep(50 * time.Millisecond) // allow Runner time to run
	assert.Len(t, output, 1)          // runner should persist new metric
	close(timer.Stop())
}

func TestRunner_NoMetricsFromSource(t *testing.T) {
	source, timer, output, r := BeforeEach(1)
	go func() { r.Run() }()
	source.ReturnArgs[0] = models.PowerMetrics{} // source mock will return empty metrics

	timer.Trig <- time.Now()
	time.Sleep(50 * time.Millisecond)       // allow Runner time to run
	assert.Equal(t, source.NumReadCalls, 1) // source.Read should be called
	assert.Len(t, output, 0)                // output channel should be empty
}
