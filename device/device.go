package device

import (
	"errors"
	"ledfx/color"
	"ledfx/config"
	"ledfx/event"
	"ledfx/logger"

	"github.com/creasty/defaults"
	"github.com/mitchellh/mapstructure"
)

// All devices take pixels and send them somewhere
type PixelPusher interface {
	initialize(base *Device, config map[string]interface{}) error
	send(p color.Pixels) error
	connect() error
	disconnect() error
	getConfig() map[string]interface{} // pointer to config
}

type Device struct {
	ID          string
	Type        string
	pixelPusher PixelPusher
	State       State
	Config      config.BaseDeviceConfig
}

func (d *Device) Initialize(id string, baseConfig map[string]interface{}, implConfig map[string]interface{}) (err error) {
	d.ID = id
	// set and validate base config
	defaults.Set(&d.Config)
	err = mapstructure.Decode(baseConfig, &d.Config)
	if err != nil {
		return err
	}
	err = validate.Struct(&d.Config)
	if err != nil {
		return err
	}
	err = d.pixelPusher.initialize(d, implConfig)
	if err != nil {
		return err
	}
	// save to config store
	base, impl := d.FullConfig()
	err = config.AddEntry(
		d.ID,
		config.DeviceEntry{
			ID:         d.ID,
			Type:       d.Type,
			BaseConfig: base,
			ImplConfig: impl,
		},
	)
	if err != nil {
		return err
	}
	// invoke event
	event.Invoke(event.DeviceUpdate,
		map[string]interface{}{
			"id":          d.ID,
			"base_config": base,
			"impl_config": impl,
			"state":       d.State,
		})
	return err
}

func (d *Device) Connect() (err error) {
	d.State = Connecting
	err = d.pixelPusher.connect()
	if err == nil {
		d.State = Connected
		// invoke event
		base, impl := d.FullConfig()
		event.Invoke(event.DeviceUpdate,
			map[string]interface{}{
				"id":          d.ID,
				"base_config": base,
				"impl_config": impl,
				"state":       d.State,
			})
	} else {
		logger.Logger.WithField("context", "Device").Errorf("Device %s failed to connect: %s", d.ID, err.Error())
	}
	return err
}

func (d *Device) Disconnect() (err error) {
	d.State = Disconnecting
	err = d.pixelPusher.disconnect()
	if err == nil {
		d.State = Disconnected
		// invoke event
		base, impl := d.FullConfig()
		event.Invoke(event.DeviceUpdate,
			map[string]interface{}{
				"id":          d.ID,
				"base_config": base,
				"impl_config": impl,
				"state":       d.State,
			})
	} else {
		logger.Logger.WithField("context", "Device").Errorf("Device %s failed to disconnect: %s", d.ID, err.Error())
	}
	return err
}

func (d *Device) Send(p color.Pixels) (err error) {
	if d.State != Connected {
		return errors.New("device isn't connected")
	}
	return d.pixelPusher.send(p)
}

func (d *Device) FullConfig() (base, impl map[string]interface{}) {
	mapstructure.Decode(&d.Config, &base)
	impl = d.pixelPusher.getConfig()
	return base, impl
}
