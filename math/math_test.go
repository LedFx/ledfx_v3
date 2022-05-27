package math

import (
	"fmt"
	"ledfx/color"
	"testing"
)

var pixelSizes = []color.Pixels{
	make(color.Pixels, 10),
	make(color.Pixels, 100),
	make(color.Pixels, 1000),
	make(color.Pixels, 10000),
}

func BenchmarkInterpolate(t *testing.B) {
	// Basis pixels
	in := make(color.Pixels, 10)
	for _, v := range pixelSizes {
		// Run the effect on some pixels
		t.Run(fmt.Sprintf("10 to %d pixels", len(v)), func(t *testing.B) {
			for i := 0; i < t.N; i++ {
				Interpolate(in, v)
			}
		})
	}
}
