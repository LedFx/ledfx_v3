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

type BaseDeviceConfig struct {
	PixelCount int    `mapstructure:"pixel_count" json:"pixel_count" description:"Number of pixels on the device" validate:"required,gte=10,lte=255"` // TODO be smarter about this
	Name       string `mapstructure:"name" json:"name" description:"Display name for the device" validate:"required"`
}

// the saved config entry for an effect
type EffectEntry struct {
	ID          string                 `mapstructure:"id" json:"id"`
	Type        string                 `mapstructure:"type" json:"type"`
	BaseConfig  map[string]interface{} `mapstructure:"base_config" json:"base_config"`
	ExtraConfig map[string]interface{} `mapstructure:"extra_config" json:"extra_config"`
}

type DeviceEntry struct {
	ID         string                 `mapstructure:"id" json:"id"`
	Type       string                 `mapstructure:"type" json:"type"`
	BaseConfig BaseDeviceConfig       `mapstructure:"base_config" json:"base_config"`
	ImplConfig map[string]interface{} `mapstructure:"impl_config" json:"impl_config"`
}
type VirtualEntry struct{}

func AddEntry(id string, entry interface{}) (err error) {
	mu.Lock()
	defer mu.Unlock()
	switch t := entry.(type) {
	case EffectEntry:
		store.Effects[id] = entry.(EffectEntry)
	case DeviceEntry:
		store.Devices[id] = entry.(DeviceEntry)
	case VirtualEntry:
		err = errors.New("virtual config entry not yet implemented")
	default:
		err = fmt.Errorf("unknown config entry type: %v", t)
	}
	if err != nil {
		return err
	}
	logger.Logger.WithField("context", "Config").Debugf("Saved %s to config", id)
	logger.Logger.WithField("context", "Config").Debug(entry)
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

func GetDevices() map[string]DeviceEntry {
	return store.Devices
}

func GetDevice(id string) (DeviceEntry, error) {
	if entry, ok := store.Devices[id]; ok {
		return entry, nil
	} else {
		return entry, fmt.Errorf("cannot retrieve device config of id: %s", id)
	}
}

// func GetDevices() map[string]DeviceEntry {
// 	return GlobalConfig.Devices
// }

// func GetVirtuals() map[string]VirtualEntry {
// 	return GlobalConfig.Virtuals
// }
