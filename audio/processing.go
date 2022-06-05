package audio

import (
	"github.com/LedFx/aubio-go"
	log "ledfx/logger"
	"math/rand"
	"sync"
)

const (
	fftSize         uint = 1024
	framesPerBuffer uint = 44100 / 60
	sampleRate      uint = 44100
)

var (
	onset *aubio.Onset
)

type Processor struct {
	mu       sync.Mutex
	OnsetNow bool
}

func init() {
	var err error

	if onset, err = aubio.NewOnset(aubio.Energy, fftSize, framesPerBuffer, sampleRate); err != nil {
		log.Logger.WithField("category", "Audio Processor Init").Fatalf("Error creating new Aubio Onset: %v", err)
	}
}

func (p *Processor) BufferCallback(buf Buffer) {
	p.mu.Lock()
	defer p.mu.Unlock()

	randNum := rand.Intn(10)
	if randNum == 1 {
		p.OnsetNow = true
	} else {
		p.OnsetNow = false
	}
}
