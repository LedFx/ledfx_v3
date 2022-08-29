package effect

import (
	"github.com/LedFx/ledfx/pkg/audio"
	"github.com/LedFx/ledfx/pkg/color"
	"github.com/LedFx/ledfx/pkg/logger"
)

type Energy struct{}

// Apply new pixels to an existing pixel array.
func (e *Energy) assembleFrame(base *Effect, p color.Pixels) {
	mel, err := audio.Analyzer.GetMelbank(base.ID)
	if err != nil {
		logger.Logger.WithField("context", "Effect Energy").Error(err)
		return
	}
	lowsCol := color.Color{0, 1, 1}
	midsCol := color.Color{0.5, 1, 1}
	highCol := color.Color{1, 1, 1}
	lowsMidsCol := color.Color{0.25, 1, 1}
	midsHighCol := color.Color{0.75, 1, 1}

	lowsAmplitude := int(mel.LowsAmplitude() * base.pixelScaler)
	midsAmplitude := int(mel.MidsAmplitude() * base.pixelScaler)
	highAmplitude := int(mel.HighAmplitude() * base.pixelScaler)

	var lows, mids, high bool
	for i := 0; i < len(p); i++ {
		lows = i < lowsAmplitude
		mids = i < midsAmplitude
		high = i < highAmplitude
		switch {
		// case !lows && !mids && !high: // none, dont update colour
		// 	// p[i] = color.Color{0, 0, 0}

		case lows && mids && high: // bass mids and high, white colour
			p[i] = color.Color{0, 0, 1}
		case lows && !mids && !high: // bass
			p[i] = lowsCol
		case lows && mids && !high: // mix bass and mids
			p[i] = lowsMidsCol
		case !lows && mids && !high: // mids
			p[i] = midsCol
		case !lows && mids && high: // mix mids and high
			p[i] = midsHighCol
		case !lows && !mids && high: // high
			p[i] = highCol
		}
	}
}
