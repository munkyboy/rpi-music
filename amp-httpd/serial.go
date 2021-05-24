package main

import (
	"time"

	"github.com/tarm/serial"
)

const baudRate int = 9600

func OpenDevice(d string) (*serial.Port, error) {
	return serial.OpenPort(&serial.Config{Name: d, Baud: baudRate, ReadTimeout: 5 * time.Millisecond})
}
