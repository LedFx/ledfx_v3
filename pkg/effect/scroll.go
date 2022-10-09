package effect

import (
	"github.com/LedFx/ledfx/pkg/audio"
	"github.com/LedFx/ledfx/pkg/color"
	"github.com/LedFx/ledfx/pkg/logger"
	"github.com/LedFx/ledfx/pkg/render"
)

type Scroll struct {
	initialised bool
	scroller    color.Pixels // hi res virtual pixels for consistent speed across different sized strips
}

// Apply new pixels to an existing pixel array.
func (e *Scroll) assembleFrame(base *Effect, pg *render.PixelGroup) {
	if !e.initialised {
		e.scroller = make(color.Pixels, 1000)
		e.initialised = true
	}

	// operate on the largest pixel output in group, then clone to others
	p := pg.Group[pg.Largest]

	mel, err := audio.Analyzer.GetMelbank(base.ID)
	if err != nil {
		logger.Logger.WithField("context", "Effect Scroll").Error(err)
		return
	}

	// make a new color based on the volume and frequency composition
	value := audio.Analyzer.Vol.Timestep
	hue := mel.LowsAmplitude() + mel.MidsAmplitude() + mel.HighAmplitude()
	newCol := color.Color{hue, 1, value}

	// rotate scroller array
	shift := int(10*base.Config.Intensity + 1)
	e.scroller = append(e.scroller[1000-shift:], e.scroller[:1000-shift]...)

	for i := 0; i < shift; i++ {
		e.scroller[i] = newCol
	}

	err = color.Interpolate(e.scroller, p)
	if err != nil {
		logger.Logger.WithField("context", "Effect Scroll").Error(err)
		return
	}

	pg.CloneToAll(pg.Largest)
}
