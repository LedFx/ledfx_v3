package device_old

import (
	"errors"
	"ledfx/color"
	"ledfx/config"
)

type Device interface {
	Init() error
	SendData(colors []color.Color, timeout byte) error
	Close() error
}

func AddDeviceToConfig(device config.Device) (err error) {
	if device.Id == "" {
		err = errors.New("device id is empty. Please provide Id to add device to config")
		return
	}

	if config.GlobalConfig.Devices == nil {
		config.GlobalConfig.Devices = make([]config.Device, 0)
	}

	for index, dev := range config.GlobalConfig.Devices {
		if dev.Id == device.Id {
			config.GlobalConfig.Devices[index] = device
			return config.GlobalViper.WriteConfig()
		}
	}

	config.GlobalConfig.Devices = append(config.GlobalConfig.Devices, device)
	config.GlobalViper.Set("devices", config.GlobalConfig.Devices)
	return config.GlobalViper.WriteConfig()
}

func AddDeviceAsVirtualToConfig(virtual config.Virtual) (exists bool, err error) {
	if virtual.Id == "" {
		return exists, errors.New("virtual id is empty. Please provide Id to add virtual to config")
	}
	/*v = config.GlobalViper
	c = config.GlobalConfig*/

	if config.GlobalConfig.Virtuals == nil {
		config.GlobalConfig.Virtuals = make([]config.Virtual, 0)
	}

	for index, virt := range config.GlobalConfig.Virtuals {
		if virt.Id == virtual.Id {
			virtual.Active = virt.Active
			config.GlobalConfig.Virtuals[index] = virtual
			config.GlobalViper.Set("virtuals", config.GlobalConfig.Virtuals)
			return true, config.GlobalViper.WriteConfig()
		}
	}

	config.GlobalConfig.Virtuals = append(config.GlobalConfig.Virtuals, virtual)
	config.GlobalViper.Set("virtuals", config.GlobalConfig.Virtuals)
	return false, config.GlobalViper.WriteConfig()
}
