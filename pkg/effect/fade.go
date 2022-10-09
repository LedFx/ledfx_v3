package effect

import (
	"github.com/LedFx/ledfx/pkg/color"
	"github.com/LedFx/ledfx/pkg/render"
)

type Fade struct{}

// Apply new pixels to an existing pixel array.
func (e *Fade) assembleFrame(base *Effect, pg *render.PixelGroup) {
	// operate on the largest pixel output in group, then clone to others
	p := pg.Group[pg.Largest]
	for i := 0; i < len(p); i++ {
		p[i] = color.Full
	}
	pg.CloneToAll(pg.Largest)
}
