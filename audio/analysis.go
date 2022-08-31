package audio

import (
	"fmt"
	log "ledfx/logger"
	"math"
	"sync"
	"time"

	"github.com/LedFx/aubio-go"
)

const (
	// fft and buffer
	fftSize         uint = 4096
	sampleRate      uint = 44100
	framesPerBuffer uint = sampleRate / 60
	// volume normalisation streams
	streamConstant     float64 = 0.1
	streamPow          float64 = 1
	normStreamSlowLen  int     = 60 * 3 // assumes 60 audio updates per second
	normStreamFastLen  int     = 60 * 2
	reactStreamSlowLen int     = 60 * 3
	reactStreamFastLen int     = 60 * 0.05
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
	RecentOnset time.Time           // onset for effects
	Vol         volumeStream        // volume stream source for effects. includes a normalised volume and a timestep.
}

func init() {
	Analyzer = &analyzer{
		mu:          sync.Mutex{},
		buf:         aubio.NewSimpleBuffer(framesPerBuffer),
		data:        make([]float32, framesPerBuffer),
		melbanks:    make(map[string]*melbank),
		RecentOnset: time.Now(),
		Vol:         NewVolumeStream(),
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

	// update volume normaliser
	a.Vol.update(aubio.DbSpl(a.buf))

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

type volumeStream struct {
	reactStream stream
	normStream  stream
	Volume      float64
	Timestep    float64
}

func NewVolumeStream() volumeStream {
	return volumeStream{
		reactStream: newStream(reactStreamFastLen, reactStreamSlowLen),
		normStream:  newStream(normStreamFastLen, normStreamSlowLen),
		Volume:      0,
		Timestep:    0,
	}
}

func (vs *volumeStream) update(volume float64) {
	vs.reactStream.update(volume)
	vs.normStream.update(volume)
	vs.Volume = vs.reactStream.volume * vs.normStream.volume
	vs.Timestep = (vs.reactStream.timeStep + vs.reactStream.timeStep) / 40
}

type stream struct {
	// volume normalisation
	fastBuffer []float64
	slowBuffer []float64
	fastBufPos int
	slowBufPos int
	volume     float64
	timeStep   float64
}

func newStream(fastBufLen, slowBufLen int) stream {
	vs := stream{
		fastBuffer: make([]float64, fastBufLen),
		slowBuffer: make([]float64, slowBufLen),
		fastBufPos: 0,
		slowBufPos: 0,
	}
	return vs
}

func (s *stream) update(volume float64) {
	// update the buffers
	// rather than rolling, reallocating, etc, we just keep the index to update looping across the slice
	s.fastBuffer[s.fastBufPos] = volume
	s.slowBuffer[s.slowBufPos] = volume
	s.fastBufPos = (s.fastBufPos + 1) % len(s.fastBuffer)
	s.slowBufPos = (s.slowBufPos + 1) % len(s.slowBuffer)

	// calculate mean and min of slow buffer
	minSlow := 0.
	avgSlow := 0.
	for i := 0; i < len(s.slowBuffer); i++ {
		val := s.slowBuffer[i]
		if val > minSlow {
			minSlow = val
		}
		avgSlow += val
	}
	avgSlow /= float64(len(s.slowBuffer))

	// calculate mean of fast buffer
	avgFast := 0.
	for i := 0; i < len(s.fastBuffer); i++ {
		avgFast += s.fastBuffer[i]
	}
	avgFast /= float64(len(s.fastBuffer))

	// scale avgFast between mean and min of slow buffer
	avgFast = (avgFast - minSlow) / (avgSlow - minSlow)
	s.volume = math.Pow(avgFast, streamPow)
	s.timeStep += math.Pow(s.volume, streamPow*3) + streamConstant
}
