package device

import (
	"errors"
	"ledfx/color"
	"ledfx/config"
	"ledfx/logger"

	"github.com/spf13/viper"
)

type Device interface {
	Init() error
	SendData(colors []color.Color, timeout byte) error
	Close() error
}

func AddDeviceToConfig(device config.Device) (err error) {
	if device.Id == "" {
		err = errors.New("Device id is empty. Please provide Id to add device to config")
		return
	}
	var c *config.Config
	var v *viper.Viper
	v = config.GlobalViper
	c = config.GlobalConfig

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

func AddDeviceAsVirtualToConfig(virtual config.Virtual) (exists bool, err error) {
	if virtual.Id == "" {
		err = errors.New("Virtual id is empty. Please provide Id to add virtual to config")
		return
	}
	var c *config.Config
	var v *viper.Viper
	v = config.GlobalViper
	c = config.GlobalConfig

	var virtualExists bool
	for _, d := range c.Virtuals {
		if d.Id == virtual.Id {
			virtualExists = true
		}
	}

	if !virtualExists {
		if c.Virtuals == nil {
			c.Virtuals = make([]config.Virtual, 0)
		}
		c.Virtuals = append(c.Virtuals, virtual)
		v.Set("virtuals", c.Virtuals)
		err = v.WriteConfig()
		if err != nil {
			logger.Logger.Warn("Failed to initialize resolver:", err.Error())
			return virtualExists, err
		}
	}
	return virtualExists, nil
}
