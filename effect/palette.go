package effect

import (
	"ledfx/color"
)

type Palette struct{}

// Apply new pixels to an existing pixel array.
func (e *Palette) assembleFrame(base *Effect, p color.Pixels) {
	for i := range p {
		p[i] = color.Full
		p[i][0] = float64(i) / base.pixelScaler
	}
}

// BOILERPLATE CODE BELOW. COPYPASTE & REPLACE CONFIG TYPE WITH THIS EFFECT'S CONFIG

/*
Updates the config of the effect. Config can be given
as Config, map[string]interface{}, or raw json
*/
func (e *Palette) UpdateExtraConfig(c interface{}) (err error) { return nil }
