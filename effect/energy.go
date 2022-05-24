package effect

import (
	"encoding/json"
	"fmt"
	"ledfx/color"

	"github.com/mitchellh/mapstructure"
)

type Energy struct {
	Effect
	Name   string
	Config EnergyConfig
}

type EnergyConfig struct {
	baseConfig EffectConfig
}

func (e *Energy) AssembleFrame(p *color.Pixels) {}
func (e *Energy) Initialize(id string)          {}
func (e *Energy) AudioUpdated()                 {}

/*
Updates the config of the effect. Config can be given
as EnergyConfig, map[string]interface{}, or raw json
*/
func (e *Energy) UpdateConfig(c interface{}) (err error) {
	var config EnergyConfig
	switch t := c.(type) {
	case EnergyConfig: // No conversion necessary
		config = c.(EnergyConfig)
	case map[string]interface{}: // Decode a map structure
		err = mapstructure.Decode(t, config)
	case []byte: // Expect a JSON type. Unmarshal and update values from json
		err = json.Unmarshal(t, &config)
	default:
		return fmt.Errorf("Invalid config type: %s", t)
	}
	// validate all values
	err = validate.Struct(&config)
	if err != nil {
		return err
	}

	// apply config to effect
	e.Config = config
	return nil
}
