package controller

import (
	"time"

	"github.com/LedFx/ledfx/pkg/config"
	"github.com/LedFx/ledfx/pkg/device"
	"github.com/LedFx/ledfx/pkg/effect"
	"github.com/LedFx/ledfx/pkg/event"
	"github.com/LedFx/ledfx/pkg/logger"
	"github.com/LedFx/ledfx/pkg/pixelgroup"

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
	pixels  *pixelgroup.PixelGroup
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
	v.pixels, err = pixelgroup.NewPixelGroup(v.Devices)
	if err != nil {
		return err
	}
	// invoke event
	event.Invoke(event.ControllerUpdate,
		map[string]interface{}{
			"id":          v.ID,
			"base_config": c,
			"active":      v.State,
		})
	return err
}

// gets the sum of device pixel counts
func (v *Controller) PixelCount() int {
	pc := 0
	for _, d := range v.Devices {
		pc += d.Config.PixelCount
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
				d.Send(v.pixels.Group[d.ID])
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
		logger.Logger.WithField("context", "Controller").Warnf("cannot start %s, it does not have an effect", v.ID)
		return nil
	}
	if len(v.Devices) == 0 {
		logger.Logger.WithField("context", "Controller").Warnf("cannot start %s, it does not have any devices", v.ID)
		return nil
	}
	for _, d := range v.Devices {
		if d.State != device.Connected {
			go d.Connect()
		}
	}
	var err error
	v.pixels, err = pixelgroup.NewPixelGroup(v.Devices)
	if err != nil {
		logger.Logger.WithField("context", "Controller").Errorf("failed to start %s: %s", v.ID, err)
	}
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
