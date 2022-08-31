package math_utils

/*
stateless 1d implementation of the pixel blurring method in color package
*/

// Blurs the data in place
func Blur1D(data []float64, radius int) {
	// copy incoming slice to working array
	w := make([]float64, len(data)+2*radius+1)
	for i, val := range data {
		w[i+radius] = val
	}
	// mirror the edges
	for i, j := 0, radius; i < radius; i, j = i+1, j-1 {
		w[i] = data[j]
	}
	for i, j := len(w)-1, len(data)-radius-1; j < len(data); i, j = i-1, j+1 {
		w[i] = data[j]
	}
	floatLen := float64(radius*2 + 1)
	// perform two passes to approximate a gaussian blur
	for i := 0; i < 2; i++ {
		// initialise the moving average box
		var sum float64
		for i := 0; i < radius*2+1; i++ {
			sum += w[i]
		}
		// normalise the box sum
		sum /= floatLen
		// clever trick: rather than compute all the values of the box for each position,
		// just add the incoming value and subtract the outgoing one
		for i := 0; i < len(data); i++ {
			// set the value to the box average
			data[i] = sum
			// compute the next box average
			// 1. add the incoming value to the average
			sum += w[i+2*radius] / floatLen
			// 2. take away the outgoing value from the average
			sum -= w[i] / floatLen
		}
	}
}
