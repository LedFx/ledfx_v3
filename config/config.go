package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"ledfx/constants"
	"ledfx/logger"
	"os"
	"path/filepath"
	"sync"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/pflag"
)

const configName string = "config"

// config values which can be set by command line args
var (
	configPath  string
	hostArg     string
	portArg     int
	noLogoArg   bool
	noTrayArg   bool
	noUpdateArg bool
	noScanArg   bool
	openUiArg   bool
	logLevelArg int
)

var mu sync.Mutex = sync.Mutex{}
var validate *validator.Validate = validator.New()
var store *config = &config{
	Settings: SettingsConfig{},
	Effects:  map[string]EffectEntry{},
	Devices:  map[string]DeviceEntry{},
	Virtuals: map[string]VirtualEntry{},
}

type BaseDeviceConfig struct {
	PixelCount int    `mapstructure:"pixel_count" json:"pixel_count" description:"Number of pixels on the device" validate:"required"` // TODO be smarter about this
	Name       string `mapstructure:"name" json:"name" description:"Display name for the device" validate:"required"`
}

type VirtualConfig struct {
	Name      string `mapstructure:"name" json:"name" description:"Display name for the virtual" validate:"required"`
	IconName  string `mapstructure:"icon_name" json:"icon_name" description:"Icon name to identify this virtual" default:"alert-circle-outline" validate:""`
	FrameRate int    `mapstructure:"framerate" json:"framerate" description:"Target framerate" default:"60" validate:"gte=5,lte=120"`
	// Span      bool            `mapstructure:"span" json:"span"`
	// Outputs   []VirtualOutput `mapstructure:"outputs" json:"outputs"`
}

type config struct {
	//Version  string                  `mapstructure:"version" json:"version"`
	Settings      SettingsConfig          `mapstructure:"core" json:"core"`
	Frontend      FrontendConfig          `mapstructure:"frontend" json:"frontend"`
	Effects       map[string]EffectEntry  `mapstructure:"effects" json:"effects"`
	Devices       map[string]DeviceEntry  `mapstructure:"devices" json:"devices"`
	Virtuals      map[string]VirtualEntry `mapstructure:"virtuals" json:"virtuals"`
	EffectsGlobal map[string]interface{}  `mapstructure:"global_effects" json:"global_effects"`
	ConnEffect    map[string]string       `mapstructure:"connections_effect" json:"connections_effect"`
	ConnDevice    map[string]string       `mapstructure:"connections_device" json:"connections_device"`
	VirtStates    map[string]bool         `mapstructure:"virtual_states" json:"virtual_states"`
	LocalInput    string                  `mapstructure:"local_input" json:"local_input"`
	// Audio    AudioEntry              `mapstructure:"audio" json:"audio"`
	// Audio    AudioConfig             `mapstructure:"audio" json:"audio"`
}

/* Populates the config store (live config in memory).
1. set config store to defaults
2. update with any values from the config file
3. update with any command line args
*/
func init() {
	// special args
	var version bool

	pflag.CommandLine.SortFlags = false

	pflag.BoolVarP(&version, "version", "v", false, "Print the version of LedFx and exit")
	pflag.StringVarP(&configPath, "config", "c", "", "Path to json configuration file")
	pflag.StringVarP(&hostArg, "host", "h", "0.0.0.0", "Web interface hostname")
	pflag.IntVarP(&portArg, "port", "p", 8080, "Web interface port")
	pflag.BoolVarP(&noLogoArg, "no_logo", "n", false, "Hide the command line logo at startup")
	pflag.BoolVarP(&noTrayArg, "no_tray", "t", false, "Disable system tray icon to access LedFx")
	pflag.BoolVarP(&noUpdateArg, "no_update", "u", false, "Disable automatic updates at startup")
	pflag.BoolVarP(&noScanArg, "no_scan", "s", false, "Disable automatic WLED scanning and configuration in LedFx")
	pflag.BoolVarP(&openUiArg, "open_ui", "o", false, "Automatically open the web interface at startup")
	pflag.IntVarP(&logLevelArg, "log_level", "l", 2, "Set log level [0: debug, 1: info, 2: warnings]")
	// pflag.BoolP("offline", "o", false, "Disable automated updates and sentry crash logger")

	pflag.Parse()

	// Just print version and exit if flag is set
	if version {
		fmt.Println("LedFx " + constants.VERSION)
		os.Exit(0)
	}

	// validate all the command line args
	SettingsConfigArgs := SettingsConfig{
		Host:     hostArg,
		Port:     portArg,
		NoLogo:   noLogoArg,
		OpenUi:   openUiArg,
		NoUpdate: noUpdateArg,
		LogLevel: logLevelArg,
	}
	err := validate.Struct(&SettingsConfigArgs)
	if err != nil {
		logger.Logger.WithField("context", "Command Line Arguments").Fatal(err)
	}

	// apply defaults to the config
	err = defaults.Set(store)
	if err != nil {
		logger.Logger.WithField("context", "Config").Fatal(err)
	}

	// load any config saved on file
	loadConfig()
	// TODO validate config loaded from json

	logger.Logger.WithField("context", "Config").Infof("Initialised config")
}

// LoadConfig reads in config file and populates the config instance.
func loadConfig() {

	// make sure config file can be opened
	if err := ensureConfigFile(); err != nil {
		logger.Logger.WithField("context", "Config").Fatal(err)
	}

	// read the contents
	logger.Logger.WithField("context", "Config").Infof("Loading config file: %s", configPath)
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		logger.Logger.WithField("context", "Config").Fatal("Error reading config file: ", err)
	}

	// parse as json
	// unknown keys will be ignored
	err = json.Unmarshal(content, &store)
	if err != nil {
		logger.Logger.WithField("context", "Config").Fatal("Error parsing config file: ", err)
	}
}

// Makes sure that a config can be opened at the configPath
func ensureConfigFile() error {
	// if user supplied a config, simply test if we can open it
	if configPath != "" {
		_, err := os.Open(configPath)
		return err
	}
	// if not supplied, make sure we have a config file in the default location
	configDir := constants.GetOsConfigDir()
	configPath = filepath.Join(configDir, configName+".json")
	// first, ensure config directory exists
	err := os.MkdirAll(configDir, 0744)
	if err != nil {
		return err
	}
	// try to open config file in the default directory
	_, err = os.Open(configPath)
	if err == nil { // if it exists and we can open it, we're good to go
		return err
	}
	// if it doesn't exist, create it
	logger.Logger.WithField("context", "Config").Warn("Config file not found")
	logger.Logger.WithField("context", "Config").Warnf("Creating blank config at %s", configPath)
	_, err = os.Create(configPath)
	if err != nil {
		logger.Logger.WithField("context", "Config").Errorf("Failed to create blank config at %s", configPath)
	}
	// finally, test we can open the new blank config and write empty config to it
	f, err := os.Open(configPath)
	f.Close()
	saveConfig()
	return err
}

func saveConfig() error {
	file, _ := json.MarshalIndent(store, "", "  ")
	err := ioutil.WriteFile(configPath, file, 0644)
	if err != nil {
		logger.Logger.WithField("context", "Config").Warnf("Failed to save config to file at %s", configPath)
	}
	logger.Logger.WithField("context", "Config").Debugf("Saved config")
	return err
}
