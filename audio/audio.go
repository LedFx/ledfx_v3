package audio

type Buffer []int16

func BufferToF64(b *Buffer) (out []float64) {
	out = make([]float64, len(*b))
	for i, x := range *b {
		out[i] = float64(x)
	}
	return out
}

func (b Buffer) Sum() int64 {
	var sum int64
	for i := range b {
		sum += int64(b[i])
	}
	return sum
}
