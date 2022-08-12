package effect

import (
	"ledfx/audio"
	"ledfx/color"
	"ledfx/logger"
	"ledfx/math_utils"
)

type Wavelegth struct{}

// Apply new pixels to an existing pixel array.
func (e *Wavelegth) assembleFrame(base *Effect, p color.Pixels) {
	mel, err := audio.Analyzer.GetMelbank(base.ID)
	if err != nil {
		logger.Logger.WithField("context", "Effect Wavelength").Error(err)
		return
	}
	scaled_mel := make([]float64, len(p))
	err = math_utils.Interpolate(mel.Data, scaled_mel)
	if err != nil {
		logger.Logger.WithField("context", "Effect Wavelength").Error(err)
		return
	}

	for i := 0; i < len(p); i++ {
		p[i][0] = float64(i) / float64(len(p))
		p[i][1] = 1
		p[i][2] = scaled_mel[i]
	}
}
