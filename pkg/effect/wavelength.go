package effect

import (
	"github.com/LedFx/ledfx/pkg/audio"
	"github.com/LedFx/ledfx/pkg/logger"
	"github.com/LedFx/ledfx/pkg/math_utils"
	"github.com/LedFx/ledfx/pkg/render"
)

type Wavelegth struct{}

// Apply new pixels to an existing pixel array.
func (e *Wavelegth) assembleFrame(base *Effect, pg *render.PixelGroup) {
	// operate on the largest pixel output in group, then clone to others
	p := pg.Group[pg.Largest]

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
		p[i][0] = float64(i) / base.pixelScaler
		p[i][1] = 1
		p[i][2] = scaled_mel[i]
	}
	pg.CloneToAll(pg.Largest)
}
