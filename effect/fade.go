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
