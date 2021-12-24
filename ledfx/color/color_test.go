package color

import (
	"testing"
)

/*
Simple module tests should be a table of test cases.
q is the question, g is the guess, a is the correct answer, e is bool if error expected
*/

func TestNewColor(t *testing.T) {
	cases := []struct {
		q string
		a [3]float64
		e bool
	}{
		{"#ffFf00", [3]float64{1, 1, 0}, false},
		{"#fF0", [3]float64{0, 0, 0}, true},
		{"RGB(0,255, 0)", [3]float64{0, 1, 0}, false},
		{"rgb(-1,0,256)", [3]float64{}, true},
		{"#efghijkl", [3]float64{}, true},
		{"nonsense color", [3]float64{}, true},
	}
	for _, c := range cases {
		guess, err := NewColor(c.q)
		if (c.a != guess) || (err == nil == c.e) { // if the answer is wrong, or the error value is unexpected
			t.Errorf("Failed to parse %s: expected (%v, %v) but got (%v, %v)", c.q, c.a, c.e, guess, err)
		}
	}
}
