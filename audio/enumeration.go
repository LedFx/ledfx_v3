package audio

import (
	"fmt"
	"ledfx/logger"
	"os"

	"github.com/gen2brain/malgo"
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
		fmt.Println(" - ", info.Name())
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
		fmt.Println(" - ", info.Name())
		logger.Logger.Debug("    %d: %v, %s, [%s], channels: %d-%d, samplerate: %d-%d\n",
			i, info.ID, info.Name(), e, full.MinChannels, full.MaxChannels, full.MinSampleRate, full.MaxSampleRate)
	}
}
