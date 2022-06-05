package audio

import (
	"fmt"

	"go.uber.org/atomic"

	aubio "github.com/LedFx/aubio-go"
)

type FxHandler struct {
	frameCount int
	pvoc       *aubio.PhaseVoc
	melbank    *aubio.FilterBank
	onset      *aubio.Onset
	highest    *atomic.Float64
}

func NewFxHandler() (fx *FxHandler, err error) {
	fx = new(FxHandler)
	if fx.pvoc, err = aubio.NewPhaseVoc(fftSize, framesPerBuffer); err != nil {
		return nil, fmt.Errorf("error initializing new Aubio phase vocoder: %w", err)
	}

	fx.melbank = aubio.NewFilterBank(40, fftSize)
	fx.melbank.SetMelCoeffsSlaney(44100)

	if fx.onset, err = aubio.NewOnset(aubio.Energy, fftSize, framesPerBuffer, sampleRate); err != nil {
		return nil, fmt.Errorf("error initializing new Aubio onset: %w", err)
	}

	fx.highest = atomic.NewFloat64(0.0)

	return fx, nil
}

func (fx *FxHandler) Callback(buf Buffer) {
	fx.frameCount += 1
	simpleBuffer := aubio.NewSimpleBufferData(uint(len(buf)), buf.AsFloat64())
	defer simpleBuffer.Free()
	fx.pvoc.Do(simpleBuffer)
	fx.melbank.Do(fx.pvoc.Grain())
	fx.onset.Do(simpleBuffer)
	// bufSlice := fx.onset.Buffer().Slice()
}
