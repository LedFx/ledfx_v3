package effect

import (
	"ledfx/color"
)

type Strobe struct{}

// Apply new pixels to an existing pixel array.
func (e *Strobe) assembleFrame(base *Effect, p color.Pixels) {
	bkgb := base.Config.BkgBrightness //eg.
	p[0][0] = bkgb
}
