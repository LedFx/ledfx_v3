package effect

import (
	"ledfx/color"
	"math"
)

type Weave struct {
	lowsPos int
	midsPos int
	highPos int
}

// Apply new pixels to an existing pixel array.
func (e *Weave) assembleFrame(base *Effect, p color.Pixels) {

	lowsNew := int(weavePosition(base.deltaStart.Seconds(), 4) * base.pixelScaler)
	midsNew := int(weavePosition(base.deltaStart.Seconds(), 2) * base.pixelScaler)
	highNew := int(weavePosition(base.deltaStart.Seconds(), 1) * base.pixelScaler)

	color.FillBetween(p, e.lowsPos, lowsNew, color.Color{0, 1, 1})
	color.FillBetween(p, e.midsPos, midsNew, color.Color{0.5, 1, 1})
	color.FillBetween(p, e.highPos, highNew, color.Color{1, 1, 1})

	for i := range p {
		p[i][1] += 0.1
		p[i][2] -= 0.01
	}

	// replace this with audio data
	e.lowsPos = lowsNew
	e.midsPos = midsNew
	e.highPos = highNew
}

// Position function. https://www.desmos.com/calculator/pacgvrebds
func weavePosition(pos float64, freq float64) float64 {
	return 1 - math.Abs(math.Mod(pos/freq, 2)-1)
}

// BOILERPLATE CODE BELOW. COPYPASTE & REPLACE CONFIG TYPE WITH THIS EFFECT'S CONFIG

/*
Updates the config of the effect. Config can be given
as Config, map[string]interface{}, or raw json
*/
func (e *Weave) UpdateExtraConfig(c interface{}) (err error) { return nil }
