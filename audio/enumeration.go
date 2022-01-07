package audio

import (
	"fmt"
	"ledfx/config"
	"ledfx/logger"
	"os"
	"strings"

	"github.com/gen2brain/malgo"
	"github.com/spf13/viper"
)

func Enumerate() {
	context, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer func() {
		_ = context.Uninit()
		context.Free()
	}()

	var c *config.Config
	var v *viper.Viper
	c = &config.GlobalConfig
	v = config.GlobalViper
	c.Audio.Outputs = nil
	c.Audio.Inputs = nil

	// Playback devices.
	infos, err := context.Devices(malgo.Playback)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("========================================================")
	fmt.Println(" Audio Playback Devices:")
	fmt.Println("========================================================")
	for i, info := range infos {
		e := "ok"
		full, err := context.DeviceInfo(malgo.Playback, info.ID, malgo.Shared)
		if err != nil {
			e = err.Error()
		}
		var device config.Audio
		s := strings.Split(info.Name(), "\u0000")
		device.Name = s[0]
		device.Id = i
		fmt.Println(" - ", device.Name)
		c.Audio.Outputs = append(c.Audio.Outputs, device)
		v.Set("audio.outputs", c.Audio.Outputs)

		logger.Logger.Debug("    %d: %v, %s, [%s], channels: %d-%d, samplerate: %d-%d\n",
			i, info.ID, info.Name(), e, full.MinChannels, full.MaxChannels, full.MinSampleRate, full.MaxSampleRate)
	}
	// Capture devices.
	infos, err = context.Devices(malgo.Capture)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("========================================================")
	fmt.Println(" Audio Capture Devices:")
	fmt.Println("========================================================")
	for i, info := range infos {
		e := "ok"
		full, err := context.DeviceInfo(malgo.Capture, info.ID, malgo.Shared)
		if err != nil {
			e = err.Error()
		}
		var device config.Audio
		s := strings.Split(info.Name(), "\u0000")
		device.Name = s[0]
		device.Id = i
		fmt.Println(" - ", device.Name)
		c.Audio.Inputs = append(c.Audio.Inputs, device)
		v.Set("audio.inputs", c.Audio.Inputs)

		logger.Logger.Debug("    %d: %v, %s, [%s], channels: %d-%d, samplerate: %d-%d\n",
			i, info.ID, info.Name(), e, full.MinChannels, full.MaxChannels, full.MinSampleRate, full.MaxSampleRate)
	}
	v.WriteConfig()
}
