package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	serialPath := os.Getenv("SERIAL_PATH")
	if serialPath == "" {
		serialPath = "/dev/ttyUSB0"
	}

	source := NewECM1240Source("power", serialPath)
	if err := source.Init(); err != nil {
		log.Fatal(err)
	}
	for {
		time.Sleep(2 * time.Second)
		metrics, err := source.Read()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("****************")
		fmt.Printf("%+v\n", metrics)
	}
}
