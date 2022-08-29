package codec

import (
	"math"
	"testing"
)

const (
	volume = 1
	val    = 128
)

func BenchmarkMinMaxInternal_1024(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for i2 := 0; i2 < 1024; i2++ {
			math.Max(-32767, math.Min(32767, volume*float64(val)))
		}
	}
}

func BenchmarkMinMaxCustom_1024(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for i2 := 0; i2 < 1024; i2++ {
			max(-32767, min(32767, int16(volume*float64(val))))
		}
	}
}
