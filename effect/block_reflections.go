package effect

import (
	"ledfx/audio"
	"ledfx/color"
	"ledfx/logger"
	"math"
)

/*
Audio reactive port of PixelBlaze effect "Block Reflections"
Credit to Ben Henke for original concept & fantastic LED art
https://electromage.com/patterns
*/

type BlockReflections struct {
	// initialised bool
	// freezeFrame color.Pixels
}

// Apply new pixels to an existing pixel array.
func (e *BlockReflections) assembleFrame(base *Effect, p color.Pixels) {
	// if !e.initialised {
	// 	e.freezeFrame = make(color.Pixels, base.pixelCount)
	// 	e.initialised = true
	// }

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
		h := math.Sin(t2) + math.Mod(((fi-base.pixelScaler/2)/base.pixelScaler)*(base.triangle(t3)*10+4*math.Sin(t4)+(mids*base.Config.Intensity)), m)
		v := math.Mod(math.Abs(h)+math.Abs(m), 1)
		v = math.Pow(v, 2)

		p[i][0] = h + high
		p[i][1] = 1 - (lows * base.Config.Intensity)
		p[i][2] = v + (mids * base.Config.Intensity)
		// p[i][0] = h + e.freezeFrame[i][0]
		// p[i][1] = 1 - (mel.MidsAmplitude() + mel.HighAmplitude()) + e.freezeFrame[i][1]
		// p[i][2] = v + e.freezeFrame[i][2]
		// if lowsAmplitude > 0.7 {
		// 	e.freezeFrame[i] = p[i]
		// }

		// e.freezeFrame[i][0] *= 0.9
		// e.freezeFrame[i][1] *= 0.9
		// e.freezeFrame[i][2] *= 0.9
	}
}
