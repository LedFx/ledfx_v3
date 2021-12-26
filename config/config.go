package config

import (
	"ledfx/constants"
	"log"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Config      string `mapstructure:"config"`
	Port        int    `mapstructure:"port"`
	Version     bool   `mapstructure:"version"`
	OpenUi      bool   `mapstructure:"open-ui"`
	Verbose     bool   `mapstructure:"verbose"`
	VeryVerbose bool   `mapstructure:"very-verbose"`
	Host        string `mapstructure:"host"`
	Offline     bool   `mapstructure:"offline"`
	SentryCrash bool   `mapstructure:"sentry-crash-test"`
}

var configPath string

func InitFlags() {
	pflag.StringVarP(&configPath, "config", "c", constants.GetOsConfigDir(), "Directory that contains the configuration files")
	pflag.IntP("port", "p", 8000, "Web interface port")
	pflag.BoolP("version", "v", false, "Print the version of ledfx")
	pflag.BoolP("open-ui", "u", false, "Automatically open the web interface")
	pflag.BoolP("verbose", "i", false, "Set log level to INFO")
	pflag.BoolP("very-verbose", "d", false, "Set log level to DEBUG")
	pflag.StringP("host", "h", "", "The hostname of the web interface")
	pflag.BoolP("offline", "o", false, "Disable automated updates and sentry crash logger")
	pflag.BoolP("sentry-crash-test", "s", false, "This crashes LedFx to test the sentry crash logger")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
}

// LoadConfig reads in config file and ENV variables if set.
func LoadConfig() (config Config, err error) {
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
		return
	}

	err = viper.Unmarshal(&config)
	return
}
