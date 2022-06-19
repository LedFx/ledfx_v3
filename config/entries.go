package config

import (
	"errors"
	"fmt"
	"ledfx/logger"
)

type EntryType int

const (
	Effect EntryType = iota
	Device
	Virtual
)

func (e EntryType) String() string {
	switch e {
	case Effect:
		return "effect"
	case Device:
		return "device"
	case Virtual:
		return "virtual"
	default:
		return "unknown"
	}
}

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
	mu.Lock()
	defer mu.Unlock()
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

func DeleteEntry(t EntryType, id string) {
	mu.Lock()
	defer mu.Unlock()
	switch t {
	case Effect:
		if _, exists := store.Effects[id]; !exists {
			return
		}
		delete(store.Effects, id)
	case Device:
		if _, exists := store.Devices[id]; !exists {
			return
		}
		delete(store.Devices, id)
	case Virtual:
		if _, exists := store.Virtuals[id]; !exists {
			return
		}
		delete(store.Virtuals, id)
	}
	logger.Logger.WithField("context", "Config").Debugf("Deleted %s %s from config", t.String(), id)
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
