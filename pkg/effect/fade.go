package effect

import (
	"github.com/LedFx/ledfx/pkg/color"
)

type Fade struct{}

// Apply new pixels to an existing pixel array.
func (e *Fade) assembleFrame(base *Effect, p color.Pixels) {
	for i := range p {
		p[i] = color.Full
	}
}
