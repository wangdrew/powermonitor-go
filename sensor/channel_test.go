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
	time.Sleep(100 * time.Millisecond) // give goroutine time to run
	for _, ch := range cloned {
		assert.Len(t, ch, numMsg)
	}
}
