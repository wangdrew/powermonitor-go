package main

import (
	"bytes"
	"github.com/tarm/serial"
	"log"
	"time"
)

func main() {
	c := &serial.Config{
		Name: "/dev/ttyUSB0",
		Baud: 19200,
		Size: serial.DefaultSize,
		Parity: serial.ParityNone,
		StopBits: serial.Stop1,
		ReadTimeout: time.Millisecond * 1000}

	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, 150)
	time.Sleep(2 * time.Second) // required to fill up the serial cache before we read it out
	n, err := s.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%v\n", buf[:n])
	log.Printf("%v\n", trim(buf))
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

