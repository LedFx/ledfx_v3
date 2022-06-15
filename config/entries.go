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
	ID          string                 `mapstructure:"id" json:"id"`
	Type        string                 `mapstructure:"type" json:"type"`
	BaseConfig  map[string]interface{} `mapstructure:"base_config" json:"base_config"`
	ExtraConfig map[string]interface{} `mapstructure:"extra_config" json:"extra_config"`
}

type DeviceEntry struct{}
type VirtualEntry struct{}

func AddEntry(id string, entry interface{}) (err error) {
	switch t := entry.(type) {
	case EffectEntry:
		store.Effects[id] = entry.(EffectEntry)
		logger.Logger.WithField("context", "Config").Debugf("Saved %s to config", id)
		logger.Logger.WithField("context", "Config").Debug(entry)
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
		logger.Logger.WithField("context", "Config").Warn(err)
		return err
	}
	err = validate.Struct(&c)
	if err != nil {
		logger.Logger.WithField("context", "Config").Warn(err)
		return err
	}
	store.Core = core
	return nil
}

func DeleteEffect(id string) {
	if _, exists := store.Effects[id]; !exists {
		return
	}
	delete(store.Effects, id)
	logger.Logger.WithField("context", "Config").Debugf("Deleted %s from config", id)
	saveConfig()
}

func GetEffects() map[string]EffectEntry {
	return store.Effects
}

func GetEffect(id string) (EffectEntry, error) {
	if entry, ok := store.Effects[id]; ok {
		return entry, nil
	} else {
		return entry, fmt.Errorf("cannot retrieve effect config of id: %s", id)
	}

}

// func GetDevices() map[string]DeviceEntry {
// 	return GlobalConfig.Devices
// }

// func GetVirtuals() map[string]VirtualEntry {
// 	return GlobalConfig.Virtuals
// }
