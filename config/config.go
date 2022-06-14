package config

import (
	"ledfx/constants"
	"ledfx/logger"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const configName string = "config"

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

type config struct {
	Version  string                  `mapstructure:"version" json:"version"`
	Host     string                  `mapstructure:"host" json:"host"`
	Port     int                     `mapstructure:"port" json:"port"`
	OpenUi   bool                    `mapstructure:"open_ui" json:"open_ui"`
	LogLevel int                     `mapstructure:"log_level" json:"log_level"`
	Effects  map[string]EffectEntry  `mapstructure:"effects" json:"effects"`
	Devices  map[string]DeviceEntry  `mapstructure:"devices" json:"devices"`
	Virtuals map[string]VirtualEntry `mapstructure:"virtuals" json:"virtuals"`
	// Audio    AudioEntry              `mapstructure:"audio" json:"audio"`
	// Audio    AudioConfig             `mapstructure:"audio" json:"audio"`
}

var config_inst *config = &config{
	Version:  "",
	Host:     "",
	Port:     0,
	OpenUi:   false,
	LogLevel: 0,
	Effects:  map[string]EffectEntry{},
	Devices:  map[string]DeviceEntry{},
	Virtuals: map[string]VirtualEntry{},
}
var GlobalViper *viper.Viper

func init() {
	GlobalViper = viper.New()
	var configPath string
	pflag.StringVarP(&configPath, "config", "c", "", "Path to json configuration file")
	pflag.StringP("host", "h", "0.0.0.0", "The hostname of the web interface")
	pflag.IntP("port", "p", 8080, "Web interface port")
	pflag.BoolP("version", "v", false, "Print the version of ledfx")
	pflag.BoolP("open-ui", "u", false, "Automatically open the web interface")
	pflag.IntP("log_level", "l", 0, "Set log level [0: warnings, 1: info, 2: debug]")
	// pflag.BoolP("offline", "o", false, "Disable automated updates and sentry crash logger")

	pflag.Parse()
	err := GlobalViper.BindPFlags(pflag.CommandLine)
	if err != nil {
		logger.Logger.WithField("context", "Config Init").Fatal(err)
	}

	// Load config
	err = loadConfig(configPath)
	if err != nil {
		logger.Logger.WithField("context", "Config Init").Fatal(err)
	}
}

// LoadConfig reads in config file and ENV variables if set.
func loadConfig(configPath string) error {

	if configPath == "" {
		configPath = filepath.Join(constants.GetOsConfigDir(), configName+".json")
	}

	err := os.MkdirAll(configPath, 0744) // ensure given config directory exists
	if err != nil {
		return err
	}

	err = createConfigIfNotExists(configPath)
	if err != nil {
		return err
	}

	GlobalViper.SetConfigName(configName)
	GlobalViper.AutomaticEnv()
	GlobalViper.AddConfigPath(configPath)
	err = GlobalViper.ReadInConfig()
	if err != nil {
		return err
	}

	err = GlobalViper.Unmarshal(&config_inst)
	return err
}

func createConfigIfNotExists(configPath string) error {
	var f *os.File
	_, err := os.Open(configPath)
	// if the error is not related to finding the path...
	if _, ok := err.(*os.PathError); !ok {
		return err
	}
	// Create config dir and file given it does not exist
	logger.Logger.WithField("context", "Config Init").Warnf("Config file not found; Creating default config at %s", configPath)
	f, err = os.Create(configPath)
	if err != nil {
		return err
	}
	// write empty json to it
	_, err = f.WriteString("{}\n")
	return err
}
