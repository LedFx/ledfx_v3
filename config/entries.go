package config

import (
	"errors"
	"fmt"
	"ledfx/logger"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
)

// the saved config entry for an effect
type EffectEntry struct {
	Type        string                 `mapstructure:"type" json:"type"`
	BaseConfig  map[string]interface{} `mapstructure:"base_config" json:"base_config"`
	ExtraConfig map[string]interface{} `mapstructure:"extra_config" json:"extra_config"`
}

type DeviceEntry struct{}
type VirtualEntry struct{}

func AddEntry(id string, entry interface{}) (err error) {
	switch t := entry.(type) {
	case EffectEntry:
		logger.Logger.WithField("context", "Config").Infof("Saving effect entry with id %s", id)
		logger.Logger.WithField("context", "Config").Debug(entry)
		store.Effects[id] = entry.(EffectEntry)
	case DeviceEntry:
		err = errors.New("device config entry not yet implemented")
	case VirtualEntry:
		err = errors.New("virtual config entry not yet implemented")
	default:
		err = fmt.Errorf("unknown config entry type: %v", t)
	}
	if err != nil {
		return err
	}
	return saveConfig()
}

func GetCore() CoreConfig {
	core := store.Core
	// apply command line args which the user specified
	host := pflag.Lookup("host")
	port := pflag.Lookup("port")
	openUi := pflag.Lookup("open_ui")
	logLevel := pflag.Lookup("log_level")

	if host.Changed {
		core.Host = hostArg
	}
	if port.Changed {
		core.Port = portArg
	}
	if openUi.Changed {
		core.OpenUi = openUiArg
	}
	if logLevel.Changed {
		core.LogLevel = logLevelArg
	}
	return core
}

func SetCore(c map[string]interface{}) error {
	core := store.Core
	err := mapstructure.Decode(c, &core)
	if err != nil {
		logger.Logger.WithField("context", "Save Core Config").Warn(err)
		return err
	}
	err = validate.Struct(&c)
	if err != nil {
		logger.Logger.WithField("context", "Save Core Config").Warn(err)
		return err
	}
	store.Core = core
	return nil
}

func DeleteEffect(id string) {
	delete(store.Effects, id)
	saveConfig()
}

func GetEffects() map[string]EffectEntry {
	return store.Effects
}

// func GetDevices() map[string]DeviceEntry {
// 	return GlobalConfig.Devices
// }

// func GetVirtuals() map[string]VirtualEntry {
// 	return GlobalConfig.Virtuals
// }
