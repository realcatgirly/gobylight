package main

import (
	"image/color"
	"os"
	"os/signal"
	"syscall"

	"github.com/realcatgirly/gobylight/device"
	"github.com/realcatgirly/gobylight/provider"
)

var (
	c    chan color.RGBA
	done chan struct{}
)

// todo flags:
// -p: provider, eg "teams"
// -d: device, eg "neotrinkey"

func main() {
	c = make(chan color.RGBA, 1)
	done = make(chan struct{})

	d, err := device.Devices["console"]()
	if err != nil {
		panic(err)
	}
	d.SetBrightness(10)
	if err := provider.Providers["teams"](c, done); err != nil {
		panic(err)
	}

	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case color := <-c:
			d.SetColor(color)
		case <-signalC:
			close(done)
			d.SetColor(color.RGBA{R: 0, G: 0, B: 0})
			return
		}
	}
}
