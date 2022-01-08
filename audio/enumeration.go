package audio

import (
	"fmt"
	"ledfx/logger"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/gen2brain/malgo"
)

func GetAudioDevices() (infos []AudioDevice, err error) {
	// Capture devices.
	backends := []malgo.Backend{
		malgo.BackendDsound,
	}

	context, err := malgo.InitContext(backends, malgo.ContextConfig{}, func(message string) {
		logger.Logger.Info(message)
	})
	if err != nil {
		logger.Logger.Error(err)
		return
	}

	defer func() {
		_ = context.Uninit()
		context.Free()
	}()

	audioDeviceTypes := []malgo.DeviceType{
		malgo.Capture,
		malgo.Loopback,
	}

	var s string
	for _, dt := range audioDeviceTypes {
		dtinfos, err := context.Devices(dt)
		if err != nil {
			logger.Logger.Errorf("Failed to get capture audio devices: %v", err)
		}
		if dt == 2 {
			s = "capture"
		} else {
			s = "loopback"
		}
		for _, info := range dtinfos {
			full, err := context.DeviceInfo(dt, info.ID, malgo.Shared)
			if err != nil {
				continue
			}

			ad := AudioDevice{
				Id:         full.ID.String(),
				SampleRate: int(full.MaxSampleRate),
				Name:       strings.ReplaceAll(full.Name(), "\u0000", ""),
				Channels:   int(full.MaxChannels),
				IsDefault:  full.IsDefault == 1,
				Source:     s,
			}
			infos = append(infos, ad)
		}
	}

	return infos, err
}

func LogAudioDevices() {
	infos, err := GetAudioDevices()
	if err != nil {
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)

	var icon rune

	for i, info := range infos {
		if info.IsDefault {
			icon = '✅'
		} else {
			icon = '❌'
		}
		fmt.Fprintf(w, "%v\t%d:\t%s,\t%s\tchannels: %d,\tsamplerate: %d,\tdefault: %q\n",
			info.Source, i, info.Name, info.Id, info.Channels, info.SampleRate, icon)
	}
	w.Flush()
}
