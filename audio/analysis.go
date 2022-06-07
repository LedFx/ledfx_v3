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
	mu              sync.Mutex
	bufMono         *aubio.SimpleBuffer // mono aubio buffer
	bufVocals       *aubio.SimpleBuffer // vocals aubio buffer
	bufInstruments  *aubio.SimpleBuffer // instruments aubio buffer
	dataLeft        []float64           // left channel
	dataRight       []float64           // right channel
	dataMono        []float64           // mono mix of left and right
	dataInstruments []float64           // non-centre panned audio (typically instruments)
	dataVocals      []float64           // centre panned audio (typically vocals)
	eq              *aubio.Filter       // balances the volume across freqs
	onset           *aubio.Onset        // detects percussinve onsets
	pvoc            *aubio.PhaseVoc     // transforms audio data to fft
	melbank         *aubio.FilterBank   // scales fft to perceptual bins
	melfreqs        []float64           // frequencies of each mel bin
	// might need individual pvoc, melbank etc. for each audio channel. Not sure if they're stateless
	// pvocMono           *aubio.PhaseVoc     // transforms audio data to fft
	// pvocVocals         *aubio.PhaseVoc     // transforms audio data to fft
	// pvocInstruments    *aubio.PhaseVoc     // transforms audio data to fft
	// melbankMono        *aubio.FilterBank   // scales fft to perceptual bins
	// melbankVocals      *aubio.FilterBank   // scales fft to perceptual bins
	// melbankInstruments *aubio.FilterBank   // scales fft to perceptual bins
	MelbankMono         []float64 // Mono melbank for effects
	MelbankVocals       []float64 // Vocal melbank for effects
	MelbankInstruments  []float64 // Instrument melbank for effects
	OnsetNowMono        bool      // Mono onset for effects
	OnsetNowVocals      bool      // Vocal onset for effects
	OnsetNowInstruments bool      // Instrument onset for effects
}

func NewAnalyzer() *Analyzer {
	a := &Analyzer{
		mu:                  sync.Mutex{},
		bufMono:             aubio.NewSimpleBuffer(framesPerBuffer),
		bufVocals:           aubio.NewSimpleBuffer(framesPerBuffer),
		bufInstruments:      aubio.NewSimpleBuffer(framesPerBuffer),
		dataLeft:            make([]float64, framesPerBuffer),
		dataRight:           make([]float64, framesPerBuffer),
		dataMono:            make([]float64, framesPerBuffer),
		dataInstruments:     make([]float64, framesPerBuffer),
		dataVocals:          make([]float64, framesPerBuffer),
		MelbankMono:         make([]float64, melBins),
		MelbankVocals:       make([]float64, melBins),
		MelbankInstruments:  make([]float64, melBins),
		OnsetNowMono:        false,
		OnsetNowVocals:      false,
		OnsetNowInstruments: false,
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
	fpbint := int(framesPerBuffer)

	// Get our left and right channels as float64
	for i := 0; i < fpbint; i++ {
		a.dataLeft[i] = float64(buf[i])
	}
	for i := fpbint; i < fpbint*2; i++ {
		a.dataRight[i-fpbint] = float64(buf[i])
	}

	// "karaoke" instrumental/vocal isolation
	// see: https://www.youtube.com/watch?v=-KfGwz1Zg6I
	// might need to divide each value by two. not sure..
	for i := 0; i < fpbint; i++ {
		a.dataMono[i] = (a.dataLeft[i] + a.dataRight[i])         // mix left and right to get mono
		a.dataInstruments[i] = (a.dataLeft[i] - a.dataRight[i])  // remove centre panned audio (vocals)
		a.dataVocals[i] = (a.dataMono[i] - a.dataInstruments[i]) // remove non centre pan (instruments) from mono to get vocals
	}

	// CGo calls are slow. I've optimised this as best I can.
	// Turns out the data conversion is a fraction of the cost of the c library calls
	// could try directly operating on the C memory in aubio-go.
	// Would be fast but also dangerous..
	// https://copyninja.info/blog/workaround-gotypesystems.html
	a.bufMono.SetData(a.dataMono)
	a.bufVocals.SetData(a.dataVocals)
	a.bufInstruments.SetData(a.dataInstruments)

	// do mono frequency analysis
	a.eq.DoOutplace(a.bufMono)
	a.pvoc.Do(a.eq.Buffer())
	a.melbank.Do(a.pvoc.Grain())
	a.MelbankMono = a.melbank.Buffer().Slice()

	// do vocal frequency analysis
	a.eq.DoOutplace(a.bufVocals)
	a.pvoc.Do(a.eq.Buffer())
	a.melbank.Do(a.pvoc.Grain())
	a.MelbankVocals = a.melbank.Buffer().Slice()

	// do instrument frequency analysis
	a.eq.DoOutplace(a.bufInstruments)
	a.pvoc.Do(a.eq.Buffer())
	a.melbank.Do(a.pvoc.Grain())
	a.MelbankInstruments = a.melbank.Buffer().Slice()

	// do onset analysis
	a.onset.Do(a.bufMono)
	a.OnsetNowMono = a.onset.Buffer().Get(uint(0)) != 0
	a.onset.Do(a.bufVocals)
	a.OnsetNowVocals = a.onset.Buffer().Get(uint(0)) != 0
	a.onset.Do(a.bufInstruments)
	a.OnsetNowInstruments = a.onset.Buffer().Get(uint(0)) != 0

	if a.OnsetNowMono {
		log.Logger.WithField("category", "Audio Analysis").Info("Onset detected")
	}
}

func (a *Analyzer) Cleanup() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.bufMono.Free()
	a.bufVocals.Free()
	a.bufInstruments.Free()

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
