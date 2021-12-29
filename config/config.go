package config

import (
	"ledfx/constants"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type DeviceConfig struct {
	CenterOffset   int    `mapstructure:"center_offset" json:"center_offset"`
	ForceRefresh   bool   `mapstructure:"force_refresh" json:"force_refresh"`
	IconName       string `mapstructure:"icon_name" json:"icon_name"`
	IncludeIndexes bool   `mapstructure:"include_indexes" json:"include_indexes"`
	IpAddress      string `mapstructure:"ip_address" json:"ip_address"`
	MaxBrightness  int    `mapstructure:"max_brightness" json:"max_brightness"`
	Name           string `mapstructure:"name" json:"name"`
	PixelCount     int    `mapstructure:"pixel_count" json:"pixel_count"`
	Port           int    `mapstructure:"port" json:"port"`
	PreviewOnly    bool   `mapstructure:"preview_only" json:"preview_only"`
	RefreshRate    int    `mapstructure:"refresh_rate" json:"refresh_rate"`
	Type           string `mapstructure:"type" json:"type"`
	UdpPacketType  string `mapstructure:"udp_packet_type" json:"udp_packet_type"`
}

type EffectConfig struct {
	BackgroundColor string `mapstructure:"background_color" json:"background_color"`
	GradientName    string `mapstructure:"gradient_name" json:"gradient_name"`
}

type Effect struct {
	Config EffectConfig `mapstructure:"config" json:"config"`
	Type   string       `mapstructure:"type" json:"type"`
}

type Device struct {
	Config DeviceConfig `mapstructure:"config" json:"config"`
	Effect Effect       `mapstructure:"effect" json:"effect"`
	Id     string       `mapstructure:"id" json:"id"`
	Type   string       `mapstructure:"type" json:"type"`
}

type Config struct {
	Config      string   `mapstructure:"config" json:"config"`
	Port        int      `mapstructure:"port" json:"port"`
	Version     bool     `mapstructure:"version" json:"version"`
	OpenUi      bool     `mapstructure:"open-ui" json:"open-ui"`
	Verbose     bool     `mapstructure:"verbose" json:"verbose"`
	VeryVerbose bool     `mapstructure:"very-verbose" json:"very-verbose"`
	Host        string   `mapstructure:"host" json:"host"`
	Offline     bool     `mapstructure:"offline" json:"offline"`
	SentryCrash bool     `mapstructure:"sentry-crash-test" json:"sentry-crash-test"`
	Devices     []Device `mapstructure:"devices" json:"devices"`
}

var configPath string
var GlobalConfig Config
var OldConfig Config
var GlobalViper *viper.Viper
var OldViper *viper.Viper

func InitConfig() error {
	GlobalViper = viper.New()
	OldViper = viper.New()

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
	return GlobalViper.BindPFlags(pflag.CommandLine)
}

// LoadConfig reads in config file and ENV variables if set.
// TODO: once we are fully backwards compatible, we can just use config.json for GlobalConfig
func LoadConfig(configName string) (err error) {

	if configPath == "" {
		configPath = constants.GetOsConfigDir()
	}

	// Create config dir and files if it does not exist
	_, err = os.Open(filepath.Join(configPath, "config.json"))
	var f *os.File
	if _, ok := err.(*os.PathError); ok {
		f, err = os.Create(filepath.Join(configPath, "config.json"))
		if err != nil {
			return err
		}
		_, err = f.WriteString("{}\n")
		if err != nil {
			return err
		}
		err = nil
	}
	_, err = os.Open(filepath.Join(configPath, "goconfig.json"))
	if _, ok := err.(*os.PathError); ok {
		f, err = os.Create(filepath.Join(configPath, "goconfig.json"))
		if err != nil {
			return err
		}
		_, err = f.WriteString("{}\n")
		if err != nil {
			return err
		}
		err = nil
	}

	var v *viper.Viper

	if configName == "goconfig" {
		v = GlobalViper
	} else if configName == "config" {
		v = OldViper
	}
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

	if configName == "goconfig" {
		err = v.Unmarshal(&GlobalConfig)
	} else if configName == "config" {
		err = v.Unmarshal(&OldConfig)
	}
	return
}
