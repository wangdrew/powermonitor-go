package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/tarm/serial"
	"log"
	"time"
)

func main() {
	c := &serial.Config{
		Name:        "/dev/ttyUSB0",
		Baud:        19200,
		Size:        serial.DefaultSize,
		Parity:      serial.ParityNone,
		StopBits:    serial.Stop1,
		ReadTimeout: time.Millisecond * 1000}

	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	//var preTs time.Time
	time.Sleep(2 * time.Second) // required to fill up the serial cache before we read it out
	var prev *PowerMeasurement
	for {
		buf := make([]byte, 150)
		time.Sleep(2 * time.Second)
		n, err := s.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("============================")
		df := trim(buf[:n])
		fmt.Printf("%v\n", df)
		ts, _ := deviceClock(df)
		v, _ := voltage(df)
		w1, _ := wattSec1(df)
		w2, _ := wattSec2(df)
		w3, _ := wattSec3(df)
		w4, _ := wattSec4(df)
		w5, _ := wattSec5(df)
		w6, _ := wattSec6(df)
		w7, _ := wattSec7(df)
		c1, _ := current1(df)
		c2, _ := current2(df)
		curr := &PowerMeasurement{
			Ts: ts,
			WattSec: []float32{w1,w2,w3,w4,w5,w6,w7},
		}
		if prev == nil {
			prev = curr
			continue
		} else {
			p := curr.calcPower(prev)
			fmt.Printf("clock:%f\nvoltage: %f\nw1: %f\nw2: %f\nw3: %f\nw4: %f\nw5: %f\nw6: %f\nw7: %f\nc1: %f\nc2: %f\n",
				ts, v, w1, w2, w3, w4, w5, w6, w7, c1, c2)
			fmt.Printf("p1:%f\np2:%f\np3:%f\np4:%f\np5:%f\np6:%f\np7:%f\n",
				p[0], p[1], p[2], p[3], p[4], p[5], p[6])
			prev = curr
		}
	}
}

type PowerMeasurement struct {
	Ts float32
	WattSec []float32
}

func (me *PowerMeasurement) calcPower(prev *PowerMeasurement) []float32 {
	pow := make([]float32, len(me.WattSec))
	diffTs := me.Ts - prev.Ts
	for i := 0; i < len(me.WattSec); i++ {
		pow[i] = (me.WattSec[i] - prev.WattSec[i]) / diffTs
	}
	return pow
}

type PowerMetric struct {
	ID      string
	Ts      time.Time
	Power   float32
	Voltage float32
}

type PowerMetrics []PowerMetric

type SerialReader struct {
	*serial.Config
}

func (me *SerialReader) Read() {

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

// AC voltage volts
func voltage(dataframe []byte) (float32, error) {
	//fmt.Printf("v: ")
	return readAsFloat(dataframe, 3, 5, 0.1, false)
}

// watt seconds for ch1
func wattSec1(dataframe []byte) (float32, error) {
	//fmt.Printf("ws1: ")
	return readAsFloat(dataframe, 5, 10, 1, true)
}

// watt seconds for ch2
func wattSec2(dataframe []byte) (float32, error) {
	//fmt.Printf("ws2: ")
	return readAsFloat(dataframe, 10, 15, 1, true)
}

// watt seconds for aux1
func wattSec3(dataframe []byte) (float32, error) {
	//fmt.Printf("ws3: ")
	return readAsFloat(dataframe, 40, 44, 1, true)
}

// watt seconds for aux2
func wattSec4(dataframe []byte) (float32, error) {
	//fmt.Printf("ws4: ")
	return readAsFloat(dataframe, 44, 48, 1, true)
}

// watt seconds for aux3
func wattSec5(dataframe []byte) (float32, error) {
	//fmt.Printf("ws5: ")
	return readAsFloat(dataframe, 48, 52, 1, true)
}

// watt seconds for aux4
func wattSec6(dataframe []byte) (float32, error) {
	//fmt.Printf("ws6: ")
	return readAsFloat(dataframe, 52, 56, 1, true)
}

// watt seconds for aux5
func wattSec7(dataframe []byte) (float32, error) {
	//fmt.Printf("ws7: ")
	return readAsFloat(dataframe, 56, 60, 1, true)
}

// amperes for ch1
func current1(dataframe []byte) (float32, error) {
	//fmt.Printf("c1: ")
	return readAsFloat(dataframe, 34, 36, .01, true)
}

// amperes for ch2
func current2(dataframe []byte) (float32, error) {
	//fmt.Printf("c2: ")
	return readAsFloat(dataframe, 36, 38, .01, true)
}

// device clock seconds
func deviceClock(dataframe []byte) (float32, error) {
	//fmt.Printf("clock: ")
	return readAsFloat(dataframe, 37, 40, 1, true)
}

// readAsFloat reads a byte slice of the dataframe dictated by start inclusive and end exclusive indices
// as an unsigned integer, multiplies it by a multiplier and returns the result as a float32
// Returns an error if the end index is out of bounds
func readAsFloat(dataframe []byte, start, end int, multiplier float32, littleEndian bool) (float32, error) {
	//fmt.Printf("%v\n", dataframe[start:end])
	if len(dataframe) <= end {
		return 0.0, fmt.Errorf("dataframe (len=%d) is missing bytes %d - %d",
			len(dataframe), start, end)
	}
	if littleEndian {
		return float32(binary.LittleEndian.Uint64(padZeros(dataframe[start:end], littleEndian))) * multiplier, nil
	}
	return float32(binary.BigEndian.Uint64(padZeros(dataframe[start:end], littleEndian))) * multiplier, nil
}

func padZeros(dataSlice []byte, littleEndian bool) []byte {
	ret := make([]byte, 8)
	if littleEndian {
		for i := 0; i < len(dataSlice); i++ {
			ret[i] = dataSlice[i] // pads [data[0]...[n],0...0]
		}
	} else {
		startIdx := 8-len(dataSlice)
		for i := startIdx; i < 8; i++ {
			ret[i] = dataSlice[i-startIdx] // pads [0...0, data[0]...[n]]
		}
	}
	return ret
}

/**
clock:1623.000000
voltage: 121.599998
w0: 1809153.000000
w1: 1173.000000
w2: 860128.000000
w3: 39518.000000
w4: 283951.000000
w5: 3.000000
w6: 9.000000
c0: 53.789997
c1: 225.279999
p0:1108.513428
p1:1.842809
p2:528.033447
p3:17.503344
p4:173.187286
p5:0.000000
p6:0.000000
 */
