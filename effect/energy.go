package effect

import (
	"encoding/json"
	"fmt"
	"ledfx/color"
	"time"

	"github.com/creasty/defaults"
	"github.com/mitchellh/mapstructure"
)

type Energy struct {
	Effect
	Name         string
	GlobalConfig GlobalEffectConfig
	ExtraConfig  EnergyConfig
}

// you can redefine defaults of base effect config to better suit the effect
// eg. here, i'm setting mirror to default to true
type EnergyGlobalConfig struct {
	EffectConfig `mapstructure:",squash"`
	Mirror       bool `mapstructure:"mirror" json:"mirror" description:"Mirror the pixels across the center" default:"true" validate:""`
}

// Apply new pixels to an existing pixel array.
func (e *Energy) assembleFrame(p color.Pixels) {
	bkgb := e.Config.BkgBrightness //eg.
	p[0][0] = bkgb
}

func (e *Energy) AudioUpdated() {}

// BOILERPLATE CODE BELOW. COPYPASTE & REPLACE CONFIG TYPE WITH THIS EFFECT'S CONFIG
func (e *Energy) Initialize(id string, npx int) error {
	e.ID = id
	e.pixelCount = npx
	e.startTime = time.Now()
	e.prevFrame = make(color.Pixels, npx)
	e.mirror = make(color.Pixels, npx)
	return defaults.Set(&e.Config)
}

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

	// READ ME
	// here you can update any stored properties that are based on the config
	// creating a new palette is expensive, should only be done if changed
	if e.Config.Palette != newConfig.Palette {
		e.palette, _ = color.NewPalette(newConfig.Palette)
	}
	// parsing a color is cheap, just do it every time
	e.bkgColor, _ = color.NewColor(e.Config.BkgColor)
	// put any of your generated effect properties below:

	// apply config to effect
	e.Config = newConfig
	return nil
}
