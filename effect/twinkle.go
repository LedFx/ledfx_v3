package effect

import (
	"ledfx/color"
	"math/rand"
)

/*
Audio reactive port of PixelBlaze effect "Glitch Bands"
Credit to Ben Henke for original concept & fantastic LED art
https://electromage.com/patterns
*/

type Twinkle struct {
	initialised bool
	hues        []float64
	periods     []float64
}

// movement speed based on beat
// modify saturation by vocals/mids
// make the bars/legs grow and shrink
// modulate t1 with highs

// Apply new pixels to an existing pixel array.
func (e *Twinkle) assembleFrame(base *Effect, p color.Pixels) {
	if !e.initialised {
		e.hues = make([]float64, len(p))
		e.periods = make([]float64, len(p))
		for i := 0; i < len(p); i++ {
			e.hues[i] = rand.Float64()
			e.periods[i] = 0.005 + rand.Float64()/65
		}
		e.initialised = true
	}

	for i := 0; i < len(p); i++ {
		p[i][0] = e.hues[i]
		p[i][1] = 1
		p[i][2] = base.triangle(base.time(e.periods[i]))
	}
}
