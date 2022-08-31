package math_utils

import (
	"fmt"
	"math"
)

// Return evenly spaced numbers over a specified interval
func Linspace(start, stop float64, num int) (ls []float64, err error) {
	if start >= stop {
		return ls, fmt.Errorf("linspace start must not be greater than stop: %v, %v", start, stop)
	}
	if num <= 0 {
		return ls, fmt.Errorf("num must be greater than 0: %v, %v", start, stop)
	}
	ls = make([]float64, num)
	delta := stop / float64(num)
	for i, x := 0, start; i < num; i, x = i+1, x+delta {
		ls[i] = x
	}
	return ls, nil
}

// Logarithm of any base
func LogN(x, n float64) float64 {
	return math.Log(x) / math.Log(n)
}
