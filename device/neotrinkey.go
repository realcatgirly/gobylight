package device

import (
	"fmt"
	"image/color"
	"strings"
	"sync"
	"time"

	"github.com/realcatgirly/gobylight/api"
	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

// The Adafruit Neotrinkey running the trinkey_busylight_firmware to display colors

func init() {
	Devices["neotrinkey"] = newNeoTrinkey
}

type NeoTrinkey struct {
	conn *serial.Port
	mu   sync.Mutex
}

func newNeoTrinkey() (api.Device, error) {
	var ports, err = enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, err
	}

	for _, port := range ports {
		if port.VID == "239A" && port.PID == "80F0" {
			fmt.Printf("found neotrinkey at vid: %s, pid: %s, name: %s\n", port.VID, port.PID, port.Name)
			buffer := make([]byte, 128)
			s, err := serial.Open(port.Name, &serial.Mode{
				BaudRate: 9600,
				DataBits: 8,
				Parity:   serial.NoParity,
				StopBits: 1,
			})
			if err != nil {
				return nil, err
			}
			if err := s.SetReadTimeout(time.Second / 2); err != nil {
				return nil, err
			}
			if _, err := s.Write([]byte("\n")); err != nil {
				return nil, err
			}
			if err := s.ResetInputBuffer(); err != nil {
				return nil, err
			}
			if _, err := s.Write([]byte("AT\n")); err != nil {
				return nil, err
			}
			time.Sleep(time.Second)
			if _, err := s.Read(buffer); err != nil {
				return nil, err
			}
			response := string(buffer)
			if response[:2] != "OK" {
				fmt.Println(buffer)
				fmt.Println(response)
				return nil, fmt.Errorf("unable to communicate with device")
			}
			return &NeoTrinkey{
				conn: &s,
				mu:   sync.Mutex{},
			}, nil
		}
	}
	return nil, fmt.Errorf("unable to find device")
}

// SetBrightness implements api.Device.
func (nt *NeoTrinkey) SetBrightness(brightness uint8) error {
	nt.mu.Lock()
	defer nt.mu.Unlock()

	if brightness > 100 {
		return fmt.Errorf("brightness out of range")
	}
	s := *nt.conn
	buffer := make([]byte, 128)
	if _, err := s.Write([]byte("\n")); err != nil {
		return err
	}
	if err := s.ResetInputBuffer(); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(s, "AT+B=%d\n", brightness); err != nil {
		return err
	}
	time.Sleep(time.Second / 2)
	if _, err := s.Read(buffer); err != nil {
		return err
	} else if string(buffer)[:2] != "OK" {
		return fmt.Errorf("%s", string(buffer))
	}
	return nil
}

// SetColor implements api.Device.
func (nt *NeoTrinkey) SetColor(color color.RGBA) error {
	nt.mu.Lock()
	defer nt.mu.Unlock()

	s := *nt.conn
	buffer := make([]byte, 128)
	if err := s.ResetInputBuffer(); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(s, "AT+C=%d,%d,%d\n", color.R, color.G, color.B); err != nil {
		return err
	}
	time.Sleep(time.Second / 2)
	if _, err := s.Read(buffer); err != nil {
		return err
	} else if string(buffer)[:2] != "OK" {
		return fmt.Errorf("%s", string(buffer))
	}
	return nil
}

// GetVersion implements api.Device.
func (nt *NeoTrinkey) GetVersion() (string, error) {
	nt.mu.Lock()
	defer nt.mu.Unlock()

	s := *nt.conn
	buffer := make([]byte, 128)
	if err := s.ResetInputBuffer(); err != nil {
		return "", err
	}
	if _, err := fmt.Fprintf(s, "AT+V\n"); err != nil {
		return "", err
	}
	time.Sleep(time.Second / 2)
	if _, err := s.Read(buffer); err != nil {
		return "", err
	}
	return strings.Replace(string(buffer), "+VER: ", "", 1), nil
}
