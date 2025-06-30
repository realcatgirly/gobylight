package device

import "github.com/realcatgirly/gobylight/api"

var (
	Devices map[string]func() (api.Device, error)
)

func init() {
	Devices = make(map[string]func() (api.Device, error))
}
