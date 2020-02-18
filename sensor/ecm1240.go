package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/tarm/serial"
	"github.com/wangdrew/powermonitor-go/models"
	"time"
)
func NewECM1240Source(deviceName, serialPath string) Source {
	return &ECM1240Source{Name: deviceName, serialPath: serialPath}
}

type ECM1240Source struct {
	Name                string
	Port                SerialDevice
	serialPath          string
	previousMetrics     map[string]*models.PowerMetric
	previousDeviceClock float64
}

type SerialDevice interface {
	Read([]byte) (int, error)
}

func (me *ECM1240Source) Init() error {
	c := &serial.Config{
		Name:        me.serialPath,
		Baud:        19200,
		Size:        serial.DefaultSize,
		Parity:      serial.ParityNone,
		StopBits:    serial.Stop1,
		ReadTimeout: time.Millisecond * 1000}
	p, err := serial.OpenPort(c)
	if err != nil {
		return err
	}
	me.Port = p
	return nil
}

func (me *ECM1240Source) Read() (models.PowerMetrics, error) {
	if me.Port == nil {
		return nil, fmt.Errorf("error reading metrics, did you forget to call init()")
	}
	now := time.Now()
	buf := make([]byte, 150) // 150 bytes should capture an entire dataframe from the ECM-1240
	n, err := me.Port.Read(buf)
	if err != nil {
		return nil, err
	}
	df := trim(buf[:n])
	clk, err := deviceClock(df)
	if err != nil {
		return nil, err
	}
	v, err := voltage(df)
	if err != nil {
		return nil, err
	}
	currMetrics := make(map[string]*models.PowerMetric)
	for k, parser := range wattSecParsers {
		ws, err := parser(df)
		if err != nil {
			return nil, err
		}
		currMetrics[k] = &models.PowerMetric{
			SensorName: me.Name + "_" + k,
			Ts:         now,
			EnergyWs:   ws,
			VoltageV:   v,
		}
	}
	// skip outputting this metric if this is the first metric, or if the ecm1240's clock has
	// overflowed its limit of 16777216
	if me.previousMetrics != nil && clk-me.previousDeviceClock > 0.0 {
		for k, v := range currMetrics {
			v.PowerW = (v.EnergyWs - me.previousMetrics[k].EnergyWs) / (clk - me.previousDeviceClock)
		}
	}
	me.previousDeviceClock = clk
	me.previousMetrics = currMetrics
	return flatten(currMetrics), nil
}

// trim returns a single data frame which represents a single measurement from the device
func trim(data []byte) []byte {
	startSeq, startIdx := []byte{254, 255, 3}, -1
	for i := 0; i <= len(data)-len(startSeq); i++ {
		if bytes.Equal(data[i:i+len(startSeq)], startSeq) {
			if startIdx == -1 {
				startIdx = i
			} else {
				return data[startIdx:i]
			}
		}
	}
	return []byte{}
}

func flatten(in map[string]*models.PowerMetric) models.PowerMetrics {
	ret := make(models.PowerMetrics, len(in))
	i := 0
	for _, v := range in {
		ret[i] = *v
		i++
	}
	return ret
}

type parsers func([]byte) (float64, error)

var wattSecParsers = map[string]parsers{
	"1": wattSec1,
	"2": wattSec2,
	"3": wattSec3,
	"4": wattSec4,
	"5": wattSec5,
	"6": wattSec6,
	"7": wattSec7,
}

/**
Parsers specific to the Brultech ECM1240 data format
https://www.brultech.com/software/files/downloadSoft/ECM1240_Packet_format_ver9.pdf
*/

// AC voltage volts
func voltage(dataframe []byte) (float64, error) {
	return readAsFloat(dataframe, 3, 5, 0.1, false)
}

// watt seconds for ch1
func wattSec1(dataframe []byte) (float64, error) {
	return readAsFloat(dataframe, 5, 10, 1, true)
}

// watt seconds for ch2
func wattSec2(dataframe []byte) (float64, error) {
	return readAsFloat(dataframe, 10, 15, 1, true)
}

// watt seconds for aux1
func wattSec3(dataframe []byte) (float64, error) {
	return readAsFloat(dataframe, 40, 44, 1, true)
}

// watt seconds for aux2
func wattSec4(dataframe []byte) (float64, error) {
	return readAsFloat(dataframe, 44, 48, 1, true)
}

// watt seconds for aux3
func wattSec5(dataframe []byte) (float64, error) {
	return readAsFloat(dataframe, 48, 52, 1, true)
}

// watt seconds for aux4
func wattSec6(dataframe []byte) (float64, error) {
	return readAsFloat(dataframe, 52, 56, 1, true)
}

// watt seconds for aux5
func wattSec7(dataframe []byte) (float64, error) {
	return readAsFloat(dataframe, 56, 60, 1, true)
}

// amperes for ch1
func current1(dataframe []byte) (float64, error) {
	return readAsFloat(dataframe, 34, 36, .01, true)
}

// amperes for ch2
func current2(dataframe []byte) (float64, error) {
	return readAsFloat(dataframe, 36, 38, .01, true)
}

// device clock seconds
func deviceClock(dataframe []byte) (float64, error) {
	return readAsFloat(dataframe, 37, 40, 1, true)
}

// readAsFloat reads a byte slice of the dataframe dictated by start inclusive and end exclusive indices
// as an unsigned integer, multiplies it by a multiplier and returns the result as a float64
// Returns an error if the end index is out of bounds
func readAsFloat(dataframe []byte, start, end int, multiplier float64, littleEndian bool) (float64, error) {
	if len(dataframe) <= end {
		return 0.0, fmt.Errorf("dataframe (len=%d) is missing bytes %d - %d",
			len(dataframe), start, end)
	}
	if littleEndian {
		return float64(binary.LittleEndian.Uint64(padZeros(dataframe[start:end], littleEndian))) * multiplier, nil
	}
	return float64(binary.BigEndian.Uint64(padZeros(dataframe[start:end], littleEndian))) * multiplier, nil
}

func padZeros(dataSlice []byte, littleEndian bool) []byte {
	ret := make([]byte, 8)
	if littleEndian {
		for i := 0; i < len(dataSlice); i++ {
			ret[i] = dataSlice[i] // pads [data[0]...[n],0...0]
		}
	} else {
		startIdx := 8 - len(dataSlice)
		for i := startIdx; i < 8; i++ {
			ret[i] = dataSlice[i-startIdx] // pads [0...0, data[0]...[n]]
		}
	}
	return ret
}
