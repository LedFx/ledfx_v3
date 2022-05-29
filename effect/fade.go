package effect

import (
	"ledfx/color"
)

type Fade struct{}

// Apply new pixels to an existing pixel array.
func (e *Fade) assembleFrame(base *Effect, p color.Pixels) {
	for i := range p {
		p[i] = color.Full
	}
}

// BOILERPLATE CODE BELOW. COPYPASTE & REPLACE CONFIG TYPE WITH THIS EFFECT'S CONFIG

/*
Updates the config of the effect. Config can be given
as Config, map[string]interface{}, or raw json
*/
func (e *Fade) UpdateExtraConfig(c interface{}) (err error) { return nil }
