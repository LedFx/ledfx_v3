package effect

import (
	"github.com/LedFx/ledfx/pkg/color"
	"github.com/LedFx/ledfx/pkg/pixelgroup"
)

type Palette struct{}

// Apply new pixels to an existing pixel array.
func (e *Palette) assembleFrame(base *Effect, pg *pixelgroup.PixelGroup) {
	// operate on the largest pixel output in group, then clone to others
	p := pg.Group[pg.Largest]

	for i := range p {
		p[i] = color.Full
		p[i][0] = float64(i) / base.pixelScaler
	}
	pg.CloneToAll(pg.Largest)
}
