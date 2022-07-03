package effect

import (
	"ledfx/audio"
	"ledfx/color"
	"math/rand"
)

type Strobe struct{}

// Apply new pixels to an existing pixel array.
func (e *Strobe) assembleFrame(base *Effect, p color.Pixels) {
	mel, err := audio.Analyzer.GetMelbank(base.ID)
	if err != nil {
		return
	}
	// set full strip to colour if bass
	if mel.LowsAmplitude() > 0.7 {
		for i := range p {
			p[i] = color.Full
		}
	}

	// if an onset has not happened since the last frame
	if !audio.Analyzer.RecentOnset.After(base.prevFrameTime) {
		return
	}

	// choose a random place to put the strobe on the strip
	strobe_width := int(base.Config.Intensity * base.pixelScaler)
	for i := rand.Intn(len(p) - strobe_width); i < strobe_width; i++ {
		p[i][1] = 0 // desaturate the colour to white
		p[i][2] = 1 // set full brightness
	}
}
