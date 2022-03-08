package audio

import (
	"fmt"
	"go.uber.org/atomic"
	"ledfx/color"
	"ledfx/config"
	"ledfx/virtual"

	aubio "github.com/simonassank/aubio-go"
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
	bufSlice := fx.onset.Buffer().Slice()
	sum := sumBufSlice(bufSlice)

	/*normalized := int((sum / 4.0) * 48)

	if sum > 0 {
		for _, d := range config.GlobalConfig.Virtuals {
			if d.Active && d.Effect.Type == "singleColor" {
				if err := virtual.RepeatNSmooth(d.Id, true, "#ff0000", normalized); err != nil {
					log.Logger.WithField("category", "Buffer FX Callback").Errorf("Error playing virtual: %v", err)
					continue
				}
			}
		}
	}
	*/

	if sum > 1.2 {
		//fmt.Printf("%0.3f\n", sum)
		for i, d := range config.GlobalConfig.Virtuals {
			// ToDo: change singleColor to audioRandom after Effect-Type-Change is possible
			if d.Active && config.GlobalConfig.Virtuals[i].Effect.Type == "singleColor" {
				//fmt.Printf("%s\n", config.GlobalConfig.Virtuals[i].Effect.Type)
				_ = virtual.PlayVirtual(config.GlobalConfig.Virtuals[i].Id, true, color.RandomColor(), "audioRandom")
			}
		}
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
