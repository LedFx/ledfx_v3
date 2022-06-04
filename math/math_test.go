package math

import (
	"fmt"
	"ledfx/color"
	"testing"
)

func BenchmarkInterpolate(t *testing.B) {
	// Basis pixels
	in := make(color.Pixels, 10)
	for _, v := range color.TestPixels {
		// Run the effect on some pixels
		t.Run(fmt.Sprintf("10 to %d pixels", len(v)), func(t *testing.B) {
			for i := 0; i < t.N; i++ {
				err := Interpolate(in, v)
				if err != nil {
					t.Error(err)
				}
			}
		})
	}
}
