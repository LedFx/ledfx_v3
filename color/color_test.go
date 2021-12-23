package color_test

import (
	"ledfx/color"
	"testing"
)

/*
Simple module tests should be a table of test cases.
q is the question, g is the guess, a is the correct answer
*/

func TestParseColor(t *testing.T) {
	cases := []struct {
		q string
		a [3]float64
	}{
		{"#ffFf00", [3]float64{1, 1, 0}},
		{"#fF0", [3]float64{1, 1, 0}},
		{"RGB(0,255,0)", [3]float64{0, 1, 0}},
	}
	for _, c := range cases {
		guess, err := color.ParseString(c.q)
		if err != nil {
			t.Errorf("Failed to parse %s: got %v, with error %v", c.q, guess, err)
		}
		if guess != c.a {
			t.Errorf("Failed to parse %s: got %v, wanted %v", c.q, guess, c.a)
		}
	}
}
