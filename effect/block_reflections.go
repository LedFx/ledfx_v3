package effect

import (
	"ledfx/color"
	"math"
)

/*
Audio reactive port of PixelBlaze effect "Block Reflections"
Credit to Ben Henke for original concept & fantastic LED art
https://electromage.com/patterns
*/

type BlockReflections struct{}

// Apply new pixels to an existing pixel array.
func (e *BlockReflections) assembleFrame(base *Effect, p color.Pixels) {

	// mel, err := audio.Analyzer.GetMelbank(base.ID)
	// if err != nil {
	// 	logger.Logger.WithField("context", "Effect Energy").Error(err)
	// 	return
	// }
	// lowsAmplitude := mel.LowsAmplitude()

	// t1 := base.time(0.1)
	t2 := base.time(0.1) * math.Pi * 2 // * lowsAmplitude
	t3 := base.time(0.5)
	t4 := base.time(0.2) * math.Pi * 2

	for i := 0; i < len(p); i++ {
		fi := float64(i)
		m := (0.3 + base.triangle(t2)*0.2)
		h := math.Sin(t2) + math.Mod(((fi-base.pixelScaler/2)/base.pixelScaler)*(base.triangle(t3)*10+4*math.Sin(t4)), m)
		v := math.Mod(math.Abs(h)+math.Abs(m), 1)
		v = math.Pow(v, 2)
		p[i][0] = h
		p[i][1] = 1
		p[i][2] = v
	}
}
