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
func Saturation(passed_color Color, saturation float64) Color {
	if saturation == 1 {
		return passed_color
	}
	// apply scaling
	saturation = math.Sqrt(saturation)
	// find brightest channel of color
	var max float64 = math.Max(passed_color[0], math.Max(passed_color[1], passed_color[2]))
	// scale each channel accordingly
	passed_color[0] += (max - passed_color[0]) * (1 - saturation)
	passed_color[1] += (max - passed_color[1]) * (1 - saturation)
	passed_color[2] += (max - passed_color[2]) * (1 - saturation)

	return passed_color
}

// Apply HSV lightness to an RGB color
// 1: color remains unchanged, 0: color is fully darkened
func Value(passed_color Color, darkness_saturation float64) Color {
	if darkness_saturation == 1 {
		return passed_color
	}
	// apply scaling
	darkness_saturation = math.Pow(darkness_saturation, 3)
	passed_color[0] *= darkness_saturation
	passed_color[1] *= darkness_saturation
	passed_color[2] *= darkness_saturation
	return passed_color
}

// Shift the hue of pixels (HSV)
func HueShiftPixels(passed_pixels Pixels, hue_shift_value float64) {
	for pixel := range passed_pixels {
		passed_pixels[pixel][0] += hue_shift_value
	}
}

// This doesn't take into account white channel temperature or relative brightness, but it'll do for now.
// Pixels should be RGB.
func (passed_pixels Pixels) ToRGBW(output_pixels PixelsRGBW) {
	for pixel := 0; pixel < len(passed_pixels); pixel++ {
		r, g, b := passed_pixels[pixel][0], passed_pixels[pixel][1], passed_pixels[pixel][2]
		if r+g+b == 0 {
			output_pixels[pixel] = ColorRGBW{0, 0, 0, 0}
		}

		// calculate luminance
		lum := math.Min(r, math.Min(g, b))
		output_pixels[pixel][0] = r - lum
		output_pixels[pixel][1] = g - lum
		output_pixels[pixel][2] = b - lum
		output_pixels[pixel][3] = lum
	}
}

// fils a colour between two indexes. use base.pixelScaler to convert 0-1 floats to integer index
func FillBetween(passed_pixels Pixels, start, stop int, col Color, blend bool) {
	// make sure ascending indexes
	if start > stop {
		start, stop = stop, start
	}
	for i := start; i <= stop; i++ {
		if blend {
			passed_pixels[i][0] = (passed_pixels[i][0] + col[0]) / 2
			passed_pixels[i][1] = (passed_pixels[i][1] + col[1]) / 2
			passed_pixels[i][2] = (passed_pixels[i][2] + col[2]) / 2
		} else {
			passed_pixels[i] = col
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

func FromBufSliceSum(_ float64) string {
	// sum divided by 37
	return ""
}

// Linear pixel interpolation. Scale input pixels onto output pixels
func Interpolate(passed_pixels Pixels, output_pixels Pixels) error {
	if len(passed_pixels) < 2 || len(output_pixels) < 2 {
		return errors.New("cannot interpolate using less than two pixels")
	}
	// handle trivial case same size in and out
	if len(passed_pixels) == len(output_pixels) {
		copy(output_pixels, passed_pixels)
		return nil
	}
	output_pixels[len(output_pixels)-1] = passed_pixels[len(passed_pixels)-1]

	ratio := float64(len(passed_pixels)-1) / float64(len(output_pixels)-1)
	for pixel := 0; pixel < len(output_pixels)-1; pixel++ {
		// What on earth are x, f and i?
		x, f := math.Modf(ratio * float64(pixel))
		ix := int(x)
		output_pixels[pixel][0] = passed_pixels[ix][0] + passed_pixels[ix+1][0]*f
		output_pixels[pixel][1] = passed_pixels[ix][1] + passed_pixels[ix+1][1]*f
		output_pixels[pixel][2] = passed_pixels[ix][2] + passed_pixels[ix+1][2]*f
	}

	return nil
}
