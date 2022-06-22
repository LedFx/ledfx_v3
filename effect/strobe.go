package effect

import (
	"ledfx/color"
)

type Strobe struct{}

/*
You can define extra config specific for this effect here.
Try to keep these to a minimum and use the base config as much as possible.
*/
type StrobeConfig struct{}

// Apply new pixels to an existing pixel array.
func (e *Strobe) assembleFrame(base *Effect, p color.Pixels) {
	bkgb := base.Config.BkgBrightness //eg.
	p[0][0] = bkgb
}

func (e *Strobe) AudioUpdated() {}

// BOILERPLATE CODE BELOW. COPYPASTE & REPLACE CONFIG TYPE WITH THIS EFFECT'S CONFIG

/*
Updates the config of the effect. Config can be given
as Config, map[string]interface{}, or raw json
*/
func (e *Strobe) UpdateExtraConfig(c interface{}) (err error) { return nil }
