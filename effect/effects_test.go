package effect

import (
	"fmt"
	"ledfx/color"
	"testing"
)

func BenchmarkEffects(t *testing.B) {
	// Default config
	c := map[string]interface{}{}
	for _, eType := range effectTypes {
		for _, p := range color.TestPixels {
			// Make a new effect
			effect, _, err := New("", eType, len(p), c)
			if err != nil {
				t.Error(err)
			}
			// Run the effect on some pixels
			t.Run(fmt.Sprintf("%s %d pixels", eType, len(p)), func(t *testing.B) {
				for i := 0; i < t.N; i++ {
					effect.Render(p)
				}
			})
			Destroy(effect.GetID())
		}
	}
}

func TestEffects(t *testing.T) {
	// Default config
	blank_c := map[string]interface{}{}

	testConfigs := []map[string]interface{}{
		{
			"intensity":      0,
			"brightness":     0,
			"palette":        "Rainbow",
			"blur":           0,
			"flip":           false,
			"mirror":         false,
			"bkg_brightness": 0,
			"bkg_color":      "black",
		},
		{
			"intensity":      1,
			"brightness":     1,
			"palette":        "linear-gradient(90deg, rgb(128, 0, 128) 0%, rgb(0, 0, 255) 25%, rgb(0, 128, 128) 50%, rgb(0, 255, 0) 75%, rgb(255, 200, 0) 100%)",
			"blur":           1,
			"flip":           true,
			"mirror":         true,
			"bkg_brightness": 1,
			"bkg_color":      "#FFFFFF",
		},
	}

	for _, eType := range effectTypes {
		for _, p := range color.TestPixels {
			// Make a new effect
			effect, _, err := New("", eType, len(p), blank_c)
			if err != nil {
				t.Error(err)
			}
			for _, c := range testConfigs {
				err = effect.UpdateBaseConfig(c) // Assign the config
				effect.Render(p)                 // Run it on some pixels
				if err != nil {
					t.Errorf("Failed on test config: %v", c)
				}
			}
			Destroy(effect.GetID())
		}
	}
}

// cases := []struct {
// 	q string
// 	a Color
// 	e bool
// }{
// 	{"#ffFf00", Color{1, 1, 0}, false},
// 	{"RGB(0,255, 0)", Color{0, 1, 0}, false},
// 	{"#fF0", Color{0, 0, 0}, true},
// 	{"rgb(-1,0,256)", Color{}, true},
// 	{"#efghij", Color{}, true},
// 	{"nonsense color", Color{}, true},
// 	{"red", Color{1, 0, 0}, false},
// }
// for _, c := range cases {
// 	guess, err := NewColor(c.q)
// 	if (c.a != guess) || (err == nil == c.e) { // if the answer is wrong, or the error value is unexpected
// 		t.Errorf("Failed to parse %s: expected (%v, %v) but got (%v, %v)", c.q, c.a, c.e, guess, err)
// 	}
// }
