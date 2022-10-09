package effect

import (
	"math"

	"github.com/LedFx/ledfx/pkg/audio"
	"github.com/LedFx/ledfx/pkg/logger"
	"github.com/LedFx/ledfx/pkg/render"
)

/*
Audio reactive port of PixelBlaze effect "Millipede"
Credit to Ben Henke for original concept & fantastic LED art
https://electromage.com/patterns
*/

type Millipede struct{}

// movement speed based on beat
// modify saturation by vocals/mids
// make the bars/legs grow and shrink
// modulate t1 with highs

// Apply new pixels to an existing pixel array.
func (e *Millipede) assembleFrame(base *Effect, pg *render.PixelGroup) {
	// operate on the largest pixel output in group, then clone to others
	p := pg.Group[pg.Largest]

	mel, err := audio.Analyzer.GetMelbank(base.ID)
	if err != nil {
		logger.Logger.WithField("context", "Effect").Error(err)
		return
	}

	t1 := base.time(0.05)
	t2 := base.time(0.1)

	for i := 0; i < len(p); i++ {
		fi := float64(i)
		h := math.Mod((fi+base.time(0.1)*base.pixelScaler)/base.pixelScaler*5, 0.5) + fi/base.pixelScaler + base.sin(t1)
		v := math.Pow(base.sin(h+t2), 2)

		p[i][0] = h + mel.HighAmplitude()
		p[i][1] = 1 - mel.LowsAmplitude()
		p[i][2] = v + mel.MidsAmplitude()
	}
	pg.CloneToAll(pg.Largest)
}
