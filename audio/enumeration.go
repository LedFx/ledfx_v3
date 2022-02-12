package audio

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"ledfx/config"
	"ledfx/logger"
	"os"
	"text/tabwriter"

	"github.com/gordonklaus/portaudio"
)

/*
Creates a hash of hostapi idx and device name
This ID should be the same regardless of device idx, meaning
it won't change when other audio devices are added or removed
*/
func createId(i int, n string) string {
	s := fmt.Sprintf("%d %s", i, n)
	id := sha1.New()
	id.Write([]byte(s))
	return hex.EncodeToString(id.Sum(nil))
}

func GetPaDeviceInfo(ad config.AudioDevice) (d *portaudio.DeviceInfo, err error) {
	hs, err := portaudio.HostApis()
	if err != nil {
		return
	}
	for i, h := range hs {
		for _, d := range h.Devices {
			if d.MaxInputChannels < 1 {
				continue
			}
			if ad.Id == createId(i, d.Name) {
				return d, nil
			}
		}
	}
	logger.Logger.Warn("Saved audio input device cannot be found. Reverting to default device.")
	d, err = portaudio.DefaultInputDevice()
	if err != nil {
		return &portaudio.DeviceInfo{}, err
	}
	return d, err
}

func GetAudioDevices() (infos []config.AudioDevice, err error) {
	err = portaudio.Initialize()
	if err != nil {
		logger.Logger.Error(err)
		return
	}
	defer portaudio.Terminate()

	hs, err := portaudio.HostApis()
	if err != nil {
		logger.Logger.Error(err)
		return
	}
	for i, h := range hs {
		for _, d := range h.Devices {
			if d.MaxInputChannels < 1 {
				continue
			}
			ad := config.AudioDevice{
				Id:         createId(i, d.Name),
				HostApi:    h.Name,
				SampleRate: d.DefaultSampleRate,
				Name:       d.Name,
				Channels:   d.MaxInputChannels,
				IsDefault:  d.Name == h.DefaultInputDevice.Name,
			}
			infos = append(infos, ad)
		}
	}
	return infos, err
}

func LogAudioDevices() {
	infos, err := GetAudioDevices()
	if err != nil {
		logger.Logger.Error(err)
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)

	var icon rune

	for _, info := range infos {
		if info.IsDefault {
			icon = '✓'
		} else {
			icon = '⨯'
		}
		fmt.Fprintf(w, "%s:\t%s,\tchannels: %d,\tsamplerate: %f,\tdefault: %c\n",
			info.HostApi, info.Name, info.Channels, info.SampleRate, icon)
	}
	w.Flush()
}
