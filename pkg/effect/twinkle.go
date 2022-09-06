package effect

import (
	"math/rand"

	"github.com/LedFx/ledfx/pkg/audio"
	"github.com/LedFx/ledfx/pkg/logger"
	"github.com/LedFx/ledfx/pkg/pixelgroup"
)

type Twinkle struct {
	initialised bool
	hues        []float64
	periods     []float64
}

// Apply new pixels to an existing pixel array.
func (e *Twinkle) assembleFrame(base *Effect, pg *pixelgroup.PixelGroup) {
	// operate on the largest pixel output in group, then clone to others
	p := pg.Group[pg.Largest]

	if !e.initialised {
		e.hues = make([]float64, len(p))
		e.periods = make([]float64, len(p))
		for i := 0; i < len(p); i++ {
			e.hues[i] = rand.Float64()
			e.periods[i] = 0.005 + rand.Float64()/65
		}
		e.initialised = true
	}

	mel, err := audio.Analyzer.GetMelbank(base.ID)
	if err != nil {
		logger.Logger.WithField("context", "Effect Energy").Error(err)
		return
	}

	var value float64
	for i := 0; i < len(p); i++ {
		hue := e.hues[i]
		switch {
		case hue < 0.3:
			value = base.triangle(base.time(e.periods[i])) * mel.LowsAmplitude()
		case hue < 0.6:
			value = base.triangle(base.time(e.periods[i])) * mel.MidsAmplitude()
		default:
			value = base.triangle(base.time(e.periods[i])) * mel.HighAmplitude()
		}
		p[i][0] = hue
		p[i][1] = 1
		p[i][2] = value
	}
	pg.CloneToAll(pg.Largest)
}
