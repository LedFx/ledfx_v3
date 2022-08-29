package color

import "math"

type Blurrer struct {
	pixelCount   int
	kernelRadius int
	kernel       []float64
	working      Pixels
}

// Create new Blurrer
// For kernel info, see: https://www.desmos.com/calculator/pjepmbj16v
func NewBlurrer(pixelCount int, blur float64) (b *Blurrer) {
	radius := int(math.Log(1+blur/10) * float64(pixelCount))
	if radius < 1 {
		radius = 1
	}
	k := make([]float64, 2*radius+1)
	for i, x := 0, -float64(radius); i < len(k); i, x = i+1, x+1 {
		k[i] = 1 - math.Abs(x/float64(radius))
	}
	// Normalise the kernel so the sum of all values is 1
	var sum float64 = 0
	for i := 0; i < len(k)-1; i++ {
		sum += k[i]
	}
	for i := 0; i < len(k)-1; i++ {
		k[i] /= sum
	}
	b = &Blurrer{
		pixelCount:   pixelCount,
		kernelRadius: radius,
		kernel:       k,
		working:      make(Pixels, pixelCount+2*radius),
	}

	return b
}

// Faster blur algorithm, linear complexity
// https://medium.com/mobile-app-development-publication/blurring-image-algorithm-example-in-android-cec81911cd5e
// https://blog.ivank.net/fastest-gaussian-blur.html
func (b *Blurrer) BoxBlur(p Pixels) {
	// copy pixels to working array
	for i, px := range p {
		b.working[i+b.kernelRadius] = px
	}
	// mirror the edges
	for i, j := 0, b.kernelRadius; i < b.kernelRadius; i, j = i+1, j-1 {
		b.working[i] = p[j]
	}
	for i, j := len(b.working)-1, len(p)-b.kernelRadius-1; j < len(p); i, j = i-1, j+1 {
		b.working[i] = p[j]
	}
	// perform two passes to approximate a gaussian blur
	for i := 0; i < 2; i++ {
		// initialise the moving average box
		// three sums, one for each color channel
		var sum0, sum1, sum2 float64 = 0, 0, 0
		for i := 0; i < len(b.kernel)-1; i++ {
			sum0 += b.working[i][0]
			sum1 += b.working[i][1]
			sum2 += b.working[i][2]
		}
		// normalise the box sum
		floatLen := float64(len(b.kernel))
		sum0 /= floatLen
		sum1 /= floatLen
		sum2 /= floatLen
		// clever trick: rather than compute all the values of the box for each pixel,
		// just add the incoming value and subtract the outgoing one
		for i := 0; i < b.pixelCount; i++ {
			// set the pixel value to the box average
			p[i][0] = sum0
			p[i][1] = sum1
			p[i][2] = sum2
			// compute the next box average
			// 1. add the incoming value to the average
			sum0 += b.working[i+len(b.kernel)-1][0] / floatLen
			sum1 += b.working[i+len(b.kernel)-1][1] / floatLen
			sum2 += b.working[i+len(b.kernel)-1][2] / floatLen
			// 2. take away the outgoing value from the average
			sum0 -= b.working[i][0] / floatLen
			sum1 -= b.working[i][1] / floatLen
			sum2 -= b.working[i][2] / floatLen
		}
	}
}

// Slower kernel blurring function, exponential complexity. Avoid.
func (b *Blurrer) KernelBlur(p Pixels) {
	/*
	   could try this kernel
	   https://dsp.stackexchange.com/questions/54375/how-to-approximate-gaussian-kernel-for-image-blur
	*/
	// copy pixels to working array
	for i, px := range p {
		b.working[i+b.kernelRadius] = px
	}
	// mirror the edges
	for i, j := 0, b.kernelRadius; i < b.kernelRadius; i, j = i+1, j-1 {
		b.working[i] = p[j]
	}
	for i, j := len(b.working)-1, len(p)-b.kernelRadius-1; j < len(p); i, j = i-1, j+1 {
		b.working[i] = p[j]
	}
	// perform convolution using filter kernel, updating pixels as we go
	// we're looping over each pixel of channel (R,G,B), applying the kernel multiplication of the surrounding pixels
	var sum float64
	var x int
	for i := b.kernelRadius + 1; i < b.pixelCount+b.kernelRadius; i++ {
		for j := 0; j < 3; j++ {
			sum = 0
			for k := 0; k < len(b.kernel); k++ {
				x = i + k - 1 - b.kernelRadius
				sum += b.kernel[k] * b.working[x][j]
			}
			// put result in pixels
			p[i-b.kernelRadius-1][j] = sum
		}
	}
}
