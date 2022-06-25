package config

type AudioDevice struct {
	Id          string  `mapstructure:"id" json:"id"`
	HostApi     string  `mapstructure:"hostapi" json:"hostapi"`
	SampleRate  float64 `mapstructure:"sample_rate" json:"sample_rate"`
	Name        string  `mapstructure:"name" json:"name"`
	ChannelsIn  int     `mapstructure:"channels_in" json:"channels_in"`
	ChannelsOut int     `mapstructure:"channels_out" json:"channels_out"`
	IsDefault   bool    `mapstructure:"is_default" json:"is_default"`
	Source      string  `mapstructure:"source" json:"source"`
}

type AudioConfig struct {
	Device    AudioDevice `mapstructure:"device" json:"device"`
	FftSize   int         `mapstructure:"fft_size" json:"fft_size"`
	FrameRate int         `mapstructure:"frame_rate" json:"frame_rate"`
}

func GetLocalInput() string {
	return store.LocalInput
}

func SetLocalInput(id string) {
	store.LocalInput = id
}
