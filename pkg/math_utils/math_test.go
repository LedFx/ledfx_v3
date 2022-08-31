package math_utils

import "testing"

func TestBlur1D(t *testing.T) {
	data, err := Linspace(0, 100, 100)
	if err != nil {
		t.Error(err)
	}
	Blur1D(data, 10)
}
