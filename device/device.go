package device

import (
	"errors"
	"ledfx/color"
)

// All devices take pixels and send them somewhere
type PixelPusher interface {
	initialize(*Device) error
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
	Name       int
}

func (d *Device) Initialize(id string) (err error) {
	d.ID = id
	return d.pixelPusher.initialize(d)
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
