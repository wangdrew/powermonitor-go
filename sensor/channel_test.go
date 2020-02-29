package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/wangdrew/powermonitor-go/models"
	"testing"
	"time"
)

func TestClone(t *testing.T) {
	n := 3
	numMsg := 10
	in := make(chan models.PowerMetrics, numMsg)
	cloned := Clone(in, n)
	assert.Len(t, cloned, n)
	for i := 0; i < numMsg; i++ {
		in <- models.PowerMetrics{{}}
	}
	time.Sleep(50 * time.Millisecond) // give goroutine time to run
	for _, ch := range cloned {
		assert.Len(t, ch, numMsg)
	}
}

func TestFullChannel(t *testing.T) {
	n := 2
	numMsg := 2
	in := make(chan models.PowerMetrics, numMsg)
	cloned := Clone(in, n)
	// fill up one of the output channels
	cloned[0] <- models.PowerMetrics{}
	cloned[0] <- models.PowerMetrics{}
	assert.Len(t, cloned[0], 2)

	// the other output channel should continue to fill up
	in <- models.PowerMetrics{}
	time.Sleep(50 * time.Millisecond)
	assert.Len(t, cloned[1], 1)
	in <- models.PowerMetrics{}
	time.Sleep(50 * time.Millisecond)
	assert.Len(t, cloned[1], 2)
}
