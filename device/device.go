package device

import (
	"errors"
	"ledfx/color"
	"ledfx/config"

	"github.com/creasty/defaults"
	"github.com/mitchellh/mapstructure"
)

// All devices take pixels and send them somewhere
type PixelPusher interface {
	initialize(base *Device, config map[string]interface{}) error
	send(p color.Pixels) error
	connect() error
	disconnect() error
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
	mapConfig := map[string]interface{}{}
	err = mapstructure.Decode(implConfig, &mapConfig)
	if err != nil {
		return err
	}
	err = config.AddEntry(
		d.ID,
		config.DeviceEntry{
			ID:         d.ID,
			Type:       d.Type,
			BaseConfig: baseConfig,
			ImplConfig: mapConfig,
		},
	)
	return err
}

func (d *Device) Connect() (err error) {
	d.State = Connecting
	err = d.pixelPusher.connect()
	if err == nil {
		d.State = Connected
	}
	return err
}

func (d *Device) Disconnect() (err error) {
	d.State = Disconnecting
	err = d.pixelPusher.disconnect()
	if err == nil {
		d.State = Disconnected
	}
	return err
}

func (d *Device) Send(p color.Pixels) (err error) {
	if d.State != Connected {
		return errors.New("device isn't connected")
	}
	return d.pixelPusher.send(p)
}
