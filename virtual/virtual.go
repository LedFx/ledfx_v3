package virtual

import (
	"errors"
	"ledfx/config"
	"ledfx/logger"

	"github.com/spf13/viper"
)

type Virtual interface {
	// device.Init() error
	// device.SendData(colors []color.Color) error
	// device.Close() error
}

func Remove(s []int, i int) []int {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func PlayVirtual(virtualid string, playState bool) (err error) {
	logger.Logger.Warn("Set PlayState of ", virtualid, " to ", playState)

	// TODO: set Active of virtualid to playState
	// search for a Virtual with id == virtualid inside Virtuals[]
	if virtualid == "" {
		err = errors.New("Virtual id is empty. Please provide Id to add virtual to config")
		return
	}
	var c *config.Config
	var v *viper.Viper

	c = &config.GlobalConfig
	v = config.GlobalViper

	var virtualExists bool

	for i, d := range c.Virtuals {
		if d.Id == virtualid {
			virtualExists = true
			c.Virtuals[i].Active = playState
		}
	}

	if virtualExists {
		v.Set("virtuals", c.Virtuals)
		err = v.WriteConfig()
	}
	return
}

func AddDeviceAsVirtualToConfig(virtual config.Virtual, configName string) (err error) {
	if virtual.Id == "" {
		err = errors.New("Virtual id is empty. Please provide Id to add virtual to config")
		return
	}
	var c *config.Config
	var v *viper.Viper
	if configName == "goconfig" {
		v = config.GlobalViper
		c = &config.GlobalConfig
	} else if configName == "config" {
		v = config.OldViper
		c = &config.OldConfig
	}

	var virtualExists bool
	for _, d := range c.Virtuals {
		if d.Id == virtual.Id {
			virtualExists = true
		}
	}

	if !virtualExists {
		if c.Virtuals == nil {
			c.Virtuals = make([]config.Virtual, 0)
		}
		c.Virtuals = append(c.Virtuals, virtual)
		v.Set("virtuals", c.Virtuals)
		err = v.WriteConfig()
	}
	return
}
