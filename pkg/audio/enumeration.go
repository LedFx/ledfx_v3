package audio

import (
	"fmt"

	"github.com/LedFx/ledfx/pkg/logger"

	"github.com/gen2brain/malgo"
)

// prints available audio devices to the console
func LogAudioDevices() {
	fmt.Println("Playback Devices")
	LogDeviceType(malgo.Playback)
	fmt.Println()
	fmt.Println("Capture Devices")
	LogDeviceType(malgo.Capture)
	fmt.Println()
	fmt.Println("Loopback Devices")
	LogDeviceType(malgo.Loopback)
	fmt.Println()
}

func LogDeviceType(deviceType malgo.DeviceType) {
	infos, err := Context.Devices(deviceType)
	if err != nil {
		logger.Logger.WithField("context", "Audio Device Enumeration").Error(err)
		return
	}
	for i, info := range infos {
		// prevent malgo bug causing sigsegv
		if info.MaxChannels == 0 {
			continue
		}
		e := "ok"
		full, err := Context.DeviceInfo(deviceType, info.ID, malgo.Shared)
		if err != nil {
			e = err.Error()
		}
		fmt.Printf("    %d: %s, [%s], channels: %d-%d, samplerate: %d-%d\n",
			i, info.Name(), e, full.MinChannels, full.MaxChannels, full.MinSampleRate, full.MaxSampleRate)
	}
}

// Get a malgo.DeviceInfo corresponding to a given ID
func GetDeviceByID(id string) (malgo.DeviceInfo, malgo.DeviceType, error) {
	info, err := SearchDeviceTypeForID(malgo.Playback, id)
	if err != nil {
		return info, malgo.Playback, err
	}
	info, err = SearchDeviceTypeForID(malgo.Capture, id)
	if err != nil {
		return info, malgo.Capture, err
	}
	info, err = SearchDeviceTypeForID(malgo.Duplex, id)
	if err != nil {
		return info, malgo.Duplex, err
	}
	info, err = SearchDeviceTypeForID(malgo.Loopback, id)
	return info, malgo.Loopback, err
}

func SearchDeviceTypeForID(deviceType malgo.DeviceType, id string) (malgo.DeviceInfo, error) {
	devices, err := Context.Devices(deviceType)
	if err != nil {
		return malgo.DeviceInfo{}, err
	}
	for _, device := range devices {
		if device.ID.String() == id {
			// return full device info
			return Context.DeviceInfo(deviceType, device.ID, malgo.Shared)
		}
	}
	return malgo.DeviceInfo{}, fmt.Errorf("could not find audio device matching id '%s'", id)
}
