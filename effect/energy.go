package effect

import (
	"encoding/json"
	"fmt"
	"ledfx/color"

	"github.com/creasty/defaults"
	"github.com/mitchellh/mapstructure"
)

type Energy struct {
	Effect
	Name   string
	Config EnergyConfig
}

type EnergyConfig struct {
	EffectConfig `mapstructure:",squash"`
}

func (e *Energy) AssembleFrame(p *color.Pixels) {
	bkgb := e.Config.BkgBrightness //eg.
	fmt.Println(bkgb)
	e.Postprocess(p)
}
func (e *Energy) Initialize() error {
	return defaults.Set(&e.Config)
}
func (e *Energy) AudioUpdated() {}

// BOILERPLATE CODE BELOW. COPYPASTE & REPLACE CONFIG TYPE WITH THIS EFFECT'S CONFIG

/*
Updates the config of the effect. Config can be given
as EnergyConfig, map[string]interface{}, or raw json
*/
func (e *Energy) UpdateConfig(c interface{}) (err error) {
	newConfig := e.Config
	switch t := c.(type) {
	case EnergyConfig: // No conversion necessary
		newConfig = c.(EnergyConfig)
	case map[string]interface{}: // Decode a map structure
		err = mapstructure.Decode(t, &newConfig)
	case []byte: // Unmarshal a json byte slice
		err = json.Unmarshal(t, &newConfig)
	default:
		err = fmt.Errorf("invalid config type: %s", t)
	}
	if err != nil {
		return err
	}

	// validate all values
	err = validate.Struct(&newConfig)
	if err != nil {
		return err
	}

	// apply config to effect
	e.Config = newConfig
	return nil
}
