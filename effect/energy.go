package effect

import (
	"encoding/json"
	"fmt"
	"ledfx/color"

	"github.com/mitchellh/mapstructure"
)

type Energy struct {
	ExtraConfig EnergyConfig
}

/*
You can define extra config specific for this effect here.
Try to keep these to a minimum and use the base config as much as possible.
*/
type EnergyConfig struct{}

// Apply new pixels to an existing pixel array.
func (e *Energy) assembleFrame(base *Effect, p color.Pixels) {
	bkgb := base.Config.BkgBrightness //eg.
	p[0][0] = bkgb
}

func (e *Energy) AudioUpdated() {}

// BOILERPLATE CODE BELOW. COPYPASTE & REPLACE CONFIG TYPE WITH THIS EFFECT'S CONFIG

/*
Updates the config of the effect. Config can be given
as EnergyConfig, map[string]interface{}, or raw json
*/
func (e *Energy) UpdateExtraConfig(c interface{}) (err error) {
	newConfig := e.ExtraConfig
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

	// READ ME
	// here you can update any stored properties that are based on the config
	// put any of your generated effect properties below:

	// apply config to effect
	e.ExtraConfig = newConfig
	return nil
}
