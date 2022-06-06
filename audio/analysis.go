package audio

import (
	log "ledfx/logger"
	"ledfx/math_utils"
	"math"
	"sync"

	"github.com/LedFx/aubio-go"
)

const (
	fftSize         uint = 4096
	sampleRate      uint = 44100
	framesPerBuffer uint = sampleRate / 60
	melBins         uint = 24
	melMin          uint = 20
	melMax          uint = sampleRate / 2
)

type Analyzer struct {
	mu       sync.Mutex
	eq       *aubio.Filter     // balances the volume across freqs
	onset    *aubio.Onset      // detects percussinve onsets
	pvoc     *aubio.PhaseVoc   // transforms audio data to fft
	melbank  *aubio.FilterBank // scales fft to perceptual bins
	melfreqs []float64         // frequencies of each mel bin
	// buf      *aubio.SimpleBuffer // not really used until I add buffer update methods to aubio-go
	Melbank  []float64
	OnsetNow bool
}

func NewAnalyzer() *Analyzer {
	a := &Analyzer{
		mu:       sync.Mutex{},
		OnsetNow: false,
	}
	var err error

	// Create EQ filter. Magic numbers to balance the audio. Boosts the bass and mid, dampens the highs.
	if a.eq, err = aubio.NewFilterBiquad(0.85870, -1.71740, 0.85870, -1.71605, 0.71874, framesPerBuffer); err != nil {
		log.Logger.WithField("category", "Audio Analyzer Init").Fatalf("Error creating new Aubio EQ Filter: %v", err)
	}

	// Create onset
	if a.onset, err = aubio.NewOnset(aubio.HFC, fftSize, framesPerBuffer, sampleRate); err != nil {
		log.Logger.WithField("category", "Audio Analyzer Init").Fatalf("Error creating new Aubio Onset: %v", err)
	}
	// Create pvoc
	if a.pvoc, err = aubio.NewPhaseVoc(fftSize, framesPerBuffer); err != nil {
		log.Logger.WithField("category", "Audio Analyzer Init").Fatalf("Error creating new Aubio Pvoc: %v", err)
	}

	// Create melbank
	a.melbank = aubio.NewFilterBank(melBins, framesPerBuffer)
	// create linear freq scale
	melbank_freqs, err := math_utils.Linspace(HzToMel(float64(melMin)), HzToMel(float64(melMax)), int(melBins)+2)
	if err != nil {
		log.Logger.WithField("category", "Audio Analyzer Init").Fatalf("Error initialising melbank: %v", err)
	}
	// convert linear freq scale to perceptually even scale.
	for i := range melbank_freqs {
		melbank_freqs[i] = MelToHz(melbank_freqs[i])
	}
	a.melfreqs = melbank_freqs
	// Set melbank bands using this scale
	b := aubio.NewSimpleBufferData(melBins+2, melbank_freqs) //[1:len(melbank_freqs)-1]
	a.melbank.SetTriangleBands(b, sampleRate)

	// Normalize the filterbank triangles to a consistent height.
	// The coeffs will be normalized by the triangles area which results in an uneven melbank
	coeffs := a.melbank.Coeffs().GetChannels()
	for ch := range coeffs {
		// find the max of the channel
		var max float64
		for pos := range coeffs[ch] {
			if coeffs[ch][pos] > max {
				max = coeffs[ch][pos]
			}
		}
		// then normalise all the heights of the channel to the maximum
		for pos := range coeffs[ch] {
			coeffs[ch][pos] /= max
		}
	}
	return a
}

func (a *Analyzer) BufferCallback(buf Buffer) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// disgusting amount of buffer allocations for each audio frame.
	// proof of concept, must tidy these away and just update the values
	f64buf := buf.AsFloat64()
	f64Left := f64buf[0:framesPerBuffer]                    // left channel
	f64Right := f64buf[framesPerBuffer : framesPerBuffer*2] // right channel
	f64Mono := make([]float64, framesPerBuffer)             // mono mix of left and right
	f64Instruments := make([]float64, framesPerBuffer)      // non-centre panned audio (typically instruments)
	f64Vocals := make([]float64, framesPerBuffer)           // centre panned audio (typically vocals)

	// "kareoke" instrumental/vocal isolation
	// see: https://www.youtube.com/watch?v=-KfGwz1Zg6I
	for i := 0; i < int(framesPerBuffer); i++ {
		f64Mono[i] = (f64Left[i] + f64Right[i]) / 2         // mix left and right to get mono
		f64Instruments[i] = (f64Left[i] - f64Right[i]) / 2  // remove centre panned audio (vocals)
		f64Vocals[i] = (f64Mono[i] - f64Instruments[i]) / 2 // remove non centre pan (instruments) from mono to get vocals
	}

	// TODO update aubio so we can reuse the buffer and update values, rather
	// than reallocating each audio frame.
	// Unnecessary type conversions. This could be optimised a lot.
	mono := aubio.NewSimpleBufferData(uint(framesPerBuffer), f64Mono)
	vocals := aubio.NewSimpleBufferData(uint(framesPerBuffer), f64Vocals)
	instruments := aubio.NewSimpleBufferData(uint(framesPerBuffer), f64Instruments)

	defer mono.Free()
	defer vocals.Free()
	defer instruments.Free()

	// filter the incoming audio. this is like applying an eq to perceptually balance the audio.
	// TODO see how applying this eq affects pitch and onset analysis of the buffer.
	// if it's no good, we can do out of place and retain the original audio sample
	a.eq.Do(mono)
	a.eq.Do(vocals)
	a.eq.Do(instruments)

	// do freq analysis on mono audio
	a.pvoc.Do(mono)              // calculate fft
	a.melbank.Do(a.pvoc.Grain()) // scale it with melbank
	a.Melbank = a.melbank.Buffer().Slice()

	// do onset analysis
	a.onset.Do(mono)
	// more useless conversions.. dont need to convert the entire slice to f64 to see if the first value is nonzero
	a.OnsetNow = a.onset.Buffer().Slice()[0] != 0
	if a.OnsetNow {
		log.Logger.WithField("category", "Audio Analysis").Info("Onset detected")
	}
}

func (a *Analyzer) Cleanup() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.onset.Free()
	a.pvoc.Free()
	a.eq.Free()
	a.melbank.Buffer().Free() // TODO properly clean up melbank
}

// Custom mel scaling.
// This scaling function is specially crafted to spread out the low range
// and compress the highs in a visually/perceptually balanced way.
func HzToMel(hz float64) float64 {
	return 3700 * math_utils.LogN(1+hz/230, 12)
}

// Inverse of HzToMel function
func MelToHz(mel float64) float64 {
	return 230*(math.Pow(12, float64(mel)/3700)) - 230
}
