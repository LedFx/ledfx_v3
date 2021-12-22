package color

import (
	"testing"
	"color"
)

func TestParseColor(t *testing.T) {
	cases := []struct {
		q string
		a [3]float64
	}{
		{"#ffFf00", {255, 255, 0}},
		{"RGB(0,10,0)", {0, 10, 0}},
	}
	for _, case := range cases {
		if color.ParseString(case.q) != case.a {
			t.Errorf("Failed to parse %s, wanted %s", case.q, case.a)
		}
	}
}
