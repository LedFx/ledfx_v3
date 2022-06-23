package virtual

import (
	"ledfx/config"
	"ledfx/device"
	"ledfx/effect"

	"github.com/creasty/defaults"
	"github.com/mitchellh/mapstructure"
)

// All virtuals map pixels to devices
type PixelMapper interface{}

type Virtual struct {
	ID     string
	Effect *effect.Effect
	Device *device.Device
	Active bool
	Config config.VirtualConfig
}

// // Points to a device, where this virtual will send its pixels to
// type VirtualOutput struct {
// 	Id    string
// 	Start int
// 	Close int
// 	// Active bool
// }

func (v *Virtual) Initialize(id string, c map[string]interface{}) (err error) {
	v.ID = id
	defaults.Set(&v.Config)
	err = mapstructure.Decode(c, &v.Config)
	if err != nil {
		return err
	}
	err = validate.Struct(&v.Config)
	if err != nil {
		return err
	}
	err = config.AddEntry(
		v.ID,
		config.VirtualEntry{
			ID:     v.ID,
			Config: c,
		},
	)
	return err
}

func (v *Virtual) Start() error {
	return nil
}
func (v *Virtual) Stop() error {
	return nil
}
