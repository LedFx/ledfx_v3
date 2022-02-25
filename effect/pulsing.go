package effect

import (
	"ledfx/color"
	"math"
)

type PulsingEffect struct {
	Done chan bool
}

func (e *PulsingEffect) AssembleFrame(phase float64, ledCount int, effectColor color.Color) (colors []color.Color) {
	colors = make([]color.Color, ledCount)
	newColor := color.Color{
		0.5 * (math.Sin(phase) + 1) * effectColor[0],
		0.5 * (math.Sin(phase) + 1) * effectColor[1],
		0.5 * (math.Sin(phase) + 1) * effectColor[2],
	}
	for i := 0; i < ledCount; i++ {
		// calculate the LED values for the effect's current frame
		// TODO: this brightness calculation might be better done using HSV instead of RGB
		colors[i] = newColor
	}
	return colors
}
