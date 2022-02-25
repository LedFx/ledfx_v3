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
	for i := range colors {
		colors[i] = newColor
	}
	return colors
}
