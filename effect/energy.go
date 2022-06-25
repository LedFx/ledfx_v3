package effect

import (
	"encoding/json"
	"fmt"
	"ledfx/audio"
	"ledfx/color"
	"ledfx/logger"

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
	mel, err := audio.Analyzer.GetMelbank(base.ID)
	if err != nil {
		logger.Logger.WithField("context", "Effect Energy").Error(err)
		return
	}
	lowsCol := color.Color{0, 1, 1}
	midsCol := color.Color{0.5, 1, 1}
	highCol := color.Color{1, 1, 1}
	lowsMidsCol := color.Color{0.25, 1, 1}
	midsHighCol := color.Color{0.75, 1, 1}

	lowsAmplitude := int(mel.LowsAmplitude() * base.pixelScaler)
	midsAmplitude := int(mel.MidsAmplitude() * base.pixelScaler)
	highAmplitude := int(mel.HighAmplitude() * base.pixelScaler)

	var lows, mids, high bool
	for i := 0; i < len(p); i++ {
		lows = i < lowsAmplitude
		mids = i < midsAmplitude
		high = i < highAmplitude
		switch {
		// case !lows && !mids && !high: // none, dont update colour
		// 	// p[i] = color.Color{0, 0, 0}

		case lows && mids && high: // bass mids and high, white colour
			p[i] = color.Color{0, 0, 1}
		case lows && !mids && !high: // bass
			p[i] = lowsCol
		case lows && mids && !high: // mix bass and mids
			p[i] = lowsMidsCol
		case !lows && mids && !high: // mids
			p[i] = midsCol
		case !lows && mids && high: // mix mids and high
			p[i] = midsHighCol
		case !lows && !mids && high: // high
			p[i] = highCol
		}
	}
}

// func (e *Energy) AudioUpdated() {}

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
