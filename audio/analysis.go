package audio

import (
	"fmt"
	log "ledfx/logger"
	"sync"

	"github.com/LedFx/aubio-go"
)

const (
	fftSize         uint = 4096
	sampleRate      uint = 44100
	framesPerBuffer uint = sampleRate / 60
)

type AudioStream int

const (
	Mono AudioStream = iota
	Vocals
)

func (a AudioStream) String() string {
	switch a {
	case Mono:
		return "Mono"
	case Vocals:
		return "Vocals"
	default:
		return "Invalid AudioStream"
	}
}

// Singleton analyzer
var Analyzer *analyzer

type analyzer struct {
	mu             sync.Mutex
	bufMono        *aubio.SimpleBuffer // mono aubio buffer
	bufVocals      *aubio.SimpleBuffer // vocals aubio buffer
	dataLeft       []float64           // left channel
	dataRight      []float64           // right channel
	dataMono       []float64           // mono mix of left and right
	dataVocals     []float64           // centre panned audio (typically vocals)
	eq             *aubio.Filter       // balances the volume across freqs. Stateless, only need one
	onsetMono      *aubio.Onset        // detects percussive onsets
	onsetVocals    *aubio.Onset        // detects percussive onsets
	pvocMono       *aubio.PhaseVoc     // transforms audio data to fft
	pvocVocals     *aubio.PhaseVoc     // transforms audio data to fft
	melbanks       map[string]*melbank // a melbank for each effect
	OnsetNowMono   bool                // Mono onset for effects
	OnsetNowVocals bool                // Vocal onset for effects
}

func init() {
	Analyzer = &analyzer{
		mu:             sync.Mutex{},
		bufMono:        aubio.NewSimpleBuffer(framesPerBuffer),
		bufVocals:      aubio.NewSimpleBuffer(framesPerBuffer),
		dataLeft:       make([]float64, framesPerBuffer),
		dataRight:      make([]float64, framesPerBuffer),
		dataMono:       make([]float64, framesPerBuffer),
		dataVocals:     make([]float64, framesPerBuffer),
		melbanks:       make(map[string]*melbank),
		OnsetNowMono:   false,
		OnsetNowVocals: false,
	}
	var err error


	// Create EQ filter. Magic numbers to balance the audio. Boosts the bass and mid, dampens the highs.
	// 0.85870, -1.71740, 0.85870, -1.71605, 0.71874
	// 0.9926, -1.985, 0.9926, -1.9852, 0.9853
	// GOOD 0.99756, -1.99512, 0.99756, -1.99511, 0.99512,
	if Analyzer.eq, err = aubio.NewFilterBiquad(1, -2,1, -2,1,  framesPerBuffer); err != nil {
		log.Logger.WithField("context", "Audio Analyzer Init").Fatalf("Error creating new Aubio EQ Filter: %v", err)
	}

	// Create onsets
	if Analyzer.onsetMono, err = aubio.NewOnset(aubio.HFC, fftSize, framesPerBuffer, sampleRate); err != nil {
		log.Logger.WithField("context", "Audio Analyzer Init").Fatalf("Error creating new Aubio Onset: %v", err)
	}
	if Analyzer.onsetVocals, err = aubio.NewOnset(aubio.SpecFlux, fftSize, framesPerBuffer, sampleRate); err != nil {
		log.Logger.WithField("context", "Audio Analyzer Init").Fatalf("Error creating new Aubio Onset: %v", err)
	}

	// Create pvocs
	if Analyzer.pvocMono, err = aubio.NewPhaseVoc(fftSize, framesPerBuffer); err != nil {
		log.Logger.WithField("context", "Audio Analyzer Init").Fatalf("Error creating new Aubio Pvoc: %v", err)
	}
	if Analyzer.pvocVocals, err = aubio.NewPhaseVoc(fftSize, framesPerBuffer); err != nil {
		log.Logger.WithField("context", "Audio Analyzer Init").Fatalf("Error creating new Aubio Pvoc: %v", err)
	}
}

