package config

import (
	"errors"
	"fmt"
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
	case DeviceEntry:
		err = errors.New("device config entry not yet implemented")
	case VirtualEntry:
		err = errors.New("virtual config entry not yet implemented")
	default:
		err = fmt.Errorf("unknown config entry type: %v", t)
	}
	return err
}

// func GetEffects() map[string]EffectEntry {
// 	return GlobalConfig.Effects
// }

// func GetDevices() map[string]DeviceEntry {
// 	return GlobalConfig.Devices
// }

// func GetVirtuals() map[string]VirtualEntry {
// 	return GlobalConfig.Virtuals
// }
