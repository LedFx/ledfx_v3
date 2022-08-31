package effect

import (
	"ledfx/audio"
	"ledfx/color"
	"math"
)

type Maelstrom struct{}

// Apply new pixels to an existing pixel array.
func (e *Maelstrom) assembleFrame(base *Effect, p color.Pixels) {
	// mel, err := audio.Analyzer.GetMelbank(base.ID)
	// if err != nil {
	// 	logger.Logger.WithField("context", "Effect").Error(err)
	// 	return
	// }

	volume := audio.Analyzer.Vol.Volume
	timestep := audio.Analyzer.Vol.Timestep

	for i := 0; i < len(p); i++ {
		fi := float64(i)
		p[i][0] = math.Pow(volume, 2) * math.Abs(math.Cos(fi/0.01-timestep)/math.Sin(fi/1.1-timestep))
		p[i][1] = math.Pow(volume, 2) * math.Abs(math.Tan(fi/0.1+timestep/0.7))
		p[i][2] = math.Pow(volume, 2) * math.Abs((math.Sin(fi/0.1-timestep/2.5)/math.Min(volume, 1))*math.Tan(fi/10.1+timestep))
	}
}
