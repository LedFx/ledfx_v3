package color

import "math"

// Shift the hue of pixels
func HueShiftPixels(p Pixels, hs float64) {
	for _, color := range p {
		h := color[0]
		h += hs
		color[0] = math.Mod(h, 1)
	}
}

// Apply desaturation to pixels (they're fully saturated by default)
func DesaturatePixels(p Pixels, ds float64) {
	for _, color := range p {
		color[1] *= ds
	}
}

// Darken pixels (they're full brightness by default)
func DarkenPixels(p Pixels, ds float64) {
	for _, color := range p {
		color[2] *= ds
	}
}

func ToRGB(p Pixels) {

}
