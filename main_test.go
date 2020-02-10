package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTrim(t *testing.T) {
	input := []byte{128, 254, 255, 3, 28, 45, 254, 255, 3}
	assert.Equal(t, []byte{254, 255, 3, 28, 45}, trim(input))

	input = []byte{254, 255, 3, 254, 255, 3}
	assert.Equal(t, []byte{254, 255, 3}, trim(input))

	input = []byte{254, 255, 3}
	assert.Equal(t, []byte{}, trim(input))

	input = []byte{}
	assert.Equal(t, []byte{}, trim(input))
}