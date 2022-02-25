package effect

import (
	"ledfx/color"
)

type Solid struct {
	Done chan bool
}

func (e *Solid) AssembleFrame(phase float64, ledCount int, effectColor color.Color) (colors []color.Color) {
	data := []color.Color{}
	for i := 0; i < ledCount; i++ {
		data = append(data, effectColor)
	}
	return data
}
