package effect

import (
	"ledfx/color"
)

type Solid struct {
	Done chan bool
}

func (e *Solid) AssembleFrame(phase float64, ledCount int, effectColor color.Color) (colors []color.Color) {
	colors = make([]color.Color, ledCount)
	for i := range colors {
		colors[i] = effectColor
	}
	return colors
}
