package color

import (
	"fmt"
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
		{"red", Color{1, 0, 0}, false},
	}
	for _, c := range cases {
		guess, err := NewColor(c.q)
		if (c.a != guess) || (err == nil == c.e) { // if the answer is wrong, or the error value is unexpected
			t.Errorf("Failed to parse %s: expected (%v, %v) but got (%v, %v)", c.q, c.a, c.e, guess, err)
		}
	}
}

func TestNewPalette(t *testing.T) {
	cases := []struct {
		test string
		pass bool
	}{
		{"this is not a palette", false},
		{"linear-gradient(#ffFf00 10%, )", false},
		{"linear-gradient(180deg, #ffgh00 10%)", false},
		{"linear-gradient(180deg, rgb(299,0,299) 10%)", false},
		{"linear-gradient(180deg, useless color 10%)", false},
		{"linear-gradient(90deg, #ffFf00 0%, rgb(255, 0, 255) 30%, rgb(255, 255, 255) 100%)", true},
		{"RGB", true},
		{"Rainbow", true},
		{"Dancefloor", true},
		{"Plasma", true},
		{"Ocean", true},
		{"Viridis", true},
		{"Jungle", true},
		{"Spring", true},
		{"Winter", true},
		{"Frost", true},
		{"Sunset", true},
		{"Borealis", true},
		{"Rust", true},
		{"Winamp", true},
	}
	for _, c := range cases {
		_, err := NewPalette(c.test)
		if err == nil != c.pass { // if the answer is wrong, or the error value is unexpected
			t.Errorf("Failed test case %s, expected error to be %t", c.test, err == nil)
		}
	}
}

func BenchmarkKernelBlur(t *testing.B) {
	for _, v := range TestPixels {
		b := NewBlurrer(len(v), 1) // use largest kernel, most demanding
		t.Run(fmt.Sprintf("%d pixels", len(v)), func(t *testing.B) {
			for i := 0; i < t.N; i++ {
				b.KernelBlur(v)
			}
		})
	}
}

func BenchmarkBoxBlur(t *testing.B) {
	for _, v := range TestPixels {
		b := NewBlurrer(len(v), 1) // use largest kernel, most demanding
		t.Run(fmt.Sprintf("%d pixels", len(v)), func(t *testing.B) {
			for i := 0; i < t.N; i++ {
				b.BoxBlur(v)
			}
		})
	}
}

func BenchmarkToRGBW(t *testing.B) {
	for _, v := range TestPixels {
		out := make(PixelsRGBW, len(v))
		for i := range v {
			v[i][0] = 1
		}
		t.Run(fmt.Sprintf("%d pixels", len(v)), func(t *testing.B) {
			for i := 0; i < t.N; i++ {
				v.ToRGBW(out)
			}
		})
	}
}

func BenchmarkInterpolate(t *testing.B) {
	// Basis pixels
	in := make(Pixels, 10)
	for _, v := range TestPixels {
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
