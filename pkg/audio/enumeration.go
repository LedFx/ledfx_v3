package audio

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"text/tabwriter"

	"github.com/LedFx/ledfx/pkg/config"
	"github.com/LedFx/ledfx/pkg/logger"

	"github.com/LedFx/portaudio"
)

/*
Creates a hash of hostapi and device name, and the channels in and out.
This ID should be the same regardless of device idx, meaning
it won't change when other audio devices are added or removed.
Numbers and symbols are removed from the device name.
*/
func createId(hostapi string, device string, chanIn, chanOut int) string {
	reg := regexp.MustCompile("[^a-zA-Z]+")
	cleanDevice := reg.ReplaceAllString(device, "")
	s := fmt.Sprintf("%s%s%d%d", hostapi, cleanDevice, chanIn, chanOut)
	id := sha1.New()
	id.Write([]byte(s))
	return hex.EncodeToString(id.Sum(nil))
}

func GetPaDeviceInfo(ad config.AudioDevice) (d *portaudio.DeviceInfo, err error) {
	hs, err := portaudio.HostApis()
	if err != nil {
		return
	}
	for _, h := range hs {
		for _, d := range h.Devices {
			if ad.Id == createId(h.Name, d.Name, d.MaxInputChannels, d.MaxOutputChannels) {
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
	hs, err := portaudio.HostApis()
	if err != nil {
		logger.Logger.Error(err)
		return
	}
	for _, h := range hs {
		for _, d := range h.Devices {
			ad := config.AudioDevice{
				Id:          createId(h.Name, d.Name, d.MaxInputChannels, d.MaxOutputChannels),
				HostApi:     h.Name,
				SampleRate:  d.DefaultSampleRate,
				Name:        d.Name,
				ChannelsIn:  d.MaxInputChannels,
				ChannelsOut: d.MaxOutputChannels,
				IsDefault:   d.Name == h.DefaultInputDevice.Name || d.Name == h.DefaultOutputDevice.Name,
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
	fmt.Println()
	fmt.Println("Audio Devices:")
	fmt.Println()

	w := tabwriter.NewWriter(logger.Logger.Out, 1, 1, 3, ' ', 0)

	var icon rune
	fmt.Fprint(w, "HostAPI\tDevice Name\tChannels In\tChannels Out\tSamplerate\tDefault\n")
	for _, info := range infos {
		if info.IsDefault {
			icon = '✓'
		} else {
			icon = '⨯'
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%f\t%c\n",
			info.HostApi, info.Name, info.ChannelsIn, info.ChannelsOut, info.SampleRate, icon)
	}
	w.Flush()
	fmt.Println()
}

func GetDeviceByID(id string) (config.AudioDevice, error) {
	devices, err := GetAudioDevices()
	if err != nil {
		return config.AudioDevice{}, err
	}
	for _, device := range devices {
		if device.Id == id {
			return device, nil
		}
	}

	return config.AudioDevice{}, errors.New("could not find saved audio device")
}
