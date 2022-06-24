package effect

import (
	"ledfx/color"
	"math"
)

type Pulse struct{}

// Apply new pixels to an existing pixel array.
func (e *Pulse) assembleFrame(base *Effect, p color.Pixels) {
	if 1-math.Mod(base.deltaStart.Seconds(), 1) < 0.1 {
		for i := range p {
			p[i] = color.Full
			p[i][0] = float64(i) / base.pixelScaler
		}
	}

}

// BOILERPLATE CODE BELOW. COPYPASTE & REPLACE CONFIG TYPE WITH THIS EFFECT'S CONFIG

/*
Updates the config of the effect. Config can be given
as Config, map[string]interface{}, or raw json
*/
func (e *Pulse) UpdateExtraConfig(c interface{}) (err error) { return nil }
