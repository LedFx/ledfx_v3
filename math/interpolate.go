package math

import (
	"errors"
	"ledfx/color"
	"math"
)

// Linear pixel interpolation. Scale input pixels onto output pixels
func Interpolate(in color.Pixels, out color.Pixels) error {
	if len(in) < 2 || len(out) < 2 {
		return errors.New("cannot interpolate using less than two pixels")
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
		out[i][0] = in[ix][0] + in[ix+1][0]*f
		out[i][1] = in[ix][1] + in[ix+1][1]*f
		out[i][2] = in[ix][2] + in[ix+1][2]*f
	}

	return nil
}
