package color

import "math"

type blurrer struct {
	pixelCount   int
	kernelRadius int
	kernel       []float64
	working      Pixels
}

// Create new blurrer
// For kernel info, see: https://www.desmos.com/calculator/pjepmbj16v
func NewBlurrer(pixelCount int, blur float64) (b *blurrer) {
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
	b = &blurrer{
		pixelCount:   pixelCount,
		kernelRadius: radius,
		kernel:       k,
		working:      make(Pixels, pixelCount+2*radius),
	}

	return b
}

func (b *blurrer) Blur(p Pixels) {
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
	for i := b.kernelRadius + 1; i < b.pixelCount+b.kernelRadius; i++ {
		for j := 0; j < 3; j++ {
			sum = 0
			for k := 0; k < len(b.kernel); k++ {
				x := i + k - 1 - b.kernelRadius
				sum += b.kernel[k] * b.working[x][j]
			}
			// put result in pixels
			p[i-b.kernelRadius-1][j] = sum
		}
	}
}
