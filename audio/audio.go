package audio

type AudioDevice struct {
	Id         string `mapstructure:"id" json:"id"`
	SampleRate int    `mapstructure:"sample_rate" json:"sample_rate"`
	Name       string `mapstructure:"name" json:"name"`
	Channels   int    `mapstructure:"channels" json:"channels"`
	IsDefault  bool   `mapstructure:"is_default" json:"is_default"`
	Source     string `mapstructure:"source" json:"source"`
}
