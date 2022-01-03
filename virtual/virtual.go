package virtual

import (
	"errors"
	"fmt"
	"ledfx/color"
	"ledfx/config"
	"ledfx/device"
	"ledfx/logger"

	"github.com/spf13/viper"
)

type Virtual interface {
	// PlayVirtual() error // is this correct? does it make sence?
}

func PlayVirtual(virtualid string, playState bool, clr string) (err error) {
	fmt.Println("Set PlayState of ", virtualid, " to ", playState)
	if clr != "" {
		fmt.Println("Set color of ", virtualid, " to ", clr)
	}

	if virtualid == "" {
		err = errors.New("Virtual id is empty. Please provide Id to add virtual to config")
		return
	}
	var c *config.Config
	var v *viper.Viper

	c = &config.GlobalConfig
	v = config.GlobalViper

	var virtualExists bool

	newColor, err := color.NewColor(clr)

	for i, d := range c.Virtuals {
		if d.Id == virtualid {
			virtualExists = true
			c.Virtuals[i].Active = playState
			if c.Virtuals[i].IsDevice != "" {
				for in, de := range c.Devices {
					if de.Id == c.Virtuals[i].IsDevice {
						var device = &device.UdpDevice{
							Name:     c.Devices[in].Config.Name,
							Port:     c.Devices[in].Config.Port,
							Protocol: device.UdpProtocols[c.Devices[in].Config.UdpPacketType],
							Config:   c.Devices[in].Config,
						}
						data := []color.Color{}
						for i := 0; i < device.Config.PixelCount; i++ {
							// newColor, err := color.NewColor(clr)
							data = append(data, newColor)
							if err != nil {
								logger.Logger.Fatal(err)
							}
						}
						err = device.Init()
						if err != nil {
							logger.Logger.Fatal(err)
						}
						var timeo byte
						if playState {
							timeo = 0xff
						} else {
							timeo = 0x00
						}
						err = device.SendData(data, timeo)
						if err != nil {
							logger.Logger.Fatal(err)
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
