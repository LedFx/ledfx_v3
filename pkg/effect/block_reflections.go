package effect

import (
	"math"

	"github.com/LedFx/ledfx/pkg/audio"
	"github.com/LedFx/ledfx/pkg/logger"
	"github.com/LedFx/ledfx/pkg/render"
)

/*
Audio reactive port of PixelBlaze effect "Block Reflections"
Credit to Ben Henke for original concept & fantastic LED art
https://electromage.com/patterns
*/

type BlockReflections struct{}

// Apply new pixels to an existing pixel array.
func (e *BlockReflections) assembleFrame(base *Effect, pg *render.PixelGroup) {
	// operate on the largest pixel output in group, then clone to others
	p := pg.Group[pg.Largest]

	mel, err := audio.Analyzer.GetMelbank(base.ID)
	if err != nil {
		logger.Logger.WithField("context", "Effect").Error(err)
		return
	}

	lows := mel.LowsAmplitude()
	mids := mel.HighAmplitude()
	high := mel.HighAmplitude()

	// t1 := base.time(0.1)
	t2 := base.time(0.1) * math.Pi * 2
	t3 := base.time(0.5)
	t4 := base.time(0.2) * math.Pi * 2

	for i := 0; i < len(p); i++ {
		fi := float64(i)
		m := (0.3 + base.triangle(t2)*0.2 + (lows * base.Config.Intensity))
		h := math.Sin(t2) + math.Mod(((fi-base.pixelScaler/2)/base.pixelScaler)*(base.triangle(t3)*10+4*math.Sin(t4)+(lows*base.Config.Intensity)), m)
		v := math.Mod(math.Abs(h)+math.Abs(m), 1)
		v = math.Pow(v, 2)

		p[i][0] = h + high
		p[i][1] = 1 - (lows * base.Config.Intensity)
		p[i][2] = v + (mids * base.Config.Intensity)
	}

	pg.CloneToAll(pg.Largest)
}