func (a *analyzer) BufferCallback(buf Buffer) {
	a.mu.Lock()
	defer a.mu.Unlock()
	fpbint := int(framesPerBuffer)

	// Get our left and right channels as float64
	for i := 0; i < fpbint; i++ {
		a.dataLeft[i] = float64(buf[i])
	}
	for i := fpbint; i < fpbint*2; i++ {
		a.dataRight[i-fpbint] = float64(buf[i])
	}

	// "karaoke" centre isolation
	// see: https://www.youtube.com/watch?v=-KfGwz1Zg6I
	// might need to divide each value by two. not sure..
	for i := 0; i < fpbint; i++ {
		a.dataMono[i] = (a.dataLeft[i] + a.dataRight[i]) / 2                   // mix left and right to get mono
		a.dataVocals[i] = (a.dataMono[i] - a.dataLeft[i] + a.dataRight[i]) / 3 // remove non centre pan (instruments) from mono to get vocals
	}

	// CGo calls are slow. I've optimised this as best I can.
	// Turns out the data conversion is a fraction of the cost of the c library calls
	// could try directly operating on the C memory in aubio-go.
	// Would be fast but also dangerous..
	// https://copyninja.info/blog/workaround-gotypesystems.html
	a.bufMono.SetData(a.dataMono)
	a.bufVocals.SetData(a.dataVocals)

	// Perform FFT of each audio stream
	a.eq.DoOutplace(a.bufMono)
	a.pvocMono.Do(a.eq.Buffer())
	a.eq.DoOutplace(a.bufVocals)
	a.pvocVocals.Do(a.eq.Buffer())

	// Perform melbank frequency analysis
	for _, mb := range a.melbanks {
		switch mb.Audio {
		case Mono:
			mb.Do(a.pvocMono.Grain())
		case Vocals:
			mb.Do(a.pvocVocals.Grain())
		}
	}

	// do onset analysis
	a.onsetMono.Do(a.bufMono)
	a.OnsetNowMono = a.onsetMono.Buffer().Get(uint(0)) != 0
	a.onsetVocals.Do(a.bufVocals)
	a.OnsetNowVocals = a.onsetVocals.Buffer().Get(uint(0)) != 0
}

func (a *analyzer) Cleanup() {
	a.mu.Lock()
	a.eq.Free()
	a.bufMono.Free()
	a.bufVocals.Free()
	a.onsetMono.Free()
	a.pvocMono.Free()
	a.onsetVocals.Free()
	a.pvocVocals.Free()
	a.mu.Unlock()

	for id := range a.melbanks {
		a.DeleteMelbank(id)
	}
}

// convenience method to get the melbank data
func (a *analyzer) GetMelbankData(id string) ([]float64, error) {
	mb, err := a.GetMelbank(id)
	return mb.Data, err
}

func (a *analyzer) GetMelbank(id string) (mb *melbank, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	mb, ok := a.melbanks[id]
	if !ok {
		err = fmt.Errorf("cannot find melbank registered for effect %s", id)
	}
	return mb, err
}

func (a *analyzer) DeleteMelbank(id string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if mb, ok := a.melbanks[id]; ok {
		log.Logger.WithField("context", "Audio Analysis").Debugf("Deleted melbank for effect %s", id)
		mb.Free()
		delete(a.melbanks, id)
	}
}

func (a *analyzer) NewMelbank(id string, audio AudioStream, min_freq, max_freq uint, intensity float64) error {
	// if a melbank is already registered to this effect id, kill it and warn
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.melbanks[id]; ok {
		log.Logger.WithField("context", "Audio Analysis").Debugf("Effect %s attempted to create a new melbank but already has one registered", id)
		a.DeleteMelbank(id)
	}
	mb, err := newMelbank(audio, min_freq, max_freq, intensity)
	if err == nil {
		log.Logger.WithField("context", "Audio Analysis").Debugf("Registered new melbank for effect %s", id)
		a.melbanks[id] = mb
	}
	return err
}
