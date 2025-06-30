package provider

import (
	"image/color"
)

var (
	Providers map[string]func(c chan color.RGBA, done chan struct{}) error
)

func init() {
	Providers = make(map[string]func(c chan color.RGBA, done chan struct{}) error)
}
