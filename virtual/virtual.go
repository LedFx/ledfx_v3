package virtual

// All virtuals map pixels to devices
type PixelMapper interface{}

type Virtual struct{}

type VirtualConfig struct {
	Name         string          `mapstructure:"name" json:"name"`
	Id           string          `mapstructure:"id" json:"id"`
	IconName     string          `mapstructure:"icon_name" json:"icon_name"`
	Span         bool            `mapstructure:"span" json:"span"`
	FrameRate    int             `mapstructure:"framerate" json:"framerate"`
	FrequencyMax int             `mapstructure:"frequency_max" json:"frequency_max"`
	FrequencyMin int             `mapstructure:"frequency_min" json:"frequency_min"`
	Outputs      []VirtualOutput `mapstructure:"outputs" json:"outputs"`
}

// Points to a device, where this virtual will send its pixels to
type VirtualOutput struct {
	Id    string
	Start int
	Close int
	// Active bool
}
