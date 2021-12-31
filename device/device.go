package device

import (
	"errors"
	"ledfx/color"
	"ledfx/config"

	"github.com/spf13/viper"
)

// Device is the base interface for all devices
type Device interface {
	Init() error
	SendData(colors []color.Color, ledOffset int) error
	Close() error
}

// AddDeviceToConfig adds a device to the config file, saves, and updates the config in memory
func AddDeviceToConfig(device config.Device, configName string) (err error) {
	if device.Id == "" {
		err = errors.New("Device id is empty. Please provide Id to add device to config")
		return
	}
	var c *config.Config
	var v *viper.Viper
	if configName == "goconfig" {
		v = config.GlobalViper
		c = &config.GlobalConfig
	} else if configName == "config" {
		v = config.OldViper
		c = &config.OldConfig
	}

	var deviceExists bool
	for _, d := range c.Devices {
		if d.Id == device.Id {
			deviceExists = true
		}
	}

	if !deviceExists {
		if c.Devices == nil {
			c.Devices = make([]config.Device, 0)
		}
		c.Devices = append(c.Devices, device)
		v.Set("devices", c.Devices)
		err = v.WriteConfig()
	}
	return
}
