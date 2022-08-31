package config

import (
	"fmt"

	"github.com/LedFx/ledfx/pkg/logger"
)

type EntryType int

const (
	Effect EntryType = iota
	Device
	Controller
)

func (e EntryType) String() string {
	switch e {
	case Effect:
		return "effect"
	case Device:
		return "device"
	case Controller:
		return "controller"
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

type DeviceEntry struct {
	ID         string                 `mapstructure:"id" json:"id"`
	Type       string                 `mapstructure:"type" json:"type"`
	BaseConfig map[string]interface{} `mapstructure:"base_config" json:"base_config"`
	ImplConfig map[string]interface{} `mapstructure:"impl_config" json:"impl_config"`
}

type ControllerEntry struct {
	ID     string                 `mapstructure:"id" json:"id"`
	Config map[string]interface{} `mapstructure:"base_config" json:"base_config"`
}

func AddEntry(id string, entry interface{}) (err error) {
	mu.Lock()
	defer mu.Unlock()
	switch t := entry.(type) {
	case EffectEntry:
		store.Effects[id] = entry.(EffectEntry)
	case DeviceEntry:
		store.Devices[id] = entry.(DeviceEntry)
	case ControllerEntry:
		store.Controllers[id] = entry.(ControllerEntry)
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
	case Controller:
		if _, exists := store.Controllers[id]; !exists {
			return
		}
		delete(store.Controllers, id)
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

func GetControllers() map[string]ControllerEntry {
	return store.Controllers
}

func GetController(id string) (ControllerEntry, error) {
	if entry, ok := store.Controllers[id]; ok {
		return entry, nil
	} else {
		return entry, fmt.Errorf("cannot retrieve controller config of id: %s", id)
	}
}
