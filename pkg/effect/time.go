package effect

import "math"

/*
time utility functions, to generate effect data based on elapsed time
*/

// constant time sawtooth, loops 0-1 every 60*interval seconds.
// a modifier of 0.016 gives a 1s sawtooth
func (e *Effect) time(interval float64) float64 {
	scaled_interval := interval * 60
	return math.Mod(e.deltaStart.Seconds(), scaled_interval) / scaled_interval
}

// constant time sinusoidal waveform, loops 0-1 every 60*interval seconds.
func (e *Effect) sin(t float64) float64 {
	return 0.5*math.Sin(t*2*math.Pi) + 0.5
}

// constant time sinusoidal waveform, loops 0-1 every 60*interval seconds.
func (e *Effect) triangle(t float64) float64 {
	return 1 - 2*math.Abs(t-0.5)
}
