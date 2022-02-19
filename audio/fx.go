package audio

import (
	"fmt"
	aubio "github.com/simonassank/aubio-go"
	"ledfx/color"
	"ledfx/config"
	"ledfx/virtual"
)

const (
	fftSize         uint = 1024
	framesPerBuffer uint = 44100 / 60
	sampleRate      uint = 44100
)

type FxHandler struct {
	frameCount int
	pvoc       *aubio.PhaseVoc
	melbank    *aubio.FilterBank
	onset      *aubio.Onset
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
	return fx, nil
}

func (fx *FxHandler) Callback(buf Buffer) {
	fx.frameCount += 1
	simpleBuffer := aubio.NewSimpleBufferData(uint(len(buf)), buf.AsFloat64())
	defer simpleBuffer.Free()
	fx.pvoc.Do(simpleBuffer)
	fx.melbank.Do(fx.pvoc.Grain())
	fx.onset.Do(simpleBuffer)
	bufSlice := fx.onset.Buffer().Slice()
	sum := sumBufSlice(bufSlice)
	if sum > 0 {
		fmt.Printf("%0.6f\n", sum)
		_ = virtual.PlayVirtual(config.GlobalConfig.Virtuals[0].Id, true, color.RandomColor())
	}
	/*sum := sumBufSlice(bufSlice)*/
	/*if err := virtual.PlayVirtual(config.GlobalConfig.Virtuals[0].Id, true, color.FromBufSliceSum(sum)); err != nil {
		log.Logger.WithField("category", "FxHandler Callback").Errorf("Error during PlayVirtual(): %v", err)
	}*/
}

func sumBufSlice(bufSlice []float64) (sum float64) {
	for i := range bufSlice {
		sum += bufSlice[i]
	}
	return sum
}
