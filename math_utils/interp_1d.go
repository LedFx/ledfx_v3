package math_utils

import (
	"errors"
	"math"
)

// Linear slice interpolation. Scale input slice onto output slice
// Similar to pixel interpolation, but for a single slice
func Interpolate(in []float64, out []float64) error {
	if len(in) < 2 || len(out) < 2 {
		return errors.New("cannot interpolate using less than two data points")
	}
	// handle trivial case same size in and out
	if len(in) == len(out) {
		copy(out, in)
		return nil
	}
	out[len(out)-1] = in[len(in)-1]

	ratio := float64(len(in)-1) / float64(len(out)-1)
	for i := 0; i < len(out)-1; i++ {
		x, f := math.Modf(ratio * float64(i))
		ix := int(x)
		out[i] = in[ix] + in[ix+1]*f
	}

	return nil
}
