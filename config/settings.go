package config

import (
	"ledfx/event"
	"ledfx/logger"
	"ledfx/util"
	"reflect"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
)

type SettingsConfig struct {
	Host     string `mapstructure:"host" json:"host" default:"0.0.0.0" validate:"ip" description:"Web interface hostname"`
	Port     int    `mapstructure:"port" json:"port" default:"8080" validate:"gte=0,lte=65535" description:"Web interface port"`
	NoLogo   bool   `mapstructure:"no_logo" json:"no_logo" default:"false" validate:"" description:"Hide the command line logo at startup"`
	NoUpdate bool   `mapstructure:"no_update" json:"no_update" default:"false" validate:"" description:"Disable automatic updates at startup"`
	NoTray   bool   `mapstructure:"no_tray" json:"no_tray" default:"false" validate:"" description:"Disable system tray icon to access LedFx"`
	NoScan   bool   `mapstructure:"no_scan" json:"no_scan" default:"false" validate:"" description:"Disable automatic WLED scanning and configuration in LedFx"`
	OpenUi   bool   `mapstructure:"open_ui" json:"open_ui" default:"false" validate:"" description:"Automatically open the web interface at startup"`
	LogLevel int    `mapstructure:"log_level" json:"log_level" default:"2" validate:"gte=0,lte=2" description:"Set log level [0: debug, 1: info, 2: warnings]"`
}

// Generate settings config schema
func SettingsSchema() (schema map[string]interface{}, err error) {
	return util.CreateSchema(reflect.TypeOf((*SettingsConfig)(nil)).Elem())
}

// Generate settings config schema as json
func CoreJsonSchema() (jsonSchema []byte, err error) {
	schema, err := SettingsSchema()
	if err != nil {
		return jsonSchema, err
	}
	jsonSchema, err = util.CreateJsonSchema(schema)
	return jsonSchema, err
}

// returns settings including those set by command line args
func GetSettings() SettingsConfig {
	settings := store.Settings
	// apply command line args which the user specified
	host := pflag.Lookup("host")
	port := pflag.Lookup("port")
	noLogo := pflag.Lookup("no_logo")
	noUpdate := pflag.Lookup("no_update")
	noScan := pflag.Lookup("no_scan")
	noTray := pflag.Lookup("no_tray")
	openUi := pflag.Lookup("open_ui")
	logLevel := pflag.Lookup("log_level")

	if host.Changed {
		settings.Host = hostArg
	}
	if port.Changed {
		settings.Port = portArg
	}
	if noLogo.Changed {
		settings.NoLogo = noLogoArg
	}
	if noUpdate.Changed {
		settings.NoUpdate = noUpdateArg
	}
	if noScan.Changed {
		settings.NoScan = noScanArg
	}
	if noTray.Changed {
		settings.NoTray = noTrayArg
	}
	if openUi.Changed {
		settings.OpenUi = openUiArg
	}
	if logLevel.Changed {
		settings.LogLevel = logLevelArg
	}
	return settings
}

func SetSettings(c map[string]interface{}) error {
	mu.Lock()
	defer mu.Unlock()
	prevSettings := store.Settings
	err := mapstructure.Decode(c, &store.Settings)
	if err != nil {
		logger.Logger.WithField("context", "Config").Warn(err)
		return err
	}
	err = validate.Struct(&store.Settings)
	if err != nil {
		store.Settings = prevSettings
		logger.Logger.WithField("context", "Config").Warn(err)
		return err
	}
	// if scan setting is changed, we need to handle it
	if GetSettings().NoScan != prevSettings.NoScan {
		switch store.Settings.NoScan {
		case false:
			util.EnableScan()
		case true:
			util.DisableScan()
		}
	}
	err = saveConfig()
	event.Invoke(event.SettingsUpdate,
		map[string]interface{}{
			"settings": store.Settings,
		})
	return err
}
