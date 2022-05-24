package color

// Desaturates a color
// 1: color remains unchanged, 0: color is fully desaturated
func Desaturate(c *Color, s float64) {
	_, max := minMaxArray(c[:])
	for _, value := range c {
		diff := max - value
		value += diff * s
	}
}

// Apply desaturation to pixels
func DesaturatePixels(p *Pixels, s float64) {
	for _, color := range *p {
		Desaturate(&color, s)
	}
}

// TODO rename and move this somewhere more sensible
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
