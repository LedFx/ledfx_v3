package color

import (
	"errors"
	"image/color"
	"math"
)

var TestPixels = []Pixels{
	make(Pixels, 10),
	make(Pixels, 100),
	make(Pixels, 1000),
	make(Pixels, 10000),
}

// Apply HSV saturation to an RGB color
// 1: color remains unchanged, 0: color is fully desaturated
func Saturation(c Color, s float64) Color {
	if s == 1 {
		return c
	}
	// find brightest channel of color
	var max float64 = math.Max(c[0], math.Max(c[1], c[2]))
	// scale each channel accordingly
	c[0] += (max - c[0]) * (1 - s)
	c[1] += (max - c[1]) * (1 - s)
	c[2] += (max - c[2]) * (1 - s)

	return c
}

// Apply HSV lightness to an RGB color
// 1: color remains unchanged, 0: color is fully darkened
func Value(c Color, ds float64) Color {
	if ds == 1 {
		return c
	}
	c[0] *= ds
	c[1] *= ds
	c[2] *= ds
	return c
}

// Shift the hue of pixels (HSV)
func HueShiftPixels(p Pixels, hs float64) {
	for i := range p {
		p[i][0] += hs
	}
}

// This doesn't take into account white channel temperature or relative brightness, but it'll do for now.
// Pixels should be RGB.
func (p Pixels) ToRGBW(out PixelsRGBW) {
	for i := 0; i < len(p); i++ {
		r, g, b := p[i][0], p[i][1], p[i][2]
		if r+g+b == 0 {
			out[i] = ColorRGBW{0, 0, 0, 0}
		}

		// calculate luminance
		lum := math.Min(r, math.Min(g, b))
		out[i][0] = r - lum
		out[i][1] = g - lum
		out[i][2] = b - lum
		out[i][3] = lum
	}
}

// fils a colour between two indexes. use base.pixelScaler to convert 0-1 floats to integer index
func FillBetween(p Pixels, start, stop int, col Color, blend bool) {
	// make sure ascending indexes
	if start > stop {
		start, stop = stop, start
	}
	for i := start; i <= stop; i++ {
		if blend {
			p[i][0] = (p[i][0] + col[0]) / 2
			p[i][1] = (p[i][1] + col[1]) / 2
			p[i][2] = (p[i][2] + col[2]) / 2
		} else {
			p[i] = col
		}
	}
}

func (col Color) NRGBA() color.NRGBA {
	return color.NRGBA{
		R: NormalizeFloat(col[0]),
		G: NormalizeFloat(col[1]),
		B: NormalizeFloat(col[2]),
		A: 255,
	}
}

func NormalizeFloat(f float64) uint8 {
	return uint8(f * 255)
}

func NormalizeColorList(cols []Color) []color.Color {
	nc := make([]color.Color, len(cols))
	for i := range cols {
		nc[i] = cols[i].NRGBA()
	}
	return nc
}

func RandomColor() string {
	for _, v := range LedFxColors {
		return v
	}
	// Return red by default if for some reason the map is empty. It won't be.
	return "#ff0000"
}

func FromBufSliceSum(sum float64) string {
	// sum divided by 37
	return ""
}

// Linear pixel interpolation. Scale input pixels onto output pixels
func Interpolate(in Pixels, out Pixels) error {
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
