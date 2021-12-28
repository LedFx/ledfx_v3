package color

import (
	"testing"
)

/*
Simple module tests should be a table of test cases.
q is the question, a is the correct answer, e is bool if error expected
*/

func TestNewColor(t *testing.T) {
	cases := []struct {
		q string
		a Color
		e bool
	}{
		{"#ffFf00", Color{1, 1, 0}, false},
		{"RGB(0,255, 0)", Color{0, 1, 0}, false},
		{"#fF0", Color{0, 0, 0}, true},
		{"rgb(-1,0,256)", Color{}, true},
		{"#efghij", Color{}, true},
		{"nonsense color", Color{}, true},
	}
	for _, c := range cases {
		guess, err := NewColor(c.q)
		if (c.a != guess) || (err == nil == c.e) { // if the answer is wrong, or the error value is unexpected
			t.Errorf("Failed to parse %s: expected (%v, %v) but got (%v, %v)", c.q, c.a, c.e, guess, err)
		}
	}
}

func TestNewGradient(t *testing.T) {
	cases := []struct {
		q string
		a Gradient
		e bool
	}{
		{"linear-gradient(#ffFf00 10%, )", Gradient{}, true},
		{"linear-gradient(180deg, #ffgh00 10%)", Gradient{}, true},
		{"linear-gradient(180deg, rgb(299,0,299) 10%)", Gradient{}, true},
		{"linear-gradient(180deg, useless color 10%)", Gradient{}, true},
		{
			"linear-gradient(90deg, #ffFf00 10%, rgb(255, 0, 255) 30%)",
			Gradient{mode: "linear", angle: 90},
			false,
		},
	}
	for _, c := range cases {
		guess, err := NewGradient(c.q)
		if (c.a.mode != guess.mode) || (c.a.angle != guess.angle) || (err == nil == c.e) { // if the answer is wrong, or the error value is unexpected
			t.Errorf("Failed to parse %s: expected (%v, %v) but got (%v, %v)", c.q, c.a, c.e, guess, err)
		}
	}
}
