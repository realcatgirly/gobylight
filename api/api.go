package api

import "image/color"

type Device interface {
	SetBrightness(brightness uint8) error
	SetColor(color color.RGBA) error
	GetVersion() (string, error)
}
