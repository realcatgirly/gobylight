package provider

import (
	"image/color"
	"math/rand/v2"
	"time"
)

// This is a random provider that provides a random color every second for testing

func init() {
	Providers["random"] = newRandom
}

func newRandom(c chan color.RGBA, done chan struct{}) error {
	go func() {
		for {
			select {
			case <-done:
				close(c)
				return
			default:
				time.Sleep(time.Second)
				c <- color.RGBA{
					R: uint8(rand.Float32() * 255),
					G: uint8(rand.Float32() * 255),
					B: uint8(rand.Float32() * 255),
				}
			}
		}
	}()
	return nil
}
