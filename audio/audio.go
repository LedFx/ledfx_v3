package audio

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type IntWriter interface {
	Write(b Buffer) (n int, err error)
}
type IntReader interface {
	Read(p Buffer) (n int, err error)
}
type ByteWriter struct {
	wrs        [8]io.Writer
	numWriters int
}

func (bw *ByteWriter) AppendWriter(writer io.Writer) {
	bw.wrs[bw.numWriters] = writer
	bw.numWriters++
}
func (bw *ByteWriter) Write(p []byte) (int, error) {
	for i := 0; i < bw.numWriters; i++ {
		bw.wrs[i].Write(p)
	}
	return len(p), nil
}

type Buffer []int16

func (b Buffer) AsFloat64() []float64 {
	out := make([]float64, len(b))
	for i, x := range b {
		out[i] = float64(x)
	}
	return out
}

func (b Buffer) AsBytes() []byte {
	byteBuf := make([]byte, len(b)*2)

	var offset int
	for i := range b {
		binary.LittleEndian.PutUint16(byteBuf[offset:], uint16(b[i]))
		offset += 2
	}
	return byteBuf
}

func (b Buffer) Sum() int64 {
	var sum int64
	for i := range b {
		sum += int64(b[i])
	}
	return sum
}

func (b Buffer) ChannelMultiplier(numChannels int) Buffer {
	b2 := Buffer(make([]int16, len(b)*numChannels))
	for i := range b {
		magnitude := b[i]
		for channel := 0; channel < numChannels; channel++ {
			b2[(i*numChannels)+channel] = magnitude
		}
	}
	return b2
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

func (b Buffer) WriteTo(filename string) error {
	fi, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0777)
	if err != nil {
		return fmt.Errorf("error creating file '%s': %w", filename, err)
	}
	defer fi.Close()
	for i := range b {
		if _, err := fi.WriteString(fmt.Sprintf("%d\n", b[i])); err != nil {
			return fmt.Errorf("error writing string to file: %w", err)
		}
	}
	return nil
}

func (b Buffer) HighestValue() int16 {
	var highest int16
	for i := range b {
		if b[i] > highest {
			highest = b[i]
		}
	}
	return highest
}

const (
	dBMax = float64(96.33)
	dBMin = float64(0)

	rawMax = int16(32767)
	rawMin = int16(-32768)
)

func (b Buffer) Decibels() []float64 {
	out := make([]float64, len(b))
	for i := range b {
		switch b[i] {
		case rawMax:
			out[i] = dBMax
		case rawMin:
			out[i] = dBMin
		default:
			out[i] = (float64(b[i]) / float64(rawMax)) * dBMax
		}
	}
	return out
}

func HighestFloat(f []float64) float64 {
	var highest float64
	for i := range f {
		if f[i] > highest {
			highest = f[i]
		}
	}
	return highest
}
