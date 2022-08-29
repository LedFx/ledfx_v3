package audio

import (
	"fmt"
	"sync"
	"time"

	log "github.com/LedFx/ledfx/pkg/logger"

	"github.com/LedFx/aubio-go"
)

const (
	fftSize         uint = 4096
	sampleRate      uint = 44100
	framesPerBuffer uint = sampleRate / 60
)

// Singleton analyzer
var Analyzer *analyzer

type analyzer struct {
	mu          sync.Mutex
	buf         *aubio.SimpleBuffer // aubio buffer
	data        []float32           // mono mix of left and right audio channels
	eq          *aubio.Filter       // balances the volume across freqs. Stateless, only need one
	onset       *aubio.Onset        // detects percussive onsets
	pvoc        *aubio.PhaseVoc     // transforms audio data to fft
	melbanks    map[string]*melbank // a melbank for each effect
	RecentOnset time.Time           //  onset for effects
}

func init() {
	Analyzer = &analyzer{
		mu:          sync.Mutex{},
		buf:         aubio.NewSimpleBuffer(framesPerBuffer),
		data:        make([]float32, framesPerBuffer),
		melbanks:    make(map[string]*melbank),
		RecentOnset: time.Now(),
	}
	var err error

	// Create EQ filter. Magic numbers to balance the audio. Boosts the bass and mid, dampens the highs.
	if Analyzer.eq, err = aubio.NewFilterBiquad(1, -2, 1, -2, 1, framesPerBuffer); err != nil {
		log.Logger.WithField("context", "Audio Analyzer Init").Fatalf("Error creating new Aubio EQ Filter: %v", err)
	}

	// Create onset
	if Analyzer.onset, err = aubio.NewOnset(aubio.HFC, fftSize, framesPerBuffer, sampleRate); err != nil {
		log.Logger.WithField("context", "Audio Analyzer Init").Fatalf("Error creating new Aubio Onset: %v", err)
	}

	// Create pvoc
	if Analyzer.pvoc, err = aubio.NewPhaseVoc(fftSize, framesPerBuffer); err != nil {
		log.Logger.WithField("context", "Audio Analyzer Init").Fatalf("Error creating new Aubio Pvoc: %v", err)
	}

}

func (a *analyzer) BufferCallback(buf Buffer) {
	a.mu.Lock()
	defer a.mu.Unlock()
	fpbint := int(framesPerBuffer)

	// Get our left and right channels as float32
	for i := 0; i < fpbint; i++ {
		a.data[i] = float32(buf[i] + buf[i+fpbint])
	}

	// set the data of the aubio buffer (optimised)
	a.buf.SetDataFast(a.data)

	// Perform FFT of each audio stream
	a.eq.DoOutplace(a.buf)
	a.pvoc.Do(a.eq.Buffer())

	// Perform melbank frequency analysis
	for _, mb := range a.melbanks {
		mb.Do(a.pvoc.Grain())
	}

	// do onset analysis
	a.onset.Do(a.buf)
	if a.onset.OnsetNow() {
		a.RecentOnset = time.Now()
	}
}

func (a *analyzer) Cleanup() {
	a.mu.Lock()
	a.eq.Free()
	a.buf.Free()
	a.onset.Free()
	a.pvoc.Free()
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

func (a *analyzer) NewMelbank(id string, min_freq, max_freq uint, intensity float64) error {
	// if a melbank is already registered to this effect id, kill it and warn
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.melbanks[id]; ok {
		log.Logger.WithField("context", "Audio Analysis").Debugf("Effect %s attempted to create a new melbank but already has one registered", id)
		a.DeleteMelbank(id)
	}
	mb, err := newMelbank(min_freq, max_freq, intensity)
	if err == nil {
		log.Logger.WithField("context", "Audio Analysis").Debugf("Registered new melbank for effect %s", id)
		a.melbanks[id] = mb
	}
	return err
}
