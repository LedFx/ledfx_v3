package virtual

import (
	"errors"
	"fmt"
	"ledfx/color"
	"ledfx/config"
	"ledfx/device"
	"ledfx/logger"
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

	var virtualExists bool

	newColor, err := color.NewColor(clr)

	for i, d := range config.GlobalConfig.Virtuals {
		if d.Id == virtualid {
			virtualExists = true
			config.GlobalConfig.Virtuals[i].Active = playState
			if config.GlobalConfig.Virtuals[i].IsDevice != "" {
				for in, de := range config.GlobalConfig.Devices {
					if de.Id == config.GlobalConfig.Virtuals[i].IsDevice {
						var device = &device.UdpDevice{
							Name:     config.GlobalConfig.Devices[in].Config.Name,
							Port:     config.GlobalConfig.Devices[in].Config.Port,
							Protocol: device.UdpProtocols[config.GlobalConfig.Devices[in].Config.UdpPacketType],
							Config:   config.GlobalConfig.Devices[in].Config,
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
		config.GlobalViper.Set("virtuals", config.GlobalConfig.Virtuals)
		err = config.GlobalViper.WriteConfig()
	}
	return
}
