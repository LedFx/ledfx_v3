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
