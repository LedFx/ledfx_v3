package effect

import (
	"math"

	"github.com/LedFx/ledfx/pkg/audio"
	"github.com/LedFx/ledfx/pkg/color"
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
		h := math.Pow(volume, 2) * math.Abs(math.Cos(fi/-(0.01+timestep))/math.Sin(fi/-(1.1+timestep)))
		s := math.Pow(volume, 2) * math.Abs(math.Tan(fi/0.1+timestep/0.7))
		v := math.Pow(volume, 2) * math.Abs((math.Sin(fi/-(0.1+timestep/2.5))/math.Min(volume, 1))*math.Tan(fi/10.1+timestep))
		if math.IsNaN(h) {
			h = 0
		}
		if math.IsNaN(s) {
			s = 0
		}
		if math.IsNaN(v) {
			v = 0
		}
		p[i][0] = h
		p[i][1] = s
		p[i][2] = v
	}
}
