package virtual

import (
	"errors"
	"fmt"
	"ledfx/color"
	"ledfx/config"
	"ledfx/device"
	log "ledfx/logger"
	"time"
)

var (
	devMap map[string]*device.UdpDevice
)

func init() {
	devMap = make(map[string]*device.UdpDevice)
}

type Virtual interface {
	// PlayVirtual() error // is this correct? does it make sence?
}

func RepeatN(virtualID string, playState bool, clr string, n int) error {
	if virtualID == "" {
		return errors.New("virtual id is empty. Please provide Id to add virtual to config")
	}

	var timeout byte
	if playState {
		timeout = 0xff
	} else {
		timeout = 0x00
	}

	newColor, err := color.NewColor(clr)
	if err != nil {
		return fmt.Errorf("error generating new color: %w", err)
	}

	dev, ok := devMap[virtualID]
	if ok {
		data := make([]color.Color, dev.Config.PixelCount)
		for i2 := 0; i2 < n-1; i2++ {
			if len(data) <= i2 {
				break
			}
			data[i2] = newColor
		}

		if err := dev.SendData(data, timeout); err != nil {
			return fmt.Errorf("error sending data to WLED: %w", err)
		}
		return nil
	}

	var virtualExists bool
	for i, d := range config.GlobalConfig.Virtuals {
		if d.Id == virtualID {
			virtualExists = true
			config.GlobalConfig.Virtuals[i].Active = playState
			if config.GlobalConfig.Virtuals[i].IsDevice != "" {
				for in, de := range config.GlobalConfig.Devices {
					if de.Id == config.GlobalConfig.Virtuals[i].IsDevice {
						devMap[virtualID] = &device.UdpDevice{
							Name:     config.GlobalConfig.Devices[in].Config.Name,
							Port:     config.GlobalConfig.Devices[in].Config.Port,
							Protocol: device.UdpProtocols[config.GlobalConfig.Devices[in].Config.UdpPacketType],
							Config:   config.GlobalConfig.Devices[in].Config,
						}

						if err := devMap[virtualID].Init(); err != nil {
							return fmt.Errorf("error during device init: %w", err)
						}

						return RepeatN(virtualID, playState, clr, n)
					}
				}
			}
		}
	}
	if virtualExists {
		config.GlobalViper.Set("virtuals", config.GlobalConfig.Virtuals)
		err = config.GlobalViper.WriteConfig()
	}
	return nil
}

func RepeatNSmooth(virtualID string, playState bool, clr string, n int) error {
	if virtualID == "" {
		return errors.New("virtual id is empty. Please provide Id to add virtual to config")
	}

	var timeout byte
	if playState {
		timeout = 0xff
	} else {
		timeout = 0x00
	}

	newColor, err := color.NewColor(clr)
	if err != nil {
		return fmt.Errorf("error generating new color: %w", err)
	}

	var virtualExists bool
	for i, d := range config.GlobalConfig.Virtuals {
		if d.Id == virtualID {
			virtualExists = true
			config.GlobalConfig.Virtuals[i].Active = playState
			if config.GlobalConfig.Virtuals[i].IsDevice != "" {
				for in, de := range config.GlobalConfig.Devices {
					if de.Id == config.GlobalConfig.Virtuals[i].IsDevice {
						var dev = &device.UdpDevice{
							Name:     config.GlobalConfig.Devices[in].Config.Name,
							Port:     config.GlobalConfig.Devices[in].Config.Port,
							Protocol: device.UdpProtocols[config.GlobalConfig.Devices[in].Config.UdpPacketType],
							Config:   config.GlobalConfig.Devices[in].Config,
						}

						if err := dev.Init(); err != nil {
							return fmt.Errorf("error initializing dev: %w", err)
						}

						data := make([]color.Color, de.Config.PixelCount)
						for i2 := 0; i2 < n-1; i2++ {
							if len(data) <= i2 {
								break
							}
							data[i2] = newColor
							time.Sleep(5 * time.Millisecond)
							if err := dev.SendData(data, timeout); err != nil {
								log.Logger.Errorf("Error sending data to WLED: %v", err)
							}
						}

						go func() {
							noColor, _ := color.NewColor("#000000")
							for i2 := len(data) - 1; ; i2-- {
								if 0 > i2 {
									break
								}
								data[i2] = noColor
								time.Sleep(5 * time.Millisecond)
								if err := dev.SendData(data, timeout); err != nil {
									log.Logger.Errorf("Error sending data to WLED: %v", err)
								}
							}
						}()
					}
				}
			}
		}
	}
	if virtualExists {
		config.GlobalViper.Set("virtuals", config.GlobalConfig.Virtuals)
		err = config.GlobalViper.WriteConfig()
	}
	return nil
}

func PlayVirtual(virtualID string, playState bool, clr string) (err error) {
	//fmt.Println("Set PlayState of ", virtualID, " to ", playState)
	if clr != "" {
		//fmt.Println("Set color of ", virtualID, " to ", clr)
	}

	if virtualID == "" {
		return errors.New("virtual id is empty. Please provide Id to add virtual to config")
	}

	var virtualExists bool

	newColor, err := color.NewColor(clr)
	if err != nil {
		return fmt.Errorf("error generating new color: %w", err)
	}

	for i, d := range config.GlobalConfig.Virtuals {
		if d.Id == virtualID {
			virtualExists = true
			config.GlobalConfig.Virtuals[i].Active = playState
			if config.GlobalConfig.Virtuals[i].IsDevice != "" {
				for in, de := range config.GlobalConfig.Devices {
					if de.Id == config.GlobalConfig.Virtuals[i].IsDevice {
						var dev = &device.UdpDevice{
							Name:     config.GlobalConfig.Devices[in].Config.Name,
							Port:     config.GlobalConfig.Devices[in].Config.Port,
							Protocol: device.UdpProtocols[config.GlobalConfig.Devices[in].Config.UdpPacketType],
							Config:   config.GlobalConfig.Devices[in].Config,
						}

						data := make([]color.Color, de.Config.PixelCount)
						for i2 := range data {
							data[i2] = newColor
						}

						if err := dev.Init(); err != nil {
							return fmt.Errorf("error initializing dev: %w", err)
						}

						var timeout byte
						if playState {
							timeout = 0xff
						} else {
							timeout = 0x00
						}

						if err := dev.SendData(data, timeout); err != nil {
							return fmt.Errorf("error sending data to WLED: %w", err)
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
