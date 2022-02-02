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

func (b Buffer) Mono2Stereo() Buffer {
	stereo := Buffer(make([]int16, len(b)*2))
	for i := range b {
		magnitude := b[i]
		stereo[i*2] = magnitude
		stereo[(i*2)+1] = magnitude
	}
	return stereo
}
