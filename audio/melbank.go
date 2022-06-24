package audio

import (
	"fmt"
	"ledfx/math_utils"
	"math"

	"github.com/LedFx/aubio-go"
)

const (
	melBins uint = 24
	melMin  uint = 20
	melMax  uint = sampleRate / 2
)

// Wrapper for filterbank which handles initialisation, normalisation
type melbank struct {
	fb           *aubio.FilterBank
	Audio        AudioStream
	Min          int
	Max          int
	Freqs        []float64
	Data         []float64
	GainFilter   *math_utils.ExpFilter
	SmoothFilter *math_utils.ExpFilterSlice
}

// Specify the min and max frequencies
func newMelbank(audio AudioStream, min, max uint, intensity float64) (*melbank, error) {

	mb := &melbank{
		fb:           aubio.NewFilterBank(melBins, framesPerBuffer),
		Audio:        audio,
		Min:          int(min),
		Max:          int(max),
		Freqs:        make([]float64, melBins+2),
		Data:         make([]float64, melBins),
		GainFilter:   math_utils.NewExpFilter(0.99, 0.01),
		SmoothFilter: math_utils.NewExpFilterSlice(0.99, 0.5+intensity/2, int(melBins)),
	}

	if min < melMin || max > melMax {
		return mb, fmt.Errorf("invalid frequency range: %d %d. Must be %d to %d", min, max, melMin, melMax)
	}

	// Build the frequency bands for the melbank
	// create linear freq scale
	freqs, err := math_utils.Linspace(HzToMel(float64(min)), HzToMel(float64(max)), int(melBins)+2)
	if err != nil {
		return mb, err
		// log.Logger.WithField("context", "melbank Init").Fatalf("Error initialising melbank: %v", err)
	}
	// convert linear freq scale to perceptually even scale.
	for i := range freqs {
		freqs[i] = MelToHz(freqs[i])
	}
	// set and normalise the bands
	mb.fb.SetTriangleBands(aubio.NewSimpleBufferData(melBins+2, freqs), sampleRate)
	mb.fb.NormalizeCoeffs()
	// save the freqs for reference
	mb.Freqs = freqs
	return mb, nil
}

// Perform mel binning on fft
func (mb *melbank) Do(fft *aubio.ComplexBuffer) {
	mb.fb.Do(fft)
	copy(mb.Data, mb.fb.Buffer().Slice())
	// Normalise the melbank gain
	// first smooth the values out to soften peaks
	gainData := make([]float64, melBins)
	copy(gainData, mb.Data)
	math_utils.Blur1D(gainData, 3)
	// get max of gainData
	var max float64
	for _, val := range gainData {
		if val > max {
			max = val
		}
	}
	// update gain filter to get smoothed value
	mb.GainFilter.Update(max)
	// divide the mel data by this smoothed gain value
	if mb.GainFilter.Value != 0 {
		for i, val := range mb.Data {
			mb.Data[i] = val / mb.GainFilter.Value
		}
	}
	// TODO this should be controlled by the effect's "intensity" config
	// Apply temporal filtering to melbank so it's not jumping around like crazy
	mb.SmoothFilter.Update(mb.Data)
	mb.Data = mb.SmoothFilter.Value
}

// Cleanup allocated C memory
func (mb *melbank) Free() {
	mb.fb.Buffer().Free()
	mb.fb.Coeffs().Free()
}

// returns the maximum value from the lowest third of the melbank
func (mb *melbank) LowsAmplitude() float64 {
	var max float64
	var val float64
	for i := 0; i < len(mb.Data)/3; i++ {
		val = mb.Data[i]
		if val > 1 {
			return 1
		}
		if val > max {
			max = val
		}
	}
	return max
}

// returns the maximum value from the middle third of the melbank
func (mb *melbank) MidsAmplitude() float64 {
	var max float64
	var val float64
	for i := len(mb.Data) / 3; i < 2*len(mb.Data)/3; i++ {
		val = mb.Data[i]
		if val > 1 {
			return 1
		}
		if val > max {
			max = val
		}
	}
	return max
}

// returns the maximum value from the highest third of the melbank
func (mb *melbank) HighAmplitude() float64 {
	var max float64
	var val float64
	for i := 2 * len(mb.Data) / 3; i < len(mb.Data); i++ {
		val = mb.Data[i]
		if val > 1 {
			return 1
		}
		if val > max {
			max = val
		}
	}
	return max
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
