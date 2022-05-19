package config

import (
	"ledfx/color"
	"ledfx/constants"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Settings applied to all effects
type GlobalEffectsConfig struct {
	Brightness     float64 `mapstructure:"center_offset" json:"center_offset"`
	Hue            float64 `mapstructure:"hue" json:"hue"`
	Saturation     float64 `mapstructure:"saturation" json:"saturation"`
	TransitionMode string  `mapstructure:"transition_mode" json:"transition_mode"`
	TransitionTime float32 `mapstructure:"transition_time" json:"transition_time"`
}

type Effect struct {
	Name    string         `mapstructure:"name" json:"name"`
	Type    string         `mapstructure:"type" json:"type"`
	Config  EffectConfig   `mapstructure:"config" json:"config"`
	Outputs []EffectOutput `mapstructure:"outputs" json:"outputs"`
}

// Points to a virtual, where this effect will send its pixels to
type EffectOutput struct {
	Id     string `mapstructure:"id" json:"id"`         // Virtual ID
	Active string `mapstructure:"active" json:"active"` // Is this output active
}

type EffectBaseConfig struct {
	Intensity     float64        `mapstructure:"intensity" json:"intensity"`
	Brightness    float64        `mapstructure:"brightness" json:"brightness"`
	Palette       color.Gradient `mapstructure:"palette" json:"palette"`
	Blur          float64        `mapstructure:"blur" json:"blur"`
	Flip          bool           `mapstructure:"flip" json:"flip"`
	Mirror        bool           `mapstructure:"mirror" json:"mirror"`
	BkgBrightness float64        `mapstructure:"bkg_brightness" json:"bkg_brightness"`
	BkgColor      string         `mapstructure:"bkg_color" json:"bkg_color"`
	// GradientName  string  `mapstructure:"gradient_name" json:"gradient_name"`
	// Color         string  `mapstructure:"color" json:"color"`
}

type EffectConfig interface {
	EffectBaseConfig
}

type EnergyConfig struct {
	EffectBaseConfig
	ColLows color.Color
	ColMids color.Color
	ColHigh color.Color
}

type Virtual struct {
	Name         string          `mapstructure:"name" json:"name"`
	Id           string          `mapstructure:"id" json:"id"`
	IconName     string          `mapstructure:"icon_name" json:"icon_name"`
	Span         bool            `mapstructure:"span" json:"span"`
	FrequencyMax int             `mapstructure:"frequency_max" json:"frequency_max"`
	FrequencyMin int             `mapstructure:"frequency_min" json:"frequency_min"`
	Outputs      []VirtualOutput `mapstructure:"outputs" json:"outputs"`
	// Config       VirtualConfig   `mapstructure:"config" json:"config"` // Virtuals are all the same "type" so don't need a config
	// MaxBrightness int    `mapstructure:"max_brightness" json:"max_brightness"`
	// PreviewOnly   bool   `mapstructure:"preview_only" json:"preview_only"`
	// CenterOffset  int    `mapstructure:"center_offset" json:"center_offset"`
}

// Points to a device, where this virtual will send its pixels to
type VirtualOutput struct {
	Id    string
	Start int
	Close int
	// Active bool
}

type Device struct {
	Name   string       `mapstructure:"name" json:"name"`
	Id     string       `mapstructure:"id" json:"id"`
	Type   string       `mapstructure:"type" json:"type"`
	Config DeviceConfig `mapstructure:"config" json:"config"`
	// Effect Effect       `mapstructure:"effect" json:"effect"` // not in old api when devicetype UDP
}

type DeviceConfig struct {
	PixelCount  int   `mapstructure:"pixel_count" json:"pixel_count"`
	RefreshRate int   `mapstructure:"refresh_rate" json:"refresh_rate"`
	Mapping     []int `mapstructure:"mapping" json:"mapping"`
	// CenterOffset  int    `mapstructure:"center_offset" json:"center_offset"`
	// Timeout       int    `mapstructure:"timeout" json:"timeout"`
	// UdpPacketType string `mapstructure:"udp_packet_type" json:"udp_packet_type"`
	// IpAddress     string `mapstructure:"ip_address" json:"ip_address"`
	// Port          int    `mapstructure:"port" json:"port"`
	// ForceRefresh    bool   `mapstructure:"force_refresh" json:"force_refresh"`     // not in old api when devicetype UDP
	// IconName        string `mapstructure:"icon_name" json:"icon_name"`             // not needed since its on virtual
	// IncludeIndexes  bool   `mapstructure:"include_indexes" json:"include_indexes"` // not in old api when devicetype UDP
	// MinimiseTraffic bool   `mapstructure:"minimise_traffic" json:"minimise_traffic"` // not in old api when devicetype UDP
	// MaxBrightness   int    `mapstructure:"max_brightness" json:"max_brightness"`     // not in old api when devicetype UDP
	// PreviewOnly     bool   `mapstructure:"preview_only" json:"preview_only"` // not needed since its on virtual
	// Type            string `mapstructure:"type" json:"type"` // not in old api when devicetype UDP
}

// func (r *Segment) AsJSON() ([]byte, error) {
// 	arr := []interface{}{r.Id, r.Start, r.Close, r.Active}
// 	return json.Marshal(arr)
// }

type PortAudioDevice struct {
	Id         string  `mapstructure:"id" json:"id"`
	HostApi    string  `mapstructure:"hostapi" json:"hostapi"`
	SampleRate float64 `mapstructure:"sample_rate" json:"sample_rate"`
	Name       string  `mapstructure:"name" json:"name"`
	Channels   int     `mapstructure:"channels" json:"channels"`
	IsDefault  bool    `mapstructure:"is_default" json:"is_default"`
	Source     string  `mapstructure:"source" json:"source"`
}

type AudioConfig struct {
	Device    PortAudioDevice `mapstructure:"device" json:"device"`
	FftSize   int             `mapstructure:"fft_size" json:"fft_size"`
	FrameRate int             `mapstructure:"frame_rate" json:"frame_rate"`
}

type Config struct {
	Version  string      `mapstructure:"version" json:"version"`
	Host     string      `mapstructure:"host" json:"host"`
	Port     int         `mapstructure:"port" json:"port"`
	OpenUi   bool        `mapstructure:"open_ui" json:"open_ui"`
	LogLevel int         `mapstructure:"log_level" json:"log_level"`
	NoSentry bool        `mapstructure:"no_sentry" json:"no_sentry"`
	Effects  []Effect    `mapstructure:"effects" json:"effects"`
	Virtuals []Virtual   `mapstructure:"virtuals" json:"virtuals"`
	Devices  []Device    `mapstructure:"devices" json:"devices"`
	Audio    AudioConfig `mapstructure:"audio" json:"audio"`
	// Config    string      `mapstructure:"config" json:"config"`
	// SentryCrash bool        `mapstructure:"sentry-crash-test" json:"sentry-crash-test"`
	// VeryVerbose bool        `mapstructure:"very-verbose" json:"very-verbose"`
}

var configPath string
var GlobalConfig *Config

var GlobalViper *viper.Viper

func InitConfig() error {
	GlobalViper = viper.New()

	pflag.StringVarP(&configPath, "config", "c", "", "Directory that contains the configuration files")
	pflag.IntP("port", "p", 8080, "Web interface port")
	pflag.BoolP("version", "v", false, "Print the version of ledfx")
	pflag.BoolP("open-ui", "u", false, "Automatically open the web interface")
	pflag.BoolP("verbose", "i", false, "Set log level to INFO")
	pflag.BoolP("very-verbose", "d", false, "Set log level to DEBUG")
	pflag.String("host", "", "The hostname of the web interface")
	pflag.BoolP("offline", "o", false, "Disable automated updates and sentry crash logger")
	pflag.BoolP("sentry-crash-test", "s", false, "This crashes LedFx to test the sentry crash logger")

	pflag.Parse()
	err := GlobalViper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return err
	}

	// Load config
	err = loadConfig("go_config")
	if err != nil {
		return err
	}

	return nil
}

func createConfigIfNotExists(configName string) error {
	// Create config dir and files if it does not exist
	_, err := os.Open(filepath.Join(configPath, configName+".json"))
	var f *os.File
	if _, ok := err.(*os.PathError); ok {
		f, err = os.Create(filepath.Join(configPath, configName+".json"))
		if err != nil {
			return err
		}
		_, err = f.WriteString("{}\n")
		if err != nil {
			return err
		}
		err = nil
	}
	return err
}

// LoadConfig reads in config file and ENV variables if set.
func loadConfig(configName string) (err error) {

	if configPath == "" {
		configPath = constants.GetOsConfigDir()
	}

	err = os.MkdirAll(configPath, 0744) // ensure given config directory exists
	if err != nil {
		return err
	}

	err = createConfigIfNotExists(configName)
	if err != nil {
		return err
	}

	v := GlobalViper

	if err != nil {
		return err
	}

	v.SetConfigName(configName)
	v.AutomaticEnv()
	v.AddConfigPath(configPath)
	err = v.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found; using defaults")
		}
		return nil
	}

	err = v.Unmarshal(&GlobalConfig)

	return
}
