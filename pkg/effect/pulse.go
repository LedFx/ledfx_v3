package effect

import (
	"math"

	"github.com/LedFx/ledfx/pkg/color"
)

type Pulse struct{}

// Apply new pixels to an existing pixel array.
func (e *Pulse) assembleFrame(base *Effect, p color.Pixels) {
	if 1-math.Mod(base.deltaStart.Seconds(), 5.1-base.Config.Intensity*5) < 0.1 {
		for i := range p {
			p[i] = color.Full
			p[i][0] = float64(i) / base.pixelScaler
		}
	}
}
