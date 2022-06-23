package virtual

import (
	"fmt"
	"ledfx/color"
	"ledfx/config"
	"ledfx/device"
	"ledfx/effect"
	"ledfx/logger"
	"time"

	"github.com/creasty/defaults"
	"github.com/mitchellh/mapstructure"
)

type Virtual struct {
	ID     string
	Effect *effect.Effect
	Device *device.Device
	Active bool
	Config config.VirtualConfig
	ticker *time.Ticker
	done   chan bool
	pixels color.Pixels
}

func (v *Virtual) Initialize(id string, c map[string]interface{}) (err error) {
	v.ID = id
	v.Active = false
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

func (v *Virtual) renderLoop() {
	for {
		select {
		case <-v.ticker.C:
			v.Effect.Render(v.pixels)
			err := v.Device.Send(v.pixels)
			if err != nil {
				logger.Logger.WithField("context", "Virtual").Error(err)
			}
		case <-v.done:
			return
		}
	}
}

func (v *Virtual) Start() error {
	if v.Effect == nil {
		err := fmt.Errorf("cannot start virtual %s, it does not have an effect", v.ID)
		logger.Logger.WithField("context", "Virtual").Error(err)
		return err
	}
	if v.Device == nil {
		err := fmt.Errorf("cannot start virtual %s, it does not have a device", v.ID)
		logger.Logger.WithField("context", "Virtual").Error(err)
		return err
	}
	if v.Device.State != device.Connected {
		err := fmt.Errorf("cannot start virtual %s, device %s is not connected", v.ID, v.Device.ID)
		logger.Logger.WithField("context", "Virtual").Error(err)
		go v.Device.Connect()
		return err
	}
	v.pixels = make(color.Pixels, v.Device.Config.PixelCount)
	v.ticker = time.NewTicker(16 * time.Millisecond)
	v.done = make(chan bool)
	go v.renderLoop()
	v.Active = true
	return nil
}

func (v *Virtual) Stop() {
	if v.ticker != nil {
		v.ticker.Stop()
	}
	if v.done != nil {
		v.done <- true
	}
	v.Active = false
}
