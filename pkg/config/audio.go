package config

import "github.com/gen2brain/malgo"

type AudioConfig struct {
	Device    malgo.DeviceInfo `mapstructure:"device" json:"device"`
	FftSize   int              `mapstructure:"fft_size" json:"fft_size"`
	FrameRate int              `mapstructure:"frame_rate" json:"frame_rate"`
}

func GetLocalInput() malgo.DeviceID {
	return store.LocalInput
}

func SetLocalInput(id malgo.DeviceID) {
	store.LocalInput = id
}
