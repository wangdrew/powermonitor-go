package main

import (
	"github.com/wangdrew/powermonitor-go/models"
)

// Clone replicates messages from the in channel into output channels which are returned in a slice
func Clone(in chan models.PowerMetrics, numCopies int) []chan models.PowerMetrics {
	bufLen := cap(in)
	out := make([]chan models.PowerMetrics, numCopies)
	for i, _ := range out {
		out[i] = make(chan models.PowerMetrics, bufLen)
	}
	go func() {
		for {
			msg := <-in
			for _, c := range out {
				c <- msg
			}
		}
	}()
	return out
}
