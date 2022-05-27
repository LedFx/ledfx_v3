package color

import (
	"fmt"
	"image/color"
	"math"
)

// Apply desaturation to pixels (RGB)
// 1: color remains unchanged, 0: color is fully desaturated
func DesaturatePixels(p Pixels, s float64) {
	var max float64

	for i := range p {
		// find brightest channel of color
		max = 0
		for _, val := range p[i] {
			if max < val {
				max = val
			}
		}
		// scale each channel accordingly
		p[i][0] += (max - p[i][0]) * s
		p[i][1] += (max - p[i][1]) * s
		p[i][2] += (max - p[i][2]) * s
	}
}

// Shift the hue of pixels (HSL)
func HueShiftPixels(p Pixels, hs float64) {
	for _, color := range p {
		color[0] += hs
	}
}

// Darken pixels (RGB) (they're full brightness by default)
func DarkenPixels(p Pixels, ds float64) {
	for i := range p {
		p[i][0] *= ds
		p[i][1] *= ds
		p[i][2] *= ds
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

func MaxOfThree(a, b, c float64) float64 {
	return math.Max(a, math.Max(b, c))
}

func MinOfThree(a, b, c float64) float64 {
	return math.Min(a, math.Min(b, c))
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
