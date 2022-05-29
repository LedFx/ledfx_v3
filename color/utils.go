package color

import (
	"fmt"
	"image/color"
	"math"
)

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
func FillBetween(p Pixels, start, stop int, col Color) {
	if stop == start {
		p[start] = col
		return
	}
	// make sure ascending indexes
	if start > stop {
		start, stop = stop, start
	}
	for i := start; i <= stop; i++ {
		p[i] = col
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

/*
Math utilities. TODO: move these to their own package?
*/

func minMax(a, b int) (min, max int) {
	if a > b {
		return b, a
	}
	return a, b
}

func minMaxArray(array []float64) (float64, float64) {
	var max float64 = array[0]
	var min float64 = array[0]
	for _, value := range array {
		if max < value {
			max = value
		}
		if min > value {
			min = value
		}
	}
	return min, max
}

// Return evenly spaced numbers over a specified interval
func linspace(start, stop float64, num int) (ls []float64, err error) {
	if start >= stop {
		return ls, fmt.Errorf("linspace start must not be greater than stop: %v, %v", start, stop)
	}
	if num <= 0 {
		return ls, fmt.Errorf("num must be greater than 0: %v, %v", start, stop)
	}
	ls = make([]float64, num)
	delta := stop / float64(num)
	for i, x := 0, start; i < num; i, x = i+1, x+delta {
		ls[i] = x
	}
	return ls, nil
}
