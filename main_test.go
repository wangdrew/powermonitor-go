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

var sampleDataFrame = []byte{
	254, 255, 3, 4, 179, 198, 62, 86, 153, 0, 73, 222, 246, 1, 0, 194, 65, 114, 128, 0, 94, 21, 217, 0, 157, 0, 0, 12, 1, 0, 0, 152, 0, 170, 2, 28, 0,
	59, 178, 194, 32, 209, 133, 39, 129, 97, 226, 11, 79, 118, 17, 20, 2, 0, 0, 0, 17, 0, 0, 0, 255, 1, 255, 254, 40, 254, 255, 3, 4, 179, 13, 63, 86,
	153, 0, 76, 222, 246, 1, 0, 194, 65, 114, 128, 0, 97, 21, 217, 0, 157, 0, 0, 12, 1, 0, 0, 144, 0, 148, 2, 28, 0, 60, 178, 194, 52, 209, 133, 39, 131,
	97, 226, 11, 231, 118, 17, 20, 2, 0, 0, 0, 17, 0, 0, 0, 255, 1, 255, 254, 7, 254, 255, 3, 4, 179, 85, 63, 86, 153, 0, 79, 222, 246, 1, 0, 10, 66, 114,
	128, 0,
}

func TestVoltage(t *testing.T) {
	val, err := voltage(sampleDataFrame)
	assert.Nil(t, err)
	assert.Equal(t, float32(120.3), val)
}
