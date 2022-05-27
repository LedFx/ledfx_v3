package device

import (
	"errors"
	"ledfx/color"
)

// All devices take pixels and send them somewhere
type PixelPusher interface {
	initialize(base *Device, config interface{}) error
	send(p color.Pixels) error
	connect() error
	disconnect() error
}

type Device struct {
	ID          string
	pixelPusher PixelPusher
	State       State
	Config      BaseDeviceConfig
}

type BaseDeviceConfig struct {
	PixelCount int
	Name       string
}

func (d *Device) Initialize(id string, baseConfig BaseDeviceConfig, implConfig interface{}) (err error) {
	d.ID = id
	d.Config = baseConfig
	return d.pixelPusher.initialize(d, implConfig)
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
