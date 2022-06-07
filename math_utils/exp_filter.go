package math_utils

/*
Exponential filters smooth the rise and fall of values
*/

// Exponential filter for a single value
type ExpFilter struct {
	alphaDecay float64
	alphaRise  float64
	Value      float64
}

func NewExpFilter(rise, decay float64) *ExpFilter {
	return &ExpFilter{
		alphaDecay: decay,
		alphaRise:  rise,
		Value:      0,
	}
}

func (e *ExpFilter) Update(value float64) {
	var alpha float64
	if value > e.Value {
		alpha = e.alphaRise
	} else {
		alpha = e.alphaDecay
	}
	e.Value = alpha*value + (1-alpha)*e.Value
}

// Exponential filter for a slice
type ExpFilterSlice struct {
	alphaDecay float64
	alphaRise  float64
	Value      []float64
}

func NewExpFilterSlice(rise, decay float64, size int) *ExpFilterSlice {
	return &ExpFilterSlice{
		alphaDecay: decay,
		alphaRise:  rise,
		Value:      make([]float64, size),
	}
}

func (e *ExpFilterSlice) Update(value []float64) {
	var alpha float64
	for i := range value {
		if value[i] > e.Value[i] {
			alpha = e.alphaRise
		} else {
			alpha = e.alphaDecay
		}
		e.Value[i] = alpha*value[i] + (1-alpha)*e.Value[i]
	}
}
