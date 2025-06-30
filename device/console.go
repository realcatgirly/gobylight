package device

import (
	"fmt"
	"image/color"

	"github.com/realcatgirly/gobylight/api"
)

// This is a console device that outputs all received commands to the console for testing

func init() {
	Devices["console"] = newConsole
}

type Console struct {
}

func newConsole() (api.Device, error) {
	return &Console{}, nil
}

// GetVersion implements api.Device.
func (c *Console) GetVersion() (string, error) {
	return "1.0.0", nil
}

// SetBrightness implements api.Device.
func (c *Console) SetBrightness(brightness uint8) error {
	if brightness > 100 {
		return fmt.Errorf("brightness out of range")
	}
	fmt.Printf("brightness: %d\n", brightness)
	return nil
}

// SetColor implements api.Device.
func (c *Console) SetColor(color color.RGBA) error {
	fmt.Printf("color: %v\n", color)
	return nil
}
