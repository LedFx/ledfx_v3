package controller

import (
	"ledfx/color"
	"ledfx/config"
	"ledfx/device"
	"ledfx/effect"
	"ledfx/event"
	"ledfx/logger"
	"time"

	"github.com/creasty/defaults"
	"github.com/mitchellh/mapstructure"
)

type Controller struct {
	ID      string
	Effect  *effect.Effect
	Devices map[string]*device.Device
	State   bool
	Config  config.ControllerConfig
	ticker  *time.Ticker
	done    chan bool
	pixels  color.Pixels
}

func (v *Controller) Initialize(id string, c map[string]interface{}) (err error) {
	v.ID = id
	v.State = false
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
		config.ControllerEntry{
			ID:     v.ID,
			Config: c,
		},
	)
	v.Devices = map[string]*device.Device{}
	// invoke event
	event.Invoke(event.ControllerUpdate,
		map[string]interface{}{
			"id":          v.ID,
			"base_config": c,
			"active":      v.State,
		})
	return err
}

// gets the largest device pixel count
func (v *Controller) PixelCount() int {
	pc := 0
	for _, d := range v.Devices {
		dpc := d.Config.PixelCount
		if dpc > pc {
			pc = dpc
		}
	}
	return pc
}

func (v *Controller) renderLoop() {
	for {
		select {
		case <-v.ticker.C:
			if v.Effect == nil {
				return
			}
			v.Effect.Render(v.pixels) // todo catch errors in send?
			for _, d := range v.Devices {
				if d.Config.PixelCount != len(v.pixels) {
					// todo maybe dont make new buffer every frame
					p := make(color.Pixels, d.Config.PixelCount)
					color.Interpolate(v.pixels, p)
					d.Send(p)
				} else {
					d.Send(v.pixels)
				}
			}
			// if err != nil {
			// 	logger.Logger.WithField("context", "Controller").Error(err)
			// }
		case <-v.done:
			return
		}
	}
}

func (v *Controller) Start() error {
	if v.Effect == nil {
		logger.Logger.WithField("context", "Controller").Warnf("cannot start controller %s, it does not have an effect", v.ID)
		return nil
	}
	if len(v.Devices) == 0 {
		logger.Logger.WithField("context", "Controller").Warnf("cannot start controller %s, it does not have any devices", v.ID)
		return nil
	}
	for _, d := range v.Devices {
		if d.State != device.Connected {
			// err := fmt.Errorf("cannot start controller %s, device %s is not connected", v.ID, d.ID)
			// logger.Logger.WithField("context", "Controller").Error(err)
			go d.Connect()
			// return err
		}
	}
	v.pixels = make(color.Pixels, v.PixelCount())
	v.ticker = time.NewTicker(time.Duration(1000/v.Config.FrameRate) * time.Millisecond)
	v.done = make(chan bool)
	go v.renderLoop()
	v.State = true
	logger.Logger.WithField("context", "Controllers").Infof("Activated %s", v.ID)
	// invoke event
	entry, _ := config.GetController(v.ID)
	event.Invoke(event.ControllerUpdate,
		map[string]interface{}{
			"id":          v.ID,
			"base_config": entry.Config,
			"active":      v.State,
		})
	return nil
}

func (v *Controller) Stop() {
	if v.ticker != nil {
		v.ticker.Stop()
	}
	if v.done != nil {
		v.done <- true
	}
	v.State = false
	for _, d := range v.Devices {
		d.Disconnect()
	}
	logger.Logger.WithField("context", "Controllers").Infof("Deactivated %s", v.ID)
	// invoke event
	entry, _ := config.GetController(v.ID)
	event.Invoke(event.ControllerUpdate,
		map[string]interface{}{
			"id":          v.ID,
			"base_config": entry.Config,
			"active":      v.State,
		})
}
