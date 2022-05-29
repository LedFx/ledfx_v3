package effect

import (
	"fmt"
	"testing"
)

func BenchmarkPalette(t *testing.B) {
	// Make a new effect
	c := map[string]interface{}{}
	for _, v := range pixelSizes {
		effect, _, err := New("palette", len(v), c)
		if err != nil {
			t.Error(err)
		}
		// Run the effect on some pixels
		t.Run(fmt.Sprintf("%d pixels", len(v)), func(t *testing.B) {
			for i := 0; i < t.N; i++ {
				effect.Render(v)
			}
		})
		Destroy(effect.GetID())
	}
}
