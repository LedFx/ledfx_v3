package effect

import (
	"ledfx/color"
)

type Strobe struct {
	ExtraConfig EnergyConfig
}

/*
You can define extra config specific for this effect here.
Try to keep these to a minimum and use the base config as much as possible.
*/
type StrobeConfig struct{}

// Apply new pixels to an existing pixel array.
func (e *Strobe) assembleFrame(base *Effect, p color.Pixels) {
	bkgb := base.Config.BkgBrightness //eg.
	p[0][0] = bkgb
}

func (e *Strobe) AudioUpdated() {

}
