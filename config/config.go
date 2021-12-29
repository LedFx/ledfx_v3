package config

import (
	"ledfx/constants"
	"log"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type DeviceConfig struct {
	CenterOffset   int    `mapstructure:"center_offset"`
	ForceRefresh   bool   `mapstructure:"force_refresh"`
	IconName       string `mapstructure:"icon_name"`
	IncludeIndexes bool   `mapstructure:"include_indexes"`
	IpAddress      string `mapstructure:"ip_address"`
	MaxBrightness  int    `mapstructure:"max_brightness"`
	Name           string `mapstructure:"name"`
	PixelCount     int    `mapstructure:"pixel_count"`
	Port           int    `mapstructure:"port"`
	PreviewOnly    bool   `mapstructure:"preview_only"`
	RefreshRate    int    `mapstructure:"refresh_rate"`
	Type           string `mapstructure:"type"`
	UdpPacketType  string `mapstructure:"udp_packet_type"`
}

type EffectConfig struct {
	BackgroundColor string `mapstructure:"background_color"`
	GradientName    string `mapstructure:"gradient_name"`
}

type Effect struct {
	Config EffectConfig `mapstructure:"config"`
	Type   string       `mapstructure:"type"`
}

type Device struct {
	Config DeviceConfig `mapstructure:"config"`
	Effect Effect       `mapstructure:"effect"`
	Id     string       `mapstructure:"id"`
	Type   string       `mapstructure:"type"`
}

type Config struct {
	Config      string   `mapstructure:"config"`
	Port        int      `mapstructure:"port"`
	Version     bool     `mapstructure:"version"`
	OpenUi      bool     `mapstructure:"open-ui"`
	Verbose     bool     `mapstructure:"verbose"`
	VeryVerbose bool     `mapstructure:"very-verbose"`
	Host        string   `mapstructure:"host"`
	Offline     bool     `mapstructure:"offline"`
	SentryCrash bool     `mapstructure:"sentry-crash-test"`
	Devices     []Device `mapstructure:"devices"`
}

var configPath string
var GlobalConfig Config

func InitFlags() error {
	pflag.StringVarP(&configPath, "config", "c", "", "Directory that contains the configuration files")
	pflag.IntP("port", "p", 8000, "Web interface port")
	pflag.BoolP("version", "v", false, "Print the version of ledfx")
	pflag.BoolP("open-ui", "u", false, "Automatically open the web interface")
	pflag.BoolP("verbose", "i", false, "Set log level to INFO")
	pflag.BoolP("very-verbose", "d", false, "Set log level to DEBUG")
	pflag.String("host", "", "The hostname of the web interface")
	pflag.BoolP("offline", "o", false, "Disable automated updates and sentry crash logger")
	pflag.BoolP("sentry-crash-test", "s", false, "This crashes LedFx to test the sentry crash logger")

	pflag.Parse()
	return viper.BindPFlags(pflag.CommandLine)
}

// LoadConfig reads in config file and ENV variables if set.
func LoadConfig() (err error) {
	viper.SetConfigName("config")
	viper.AutomaticEnv()
	if configPath != "" {
		viper.AddConfigPath(configPath)
	} else {
		viper.AddConfigPath(constants.GetOsConfigDir())
	}
	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found; using defaults")
		}
		return nil
	}

	err = viper.Unmarshal(&GlobalConfig)
	return
}

func AddDevice(device Device) (err error) {
	if GlobalConfig.Devices == nil {
		GlobalConfig.Devices = make([]Device, 0)
	}
	GlobalConfig.Devices = append(GlobalConfig.Devices, device)
	viper.Set("devices", GlobalConfig.Devices)
	log.Println(GlobalConfig.Devices)
	err = viper.WriteConfig()
	return
}
