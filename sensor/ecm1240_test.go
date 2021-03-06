package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const delta = 1e-3 // used to assert floating-point "equality"

func TestECM1240Source_Read(t *testing.T) {
	s := ECM1240Source{Name: "power", Port: &fakeSerial{}}
	wattSecParsers = map[string]parsers{"1": wattSec1}
	metrics, err := s.Read()
	assert.Nil(t, err)
	assert.Len(t, metrics, 1)
	assert.InDelta(t, 0.0, metrics[0].PowerW, delta)
	assert.InDelta(t, 100.0, metrics[0].EnergyWs, delta)
	assert.Equal(t, "power_1", metrics[0].SensorName)
	assert.InDelta(t, 120.3, metrics[0].VoltageV, delta)
	assert.NotEmpty(t, metrics[0].Ts)

	// 2nd metric
	dataFrame2 := sampleDataFrame
	dataFrame2[37] += 1  // advance clock by 1 second
	dataFrame2[5] += 100 // add 100 wattsecs

	metrics, err = s.Read()
	assert.Nil(t, err)
	assert.Len(t, metrics, 1)
	assert.InDelta(t, 100.0, metrics[0].PowerW, delta)   // (200-100) wattsec / 1 sec = 100 watts
	assert.InDelta(t, 200.0, metrics[0].EnergyWs, delta) // 200 wattsec
	assert.Equal(t, "power_1", metrics[0].SensorName)
	assert.InDelta(t, 120.3, metrics[0].VoltageV, delta)
	assert.NotEmpty(t, metrics[0].Ts)
}

func TestTrim(t *testing.T) {
	input := []byte{127, 254, 255, 3, 28, 45, 254, 255, 3}
	assert.Equal(t, []byte{254, 255, 3, 28, 45}, trim(input))

	input = []byte{254, 255, 3, 254, 255, 3}
	assert.Equal(t, []byte{254, 255, 3}, trim(input))

	input = []byte{253, 255, 3}
	assert.Equal(t, []byte{}, trim(input))

	input = []byte{}
	assert.Equal(t, []byte{}, trim(input))
}

func TestReadAsFloat(t *testing.T) {
	_, err := readAsFloat([]byte{}, 3, 5, 1.0, false)
	assert.Error(t, err, "dataframe (len=0) is missing bytes 3 - 5")

	val, err := readAsFloat(sampleDataFrame, 3, 5, 0.1, false)
	assert.Nil(t, err)
	assert.InDelta(t, 120.3, val, delta)

	val, err = readAsFloat(sampleDataFrame, 3, 5, 0.1, true)
	assert.Nil(t, err)
	assert.InDelta(t, 4582.8003, val, delta)
}

func TestPadZeros(t *testing.T) {
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0}, padZeros([]byte{}, false))
	assert.Equal(t, []byte{1, 2, 3, 0, 0, 0, 0, 0}, padZeros([]byte{1, 2, 3}, true))
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 1, 2, 3}, padZeros([]byte{1, 2, 3}, false))
}

func TestVoltage(t *testing.T) {
	val, err := voltage(sampleDataFrame)
	assert.Nil(t, err)
	assert.InDelta(t, 120.3, val, delta)
}

var sampleDataFrame = []byte{
	254, 255, 3, // start byte sequence
	4, 179, // voltage
	100, 0, 0, 0, 0, //ch1 ws
	73, 222, 246, 1, 0, // ch2 ws
	194, 65, 114, 128, 0, // ch1 polarized ws
	94, 21, 217, 0, 157, // ch2 polarized ws
	0, 0, 12, 1, 0, 0, 152, 0, // reserved, serial number, unit info, pre-programmed stuff
	170, 2, // ch1 current
	28, 0, // ch2 current
	0, 0, 0, // device clock
	32, 209, 133, 39, // aux1 ws
	129, 97, 226, 11, // aux2 ws
	79, 118, 17, 20, // aux3 ws
	2, 0, 0, 0, // aux4 ws
	17, 0, 0, 0, // aux5 ws
	255, 1, // dc voltage
	255, 254, 40, // end frame and checksum
	254, 255, 3, 2, // start byte sequence
}

type fakeSerial struct{}

func (me *fakeSerial) Read(buf []byte) (int, error) {
	for i, v := range sampleDataFrame {
		buf[i] = v
	}
	return len(sampleDataFrame), nil
}
