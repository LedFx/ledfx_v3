package effect

import (
	"math"

	"github.com/LedFx/ledfx/pkg/audio"
	"github.com/LedFx/ledfx/pkg/color"
	"github.com/LedFx/ledfx/pkg/logger"
	"github.com/LedFx/ledfx/pkg/pixelgroup"
)

type Weave struct {
	lowsPos int
	midsPos int
	highPos int
}

// Apply new pixels to an existing pixel array.
func (e *Weave) assembleFrame(base *Effect, pg *pixelgroup.PixelGroup) {
	// operate on the largest pixel output in group, then clone to others
	p := pg.Group[pg.Largest]

	mel, err := audio.Analyzer.GetMelbank(base.ID)
	if err != nil {
		logger.Logger.WithField("context", "Effect Weave").Error(err)
		return
	}
	lowsStep := base.deltaStart.Seconds()*base.Config.Intensity*3 + 0
	midsStep := base.deltaStart.Seconds()*base.Config.Intensity*3 + 0.5
	highStep := base.deltaStart.Seconds()*base.Config.Intensity*3 + 1

	lowsNew := int(weavePosition(lowsStep, 1) * base.pixelScaler)
	midsNew := int(weavePosition(midsStep, 1) * base.pixelScaler)
	highNew := int(weavePosition(highStep, 1) * base.pixelScaler)

	color.FillBetween(p, e.highPos, highNew, color.Color{1, 0.4, mel.HighAmplitude()}, true)
	color.FillBetween(p, e.midsPos, midsNew, color.Color{0.5, 0.4, mel.MidsAmplitude()}, true)
	color.FillBetween(p, e.lowsPos, lowsNew, color.Color{0, 0.4, mel.LowsAmplitude()}, true)

	for i := range p {
		if p[i][1] < 1 {
			p[i][1] += 0.1
		}
	}

	// replace this with audio data
	e.lowsPos = lowsNew
	e.midsPos = midsNew
	e.highPos = highNew

	pg.CloneToAll(pg.Largest)
}

// Position function. https://www.desmos.com/calculator/pacgvrebds
func weavePosition(pos float64, freq float64) float64 {
	return 1 - math.Abs(math.Mod(pos/freq, 2)-1)
}
