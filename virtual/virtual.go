package virtual

import (
	"errors"
	"fmt"
	"ledfx/config"
	"ledfx/effect"
	"ledfx/logger"
)

// Virtual represents a virtual device which can be mapped to one or more devices or segments
type Virtual interface {
	// PlayVirtual() error // is this correct? does it make sence?
}

// TODO: this should belong to the virtual instance
var done chan bool

// FindAndPlayVirtual finds the virtual with the given name and sets any effects that are on it to active
func FindAndPlayVirtual(virtualid string, playState bool, clr string) (err error) {
	fmt.Println("Set PlayState of ", virtualid, " to ", playState)
	if clr != "" {
		fmt.Println("Set color of ", virtualid, " to ", clr)
	}

	if virtualid == "" {
		err = errors.New("Virtual id is empty. Please provide Id to add virtual to config")
		return
	}

	c := &config.GlobalConfig
	v := config.GlobalViper

	var virtualExists bool

	for i, d := range c.Virtuals {
		if d.Id == virtualid {
			virtualExists = true
			c.Virtuals[i].Active = playState
			if c.Virtuals[i].IsDevice != "" {
				for in, de := range c.Devices {
					if de.Id == c.Virtuals[i].IsDevice {

						// FOR TESTING: solid color
						// var device = &device.UdpDevice{
						// 	Name:     c.Devices[in].Config.Name,
						// 	Port:     c.Devices[in].Config.Port,
						// 	Protocol: device.UdpProtocols[c.Devices[in].Config.UdpPacketType],
						// 	Config:   c.Devices[in].Config,
						// }
						// data := []color.Color{}
						// for i := 0; i < device.Config.PixelCount; i++ {
						// 	newColor, err := color.NewColor(clr)
						// 	data = append(data, newColor)
						// 	if err != nil {
						// 		logger.Logger.Fatal(err)
						// 	}
						// }
						// err = device.Init()
						// if err != nil {
						// 	logger.Logger.Fatal(err)
						// }
						// var timeo byte
						// if playState {
						// 	timeo = 0xff
						// } else {
						// 	timeo = 0x00
						// }
						// err = device.SendData(data, timeo)
						// if err != nil {
						// 	logger.Logger.Fatal(err)
						// }

						// FOR TESTING: pulse effect
						var currentEffect effect.Effect = &effect.PulsingEffect{}

						if playState {
							if done == nil {
								done = make(chan bool)
							}
							go func() {
								err := effect.StartEffect(c.Devices[in].Config, currentEffect, clr, 60, done)
								if err != nil {
									logger.Logger.Warn(err)
								}
							}()
						} else if !playState {
							if done != nil {
								done <- true
							}
						}
					}

				}
			}
		}
	}

	if virtualExists {
		v.Set("virtuals", c.Virtuals)
		err = v.WriteConfig()
	}
	return
}

func FindAndStopVirtual(virtualid string) (err error) {
	fmt.Println("Clear Effect of ", virtualid)

	if virtualid == "" {
		err = errors.New("Virtual id is empty. Please provide Id to add virtual to config")
		return
	}

	c := &config.GlobalConfig
	v := config.GlobalViper

	var virtualExists bool

	for i, d := range c.Virtuals {
		if d.Id == virtualid {
			virtualExists = true
			if c.Virtuals[i].IsDevice != "" {
				fmt.Println("WTF Clear Effect of ", c.Virtuals[i].Effect.Name)
				for in, de := range c.Devices {
					if de.Id == c.Virtuals[i].IsDevice {
						var currentEffect effect.Effect = &effect.PulsingEffect{}
						go func() {
							// err := effect.StartEffect(c.Devices[in].Config, currentEffect, "#fff000", 60, done)
							err := effect.StopEffect(c.Devices[in].Config, currentEffect, "#000000", 60, done)
							if err != nil {
								logger.Logger.Warn(err)
							}
						}()
						// c.Virtuals[i].Effect = config.Effect{
						// 	Config: config.EffectConfig{
						// 		BackgroundColor: "#000000",
						// 		Color:           "#eee000",
						// 	},
						// 	Name: "Single Color",
						// 	Type: "singleColor",
						// }
					}

				}
			}
		}
	}

	if virtualExists {
		v.Set("virtuals", c.Virtuals)
		err = v.WriteConfig()
	}
	return
}

// LoadVirtuals loads the virtuals from the config file and plays any effects that are active on them
func LoadVirtuals() (err error) {
	// TODO: load all virtuals from config

	// c := &config.GlobalConfig
	// v := config.GlobalViper

	// for i, virtualConfig := range c.Virtuals {
	// 	if c.Virtuals[i].Active == true {
	// 		if c.Virtuals[i].IsDevice != "" {
	//       // TODO: instantiate a virtual
	// 			// PlayVirtual(virtual)
	// 		}
	// 	}
	// }

	return nil
}

// PlayVirtual sets any effects that are on the given virtual to active
func PlayVirtual(virtual *Virtual) (err error) {
	// TODO: start the effect on an active virtual
	return nil
}
