package audio

type Buffer []float32

func bufferToF64(b *Buffer) (out []float64) {
	out = make([]float64, len(*b))
	for i, x := range *b {
		out[i] = float64(x)
	}
	return out
}
