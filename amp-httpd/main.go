package main

import (
	"github.com/labstack/gommon/log"
)

func main() {
	d, err := OpenDevice("/dev/ttyUSB0")
	if err != nil {
		panic(err)
	}
	defer d.Close()

	e := NewWeb(NewAmp(d))
	e.Logger.SetLevel(log.DEBUG)
	if err := e.Start(":8080"); err != nil {
		panic(err)
	}
}
